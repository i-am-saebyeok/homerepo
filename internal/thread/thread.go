package thread

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/i-am-saebyeok/homerepo/internal/herdr"
	"github.com/i-am-saebyeok/homerepo/internal/home"
	"github.com/i-am-saebyeok/homerepo/internal/workspace"
)

const EnvironmentVariable = "HOMEREPO_THREAD"

type Threads struct {
	Home       home.Home
	Workspaces workspace.Workspaces
	Herdr      herdr.Client
	Records    Records
	Now        func() time.Time
}

type CreateRequest struct {
	Name         string
	Repositories []string
	Mode         string
	Caller       Caller
}

type Summary struct {
	Record    Record
	TabOpen   bool
	Worktrees []workspace.Worktree
}

func (t Threads) Create(ctx context.Context, request CreateRequest) (Record, error) {
	adapter, bundle, err := t.checkCreate(request)
	if err != nil {
		return Record{}, err
	}
	for _, repository := range request.Repositories {
		if _, err := t.Workspaces.Create(ctx, repository, request.Name); err != nil {
			return Record{}, err
		}
	}
	attachDirs, err := t.attachDirs()
	if err != nil {
		return Record{}, err
	}
	tab, err := t.Herdr.CreateTab(ctx, herdr.TabRequest{
		Workspace:   request.Caller.Workspace,
		Cwd:         t.Home.Root,
		Label:       request.Name,
		Environment: []string{EnvironmentVariable + "=" + request.Name},
	})
	if err != nil {
		return Record{}, err
	}
	record := Record{Name: request.Name, Runtime: request.Caller.Runtime, TabID: tab.ID, CreatedAt: t.Now().UTC()}
	if err := t.Records.Write(record); err != nil {
		return record, err
	}
	if err := t.Herdr.RunInPane(ctx, tab.Pane, adapter.program, launchArgs(adapter, request.Caller.Session, attachDirs, bundle)); err != nil {
		return record, fmt.Errorf("tab %s was created but the session did not start (close it with `homerepo thread close %s`): %w", tab.ID, request.Name, err)
	}
	return record, nil
}

func (t Threads) checkCreate(request CreateRequest) (runtimeAdapter, []string, error) {
	if err := workspace.CheckThreadName(request.Name); err != nil {
		return runtimeAdapter{}, nil, err
	}
	if err := t.checkNameIsFree(request.Name); err != nil {
		return runtimeAdapter{}, nil, err
	}
	for _, repository := range request.Repositories {
		if _, err := t.Home.Repository(repository); err != nil {
			return runtimeAdapter{}, nil, err
		}
	}
	adapter, err := adapterFor(request.Caller.Runtime)
	if err != nil {
		return runtimeAdapter{}, nil, err
	}
	bundle, err := argumentBundle(t.Home.Local, request.Caller.Runtime, request.Mode)
	return adapter, bundle, err
}

func (t Threads) checkNameIsFree(name string) error {
	records, err := t.Records.List()
	if err != nil {
		return err
	}
	for _, record := range records {
		if record.Name == name {
			return fmt.Errorf("thread %q already exists", name)
		}
		if strings.HasPrefix(name, record.Name+"/") || strings.HasPrefix(record.Name, name+"/") {
			return fmt.Errorf("thread %q conflicts with thread %q: Git cannot keep a branch and another branch inside its name", name, record.Name)
		}
	}
	return nil
}

func (t Threads) attachDirs() ([]string, error) {
	var dirs []string
	for _, repository := range t.Home.Repositories {
		if !home.IsDir(t.Home.ClonePath(repository.Name)) {
			continue
		}
		dir := t.Home.WorktreesPath(repository.Name)
		if _, err := home.EnsureIgnoredDir(dir); err != nil {
			return nil, err
		}
		dirs = append(dirs, dir)
	}
	return dirs, nil
}

func (t Threads) List(ctx context.Context) ([]Summary, error) {
	records, err := t.Records.List()
	if err != nil {
		return nil, err
	}
	worktrees, err := t.Workspaces.List(ctx)
	if err != nil {
		return nil, err
	}
	summaries := make([]Summary, 0, len(records))
	for _, record := range records {
		open, err := t.Herdr.TabExists(ctx, record.TabID)
		if err != nil {
			return nil, err
		}
		summaries = append(summaries, Summary{Record: record, TabOpen: open, Worktrees: workspace.OwnedBy(worktrees, record.Name)})
	}
	return summaries, nil
}

func (t Threads) Close(ctx context.Context, name string) ([]string, error) {
	record, err := t.Records.Read(name)
	if err != nil {
		return nil, err
	}
	worktrees, err := t.Workspaces.OfThread(ctx, name)
	if err != nil {
		return nil, err
	}
	var problems []error
	for _, worktree := range worktrees {
		problems = append(problems, t.Workspaces.CheckClean(ctx, worktree))
	}
	if err := errors.Join(problems...); err != nil {
		return nil, fmt.Errorf("thread %q was not closed; its tab and worktrees are kept:\n%w", name, err)
	}
	var released []string
	for _, worktree := range worktrees {
		path, err := t.Workspaces.Release(ctx, worktree.Repository, name)
		if err != nil {
			return released, err
		}
		released = append(released, path)
	}
	if err := t.Records.Delete(name); err != nil {
		return released, err
	}
	return released, t.Herdr.CloseTab(ctx, record.TabID)
}

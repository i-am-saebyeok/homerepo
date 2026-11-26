package thread

import (
	"context"
	"errors"
	"fmt"
	"slices"
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
	Hooks      Hooks
	Now        func() time.Time
}

type CreateRequest struct {
	Name         string
	Repositories []string
	Mode         string
	Caller       Caller
}

type Created struct {
	Record    Record
	Worktrees []CreatedWorktree
}

type CreatedWorktree struct {
	Repository string
	Path       string
}

type Summary struct {
	Record    Record
	TabOpen   bool
	Worktrees []workspace.Worktree
}

func (t Threads) Create(ctx context.Context, request CreateRequest) (Created, error) {
	plan, err := t.checkCreate(request)
	if err != nil {
		return Created{}, err
	}
	values := hookValues{thread: request.Name, runtime: request.Caller.Runtime, attachDirs: t.attachDirs(request.Repositories)}
	if err := t.Hooks.run(ctx, preCreateHook, values); err != nil {
		return Created{}, err
	}
	var created Created
	for _, repository := range request.Repositories {
		path, err := t.Workspaces.Create(ctx, repository, request.Name)
		if err != nil {
			return created, err
		}
		created.Worktrees = append(created.Worktrees, CreatedWorktree{Repository: repository, Path: path})
	}
	if err := ensureIgnoredDirs(values.attachDirs); err != nil {
		return created, err
	}
	tab, err := t.Herdr.CreateTab(ctx, herdr.TabRequest{
		Workspace:   request.Caller.Workspace,
		Cwd:         t.Home.Root,
		Label:       request.Name,
		Environment: []string{EnvironmentVariable + "=" + request.Name},
	})
	if err != nil {
		return created, err
	}
	created.Record = Record{Name: request.Name, Runtime: request.Caller.Runtime, TabID: tab.ID, CreatedAt: t.Now().UTC()}
	if err := t.Records.Write(created.Record); err != nil {
		return created, err
	}
	if err := t.Herdr.RunInPane(ctx, tab.Pane, plan.adapter.program, plan.args(request.Caller.Session, values.attachDirs)); err != nil {
		return created, fmt.Errorf("tab %s was created but the session did not start (close it with `homerepo thread close %s`): %w", tab.ID, request.Name, err)
	}
	if err := t.Hooks.run(ctx, postCreateHook, values.inTab(tab.ID, tab.Pane)); err != nil {
		return created, fmt.Errorf("thread %q was created in tab %s, but %w", request.Name, tab.ID, err)
	}
	return created, nil
}

func (t Threads) checkCreate(request CreateRequest) (launch, error) {
	if err := workspace.CheckThreadName(request.Name); err != nil {
		return launch{}, err
	}
	if err := t.checkNameIsFree(request.Name); err != nil {
		return launch{}, err
	}
	for _, repository := range request.Repositories {
		if _, err := t.Home.Repository(repository); err != nil {
			return launch{}, err
		}
	}
	return prepareLaunch(t.Home.Local, request)
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

func (t Threads) attachDirs(requested []string) []string {
	var dirs []string
	for _, repository := range t.Home.Repositories {
		if home.IsDir(t.Home.ClonePath(repository.Name)) || slices.Contains(requested, repository.Name) {
			dirs = append(dirs, t.Home.WorktreesPath(repository.Name))
		}
	}
	return dirs
}

func ensureIgnoredDirs(dirs []string) error {
	for _, dir := range dirs {
		if _, err := home.EnsureIgnoredDir(dir); err != nil {
			return err
		}
	}
	return nil
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
	record, worktrees, err := t.checkClose(ctx, name)
	if err != nil {
		return nil, err
	}
	values := hookValues{thread: name, runtime: record.Runtime, attachDirs: t.attachDirs(nil), tab: record.TabID}
	if err := t.runBeforeClose(ctx, values); err != nil {
		return nil, fmt.Errorf("thread %q was not closed: %w", name, err)
	}
	released, err := t.release(ctx, record, worktrees)
	if err != nil {
		return released, err
	}
	if err := t.Hooks.run(ctx, postCloseHook, values); err != nil {
		return released, fmt.Errorf("thread %q was closed, but %w", name, err)
	}
	return released, nil
}

func (t Threads) checkClose(ctx context.Context, name string) (Record, []workspace.Worktree, error) {
	record, err := t.Records.Read(name)
	if err != nil {
		return Record{}, nil, err
	}
	worktrees, err := t.Workspaces.OfThread(ctx, name)
	if err != nil {
		return Record{}, nil, err
	}
	var problems []error
	for _, worktree := range worktrees {
		problems = append(problems, t.Workspaces.CheckClean(ctx, worktree))
	}
	if err := errors.Join(problems...); err != nil {
		return Record{}, nil, fmt.Errorf("thread %q was not closed; its tab and worktrees are kept:\n%w", name, err)
	}
	return record, worktrees, nil
}

func (t Threads) runBeforeClose(ctx context.Context, values hookValues) error {
	if !t.Hooks.has(preCloseHook) {
		return nil
	}
	pane, err := t.Herdr.PaneOfTab(ctx, values.tab)
	if err != nil {
		return err
	}
	return t.Hooks.run(ctx, preCloseHook, values.inTab(values.tab, pane))
}

func (t Threads) release(ctx context.Context, record Record, worktrees []workspace.Worktree) ([]string, error) {
	var released []string
	for _, worktree := range worktrees {
		path, err := t.Workspaces.Release(ctx, worktree.Repository, record.Name)
		if err != nil {
			return released, err
		}
		released = append(released, path)
	}
	if err := t.Records.Delete(record.Name); err != nil {
		return released, err
	}
	return released, t.Herdr.CloseTab(ctx, record.TabID)
}

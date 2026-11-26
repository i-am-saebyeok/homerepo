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

const (
	EnvironmentVariable = "HOMEREPO_THREAD"
	agentNameLimit      = 32
)

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

type Created struct {
	Record    Record
	Worktrees []string
	Readiness herdr.Readiness
}

type Summary struct {
	Record    Record
	TabOpen   bool
	Worktrees []workspace.Worktree
}

func (t Threads) Create(ctx context.Context, request CreateRequest) (Created, error) {
	adapter, bundle, err := t.checkCreate(request)
	if err != nil {
		return Created{}, err
	}
	var created Created
	for _, repository := range request.Repositories {
		path, err := t.Workspaces.Create(ctx, repository, request.Name)
		if err != nil {
			return created, err
		}
		created.Worktrees = append(created.Worktrees, path)
	}
	attachDirs, err := t.attachDirs()
	if err != nil {
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
	agent := AgentName(request.Name)
	created.Readiness, err = t.Herdr.StartAgent(ctx, herdr.AgentRequest{
		Name: agent,
		Kind: string(request.Caller.Runtime),
		Pane: tab.Pane,
		Args: launchArgs(adapter, request.Caller.Session, attachDirs, bundle),
	})
	if err != nil {
		return created, fmt.Errorf("tab %s was created but the agent did not start (close it with `homerepo thread close %s`): %w", tab.ID, request.Name, err)
	}
	if created.Readiness != herdr.Ready {
		return created, nil
	}
	return created, t.Herdr.Prompt(ctx, agent, FirstMessage(request.Name))
}

func (t Threads) checkCreate(request CreateRequest) (runtimeAdapter, []string, error) {
	if err := workspace.CheckThreadName(request.Name); err != nil {
		return runtimeAdapter{}, nil, err
	}
	exists, err := t.Records.Exists(request.Name)
	if err != nil {
		return runtimeAdapter{}, nil, err
	}
	if exists {
		return runtimeAdapter{}, nil, fmt.Errorf("thread %q already exists", request.Name)
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

func AgentName(thread string) string {
	var name strings.Builder
	for _, character := range strings.ToLower(thread) {
		if isAgentNameCharacter(character) {
			name.WriteRune(character)
		} else {
			name.WriteRune('-')
		}
	}
	result := name.String()
	if result == "" || result[0] < 'a' || result[0] > 'z' {
		result = "thread-" + result
	}
	if len(result) > agentNameLimit {
		result = result[:agentNameLimit]
	}
	return result
}

func isAgentNameCharacter(character rune) bool {
	return (character >= 'a' && character <= 'z') || (character >= '0' && character <= '9') || character == '_' || character == '-'
}

func FirstMessage(name string) string {
	return fmt.Sprintf("You are the homerepo thread %q. This session was just forked from another session into its own Herdr tab; continue here with the work that was split off. "+
		"The thread already exists, so do not run `homerepo thread create` again, even if the conversation ends with that call unfinished. "+
		"Do code work in a worktree, not in the original clone: run `homerepo workspace create <repository>` and work in the path it prints.", name)
}

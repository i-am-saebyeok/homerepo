package cli

import (
	"context"
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/i-am-saebyeok/homerepo/internal/herdr"
	"github.com/i-am-saebyeok/homerepo/internal/thread"
	"github.com/i-am-saebyeok/homerepo/internal/workspace"
)

func (app App) threads() (thread.Threads, error) {
	loaded, err := app.loadHome()
	if err != nil {
		return thread.Threads{}, err
	}
	return thread.Threads{
		Home:       loaded,
		Workspaces: workspace.Workspaces{Home: loaded, Runner: app.Runner},
		Herdr:      herdr.Client{Runner: app.Runner, Sleep: app.Sleep},
		Records:    thread.Records{HomeRoot: loaded.Root},
		Hooks:      thread.Hooks{HomeRoot: loaded.Root, Runner: app.Runner, Output: app.Stderr},
		Now:        app.Now,
	}, nil
}

func runThreadCreate(ctx context.Context, app App, args []string) error {
	flags := newFlags("thread create")
	var repositories repeated
	flags.Var(&repositories, "repo", "create the thread's worktree of this `repository` before the fork starts; repeat for more")
	mode := flags.String("mode", "", "add the argument `bundle` of this name from runtimes.<runtime>.thread in .homerepo/config.yaml (default: default)")
	runtimeName := flags.String("runtime", "", "`runtime` of the calling session: claude or codex (default: read from the environment)")
	session := flags.String("session", "", "`id` of the session to fork (default: read from the environment)")
	positionals, err := parse(flags, args, 1)
	if err != nil {
		return err
	}
	caller, err := thread.DetectCaller(app.Getenv, thread.CallerOverrides{Runtime: *runtimeName, Session: *session})
	if err != nil {
		return err
	}
	threads, err := app.threads()
	if err != nil {
		return err
	}
	created, err := threads.Create(ctx, thread.CreateRequest{Name: positionals[0], Repositories: repositories, Mode: *mode, Caller: caller})
	if err != nil {
		return err
	}
	app.println("created thread", created.Record.Name, "in tab", created.Record.TabID)
	for _, worktree := range created.Worktrees {
		app.println("worktree", worktree.Repository, worktree.Path)
	}
	return nil
}

func runThreadList(ctx context.Context, app App, args []string) error {
	if _, err := parse(newFlags("thread list"), args, 0); err != nil {
		return err
	}
	threads, err := app.threads()
	if err != nil {
		return err
	}
	summaries, err := threads.List(ctx)
	if err != nil {
		return err
	}
	table := tabwriter.NewWriter(app.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(table, "NAME\tRUNTIME\tTAB\tWORKTREES")
	for _, summary := range summaries {
		fmt.Fprintf(table, "%s\t%s\t%s\t%s\n", summary.Record.Name, summary.Record.Runtime, tabLabel(summary), worktreeLabel(summary.Worktrees))
	}
	return table.Flush()
}

func tabLabel(summary thread.Summary) string {
	if summary.TabOpen {
		return summary.Record.TabID
	}
	return "no tab"
}

func worktreeLabel(worktrees []workspace.Worktree) string {
	if len(worktrees) == 0 {
		return "-"
	}
	labels := make([]string, 0, len(worktrees))
	for _, worktree := range worktrees {
		labels = append(labels, worktree.Repository+":"+worktree.Path)
	}
	return strings.Join(labels, ", ")
}

func runThreadClose(ctx context.Context, app App, args []string) error {
	positionals, err := parse(newFlags("thread close"), args, 1)
	if err != nil {
		return err
	}
	threads, err := app.threads()
	if err != nil {
		return err
	}
	released, err := threads.Close(ctx, positionals[0])
	for _, path := range released {
		app.println("released", path)
	}
	if err != nil {
		return err
	}
	app.println("closed thread", positionals[0])
	return nil
}

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
		Now:        app.Now,
	}, nil
}

func runThreadCreate(ctx context.Context, app App, args []string) error {
	flags := newFlags("thread create")
	var repositories repeated
	flags.Var(&repositories, "repo", "repository to create a worktree in")
	mode := flags.String("mode", "", "argument bundle")
	runtimeName := flags.String("runtime", "", "runtime of the calling session")
	session := flags.String("session", "", "session ID of the calling session")
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
	for _, path := range created.Worktrees {
		app.println("worktree", path)
	}
	if err != nil {
		return err
	}
	app.println("created thread", created.Record.Name, "in tab", created.Record.TabID)
	if created.Readiness == herdr.Ready {
		app.println("sent the first message")
		return nil
	}
	app.println("the session in tab", created.Record.TabID, "is waiting for a confirmation; answer it in that tab. The first message was not sent.")
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

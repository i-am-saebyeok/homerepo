package cli

import (
	"context"
	"errors"
	"fmt"
	"text/tabwriter"

	"github.com/i-am-saebyeok/homerepo/internal/workspace"
)

const threadVariable = "HOMEREPO_THREAD"

func (app App) workspaces() (workspace.Workspaces, error) {
	loaded, err := app.loadHome()
	if err != nil {
		return workspace.Workspaces{}, err
	}
	return workspace.Workspaces{Home: loaded, Runner: app.Runner}, nil
}

func (app App) threadName(flagValue string) (string, error) {
	if flagValue != "" {
		return flagValue, nil
	}
	if fromEnvironment := app.Getenv(threadVariable); fromEnvironment != "" {
		return fromEnvironment, nil
	}
	return "", errors.New("thread name required: run this inside a thread (" + threadVariable + " is not set) or pass --thread <name>")
}

func parseWorkspaceArgs(name string, args []string) (string, string, error) {
	flags := newFlags(name)
	thread := flags.String("thread", "", "thread name")
	positionals, err := parse(flags, args, 1)
	if err != nil {
		return "", "", err
	}
	return positionals[0], *thread, nil
}

func runWorkspaceCreate(ctx context.Context, app App, args []string) error {
	repository, threadFlag, err := parseWorkspaceArgs("workspace create", args)
	if err != nil {
		return err
	}
	thread, err := app.threadName(threadFlag)
	if err != nil {
		return err
	}
	workspaces, err := app.workspaces()
	if err != nil {
		return err
	}
	path, err := workspaces.Create(ctx, repository, thread)
	if err != nil {
		return err
	}
	app.println(path)
	return nil
}

func runWorkspaceRelease(ctx context.Context, app App, args []string) error {
	repository, threadFlag, err := parseWorkspaceArgs("workspace release", args)
	if err != nil {
		return err
	}
	thread, err := app.threadName(threadFlag)
	if err != nil {
		return err
	}
	workspaces, err := app.workspaces()
	if err != nil {
		return err
	}
	path, err := workspaces.Release(ctx, repository, thread)
	if err != nil {
		return err
	}
	app.println("released", path, "(branch", thread, "kept)")
	return nil
}

func runWorkspaceList(ctx context.Context, app App, args []string) error {
	if _, err := parse(newFlags("workspace list"), args, 0); err != nil {
		return err
	}
	workspaces, err := app.workspaces()
	if err != nil {
		return err
	}
	worktrees, err := workspaces.List(ctx)
	if err != nil {
		return err
	}
	table := tabwriter.NewWriter(app.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(table, "REPOSITORY\tTHREAD\tPATH\tBRANCH\tCHANGES")
	for _, worktree := range worktrees {
		fmt.Fprintf(table, "%s\t%s\t%s\t%s\t%s\n", worktree.Repository, threadLabel(worktree), worktree.Path, worktree.Branch, worktree.Changes)
	}
	return table.Flush()
}

func threadLabel(worktree workspace.Worktree) string {
	if worktree.Outside {
		return "outside"
	}
	return worktree.Thread
}

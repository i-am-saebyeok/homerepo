package cli

import (
	"context"
	"fmt"
	"text/tabwriter"

	"github.com/i-am-saebyeok/homerepo/internal/home"
	"github.com/i-am-saebyeok/homerepo/internal/workspace"
)

func runInit(ctx context.Context, app App, args []string) error {
	if _, err := parse(newFlags("init"), args, 0); err != nil {
		return err
	}
	root, err := app.Getwd()
	if err != nil {
		return err
	}
	result, err := home.Init(root)
	if err != nil {
		return err
	}
	for _, path := range result.Created {
		app.println("created", path)
	}
	for _, path := range result.Skipped {
		app.println("skipped", path, "(exists)")
	}
	return nil
}

func runRepoList(ctx context.Context, app App, args []string) error {
	if _, err := parse(newFlags("repo list"), args, 0); err != nil {
		return err
	}
	loaded, err := app.loadHome()
	if err != nil {
		return err
	}
	table := tabwriter.NewWriter(app.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(table, "NAME\tREMOTE\tCLONE\tSTATE")
	for _, repository := range loaded.Repositories {
		clone := loaded.ClonePath(repository.Name)
		fmt.Fprintf(table, "%s\t%s\t%s\t%s\n", repository.Name, repository.Remote, clone, cloneState(clone))
	}
	return table.Flush()
}

func cloneState(path string) string {
	if home.IsDir(path) {
		return "present"
	}
	return "missing"
}

func runRepoPath(ctx context.Context, app App, args []string) error {
	positionals, err := parse(newFlags("repo path"), args, 1)
	if err != nil {
		return err
	}
	loaded, err := app.loadHome()
	if err != nil {
		return err
	}
	if _, err := loaded.Repository(positionals[0]); err != nil {
		return err
	}
	app.println(loaded.ClonePath(positionals[0]))
	return nil
}

func runRepoAdd(ctx context.Context, app App, args []string) error {
	flags := newFlags("repo add")
	remote := flags.String("remote", "", "`address` of the remote, as given to git clone (required)")
	branch := flags.String("branch", "", "`branch` that worktrees start from (default: the remote's default branch)")
	positionals, err := parse(flags, args, 1)
	if err != nil {
		return err
	}
	if *remote == "" {
		return usageOfFlags(flags, "--remote is required")
	}
	loaded, err := app.loadHome()
	if err != nil {
		return err
	}
	if err := loaded.CheckNewRepository(positionals[0]); err != nil {
		return err
	}
	if *branch == "" {
		if *branch, err = workspace.DefaultBranch(ctx, app.Runner, *remote); err != nil {
			return err
		}
	}
	if err := loaded.AddRepository(home.Repository{Name: positionals[0], Remote: *remote, Branch: *branch}); err != nil {
		return err
	}
	app.println("added", positionals[0], "to", home.DeclarationFile, "(remote", *remote+",", "branch", *branch+")")
	return nil
}

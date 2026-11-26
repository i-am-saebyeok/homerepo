package cli

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"strings"
	"text/tabwriter"
)

type handler func(ctx context.Context, app App, args []string) error

type commandSpec struct {
	name      string
	arguments string
	summary   string
	run       handler
}

var commands = []commandSpec{
	{name: "init", summary: "Create homerepo.yaml, .homerepo/ and the agent skill in the current folder. Existing files are kept.", run: runInit},
	{name: "repo list", summary: "List the declared repositories, their remotes and clones.", run: runRepoList},
	{name: "repo path", arguments: "<repository>", summary: "Print the path of a repository's clone.", run: runRepoPath},
	{name: "repo add", arguments: "<name> --remote <address> [--branch <branch>]", summary: "Declare a repository in homerepo.yaml. It is not cloned until a worktree needs it.", run: runRepoAdd},
	{name: "workspace create", arguments: "<repository> [--thread <name>]", summary: "Create the thread's worktree of a repository and print its path. If it exists, print the same path.", run: runWorkspaceCreate},
	{name: "workspace list", summary: "List the worktrees of the declared repositories.", run: runWorkspaceList},
	{name: "workspace release", arguments: "<repository> [--thread <name>]", summary: "Remove the thread's worktree of a repository if it has no uncommitted changes. The branch is kept.", run: runWorkspaceRelease},
	{name: "thread create", arguments: "<name> [--repo <repository>]... [--mode <bundle>] [--runtime claude|codex] [--session <id>]", summary: "Fork the calling session into a new Herdr tab. Prints the tab, then one line per worktree made with --repo.", run: runThreadCreate},
	{name: "thread list", summary: "List threads with their runtime, tab and worktrees.", run: runThreadList},
	{name: "thread close", arguments: "<name>", summary: "Release the thread's worktrees and close its tab.", run: runThreadClose},
}

func findCommand(name string) (commandSpec, bool) {
	for _, spec := range commands {
		if spec.name == name {
			return spec, true
		}
	}
	return commandSpec{}, false
}

func groupOf(args []string) string {
	if len(args) == 0 {
		return ""
	}
	for _, spec := range commands {
		if strings.HasPrefix(spec.name, args[0]+" ") {
			return args[0]
		}
	}
	return ""
}

func usageOf(group string) string {
	var usage strings.Builder
	usage.WriteString("usage:\n")
	for _, spec := range commands {
		if group == "" || strings.HasPrefix(spec.name, group+" ") {
			usage.WriteString("  " + spec.synopsis() + "\n")
		}
	}
	return usage.String()
}

func (c commandSpec) synopsis() string {
	return strings.TrimSpace("homerepo " + c.name + " " + c.arguments)
}

func (c commandSpec) help(flags *flag.FlagSet) string {
	var help bytes.Buffer
	fmt.Fprintf(&help, "usage: %s\n\n%s\n", c.synopsis(), c.summary)
	options := tabwriter.NewWriter(&help, 0, 0, 2, ' ', 0)
	first := true
	flags.VisitAll(func(option *flag.Flag) {
		if first {
			fmt.Fprint(&help, "\noptions:\n")
			first = false
		}
		value, description := flag.UnquoteUsage(option)
		fmt.Fprintf(options, "  --%s <%s>\t%s\n", option.Name, value, description)
	})
	options.Flush()
	return help.String()
}

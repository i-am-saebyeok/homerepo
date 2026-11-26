package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/i-am-saebyeok/homerepo/internal/command"
)

const (
	exitOK    = 0
	exitError = 1
	exitUsage = 2
)

type App struct {
	Stdout io.Writer
	Stderr io.Writer
	Getenv func(string) string
	Getwd  func() (string, error)
	Runner command.Runner
	Now    func() time.Time
	Sleep  func(time.Duration)
}

func (app App) Run(ctx context.Context, args []string) int {
	if len(args) == 1 && isHelpFlag(args[0]) {
		fmt.Fprint(app.Stdout, usageOf(""))
		return exitOK
	}
	spec, rest, found := lookup(args)
	if !found {
		return app.runGroup(args)
	}
	return app.finish(spec, spec.run(ctx, app, rest))
}

func (app App) runGroup(args []string) int {
	group := groupOf(args)
	if group != "" && len(args) == 2 && isHelpFlag(args[1]) {
		fmt.Fprint(app.Stdout, usageOf(group))
		return exitOK
	}
	if len(args) > 0 {
		fmt.Fprintf(app.Stderr, "homerepo: unknown command: %s\n", strings.Join(args, " "))
	}
	fmt.Fprint(app.Stderr, usageOf(group))
	return exitUsage
}

func (app App) finish(spec commandSpec, err error) int {
	if err == nil {
		return exitOK
	}
	var help helpRequested
	if errors.As(err, &help) {
		fmt.Fprint(app.Stdout, spec.help(help.flags))
		return exitOK
	}
	fmt.Fprintf(app.Stderr, "homerepo: %s\n", err)
	var usageErr usageError
	if errors.As(err, &usageErr) {
		fmt.Fprint(app.Stderr, spec.help(usageErr.flags))
		return exitUsage
	}
	return exitError
}

func lookup(args []string) (commandSpec, []string, bool) {
	if len(args) >= 2 {
		if spec, ok := findCommand(args[0] + " " + args[1]); ok {
			return spec, args[2:], true
		}
	}
	if len(args) >= 1 {
		if spec, ok := findCommand(args[0]); ok {
			return spec, args[1:], true
		}
	}
	return commandSpec{}, nil, false
}

func isHelpFlag(arg string) bool {
	return arg == "-h" || arg == "--help"
}

func (app App) println(values ...string) {
	fmt.Fprintln(app.Stdout, strings.Join(values, " "))
}

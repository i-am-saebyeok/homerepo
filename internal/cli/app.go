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
}

type handler func(ctx context.Context, app App, args []string) error

type usageError struct {
	message string
}

func (e usageError) Error() string {
	return e.message
}

func usagef(format string, values ...any) error {
	return usageError{message: fmt.Sprintf(format, values...)}
}

const usage = `usage:
  homerepo init
  homerepo repo list
  homerepo repo path <repository>
  homerepo workspace create <repository> [--thread <name>]
  homerepo workspace list
  homerepo workspace release <repository> [--thread <name>]
  homerepo thread create <name> [--repo <repository>]... [--mode <bundle>] [--runtime claude|codex] [--session <id>]
  homerepo thread list
  homerepo thread close <name>`

func (app App) Run(ctx context.Context, args []string) int {
	run, rest, ok := lookup(args)
	if !ok {
		fmt.Fprintln(app.Stderr, usage)
		return exitUsage
	}
	err := run(ctx, app, rest)
	if err == nil {
		return exitOK
	}
	fmt.Fprintf(app.Stderr, "homerepo: %s\n", err)
	var usageErr usageError
	if errors.As(err, &usageErr) {
		return exitUsage
	}
	return exitError
}

func lookup(args []string) (handler, []string, bool) {
	if len(args) >= 2 {
		if run, ok := commands[args[0]+" "+args[1]]; ok {
			return run, args[2:], true
		}
	}
	if len(args) >= 1 {
		if run, ok := commands[args[0]]; ok {
			return run, args[1:], true
		}
	}
	return nil, nil, false
}

func (app App) println(values ...string) {
	fmt.Fprintln(app.Stdout, strings.Join(values, " "))
}

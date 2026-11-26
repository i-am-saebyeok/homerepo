package dogfood

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/i-am-saebyeok/homerepo/internal/command"
	"github.com/i-am-saebyeok/homerepo/internal/herdr"
)

const (
	exitPassed = 0
	exitFailed = 1
	exitUsage  = 2
)

const usage = `usage: go run ./dogfood --root <folder> --runtime claude|codex [--timeout <duration>] [--runtime-arg <arg>]... <scenario>`

const defaultTimeout = 30 * time.Minute

type App struct {
	Stdout   io.Writer
	Stderr   io.Writer
	Getenv   func(string) string
	LookPath func(string) (string, error)
	Runner   command.Runner
	Now      func() time.Time
	Sleep    func(time.Duration)
}

type options struct {
	root        string
	runtime     Runtime
	timeout     time.Duration
	runtimeArgs []string
	scenario    Scenario
}

type repeated []string

func (r *repeated) String() string {
	return fmt.Sprint([]string(*r))
}

func (r *repeated) Set(value string) error {
	*r = append(*r, value)
	return nil
}

func (app App) Run(ctx context.Context, args []string) int {
	chosen, err := parseOptions(args)
	if err != nil {
		fmt.Fprintf(app.Stderr, "dogfood: %s\n%s\n", err, usage)
		return exitUsage
	}
	if err := app.checkRequirements(chosen); err != nil {
		fmt.Fprintf(app.Stderr, "dogfood: nothing was created:\n%s\n", err)
		return exitFailed
	}
	started := app.Now()
	session, err := app.start(ctx, chosen, started.Add(chosen.timeout))
	if err != nil {
		fmt.Fprintf(app.Stderr, "dogfood: %s\n", err)
		return exitFailed
	}
	observation := session.observe(ctx)
	checks := session.check(ctx, observation)
	session.report(observation, checks, app.Now().Sub(started))
	if observation.timedOut || !allPassed(checks) {
		return exitFailed
	}
	return exitPassed
}

func parseOptions(args []string) (options, error) {
	flags := flag.NewFlagSet("dogfood", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	root := flags.String("root", "", "folder that holds the test Home, the new repository and the remote")
	runtimeName := flags.String("runtime", "", "claude or codex")
	timeout := flags.Duration("timeout", defaultTimeout, "time limit for the whole run")
	var runtimeArgs repeated
	flags.Var(&runtimeArgs, "runtime-arg", "argument added when a runtime session starts")
	var positionals []string
	for {
		if err := flags.Parse(args); err != nil {
			return options{}, err
		}
		if flags.NArg() == 0 {
			break
		}
		positionals = append(positionals, flags.Arg(0))
		args = flags.Args()[1:]
	}
	var problems []error
	if *root == "" {
		problems = append(problems, errors.New("--root is required"))
	}
	if *timeout <= 0 {
		problems = append(problems, errors.New("--timeout must be positive"))
	}
	runtime, err := checkRuntime(*runtimeName)
	problems = append(problems, err)
	if len(positionals) != 1 {
		return options{}, errors.Join(append(problems, fmt.Errorf("expected one scenario, got %d", len(positionals)))...)
	}
	scenario, err := LookupScenario(positionals[0])
	problems = append(problems, err)
	if err := errors.Join(problems...); err != nil {
		return options{}, err
	}
	absoluteRoot, err := filepath.Abs(*root)
	if err != nil {
		return options{}, err
	}
	return options{root: absoluteRoot, runtime: runtime, timeout: *timeout, runtimeArgs: runtimeArgs, scenario: scenario}, nil
}

func (app App) checkRequirements(chosen options) error {
	var problems []error
	if app.Getenv("HERDR_ENV") != "1" {
		problems = append(problems, errors.New("Herdr not found: HERDR_ENV is not 1; run this inside a Herdr pane"))
	}
	if app.Getenv("HERDR_WORKSPACE_ID") == "" {
		problems = append(problems, errors.New("Herdr workspace not found: HERDR_WORKSPACE_ID is empty"))
	}
	for _, program := range []string{"herdr", "git", "go", string(chosen.runtime)} {
		if _, err := app.LookPath(program); err != nil {
			problems = append(problems, fmt.Errorf("%s not found in PATH", program))
		}
	}
	problems = append(problems, checkRootIsOurs(Layout{Root: chosen.root}))
	return errors.Join(problems...)
}

func (app App) herdr() herdr.Client {
	return herdr.Client{Runner: app.Runner, Sleep: app.Sleep}
}

func (app App) remaining(deadline time.Time) time.Duration {
	return deadline.Sub(app.Now())
}

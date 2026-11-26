package command_test

import (
	"context"
	"errors"
	"testing"

	"github.com/i-am-saebyeok/homerepo/internal/command"
	"github.com/i-am-saebyeok/homerepo/internal/command/commandtest"
)

func TestExecReturnsStdout(t *testing.T) {
	output, err := command.Exec{}.Run(context.Background(), command.Command{Name: "git", Args: []string{"--version"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(output.Stdout) == 0 {
		t.Fatal("expected output from git --version")
	}
}

func TestExecReportsFailureWithStderr(t *testing.T) {
	_, err := command.Exec{}.Run(context.Background(), command.Command{Name: "git", Args: []string{"no-such-subcommand"}, Dir: t.TempDir()})
	var failure *command.Failure
	if !errors.As(err, &failure) {
		t.Fatalf("expected a failure, got %v", err)
	}
	if failure.ExitCode == 0 || failure.Stderr == "" {
		t.Fatalf("expected exit code and stderr, got %+v", failure)
	}
}

func TestFakeRoutesByProgramName(t *testing.T) {
	fake := &commandtest.Fake{Handlers: map[string]commandtest.Handler{
		"herdr": func(cmd command.Command) (command.Output, error) {
			return command.Output{Stdout: []byte("ok")}, nil
		},
	}}
	output, err := fake.Run(context.Background(), command.Command{Name: "herdr", Args: []string{"tab", "list"}})
	if err != nil || string(output.Stdout) != "ok" {
		t.Fatalf("unexpected result %q, %v", output.Stdout, err)
	}
	if _, err := fake.Run(context.Background(), command.Command{Name: "git"}); err == nil {
		t.Fatal("expected an error for a command without a handler or fallback")
	}
	if len(fake.Handled) != 1 {
		t.Fatalf("expected one handled call, got %d", len(fake.Handled))
	}
}

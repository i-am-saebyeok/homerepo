package command_test

import (
	"bytes"
	"context"
	"errors"
	"path/filepath"
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

func TestExecAddsEnvironmentAndRunsInDir(t *testing.T) {
	dir := t.TempDir()
	output, err := command.Exec{}.Run(context.Background(), command.Command{
		Name:        "sh",
		Args:        []string{"-c", `printf '%s|%s' "$HOMEREPO_TEST_VALUE" "$(pwd -P)"`},
		Dir:         dir,
		Environment: []string{"HOMEREPO_TEST_VALUE=two words"},
	})
	if err != nil {
		t.Fatal(err)
	}
	resolved, err := filepath.EvalSymlinks(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(output.Stdout), "two words|"+resolved; got != want {
		t.Fatalf("output %q, want %q", got, want)
	}
}

func TestExecStreamsBothOutputsWhenAsked(t *testing.T) {
	var stream bytes.Buffer
	output, err := command.Exec{}.Run(context.Background(), command.Command{
		Name:   "sh",
		Args:   []string{"-c", "echo out; echo err 1>&2; exit 3"},
		Stream: &stream,
	})
	var failure *command.Failure
	if !errors.As(err, &failure) || failure.ExitCode != 3 {
		t.Fatalf("expected exit 3, got %v", err)
	}
	if stream.String() != "out\nerr\n" {
		t.Fatalf("stream %q", stream.String())
	}
	if len(output.Stdout) != 0 || len(output.Stderr) != 0 {
		t.Fatalf("streamed output was also captured: %+v", output)
	}
}

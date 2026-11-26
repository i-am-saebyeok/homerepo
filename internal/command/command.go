package command

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

type Command struct {
	Name        string
	Args        []string
	Dir         string
	Environment []string
	Stream      io.Writer
}

func (c Command) String() string {
	return strings.Join(append([]string{c.Name}, c.Args...), " ")
}

type Output struct {
	Stdout []byte
	Stderr []byte
}

type Runner interface {
	Run(ctx context.Context, cmd Command) (Output, error)
}

type Failure struct {
	Command  Command
	ExitCode int
	Stderr   string
}

func (f *Failure) Error() string {
	return fmt.Sprintf("%s: exit %d: %s", f.Command, f.ExitCode, strings.TrimSpace(f.Stderr))
}

type Exec struct{}

func (Exec) Run(ctx context.Context, cmd Command) (Output, error) {
	process := exec.CommandContext(ctx, cmd.Name, cmd.Args...)
	process.Dir = cmd.Dir
	if len(cmd.Environment) > 0 {
		process.Env = append(os.Environ(), cmd.Environment...)
	}
	var stdout, stderr bytes.Buffer
	process.Stdout = outputTarget(cmd.Stream, &stdout)
	process.Stderr = outputTarget(cmd.Stream, &stderr)
	err := process.Run()
	output := Output{Stdout: stdout.Bytes(), Stderr: stderr.Bytes()}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return output, &Failure{Command: cmd, ExitCode: exitErr.ExitCode(), Stderr: stderr.String()}
	}
	if err != nil {
		return output, fmt.Errorf("%s: %w", cmd.Name, err)
	}
	return output, nil
}

func outputTarget(stream io.Writer, capture *bytes.Buffer) io.Writer {
	if stream != nil {
		return stream
	}
	return capture
}

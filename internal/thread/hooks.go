package thread

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/i-am-saebyeok/homerepo/internal/command"
	"github.com/i-am-saebyeok/homerepo/internal/herdr"
	"github.com/i-am-saebyeok/homerepo/internal/home"
)

const (
	preCreateHook  = "pre-thread-create"
	postCreateHook = "post-thread-create"
)

type Hooks struct {
	HomeRoot string
	Runner   command.Runner
	Output   io.Writer
}

type HookFailure struct {
	Hook string
	Err  error
}

func (e *HookFailure) Error() string {
	var failure *command.Failure
	if errors.As(e.Err, &failure) {
		return fmt.Sprintf("hook %s exited with code %d", e.Hook, failure.ExitCode)
	}
	return fmt.Sprintf("hook %s could not run: %s", e.Hook, e.Err)
}

func (e *HookFailure) Unwrap() error {
	return e.Err
}

type hookValues struct {
	thread     string
	runtime    Runtime
	attachDirs []string
}

func (v hookValues) environment() []string {
	return []string{
		"HOMEREPO_HOOK_THREAD=" + v.thread,
		"HOMEREPO_HOOK_RUNTIME=" + string(v.runtime),
		"HOMEREPO_HOOK_DIRS=" + strings.Join(v.attachDirs, "\n"),
	}
}

func (h Hooks) beforeCreate(ctx context.Context, values hookValues) error {
	return h.run(ctx, preCreateHook, values.environment())
}

func (h Hooks) afterCreate(ctx context.Context, values hookValues, tab herdr.Tab) error {
	return h.run(ctx, postCreateHook, append(values.environment(), "HOMEREPO_HOOK_TAB="+tab.ID, "HOMEREPO_HOOK_PANE="+tab.Pane))
}

func (h Hooks) run(ctx context.Context, name string, environment []string) error {
	path := filepath.Join(h.HomeRoot, home.LocalDir, "hooks", name)
	if _, err := os.Stat(path); errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	_, err := h.Runner.Run(ctx, command.Command{Name: path, Dir: h.HomeRoot, Environment: environment, Stream: h.Output})
	if err != nil {
		return &HookFailure{Hook: name, Err: err}
	}
	return nil
}

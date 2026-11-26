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
	"github.com/i-am-saebyeok/homerepo/internal/home"
)

const (
	preCreateHook  = "pre-thread-create"
	postCreateHook = "post-thread-create"
	preCloseHook   = "pre-thread-close"
	postCloseHook  = "post-thread-close"
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
	tab        string
	pane       string
}

func (v hookValues) environment() []string {
	environment := []string{
		"HOMEREPO_HOOK_THREAD=" + v.thread,
		"HOMEREPO_HOOK_RUNTIME=" + string(v.runtime),
		"HOMEREPO_HOOK_DIRS=" + strings.Join(v.attachDirs, "\n"),
	}
	if v.tab != "" {
		environment = append(environment, "HOMEREPO_HOOK_TAB="+v.tab)
	}
	if v.pane != "" {
		environment = append(environment, "HOMEREPO_HOOK_PANE="+v.pane)
	}
	return environment
}

func (v hookValues) inTab(tab string, pane string) hookValues {
	v.tab = tab
	v.pane = pane
	return v
}

func (h Hooks) path(name string) string {
	return filepath.Join(h.HomeRoot, home.LocalDir, "hooks", name)
}

func (h Hooks) has(name string) bool {
	_, err := os.Stat(h.path(name))
	return !errors.Is(err, fs.ErrNotExist)
}

func (h Hooks) run(ctx context.Context, name string, values hookValues) error {
	if !h.has(name) {
		return nil
	}
	_, err := h.Runner.Run(ctx, command.Command{Name: h.path(name), Dir: h.HomeRoot, Environment: values.environment(), Stream: h.Output})
	if err != nil {
		return &HookFailure{Hook: name, Err: err}
	}
	return nil
}

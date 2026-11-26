package thread

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/i-am-saebyeok/homerepo/internal/home"
)

type Runtime string

const (
	Claude Runtime = "claude"
	Codex  Runtime = "codex"
)

type runtimeAdapter struct {
	runtime         Runtime
	program         string
	sessionVariable string
	present         func(getenv func(string) string) bool
	fork            func(session string) []string
}

var runtimeAdapters = []runtimeAdapter{
	{
		runtime:         Claude,
		program:         "claude",
		sessionVariable: "CLAUDE_CODE_SESSION_ID",
		present:         func(getenv func(string) string) bool { return getenv("CLAUDECODE") == "1" },
		fork:            func(session string) []string { return []string{"--resume", session, "--fork-session"} },
	},
	{
		runtime:         Codex,
		program:         "codex",
		sessionVariable: "CODEX_SESSION_ID",
		present:         func(getenv func(string) string) bool { return getenv("CODEX_SESSION_ID") != "" },
		fork:            func(session string) []string { return []string{"fork", session} },
	},
}

type Caller struct {
	Runtime   Runtime
	Session   string
	Workspace string
	Pane      string
}

type CallerOverrides struct {
	Runtime string
	Session string
}

func DetectCaller(getenv func(string) string, overrides CallerOverrides) (Caller, error) {
	if current := getenv(EnvironmentVariable); current != "" {
		return Caller{}, fmt.Errorf("already inside thread %q, so no thread was created; run `homerepo workspace create <repository>` and work in the path it prints", current)
	}
	var problems []error
	var caller Caller
	adapter, err := detectRuntime(getenv, overrides.Runtime)
	if err != nil {
		problems = append(problems, err)
	} else {
		caller.Runtime = adapter.runtime
		caller.Session, err = detectSession(getenv, adapter, overrides.Session)
		problems = append(problems, err)
	}
	caller.Workspace, caller.Pane, err = detectHerdr(getenv)
	problems = append(problems, err)
	return caller, errors.Join(problems...)
}

func detectRuntime(getenv func(string) string, override string) (runtimeAdapter, error) {
	if override != "" {
		return adapterFor(Runtime(override))
	}
	var present []runtimeAdapter
	for _, adapter := range runtimeAdapters {
		if adapter.present(getenv) {
			present = append(present, adapter)
		}
	}
	if len(present) == 1 {
		return present[0], nil
	}
	if len(present) > 1 {
		return runtimeAdapter{}, errors.New("runtime is ambiguous: both CLAUDECODE=1 and CODEX_SESSION_ID are set; pass --runtime " + runtimeNames())
	}
	return runtimeAdapter{}, errors.New("runtime not found: neither CLAUDECODE=1 nor CODEX_SESSION_ID is set; pass --runtime " + runtimeNames())
}

func detectSession(getenv func(string) string, adapter runtimeAdapter, override string) (string, error) {
	if override != "" {
		return override, nil
	}
	if session := getenv(adapter.sessionVariable); session != "" {
		return session, nil
	}
	return "", fmt.Errorf("session ID not found: %s is empty; pass --session <id>", adapter.sessionVariable)
}

func detectHerdr(getenv func(string) string) (string, string, error) {
	if getenv("HERDR_ENV") != "1" {
		return "", "", errors.New("Herdr not found: HERDR_ENV is not 1; run this inside a Herdr pane")
	}
	workspace := getenv("HERDR_WORKSPACE_ID")
	pane := getenv("HERDR_PANE_ID")
	var problems []error
	if workspace == "" {
		problems = append(problems, errors.New("Herdr workspace not found: HERDR_WORKSPACE_ID is empty"))
	}
	if pane == "" {
		problems = append(problems, errors.New("Herdr pane not found: HERDR_PANE_ID is empty"))
	}
	return workspace, pane, errors.Join(problems...)
}

func adapterFor(runtime Runtime) (runtimeAdapter, error) {
	for _, adapter := range runtimeAdapters {
		if adapter.runtime == runtime {
			return adapter, nil
		}
	}
	return runtimeAdapter{}, fmt.Errorf("unknown runtime %q; use %s", runtime, runtimeNames())
}

func runtimeNames() string {
	names := make([]string, 0, len(runtimeAdapters))
	for _, adapter := range runtimeAdapters {
		names = append(names, string(adapter.runtime))
	}
	return strings.Join(names, " or ")
}

func launchArgs(adapter runtimeAdapter, session string, attachDirs []string, bundle []string) []string {
	args := adapter.fork(session)
	for _, dir := range attachDirs {
		args = append(args, "--add-dir", dir)
	}
	return append(args, bundle...)
}

func argumentBundle(local home.LocalConfig, runtime Runtime, mode string) ([]string, error) {
	bundles := local.Runtimes[string(runtime)].Thread
	if mode == "" {
		return bundles["default"], nil
	}
	if args, ok := bundles[mode]; ok {
		return args, nil
	}
	return nil, fmt.Errorf("no argument bundle %q for %s in runtimes.%s.thread; available: %s", mode, runtime, runtime, bundleNames(bundles))
}

func bundleNames(bundles map[string][]string) string {
	if len(bundles) == 0 {
		return "none"
	}
	names := make([]string, 0, len(bundles))
	for name := range bundles {
		names = append(names, name)
	}
	sort.Strings(names)
	return strings.Join(names, ", ")
}

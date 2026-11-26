package cli_test

import (
	"bytes"
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/i-am-saebyeok/homerepo/internal/cli"
	"github.com/i-am-saebyeok/homerepo/internal/command"
	"github.com/i-am-saebyeok/homerepo/internal/command/commandtest"
	"github.com/i-am-saebyeok/homerepo/internal/gittest"
)

type fixture struct {
	t       *testing.T
	root    string
	home    string
	cwd     string
	env     map[string]string
	runner  *commandtest.Fake
	remotes map[string]string
}

type result struct {
	stdout string
	stderr string
	code   int
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	gittest.Isolate(t)
	root := t.TempDir()
	f := &fixture{
		t:       t,
		root:    root,
		home:    filepath.Join(root, "home"),
		env:     map[string]string{},
		runner:  &commandtest.Fake{Handlers: map[string]commandtest.Handler{}, Fallback: command.Exec{}},
		remotes: map[string]string{},
	}
	f.cwd = f.home
	declaration := "repositories:\n"
	for _, name := range []string{"frontend", "backend"} {
		f.remotes[name] = gittest.BareRemote(t, filepath.Join(root, "remotes", name+".git"))
		declaration += "  - name: " + name + "\n    remote: " + f.remotes[name] + "\n    branch: main\n"
	}
	gittest.WriteFile(t, filepath.Join(f.home, "homerepo.yaml"), declaration)
	return f
}

func (f *fixture) run(args ...string) result {
	f.t.Helper()
	var stdout, stderr bytes.Buffer
	app := cli.App{
		Stdout: &stdout,
		Stderr: &stderr,
		Getenv: func(key string) string { return f.env[key] },
		Getwd:  func() (string, error) { return f.cwd, nil },
		Runner: f.runner,
		Now:    func() time.Time { return time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC) },
	}
	code := app.Run(context.Background(), args)
	return result{stdout: stdout.String(), stderr: stderr.String(), code: code}
}

func (f *fixture) mustRun(args ...string) result {
	f.t.Helper()
	got := f.run(args...)
	if got.code != 0 {
		f.t.Fatalf("%v exited %d\nstdout:\n%s\nstderr:\n%s", args, got.code, got.stdout, got.stderr)
	}
	return got
}

func (f *fixture) clonePath(name string) string {
	return filepath.Join(f.root, name)
}

func (f *fixture) clone(name string) string {
	f.t.Helper()
	gittest.Git(f.t, f.root, "clone", "--quiet", f.remotes[name], f.clonePath(name))
	return f.clonePath(name)
}

func (f *fixture) writeLocalConfig(content string) {
	f.t.Helper()
	gittest.WriteFile(f.t, filepath.Join(f.home, ".homerepo", "config.yaml"), content)
}

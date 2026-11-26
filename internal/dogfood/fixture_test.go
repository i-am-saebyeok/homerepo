package dogfood_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/i-am-saebyeok/homerepo/internal/cli"
	"github.com/i-am-saebyeok/homerepo/internal/command"
	"github.com/i-am-saebyeok/homerepo/internal/command/commandtest"
	"github.com/i-am-saebyeok/homerepo/internal/dogfood"
	"github.com/i-am-saebyeok/homerepo/internal/gittest"
	"github.com/i-am-saebyeok/homerepo/internal/herdr"
	"github.com/i-am-saebyeok/homerepo/internal/thread"
)

const (
	workspaceID  = "w1"
	repository   = "notes-app"
	threadName   = "first-change"
	searchPath   = "/usr/bin"
	scenarioName = "new-repository"
)

type world struct {
	t          *testing.T
	root       string
	layout     dogfood.Layout
	env        map[string]string
	missing    []string
	herdr      *fakeHerdr
	runner     *commandtest.Fake
	now        time.Time
	coordinate func(w *world)
	work       func(w *world) error
}

type outcome struct {
	stdout string
	stderr string
	code   int
}

func newWorld(t *testing.T) *world {
	t.Helper()
	gittest.Isolate(t)
	root := filepath.Join(t.TempDir(), "dogfood-root")
	w := &world{
		t:      t,
		root:   root,
		layout: dogfood.Layout{Root: root},
		env:    map[string]string{"HERDR_ENV": "1", "HERDR_WORKSPACE_ID": workspaceID, "PATH": searchPath},
		now:    time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
	}
	w.herdr = newFakeHerdr(w)
	w.runner = &commandtest.Fake{Fallback: command.Exec{}, Handlers: map[string]commandtest.Handler{
		"herdr":             w.herdr.handle,
		"go":                w.buildHomerepo,
		w.layout.Homerepo(): w.runHomerepo,
	}}
	w.coordinate = func(w *world) { w.createRepositoryAndThread() }
	w.work = func(w *world) error {
		w.commitInWorktree()
		w.push(threadName)
		return nil
	}
	return w
}

func (w *world) run(args ...string) outcome {
	w.t.Helper()
	var stdout, stderr bytes.Buffer
	app := dogfood.App{
		Stdout: &stdout,
		Stderr: &stderr,
		Getenv: func(key string) string { return w.env[key] },
		LookPath: func(program string) (string, error) {
			if slices.Contains(w.missing, program) {
				return "", errors.New("not found")
			}
			return "/fake/" + program, nil
		},
		Runner: w.runner,
		Now:    func() time.Time { return w.now },
		Sleep:  func(d time.Duration) { w.now = w.now.Add(d) },
	}
	code := app.Run(context.Background(), args)
	return outcome{stdout: stdout.String(), stderr: stderr.String(), code: code}
}

func (w *world) runScenario(extra ...string) outcome {
	w.t.Helper()
	args := append([]string{"--root", w.root, "--runtime", "claude", "--timeout", "10m"}, extra...)
	return w.run(append(args, scenarioName)...)
}

func (w *world) buildHomerepo(cmd command.Command) (command.Output, error) {
	output := cmd.Args[slices.Index(cmd.Args, "-o")+1]
	if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil {
		return command.Output{}, err
	}
	return command.Output{}, os.WriteFile(output, []byte("built\n"), 0o755)
}

func (w *world) runHomerepo(cmd command.Command) (command.Output, error) {
	var stdout, stderr bytes.Buffer
	app := cli.App{
		Stdout: &stdout,
		Stderr: &stderr,
		Getenv: func(string) string { return "" },
		Getwd:  func() (string, error) { return cmd.Dir, nil },
		Runner: command.Exec{},
		Now:    func() time.Time { return w.now },
	}
	if code := app.Run(context.Background(), cmd.Args); code != 0 {
		return command.Output{Stdout: stdout.Bytes()}, commandtest.Fail(cmd, stderr.String())
	}
	return command.Output{Stdout: stdout.Bytes()}, nil
}

func (w *world) clone() string {
	return filepath.Join(w.root, repository)
}

func (w *world) worktree() string {
	return filepath.Join(w.clone(), ".worktrees", threadName)
}

func (w *world) createRepositoryAndThread() {
	w.createRepository()
	w.createThread()
}

func (w *world) createRepository() {
	gittest.Repository(w.t, w.clone())
	gittest.Git(w.t, w.clone(), "remote", "add", "origin", w.layout.Remote(repository))
	gittest.Git(w.t, w.clone(), "push", "--quiet", "origin", "main")
	gittest.WriteFile(w.t, filepath.Join(w.layout.Home(), "homerepo.yaml"), fmt.Sprintf("repositories:\n  - name: %s\n    remote: %s\n    branch: main\n", repository, w.layout.Remote(repository)))
}

func (w *world) createThread() {
	tab := w.herdr.newTab()
	w.herdr.statuses[tab.Pane] = []herdr.AgentStatus{herdr.AgentAbsent, herdr.AgentIdle}
	w.herdr.threadPane = tab.Pane
	record := thread.Record{Name: threadName, Runtime: thread.Claude, TabID: tab.ID, CreatedAt: w.now}
	if err := (thread.Records{HomeRoot: w.layout.Home()}).Write(record); err != nil {
		w.t.Fatal(err)
	}
}

func (w *world) createThreadWithHomerepo() {
	var stdout, stderr bytes.Buffer
	env := map[string]string{"CLAUDECODE": "1", "CLAUDE_CODE_SESSION_ID": "session-1", "HERDR_ENV": "1", "HERDR_WORKSPACE_ID": workspaceID, "HERDR_PANE_ID": w.herdr.coordinatorPane}
	app := cli.App{
		Stdout: &stdout,
		Stderr: &stderr,
		Getenv: func(key string) string { return env[key] },
		Getwd:  func() (string, error) { return w.layout.Home(), nil },
		Runner: w.runner,
		Now:    func() time.Time { return w.now },
	}
	if code := app.Run(context.Background(), []string{"thread", "create", threadName, "--repo", repository}); code != 0 {
		w.t.Fatalf("thread create exited %d: %s", code, stderr.String())
	}
}

func (w *world) commitInWorktree() {
	gittest.Git(w.t, w.clone(), "worktree", "add", "--quiet", "-b", threadName, w.worktree(), "main")
	gittest.WriteFile(w.t, filepath.Join(w.worktree(), "hello.txt"), "hello\n")
	gittest.Git(w.t, w.worktree(), "add", "hello.txt")
	gittest.Git(w.t, w.worktree(), "commit", "--quiet", "--message", "first change")
	gittest.WriteFile(w.t, filepath.Join(w.worktree(), dogfood.NotesFile), "notes\n")
}

func (w *world) push(branch string) {
	gittest.Git(w.t, w.worktree(), "push", "--quiet", "origin", branch)
}

func (w *world) readFile(path string) string {
	w.t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		w.t.Fatal(err)
	}
	return string(content)
}

func exists(path string) bool {
	_, err := os.Lstat(path)
	return err == nil
}

func mustContain(t *testing.T, text string, parts ...string) {
	t.Helper()
	for _, part := range parts {
		if !strings.Contains(text, part) {
			t.Fatalf("missing %q in:\n%s", part, text)
		}
	}
}

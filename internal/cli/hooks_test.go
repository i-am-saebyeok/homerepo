package cli_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type hookRecorder struct {
	t   *testing.T
	dir string
}

func (f *fixture) hookPath(name string) string {
	return filepath.Join(f.home, ".homerepo", "hooks", name)
}

func (f *fixture) writeHook(name string, body string, mode os.FileMode) {
	f.t.Helper()
	path := f.hookPath(name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		f.t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+body), mode); err != nil {
		f.t.Fatal(err)
	}
}

func (f *fixture) recordingHook(name string, exitCode string) *hookRecorder {
	f.t.Helper()
	recorder := &hookRecorder{t: f.t, dir: f.t.TempDir()}
	out := filepath.Join(recorder.dir, name)
	f.writeHook(name, `out='`+out+`'
touch "$out.ran"
printf '%s' "$HOMEREPO_HOOK_THREAD" > "$out.THREAD"
printf '%s' "$HOMEREPO_HOOK_RUNTIME" > "$out.RUNTIME"
printf '%s' "$HOMEREPO_HOOK_DIRS" > "$out.DIRS"
if [ -n "${HOMEREPO_HOOK_TAB+set}" ]; then printf '%s' "$HOMEREPO_HOOK_TAB" > "$out.TAB"; fi
if [ -n "${HOMEREPO_HOOK_PANE+set}" ]; then printf '%s' "$HOMEREPO_HOOK_PANE" > "$out.PANE"; fi
pwd -P > "$out.PWD"
for target in "$@"; do printf '%s' "$target" > "$out.ARGS"; done
exit `+exitCode+`
`, 0o755)
	return recorder
}

func (r *hookRecorder) ran(name string) bool {
	return exists(filepath.Join(r.dir, name+".ran"))
}

func (r *hookRecorder) value(name string, key string) (string, bool) {
	content, err := os.ReadFile(filepath.Join(r.dir, name+"."+key))
	if err != nil {
		return "", false
	}
	return string(content), true
}

func (r *hookRecorder) mustValue(name string, key string) string {
	r.t.Helper()
	value, ok := r.value(name, key)
	if !ok {
		r.t.Fatalf("hook %s did not record %s", name, key)
	}
	return value
}

func resolvedPath(t *testing.T, path string) string {
	t.Helper()
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		t.Fatal(err)
	}
	return resolved
}

func TestThreadCreateWithoutHooksIsUnchanged(t *testing.T) {
	f := newFixture(t)
	fake := f.useHerdr()
	f.enterClaudeSession()

	got := f.mustRun("thread", "create", "login-fix")

	if got.stdout != "created thread login-fix in tab w1:t1\n" || got.stderr != "" {
		t.Fatalf("stdout %q, stderr %q", got.stdout, got.stderr)
	}
	if len(fake.calls) != 2 {
		t.Fatalf("herdr calls %q", fake.calls)
	}
}

func TestThreadCreateRunsPreHookBeforeWorktreesAndTab(t *testing.T) {
	f := newFixture(t)
	marks := t.TempDir()
	ranMark := filepath.Join(marks, "ran")
	worktreeMark := filepath.Join(marks, "worktree-existed")
	f.writeHook("pre-thread-create", `[ -e '`+f.worktreePath("frontend", "login-fix")+`' ] && touch '`+worktreeMark+`'
touch '`+ranMark+`'
exit 0
`, 0o755)
	fake := f.useHerdr()
	hookRanBeforeTab := false
	fake.onTabCreate = func() { hookRanBeforeTab = exists(ranMark) }
	f.enterClaudeSession()

	f.mustRun("thread", "create", "login-fix", "--repo", "frontend")

	if !hookRanBeforeTab {
		t.Fatal("the hook did not run before the tab was created")
	}
	if exists(worktreeMark) {
		t.Fatal("the worktree existed when the hook ran")
	}
	if !exists(f.worktreePath("frontend", "login-fix")) || len(fake.callsOf("pane run")) != 1 {
		t.Fatal("the thread was not created after the hook")
	}
}

func TestThreadCreateStopsWhenPreHookFails(t *testing.T) {
	f := newFixture(t)
	f.clone("backend")
	f.recordingHook("pre-thread-create", "3")
	fake := f.useHerdr()
	f.enterClaudeSession()

	got := f.run("thread", "create", "login-fix", "--repo", "frontend")

	if got.code == 0 || !strings.Contains(got.stderr, "hook pre-thread-create exited with code 3") {
		t.Fatalf("code %d, stderr:\n%s", got.code, got.stderr)
	}
	if len(fake.calls) != 0 || exists(f.recordPath("login-fix")) || exists(f.clonePath("frontend")) {
		t.Fatal("a tab, record or clone was created")
	}
	if exists(filepath.Join(f.clonePath("backend"), ".worktrees")) {
		t.Fatal("a worktrees folder was created")
	}
}

func TestThreadCreateRejectedRequestsRunNoHook(t *testing.T) {
	cases := map[string]struct {
		prepare func(f *fixture)
		args    []string
	}{
		"inside a thread":  {prepare: func(f *fixture) { f.env["HOMEREPO_THREAD"] = "other" }, args: []string{"login-fix"}},
		"overlapping name": {prepare: func(f *fixture) { f.mustRun("thread", "create", "fix") }, args: []string{"fix/login"}},
		"unknown repo":     {prepare: func(f *fixture) {}, args: []string{"login-fix", "--repo", "mobile"}},
		"unknown bundle":   {prepare: func(f *fixture) {}, args: []string{"login-fix", "--mode", "careful"}},
		"unusable name":    {prepare: func(f *fixture) {}, args: []string{"../escape"}},
		"existing name":    {prepare: func(f *fixture) { f.mustRun("thread", "create", "login-fix") }, args: []string{"login-fix"}},
	}
	for name, testCase := range cases {
		t.Run(name, func(t *testing.T) {
			f := newFixture(t)
			f.useHerdr()
			f.enterClaudeSession()
			testCase.prepare(f)
			hooks := f.recordingHook("pre-thread-create", "0")
			post := f.recordingHook("post-thread-create", "0")

			got := f.run(append([]string{"thread", "create"}, testCase.args...)...)

			if got.code == 0 {
				t.Fatal("the request was accepted")
			}
			if hooks.ran("pre-thread-create") || post.ran("post-thread-create") {
				t.Fatal("a hook ran for a rejected request")
			}
		})
	}
}

func TestPostHookReceivesThreadTabAndPane(t *testing.T) {
	f := newFixture(t)
	hooks := f.recordingHook("post-thread-create", "0")
	fake := f.useHerdr()
	hookRanBeforePaneRun := false
	fake.onPaneRun = func() { hookRanBeforePaneRun = hooks.ran("post-thread-create") }
	f.enterClaudeSession()

	f.mustRun("thread", "create", "login-fix")

	if hookRanBeforePaneRun || !hooks.ran("post-thread-create") {
		t.Fatal("the hook did not run after the fork command")
	}
	want := map[string]string{"THREAD": "login-fix", "RUNTIME": "claude", "TAB": "w1:t1", "PANE": "w1:t1:p1", "PWD": resolvedPath(t, f.home) + "\n"}
	for key, value := range want {
		if got := hooks.mustValue("post-thread-create", key); got != value {
			t.Fatalf("%s = %q, want %q", key, got, value)
		}
	}
	if _, ok := hooks.value("post-thread-create", "ARGS"); ok {
		t.Fatal("the hook received arguments")
	}
}

func TestPreHookGetsNoTabOrPane(t *testing.T) {
	f := newFixture(t)
	hooks := f.recordingHook("pre-thread-create", "0")
	f.useHerdr()
	f.env["CODEX_SESSION_ID"] = "codex-session-9"
	f.enterHerdr()

	f.mustRun("thread", "create", "login-fix")

	if hooks.mustValue("pre-thread-create", "THREAD") != "login-fix" || hooks.mustValue("pre-thread-create", "RUNTIME") != "codex" {
		t.Fatal("the hook did not receive the thread and runtime")
	}
	for _, key := range []string{"TAB", "PANE"} {
		if _, ok := hooks.value("pre-thread-create", key); ok {
			t.Fatalf("the hook received HOMEREPO_HOOK_%s", key)
		}
	}
}

func TestHooksListEachAttachedFolderOnItsOwnLine(t *testing.T) {
	f := newFixture(t)
	f.clone("frontend")
	pre := f.recordingHook("pre-thread-create", "0")
	post := f.recordingHook("post-thread-create", "0")
	fake := f.useHerdr()
	f.enterClaudeSession()

	f.mustRun("thread", "create", "login-fix", "--repo", "backend")

	want := filepath.Join(f.clonePath("frontend"), ".worktrees") + "\n" + filepath.Join(f.clonePath("backend"), ".worktrees")
	for _, recorded := range []string{pre.mustValue("pre-thread-create", "DIRS"), post.mustValue("post-thread-create", "DIRS")} {
		if recorded != want {
			t.Fatalf("HOMEREPO_HOOK_DIRS %q, want %q", recorded, want)
		}
	}
	args := strings.Join(fake.runtimeArgs(t), "\n")
	if !strings.Contains(args, "--add-dir\n"+filepath.Join(f.clonePath("frontend"), ".worktrees")+"\n--add-dir\n"+filepath.Join(f.clonePath("backend"), ".worktrees")) {
		t.Fatalf("the session got other folders: %q", fake.runtimeArgs(t))
	}
}

func TestHookOutputGoesToStderr(t *testing.T) {
	f := newFixture(t)
	f.writeHook("pre-thread-create", "echo pre-out\necho pre-err 1>&2\n", 0o755)
	f.writeHook("post-thread-create", "echo post-out\necho post-err 1>&2\n", 0o755)
	f.useHerdr()
	f.enterClaudeSession()

	got := f.mustRun("thread", "create", "login-fix")

	if got.stdout != "created thread login-fix in tab w1:t1\n" {
		t.Fatalf("stdout %q", got.stdout)
	}
	if got.stderr != "pre-out\npre-err\npost-out\npost-err\n" {
		t.Fatalf("stderr %q", got.stderr)
	}
}

func TestPostHookFailureKeepsTheThread(t *testing.T) {
	f := newFixture(t)
	f.recordingHook("post-thread-create", "4")
	fake := f.useHerdr()
	f.enterClaudeSession()

	got := f.run("thread", "create", "login-fix", "--repo", "frontend")

	if got.code == 0 || !strings.Contains(got.stderr, `thread "login-fix" was created in tab w1:t1, but hook post-thread-create exited with code 4`) {
		t.Fatalf("code %d, stderr:\n%s", got.code, got.stderr)
	}
	if !fake.openTabs["w1:t1"] || !exists(f.recordPath("login-fix")) || !exists(f.worktreePath("frontend", "login-fix")) {
		t.Fatal("the tab, record or worktree was removed")
	}
}

func TestUnrunnableHookCreatesNothing(t *testing.T) {
	f := newFixture(t)
	f.writeHook("pre-thread-create", "exit 0\n", 0o644)
	fake := f.useHerdr()
	f.enterClaudeSession()

	got := f.run("thread", "create", "login-fix", "--repo", "frontend")

	if got.code == 0 || !strings.Contains(got.stderr, "hook pre-thread-create could not run") {
		t.Fatalf("code %d, stderr:\n%s", got.code, got.stderr)
	}
	if len(fake.calls) != 0 || exists(f.recordPath("login-fix")) || exists(f.clonePath("frontend")) {
		t.Fatal("something was created")
	}
}

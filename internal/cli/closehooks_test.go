package cli_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/i-am-saebyeok/homerepo/internal/gittest"
)

func (f *fixture) createThreadForClose(fake *fakeHerdr) {
	f.t.Helper()
	f.enterClaudeSession()
	f.mustRun("thread", "create", "login-fix", "--repo", "frontend")
	fake.calls = nil
}

func TestThreadCloseWithoutHooksIsUnchanged(t *testing.T) {
	f := newFixture(t)
	fake := f.useHerdr()
	f.createThreadForClose(fake)

	got := f.mustRun("thread", "close", "login-fix")

	if len(fake.calls) != 1 || strings.Join(fake.calls[0], " ") != "tab close w1:t1" {
		t.Fatalf("herdr calls %q", fake.calls)
	}
	if got.stderr != "" || !strings.Contains(got.stdout, "closed thread login-fix") {
		t.Fatalf("stdout %q, stderr %q", got.stdout, got.stderr)
	}
}

func TestPreCloseHookRunsBeforeReleaseAndTabClose(t *testing.T) {
	f := newFixture(t)
	fake := f.useHerdr()
	f.createThreadForClose(fake)
	marks := t.TempDir()
	worktreeMark := filepath.Join(marks, "worktree-existed")
	recordMark := filepath.Join(marks, "record-existed")
	ranMark := filepath.Join(marks, "ran")
	f.writeHook("pre-thread-close", `[ -e '`+f.worktreePath("frontend", "login-fix")+`' ] && touch '`+worktreeMark+`'
[ -e '`+f.recordPath("login-fix")+`' ] && touch '`+recordMark+`'
touch '`+ranMark+`'
`, 0o755)
	hookRanBeforeTabClose := false
	fake.onTabClose = func() { hookRanBeforeTabClose = exists(ranMark) }

	f.mustRun("thread", "close", "login-fix")

	if !exists(worktreeMark) || !exists(recordMark) || !hookRanBeforeTabClose {
		t.Fatal("the hook did not run before the worktree, record and tab were removed")
	}
	if exists(f.worktreePath("frontend", "login-fix")) || exists(f.recordPath("login-fix")) || fake.openTabs["w1:t1"] {
		t.Fatal("the thread was not closed after the hook")
	}
}

func TestThreadCloseStopsWhenPreCloseHookFails(t *testing.T) {
	f := newFixture(t)
	fake := f.useHerdr()
	f.createThreadForClose(fake)
	f.recordingHook("pre-thread-close", "5")

	got := f.run("thread", "close", "login-fix")

	if got.code == 0 || !strings.Contains(got.stderr, `thread "login-fix" was not closed: hook pre-thread-close exited with code 5`) {
		t.Fatalf("code %d, stderr:\n%s", got.code, got.stderr)
	}
	if !exists(f.worktreePath("frontend", "login-fix")) || !exists(f.recordPath("login-fix")) || !fake.openTabs["w1:t1"] {
		t.Fatal("the worktree, record or tab was removed")
	}
	if len(fake.callsOf("tab close")) != 0 {
		t.Fatal("the tab was closed")
	}
}

func TestThreadCloseRejectedRequestsRunNoHook(t *testing.T) {
	cases := map[string]struct {
		prepare func(f *fixture)
		name    string
	}{
		"uncommitted change": {prepare: func(f *fixture) {
			gittest.WriteFile(f.t, filepath.Join(f.worktreePath("frontend", "login-fix"), "unsaved.txt"), "work\n")
		}, name: "login-fix"},
		"unknown thread": {prepare: func(f *fixture) {}, name: "nothing"},
	}
	for name, testCase := range cases {
		t.Run(name, func(t *testing.T) {
			f := newFixture(t)
			fake := f.useHerdr()
			f.createThreadForClose(fake)
			testCase.prepare(f)
			pre := f.recordingHook("pre-thread-close", "0")
			post := f.recordingHook("post-thread-close", "0")

			got := f.run("thread", "close", testCase.name)

			if got.code == 0 {
				t.Fatal("the close was accepted")
			}
			if pre.ran("pre-thread-close") || post.ran("post-thread-close") {
				t.Fatal("a hook ran for a thread that cannot be closed")
			}
		})
	}
}

func TestThreadCloseReportsFailedPostCloseHook(t *testing.T) {
	f := newFixture(t)
	fake := f.useHerdr()
	f.createThreadForClose(fake)
	f.recordingHook("post-thread-close", "6")

	got := f.run("thread", "close", "login-fix")

	if got.code == 0 || !strings.Contains(got.stderr, `thread "login-fix" was closed, but hook post-thread-close exited with code 6`) {
		t.Fatalf("code %d, stderr:\n%s", got.code, got.stderr)
	}
	if exists(f.worktreePath("frontend", "login-fix")) || exists(f.recordPath("login-fix")) || fake.openTabs["w1:t1"] {
		t.Fatal("the thread was not left closed")
	}
}

func TestThreadCloseStillClosesTabAfterHookEndsSession(t *testing.T) {
	f := newFixture(t)
	fake := f.useHerdr()
	f.createThreadForClose(fake)
	endedMark := filepath.Join(t.TempDir(), "session-ended")
	f.writeHook("pre-thread-close", `[ "$HOMEREPO_HOOK_PANE" = w1:t1:p1 ] && touch '`+endedMark+`'
exit 0
`, 0o755)
	sessionEndedBeforeTabClose := false
	fake.onTabClose = func() { sessionEndedBeforeTabClose = exists(endedMark) }

	got := f.mustRun("thread", "close", "login-fix")

	if !sessionEndedBeforeTabClose || len(fake.callsOf("tab close")) != 1 {
		t.Fatalf("tab close calls %q", fake.callsOf("tab close"))
	}
	if !strings.Contains(got.stdout, "closed thread login-fix") || exists(f.recordPath("login-fix")) {
		t.Fatalf("the thread was not closed: %q", got.stdout)
	}
}

func TestCloseHooksReceiveThreadTabAndPane(t *testing.T) {
	f := newFixture(t)
	f.clone("backend")
	fake := f.useHerdr()
	f.createThreadForClose(fake)
	pre := f.recordingHook("pre-thread-close", "0")
	post := f.recordingHook("post-thread-close", "0")

	f.mustRun("thread", "close", "login-fix")

	dirs := f.worktreesDir("frontend") + "\n" + f.worktreesDir("backend")
	common := map[string]string{"THREAD": "login-fix", "RUNTIME": "claude", "TAB": "w1:t1", "DIRS": dirs, "PWD": resolvedPath(t, f.home) + "\n"}
	for _, hook := range []struct {
		recorder *hookRecorder
		name     string
	}{{pre, "pre-thread-close"}, {post, "post-thread-close"}} {
		for key, value := range common {
			if got := hook.recorder.mustValue(hook.name, key); got != value {
				t.Fatalf("%s %s = %q, want %q", hook.name, key, got, value)
			}
		}
		if _, ok := hook.recorder.value(hook.name, "ARGS"); ok {
			t.Fatalf("%s received arguments", hook.name)
		}
	}
	if got := pre.mustValue("pre-thread-close", "PANE"); got != "w1:t1:p1" {
		t.Fatalf("pre-thread-close PANE = %q", got)
	}
	if _, ok := post.value("post-thread-close", "PANE"); ok {
		t.Fatal("post-thread-close received HOMEREPO_HOOK_PANE")
	}
}

func TestPreCloseHookGetsNoPaneWhenTabIsGone(t *testing.T) {
	f := newFixture(t)
	fake := f.useHerdr()
	f.createThreadForClose(fake)
	delete(fake.openTabs, "w1:t1")
	pre := f.recordingHook("pre-thread-close", "0")

	f.mustRun("thread", "close", "login-fix")

	if !pre.ran("pre-thread-close") || pre.mustValue("pre-thread-close", "TAB") != "w1:t1" {
		t.Fatal("the hook did not run with the tab ID")
	}
	if _, ok := pre.value("pre-thread-close", "PANE"); ok {
		t.Fatal("the hook received a pane for a closed tab")
	}
	if exists(f.recordPath("login-fix")) {
		t.Fatal("the thread was not closed")
	}
}

func TestPreCloseHookGetsNoPaneWhenTabHasSeveralPanes(t *testing.T) {
	f := newFixture(t)
	fake := f.useHerdr()
	f.createThreadForClose(fake)
	fake.splitTabs = map[string]bool{"w1:t1": true}
	pre := f.recordingHook("pre-thread-close", "0")

	f.mustRun("thread", "close", "login-fix")

	if _, ok := pre.value("pre-thread-close", "PANE"); ok || !pre.ran("pre-thread-close") {
		t.Fatal("the hook did not run without a pane")
	}
}

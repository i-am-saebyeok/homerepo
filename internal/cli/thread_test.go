package cli_test

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/i-am-saebyeok/homerepo/internal/command"
	"github.com/i-am-saebyeok/homerepo/internal/command/commandtest"
	"github.com/i-am-saebyeok/homerepo/internal/gittest"
	"github.com/i-am-saebyeok/homerepo/internal/herdr"
)

const callerPane = "w1:p0"

type fakeHerdr struct {
	openTabs    map[string]bool
	createdTabs int
	startCodes  []string
	calls       [][]string
	onTabCreate func()
}

func (f *fixture) useHerdr() *fakeHerdr {
	fake := &fakeHerdr{openTabs: map[string]bool{}}
	f.runner.Handlers["herdr"] = fake.handle
	return fake
}

func (f *fixture) enterClaudeSession() {
	f.env["CLAUDECODE"] = "1"
	f.env["CLAUDE_CODE_SESSION_ID"] = "session-1"
	f.enterHerdr()
}

func (f *fixture) enterHerdr() {
	f.env["HERDR_ENV"] = "1"
	f.env["HERDR_WORKSPACE_ID"] = "w1"
	f.env["HERDR_PANE_ID"] = callerPane
}

func herdrError(cmd command.Command, code string) error {
	return commandtest.Fail(cmd, `{"error":{"code":"`+code+`","message":"fake"},"id":"cli"}`)
}

func reply(json string) (command.Output, error) {
	return command.Output{Stdout: []byte(json)}, nil
}

func (h *fakeHerdr) handle(cmd command.Command) (command.Output, error) {
	h.calls = append(h.calls, cmd.Args)
	action := strings.Join(cmd.Args[:2], " ")
	switch action {
	case "tab create":
		if h.onTabCreate != nil {
			h.onTabCreate()
		}
		h.createdTabs++
		tab := fmt.Sprintf("w1:t%d", h.createdTabs)
		h.openTabs[tab] = true
		return reply(`{"id":"cli:tab:create","result":{"tab":{"tab_id":"` + tab + `"},"root_pane":{"pane_id":"` + tab + `:p1"}}}`)
	case "agent start":
		if len(h.startCodes) > 0 {
			code := h.startCodes[0]
			h.startCodes = h.startCodes[1:]
			return command.Output{}, herdrError(cmd, code)
		}
		return reply(`{"id":"cli:agent:start","result":{}}`)
	case "agent prompt":
		return reply(`{"id":"cli:agent:prompt","result":{}}`)
	case "tab get", "tab close":
		tab := cmd.Args[2]
		if !h.openTabs[tab] {
			return command.Output{}, herdrError(cmd, "tab_not_found")
		}
		if action == "tab close" {
			delete(h.openTabs, tab)
		}
		return reply(`{"result":{}}`)
	}
	return command.Output{}, fmt.Errorf("unexpected herdr call: %v", cmd.Args)
}

func (h *fakeHerdr) callsOf(action string) [][]string {
	var found [][]string
	for _, call := range h.calls {
		if strings.Join(call[:2], " ") == action {
			found = append(found, call)
		}
	}
	return found
}

func (h *fakeHerdr) agentArgs(t *testing.T) []string {
	t.Helper()
	starts := h.callsOf("agent start")
	if len(starts) == 0 {
		t.Fatal("no agent was started")
	}
	last := starts[len(starts)-1]
	return last[slices.Index(last, "--")+1:]
}

func (f *fixture) recordPath(name string) string {
	return filepath.Join(f.home, ".homerepo", "threads", filepath.FromSlash(name)+".json")
}

func TestThreadCreateFromClaudeSession(t *testing.T) {
	f := newFixture(t)
	f.clone("frontend")
	fake := f.useHerdr()
	f.enterClaudeSession()

	got := f.mustRun("thread", "create", "login-fix")

	frontendWorktrees := filepath.Join(f.clonePath("frontend"), ".worktrees")
	wantCalls := [][]string{
		{"tab", "create", "--workspace", "w1", "--cwd", f.home, "--label", "login-fix", "--env", "HOMEREPO_THREAD=login-fix", "--no-focus"},
		{"agent", "start", "login-fix", "--kind", "claude", "--pane", "w1:t1:p1", "--timeout", "60000", "--", "--resume", "session-1", "--fork-session", "--add-dir", frontendWorktrees},
	}
	for index, want := range wantCalls {
		if !slices.Equal(fake.calls[index], want) {
			t.Fatalf("call %d:\n got %q\nwant %q", index, fake.calls[index], want)
		}
	}
	if len(fake.callsOf("agent prompt")) != 1 {
		t.Fatalf("first message was not sent once: %q", fake.calls)
	}
	for _, call := range fake.calls {
		if slices.Contains(call, callerPane) || slices.Contains(call, "--focus") || call[1] == "focus" {
			t.Fatalf("the calling tab was touched: %q", call)
		}
	}
	if _, err := os.Stat(filepath.Join(frontendWorktrees, ".gitignore")); err != nil {
		t.Fatalf("attached folder was not created as an ignored folder: %v", err)
	}
	if status := gittest.Git(t, f.clonePath("frontend"), "status", "--porcelain", "--untracked-files=all"); status != "" {
		t.Fatalf("clone status is not clean:\n%s", status)
	}
	record, err := os.ReadFile(f.recordPath("login-fix"))
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{`"name": "login-fix"`, `"runtime": "claude"`, `"tab_id": "w1:t1"`, `"created_at"`} {
		if !strings.Contains(string(record), expected) {
			t.Fatalf("record misses %s:\n%s", expected, record)
		}
	}
	if !strings.Contains(got.stdout, "created thread login-fix in tab w1:t1") {
		t.Fatalf("stdout:\n%s", got.stdout)
	}
}

func TestThreadCreateFromCodexSession(t *testing.T) {
	f := newFixture(t)
	f.clone("frontend")
	fake := f.useHerdr()
	f.enterHerdr()

	f.mustRun("thread", "create", "login-fix", "--runtime", "codex", "--session", "codex-session-9")

	start := fake.callsOf("agent start")[0]
	if start[slices.Index(start, "--kind")+1] != "codex" {
		t.Fatalf("kind is not codex: %q", start)
	}
	want := []string{"fork", "codex-session-9", "--add-dir", filepath.Join(f.clonePath("frontend"), ".worktrees")}
	if got := fake.agentArgs(t); !slices.Equal(got, want) {
		t.Fatalf("runtime args %q, want %q", got, want)
	}
}

func TestThreadCreateCodexNeedsSession(t *testing.T) {
	f := newFixture(t)
	fake := f.useHerdr()
	f.enterHerdr()
	got := f.run("thread", "create", "login-fix", "--runtime", "codex")
	if got.code == 0 || !strings.Contains(got.stderr, "--session") {
		t.Fatalf("code %d, stderr:\n%s", got.code, got.stderr)
	}
	if len(fake.calls) != 0 {
		t.Fatalf("herdr was called: %q", fake.calls)
	}
}

func TestThreadCreateWithoutRuntimeOrHerdrCreatesNothing(t *testing.T) {
	cases := map[string]struct {
		env      map[string]string
		expected []string
	}{
		"outside any session": {env: map[string]string{}, expected: []string{"runtime not found", "Herdr not found"}},
		"outside Herdr":       {env: map[string]string{"CLAUDECODE": "1", "CLAUDE_CODE_SESSION_ID": "session-1"}, expected: []string{"Herdr not found"}},
		"no session ID":       {env: map[string]string{"CLAUDECODE": "1", "HERDR_ENV": "1", "HERDR_WORKSPACE_ID": "w1", "HERDR_PANE_ID": callerPane}, expected: []string{"CLAUDE_CODE_SESSION_ID"}},
		"no Herdr pane":       {env: map[string]string{"CLAUDECODE": "1", "CLAUDE_CODE_SESSION_ID": "session-1", "HERDR_ENV": "1", "HERDR_WORKSPACE_ID": "w1"}, expected: []string{"HERDR_PANE_ID"}},
	}
	for name, testCase := range cases {
		t.Run(name, func(t *testing.T) {
			f := newFixture(t)
			f.clone("frontend")
			fake := f.useHerdr()
			f.env = testCase.env

			got := f.run("thread", "create", "login-fix", "--repo", "frontend")

			if got.code == 0 {
				t.Fatal("create succeeded")
			}
			for _, expected := range testCase.expected {
				if !strings.Contains(got.stderr, expected) {
					t.Fatalf("stderr misses %q:\n%s", expected, got.stderr)
				}
			}
			if len(fake.calls) != 0 || exists(f.recordPath("login-fix")) || exists(f.worktreePath("frontend", "login-fix")) {
				t.Fatal("something was created")
			}
		})
	}
}

func TestThreadCreateWithExistingNameCreatesNothing(t *testing.T) {
	f := newFixture(t)
	fake := f.useHerdr()
	f.enterClaudeSession()
	f.mustRun("thread", "create", "login-fix")

	got := f.run("thread", "create", "login-fix", "--repo", "frontend")

	if got.code == 0 || !strings.Contains(got.stderr, "already exists") {
		t.Fatalf("code %d, stderr:\n%s", got.code, got.stderr)
	}
	if len(fake.callsOf("tab create")) != 1 || exists(f.clonePath("frontend")) {
		t.Fatal("something was created for the second thread")
	}
}

func TestThreadFirstMessageNamesThreadAndWorktreeCommand(t *testing.T) {
	f := newFixture(t)
	fake := f.useHerdr()
	f.enterClaudeSession()
	f.mustRun("thread", "create", "fix/login")

	prompts := fake.callsOf("agent prompt")
	if len(prompts) != 1 {
		t.Fatalf("prompts %q", prompts)
	}
	if prompts[0][2] != "fix-login" {
		t.Fatalf("prompt went to agent %q", prompts[0][2])
	}
	message := prompts[0][3]
	for _, expected := range []string{"thread", `"fix/login"`, "worktree", "homerepo workspace create"} {
		if !strings.Contains(message, expected) {
			t.Fatalf("first message misses %q: %s", expected, message)
		}
	}
	promptIndex := slices.IndexFunc(fake.calls, func(call []string) bool { return call[1] == "prompt" })
	startIndex := slices.IndexFunc(fake.calls, func(call []string) bool { return call[1] == "start" })
	if promptIndex < startIndex {
		t.Fatal("the first message was sent before the agent was ready")
	}
}

func TestThreadCreateWaitingForConfirmationSendsNoMessage(t *testing.T) {
	f := newFixture(t)
	fake := f.useHerdr()
	fake.startCodes = []string{"agent_not_ready"}
	f.enterClaudeSession()

	got := f.mustRun("thread", "create", "login-fix")

	if len(fake.callsOf("agent prompt")) != 0 {
		t.Fatal("a message was sent to a session waiting for confirmation")
	}
	for _, call := range fake.calls {
		if call[1] == "send-keys" {
			t.Fatalf("keys were sent: %q", call)
		}
	}
	if !strings.Contains(got.stdout, "waiting for a confirmation; answer it in that tab") {
		t.Fatalf("stdout:\n%s", got.stdout)
	}
}

func TestThreadCreateRetriesBusyPane(t *testing.T) {
	f := newFixture(t)
	fake := f.useHerdr()
	fake.startCodes = []string{"agent_pane_busy", "agent_pane_busy"}
	f.enterClaudeSession()

	f.mustRun("thread", "create", "login-fix")

	if len(fake.callsOf("agent start")) != 3 || len(fake.callsOf("agent prompt")) != 1 {
		t.Fatalf("calls %q", fake.calls)
	}
}

func TestThreadCreateGivesUpOnBusyPane(t *testing.T) {
	f := newFixture(t)
	fake := f.useHerdr()
	for range herdr.BusyAttempts {
		fake.startCodes = append(fake.startCodes, "agent_pane_busy")
	}
	f.enterClaudeSession()

	got := f.run("thread", "create", "login-fix")

	if got.code == 0 || !strings.Contains(got.stderr, "agent_pane_busy") || !strings.Contains(got.stderr, "homerepo thread close login-fix") {
		t.Fatalf("code %d, stderr:\n%s", got.code, got.stderr)
	}
	if len(fake.callsOf("agent start")) != herdr.BusyAttempts {
		t.Fatalf("started %d times", len(fake.callsOf("agent start")))
	}
}

func TestThreadCreateUsesArgumentBundles(t *testing.T) {
	f := newFixture(t)
	f.writeLocalConfig("runtimes:\n  claude:\n    thread:\n      default: [--model, opus]\n      yolo: [--dangerously-skip-permissions]\n")
	fake := f.useHerdr()
	f.enterClaudeSession()

	f.mustRun("thread", "create", "with-default")
	if args := fake.agentArgs(t); !slices.Equal(args[len(args)-2:], []string{"--model", "opus"}) {
		t.Fatalf("default bundle missing: %q", args)
	}

	f.mustRun("thread", "create", "with-yolo", "--mode", "yolo")
	args := fake.agentArgs(t)
	if args[len(args)-1] != "--dangerously-skip-permissions" || slices.Contains(args, "opus") {
		t.Fatalf("yolo bundle not used alone: %q", args)
	}
}

func TestThreadCreateWithUnknownBundleCreatesNothing(t *testing.T) {
	f := newFixture(t)
	f.writeLocalConfig("runtimes:\n  claude:\n    thread:\n      default: [--model, opus]\n      yolo: [--dangerously-skip-permissions]\n")
	fake := f.useHerdr()
	f.enterClaudeSession()

	got := f.run("thread", "create", "login-fix", "--mode", "careful", "--repo", "frontend")

	if got.code == 0 || !strings.Contains(got.stderr, "default, yolo") {
		t.Fatalf("code %d, stderr:\n%s", got.code, got.stderr)
	}
	if len(fake.calls) != 0 || exists(f.clonePath("frontend")) {
		t.Fatal("something was created")
	}
}

func TestThreadCreateWithRepositoriesMakesWorktreesFirst(t *testing.T) {
	f := newFixture(t)
	f.clone("frontend")
	fake := f.useHerdr()
	worktreesBeforeTab := map[string]bool{}
	fake.onTabCreate = func() {
		for _, repository := range []string{"frontend", "backend"} {
			worktreesBeforeTab[repository] = exists(f.worktreePath(repository, "login-fix"))
		}
	}
	f.enterClaudeSession()

	got := f.mustRun("thread", "create", "login-fix", "--repo", "frontend", "--repo", "backend")

	if !worktreesBeforeTab["frontend"] || !worktreesBeforeTab["backend"] {
		t.Fatalf("worktrees before the tab: %v", worktreesBeforeTab)
	}
	if branch := strings.TrimSpace(gittest.Git(t, f.worktreePath("backend", "login-fix"), "branch", "--show-current")); branch != "login-fix" {
		t.Fatalf("branch %q", branch)
	}
	args := fake.agentArgs(t)
	for _, repository := range []string{"frontend", "backend"} {
		if !slices.Contains(args, filepath.Join(f.clonePath(repository), ".worktrees")) {
			t.Fatalf("%s worktrees folder not attached: %q", repository, args)
		}
	}
	if !strings.Contains(got.stdout, "worktree "+f.worktreePath("backend", "login-fix")) {
		t.Fatalf("stdout:\n%s", got.stdout)
	}
}

func TestThreadCreateWithUndeclaredRepositoryCreatesNothing(t *testing.T) {
	f := newFixture(t)
	fake := f.useHerdr()
	f.enterClaudeSession()
	got := f.run("thread", "create", "login-fix", "--repo", "mobile")
	if got.code == 0 || len(fake.calls) != 0 {
		t.Fatalf("code %d, calls %q", got.code, fake.calls)
	}
}

func TestThreadListShowsRuntimeTabAndWorktrees(t *testing.T) {
	f := newFixture(t)
	fake := f.useHerdr()
	f.enterClaudeSession()
	f.mustRun("thread", "create", "login-fix", "--repo", "frontend")
	f.mustRun("thread", "create", "other")

	got := f.mustRun("thread", "list")

	lines := strings.Split(strings.TrimSpace(got.stdout), "\n")
	if len(lines) != 3 {
		t.Fatalf("stdout:\n%s", got.stdout)
	}
	for _, expected := range []string{"login-fix", "claude", "w1:t1", "frontend:" + f.worktreePath("frontend", "login-fix")} {
		if !strings.Contains(lines[1], expected) {
			t.Fatalf("row misses %q: %s", expected, lines[1])
		}
	}

	delete(fake.openTabs, "w1:t2")
	got = f.mustRun("thread", "list")
	if !strings.Contains(strings.Split(got.stdout, "\n")[2], "no tab") {
		t.Fatalf("closed tab is not shown as missing:\n%s", got.stdout)
	}
}

func TestThreadCloseCleanThread(t *testing.T) {
	f := newFixture(t)
	fake := f.useHerdr()
	f.enterClaudeSession()
	f.mustRun("thread", "create", "login-fix", "--repo", "frontend", "--repo", "backend")

	got := f.mustRun("thread", "close", "login-fix")

	for _, repository := range []string{"frontend", "backend"} {
		if exists(f.worktreePath(repository, "login-fix")) {
			t.Fatalf("%s worktree still exists", repository)
		}
		gittest.Git(t, f.clonePath(repository), "rev-parse", "--verify", "refs/heads/login-fix")
	}
	if closes := fake.callsOf("tab close"); len(closes) != 1 || closes[0][2] != "w1:t1" {
		t.Fatalf("tab close calls %q", closes)
	}
	if exists(f.recordPath("login-fix")) {
		t.Fatal("thread record still exists")
	}
	if list := f.mustRun("thread", "list"); strings.Contains(list.stdout, "login-fix") {
		t.Fatalf("closed thread is still listed:\n%s", list.stdout)
	}
	if !strings.Contains(got.stdout, "closed thread login-fix") {
		t.Fatalf("stdout:\n%s", got.stdout)
	}
}

func TestThreadCloseWithChangesKeepsEverything(t *testing.T) {
	f := newFixture(t)
	fake := f.useHerdr()
	f.enterClaudeSession()
	f.mustRun("thread", "create", "login-fix", "--repo", "frontend", "--repo", "backend")
	dirty := f.worktreePath("backend", "login-fix")
	gittest.WriteFile(t, filepath.Join(dirty, "unsaved.txt"), "work\n")

	got := f.run("thread", "close", "login-fix")

	if got.code == 0 || !strings.Contains(got.stderr, dirty) || !strings.Contains(got.stderr, "unsaved.txt") {
		t.Fatalf("code %d, stderr:\n%s", got.code, got.stderr)
	}
	if strings.Contains(got.stderr, f.worktreePath("frontend", "login-fix")) {
		t.Fatalf("the clean worktree was reported as changed:\n%s", got.stderr)
	}
	for _, repository := range []string{"frontend", "backend"} {
		if !exists(f.worktreePath(repository, "login-fix")) {
			t.Fatalf("%s worktree was removed", repository)
		}
	}
	if len(fake.callsOf("tab close")) != 0 || !exists(f.recordPath("login-fix")) {
		t.Fatal("the thread was closed")
	}
}

func TestThreadCloseUnknownThread(t *testing.T) {
	f := newFixture(t)
	f.useHerdr()
	if got := f.run("thread", "close", "nothing"); got.code == 0 || !strings.Contains(got.stderr, "no thread named") {
		t.Fatalf("code %d, stderr:\n%s", got.code, got.stderr)
	}
}

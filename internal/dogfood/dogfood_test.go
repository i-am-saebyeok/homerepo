package dogfood_test

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/i-am-saebyeok/homerepo/internal/command"
	"github.com/i-am-saebyeok/homerepo/internal/dogfood"
	"github.com/i-am-saebyeok/homerepo/internal/gittest"
	"github.com/i-am-saebyeok/homerepo/internal/herdr"
	"github.com/i-am-saebyeok/homerepo/internal/home"
)

func TestScenarioSpeaksLikeAUser(t *testing.T) {
	scenario, err := dogfood.LookupScenario(scenarioName)
	if err != nil {
		t.Fatal(err)
	}
	layout := dogfood.Layout{Root: filepath.Join(t.TempDir(), "root")}
	text, err := scenario.Instructions(layout)
	if err != nil {
		t.Fatal(err)
	}
	mustContain(t, text, scenario.Repository, layout.Remote(scenario.Repository), "thread", "worktree", dogfood.NotesFile)
	for _, answer := range []string{"homerepo", "homerepo.yaml", "repositories:", "remote:", "branch:", "--repo", "thread create", "workspace create", "token", "gh ", "ssh"} {
		if strings.Contains(strings.ToLower(text), answer) {
			t.Fatalf("scenario gives away %q:\n%s", answer, text)
		}
	}
	if strings.Contains(text, "{{") {
		t.Fatalf("unfilled placeholder:\n%s", text)
	}
}

func TestFirstRunSetsUpHomeAndSendsScenario(t *testing.T) {
	w := newWorld(t)
	got := w.runScenario()
	if got.code != 0 {
		t.Fatalf("exit %d\nstdout:\n%s\nstderr:\n%s", got.code, got.stdout, got.stderr)
	}
	mustContain(t, got.stdout, "PASS  thread started", "PASS  repository declared", "PASS  worktree", "PASS  remote", "home: "+w.layout.Home(), "coordinator: tab w1:t1, pane w1:p1", "thread: "+threadName+", tab w1:t2, pane w1:p2", dogfood.NotesFile+" (present)")
	for _, path := range []string{".git", "homerepo.yaml", ".agents/skills/homerepo-thread/SKILL.md", ".homerepo/config.yaml"} {
		if !exists(filepath.Join(w.layout.Home(), path)) {
			t.Fatalf("Home has no %s", path)
		}
	}
	if len(w.herdr.paneRuns) != 1 || w.herdr.paneRuns[0][0] != w.herdr.coordinatorPane || w.herdr.paneRuns[0][1] != "claude" {
		t.Fatalf("pane runs %v", w.herdr.paneRuns)
	}
	scenario, _ := dogfood.LookupScenario(scenarioName)
	instructions, _ := scenario.Instructions(w.layout)
	coordinatorPrompts := w.herdr.promptsTo(w.herdr.coordinatorPane)
	if len(coordinatorPrompts) != 1 || coordinatorPrompts[0].text != instructions {
		t.Fatalf("coordinator prompts %v", coordinatorPrompts)
	}
	if threadPrompts := w.herdr.promptsTo(w.herdr.threadPane); len(threadPrompts) != 1 {
		t.Fatalf("thread prompts %v", threadPrompts)
	}
}

func TestSessionsUseTheBuiltHomerepo(t *testing.T) {
	w := newWorld(t)
	w.runScenario()
	if !exists(w.layout.Homerepo()) {
		t.Fatal("homerepo was not built into the root")
	}
	built := slices.IndexFunc(w.runner.Handled, func(cmd command.Command) bool { return cmd.Name == "go" })
	if built < 0 || !slices.Contains(w.runner.Handled[built].Args, w.layout.Homerepo()) {
		t.Fatalf("handled %v", w.runner.Handled)
	}
	var settings struct {
		Permissions struct {
			AdditionalDirectories []string `json:"additionalDirectories"`
		} `json:"permissions"`
		Env map[string]string `json:"env"`
	}
	if err := json.Unmarshal([]byte(w.readFile(filepath.Join(w.layout.Home(), ".claude", "settings.local.json"))), &settings); err != nil {
		t.Fatal(err)
	}
	if settings.Env["PATH"] != w.layout.Bin()+string(os.PathListSeparator)+searchPath {
		t.Fatalf("PATH %q", settings.Env["PATH"])
	}
	if !slices.Equal(settings.Permissions.AdditionalDirectories, []string{w.root}) {
		t.Fatalf("additional directories %v", settings.Permissions.AdditionalDirectories)
	}
}

func TestHomeTurnsOffAttachedFolders(t *testing.T) {
	w := newWorld(t)
	w.runScenario("--runtime-arg", "--model", "--runtime-arg", "some-model")
	loaded, err := home.Load(w.layout.Home())
	if err != nil {
		t.Fatal(err)
	}
	claude := loaded.Local.Runtimes["claude"]
	if claude.AttachesWorktrees() {
		t.Fatal("attach is on")
	}
	if !slices.Equal(claude.Thread["default"], []string{"--model", "some-model"}) {
		t.Fatalf("thread bundle %v", claude.Thread)
	}
	if !slices.Equal(w.herdr.paneRuns[0][1:], []string{"claude", "--model", "some-model"}) {
		t.Fatalf("pane run %v", w.herdr.paneRuns[0])
	}
}

func TestThreadForkGetsNoWorktreeFolders(t *testing.T) {
	w := newWorld(t)
	w.coordinate = func(w *world) {
		w.createRepository()
		w.createThreadWithHomerepo()
	}
	w.work = func(w *world) error { return nil }
	w.runScenario()
	if len(w.herdr.paneRuns) != 2 {
		t.Fatalf("pane runs %v", w.herdr.paneRuns)
	}
	fork := w.herdr.paneRuns[1]
	if !slices.Equal(fork[1:], []string{"claude", "--resume", "session-1", "--fork-session"}) {
		t.Fatalf("fork %v", fork)
	}
	if len(w.herdr.promptsTo(w.herdr.threadPane)) != 1 {
		t.Fatal("the thread made by homerepo was not started")
	}
}

func TestCodexOpensTheRootAndPutsBuiltHomerepoOnPath(t *testing.T) {
	w := newWorld(t)
	w.run("--root", w.root, "--runtime", "codex", "--runtime-arg", "--model", "--runtime-arg", "some-model", scenarioName)
	loaded, err := home.Load(w.layout.Home())
	if err != nil {
		t.Fatal(err)
	}
	codex := loaded.Local.Runtimes["codex"]
	path, _ := json.Marshal(w.layout.Bin() + string(os.PathListSeparator) + searchPath)
	expected := []string{"--add-dir", w.root, "-c", "shell_environment_policy.set.PATH=" + string(path), "--model", "some-model"}
	if codex.AttachesWorktrees() || !slices.Equal(codex.Thread["default"], expected) {
		t.Fatalf("codex config %+v", codex)
	}
	if !slices.Equal(w.herdr.paneRuns[0][1:], append([]string{"codex"}, expected...)) {
		t.Fatalf("pane run %v", w.herdr.paneRuns[0])
	}
	if exists(filepath.Join(w.layout.Home(), ".claude", "settings.local.json")) {
		t.Fatal("codex run wrote Claude settings")
	}
}

func TestRemoteIsAnEmptyBareRepositoryUnderRoot(t *testing.T) {
	w := newWorld(t)
	var branchesAtStart string
	w.coordinate = func(w *world) {
		branchesAtStart = gittest.Git(t, w.layout.Remote(repository), "branch", "--list")
		w.createRepositoryAndThread()
	}
	w.runScenario()
	remote := w.layout.Remote(repository)
	if !strings.HasPrefix(remote, w.root+string(filepath.Separator)) {
		t.Fatalf("remote %s is outside the root", remote)
	}
	if strings.TrimSpace(gittest.Git(t, remote, "rev-parse", "--is-bare-repository")) != "true" {
		t.Fatal("remote is not bare")
	}
	if branchesAtStart != "" {
		t.Fatalf("remote had branches before the scenario: %q", branchesAtStart)
	}
}

func TestRunAgainClosesPreviousTabsAndEmptiesRoot(t *testing.T) {
	w := newWorld(t)
	w.runScenario()
	leftover := filepath.Join(w.root, "stray")
	gittest.WriteFile(t, leftover, "x\n")
	firstTabs := []string{"w1:t1", "w1:t2"}
	got := w.runScenario()
	if got.code != 0 {
		t.Fatalf("exit %d\n%s\n%s", got.code, got.stdout, got.stderr)
	}
	for _, tab := range firstTabs {
		if !slices.Contains(w.herdr.closedTabs, tab) {
			t.Fatalf("tab %s was not closed: %v", tab, w.herdr.closedTabs)
		}
	}
	if exists(leftover) {
		t.Fatal("root was not emptied")
	}
	mustContain(t, got.stdout, "home: "+w.layout.Home(), "coordinator: tab w1:t3")
}

func TestRefusesARootItDidNotMake(t *testing.T) {
	w := newWorld(t)
	other := filepath.Join(w.root, "someone-elses.txt")
	gittest.WriteFile(t, other, "keep\n")
	got := w.runScenario()
	if got.code == 0 || !exists(other) || exists(w.layout.Home()) {
		t.Fatalf("exit %d, stderr %s", got.code, got.stderr)
	}
	mustContain(t, got.stderr, "not empty")
}

func TestMissingRequirementsCreateNothing(t *testing.T) {
	cases := map[string]struct {
		prepare func(w *world)
		args    func(w *world) []string
		message string
	}{
		"outside Herdr": {
			prepare: func(w *world) { delete(w.env, "HERDR_ENV") },
			message: "HERDR_ENV",
		},
		"no runtime program": {
			prepare: func(w *world) { w.missing = []string{"claude"} },
			message: "claude not found",
		},
		"unknown scenario": {
			args:    func(w *world) []string { return []string{"--root", w.root, "--runtime", "claude", "no-such-scenario"} },
			message: "unknown scenario",
		},
		"unknown runtime": {
			args:    func(w *world) []string { return []string{"--root", w.root, "--runtime", "other", scenarioName} },
			message: "unknown runtime",
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			w := newWorld(t)
			if tc.prepare != nil {
				tc.prepare(w)
			}
			args := []string{"--root", w.root, "--runtime", "claude", scenarioName}
			if tc.args != nil {
				args = tc.args(w)
			}
			got := w.run(args...)
			if got.code == 0 {
				t.Fatal("expected failure")
			}
			mustContain(t, got.stderr, tc.message)
			if exists(w.root) || len(w.herdr.paneRuns) > 0 || w.herdr.created > 0 {
				t.Fatal("something was created")
			}
		})
	}
}

func TestCoordinatorIsSentAgainUntilShellIsReady(t *testing.T) {
	w := newWorld(t)
	w.herdr.busyRuns = 2
	if got := w.runScenario(); got.code != 0 {
		t.Fatalf("exit %d\n%s\n%s", got.code, got.stdout, got.stderr)
	}
}

func TestCoordinatorThatNeverStartsIsReported(t *testing.T) {
	w := newWorld(t)
	w.herdr.busyRuns = 1000
	got := w.runScenario()
	if got.code == 0 {
		t.Fatal("expected failure")
	}
	mustContain(t, got.stderr, "could not start the coordinator", "pane_send_failed")
}

func TestNoThreadPointsAtCoordinator(t *testing.T) {
	w := newWorld(t)
	w.coordinate = func(w *world) { w.createRepository() }
	got := w.runScenario()
	if got.code == 0 {
		t.Fatal("expected failure")
	}
	mustContain(t, got.stdout, "FAIL  thread started: no thread was created", "no thread was created; see the coordinator in pane w1:p1", "PASS  repository declared")
}

func TestNoThreadBeforeTimeLimit(t *testing.T) {
	w := newWorld(t)
	w.coordinate = func(w *world) {
		w.herdr.statuses[w.herdr.coordinatorPane] = []herdr.AgentStatus{herdr.AgentWorking}
	}
	got := w.runScenario()
	if got.code == 0 {
		t.Fatal("expected failure")
	}
	mustContain(t, got.stdout, "time limit reached: no thread was created", "timed out after 10m0s", "coordinator: tab w1:t1, pane w1:p1")
}

func TestFailedCheckIsReported(t *testing.T) {
	w := newWorld(t)
	w.work = func(w *world) error {
		w.commitInWorktree()
		return nil
	}
	got := w.runScenario()
	if got.code == 0 {
		t.Fatal("expected failure")
	}
	mustContain(t, got.stdout, "PASS  thread started", "PASS  repository declared", "PASS  worktree", "FAIL  remote: no branch "+threadName, "home: "+w.layout.Home(), "pane w1:p2")
}

func TestRemoteBranchWithoutNewCommitFails(t *testing.T) {
	w := newWorld(t)
	w.work = func(w *world) error {
		gittest.Git(t, w.clone(), "worktree", "add", "--quiet", "-b", threadName, w.worktree(), "main")
		w.push(threadName)
		return nil
	}
	got := w.runScenario()
	mustContain(t, got.stdout, "FAIL  remote: branch "+threadName)
}

func TestMissingWorktreeAndDeclarationFail(t *testing.T) {
	w := newWorld(t)
	w.coordinate = func(w *world) { w.createThread() }
	w.work = func(w *world) error { return nil }
	got := w.runScenario()
	if got.code == 0 {
		t.Fatal("expected failure")
	}
	mustContain(t, got.stdout, "PASS  thread started", "FAIL  repository declared", "FAIL  worktree", "FAIL  remote")
}

func TestThreadWaitingForConfirmationGetsNoInput(t *testing.T) {
	w := newWorld(t)
	w.coordinate = func(w *world) {
		w.createRepositoryAndThread()
		w.herdr.statuses[w.herdr.threadPane] = []herdr.AgentStatus{herdr.AgentAbsent, herdr.AgentBlocked}
		w.herdr.screens[w.herdr.threadPane] = "Allow running this command?\n"
	}
	got := w.runScenario()
	if got.code == 0 {
		t.Fatal("expected failure")
	}
	mustContain(t, got.stdout, "FAIL  thread started", "is waiting for a confirmation; no keys were sent", "  | Allow running this command?")
	if strings.Contains(got.stdout, "trust question") {
		t.Fatalf("thread note guesses at the prompt:\n%s", got.stdout)
	}
	if len(w.herdr.promptsTo(w.herdr.threadPane)) > 0 || w.herdr.keysSent {
		t.Fatal("input was sent to the blocked thread")
	}
}

func TestCoordinatorWaitingForTrustStops(t *testing.T) {
	w := newWorld(t)
	w.herdr.handleTabCreate = func(pane string) {
		w.herdr.statuses[pane] = []herdr.AgentStatus{herdr.AgentAbsent, herdr.AgentBlocked}
		w.herdr.screens[pane] = "Do you trust the files in this folder?\n> Yes\n  No\n\n"
	}
	got := w.runScenario()
	if got.code == 0 {
		t.Fatal("expected failure")
	}
	mustContain(t, got.stdout, "answer the trust question", w.layout.Home(), "  | Do you trust the files in this folder?\n  | > Yes\n  |   No\n")
	if len(w.herdr.prompts) > 0 || w.herdr.keysSent {
		t.Fatal("input was sent to the blocked coordinator")
	}
}

func TestThreadPastTimeLimit(t *testing.T) {
	w := newWorld(t)
	w.work = func(w *world) error { return errors.New("timeout") }
	got := w.runScenario()
	if got.code == 0 {
		t.Fatal("expected failure")
	}
	mustContain(t, got.stdout, `time limit reached: thread "`+threadName, "timed out after 10m0s", "PASS  thread started", "PASS  repository declared", "FAIL  worktree")
	threadPrompt := w.herdr.promptsTo(w.herdr.threadPane)[0]
	mustContain(t, strings.Join(threadPrompt.args, " "), "--wait", "--until done", "--until idle", "--until blocked", "--timeout")
}

const startingStatus = herdr.AgentStatus("unknown")

func TestCoordinatorBlockedAfterScenarioShowsScreenWithoutTrustHint(t *testing.T) {
	w := newWorld(t)
	w.coordinate = func(w *world) {
		w.herdr.statuses[w.herdr.coordinatorPane] = []herdr.AgentStatus{herdr.AgentBlocked}
		w.herdr.screens[w.herdr.coordinatorPane] = "Run homerepo thread create?\n"
	}
	got := w.runScenario()
	mustContain(t, got.stdout, "no thread was created; see the coordinator in pane w1:p1: waiting for a confirmation", "  | Run homerepo thread create?")
	if strings.Contains(got.stdout, "trust question") {
		t.Fatalf("note guesses at the prompt:\n%s", got.stdout)
	}
}

func TestCoordinatorThatExitsRightAwayFailsAtOnce(t *testing.T) {
	w := newWorld(t)
	w.herdr.handleTabCreate = func(pane string) {
		w.herdr.statuses[pane] = []herdr.AgentStatus{herdr.AgentAbsent, startingStatus, herdr.AgentAbsent}
		w.herdr.screens[pane] = "Error adding directories\n$ \n"
	}
	started := w.now
	got := w.runScenario()
	if got.code == 0 {
		t.Fatal("expected failure")
	}
	mustContain(t, got.stdout, "the coordinator in pane w1:p1 did not start: the runtime exited and the pane is back at its shell", "  | Error adding directories")
	if strings.Contains(got.stdout, "timed out") || w.now.Sub(started) > time.Minute || len(w.herdr.prompts) > 0 {
		t.Fatalf("waited %s, prompts %v\n%s", w.now.Sub(started), w.herdr.prompts, got.stdout)
	}
}

func TestRuntimeThatNeverAppearsFailsBeforeTimeLimit(t *testing.T) {
	w := newWorld(t)
	w.herdr.handleTabCreate = func(pane string) {
		w.herdr.statuses[pane] = []herdr.AgentStatus{herdr.AgentAbsent}
		w.herdr.screens[pane] = "$ \n"
	}
	got := w.runScenario()
	mustContain(t, got.stdout, "did not start: no runtime appeared in the pane within 1m0s", "  | $")
	if strings.Contains(got.stdout, "timed out") {
		t.Fatalf("waited for the time limit:\n%s", got.stdout)
	}
}

func TestThreadThatExitsRightAwayFailsAtOnce(t *testing.T) {
	w := newWorld(t)
	w.coordinate = func(w *world) {
		w.createRepositoryAndThread()
		w.herdr.statuses[w.herdr.threadPane] = []herdr.AgentStatus{startingStatus, herdr.AgentAbsent}
		w.herdr.screens[w.herdr.threadPane] = "session not found\n"
	}
	got := w.runScenario()
	mustContain(t, got.stdout, `thread "first-change" in pane w1:p2 did not start: the runtime exited`, "  | session not found", "FAIL  thread started")
	if strings.Contains(got.stdout, "timed out") || len(w.herdr.promptsTo(w.herdr.threadPane)) > 0 {
		t.Fatalf("unexpected:\n%s", got.stdout)
	}
}

package thread

import (
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/i-am-saebyeok/homerepo/internal/home"
)

var herdrAgentName = regexp.MustCompile(`^[a-z][a-z0-9_-]{0,31}$`)

func TestAgentNameFitsHerdr(t *testing.T) {
	cases := map[string]string{
		"login-fix":             "login-fix",
		"fix/Login.Page":        "fix-login-page",
		"2026-cleanup":          "thread-2026-cleanup",
		strings.Repeat("a", 40): strings.Repeat("a", 32),
	}
	for thread, want := range cases {
		got := AgentName(thread)
		if got != want || !herdrAgentName.MatchString(got) {
			t.Errorf("AgentName(%q) = %q, want %q", thread, got, want)
		}
	}
}

func environment(values map[string]string) func(string) string {
	return func(key string) string { return values[key] }
}

func TestDetectCallerFromClaudeEnvironment(t *testing.T) {
	caller, err := DetectCaller(environment(map[string]string{
		"CLAUDECODE": "1", "CLAUDE_CODE_SESSION_ID": "s1",
		"HERDR_ENV": "1", "HERDR_WORKSPACE_ID": "w1", "HERDR_PANE_ID": "p1",
	}), CallerOverrides{})
	if err != nil {
		t.Fatal(err)
	}
	if caller != (Caller{Runtime: Claude, Session: "s1", Workspace: "w1", Pane: "p1"}) {
		t.Fatalf("caller %+v", caller)
	}
}

func TestDetectCallerOverridesWin(t *testing.T) {
	caller, err := DetectCaller(environment(map[string]string{
		"CLAUDECODE": "1", "CLAUDE_CODE_SESSION_ID": "s1",
		"HERDR_ENV": "1", "HERDR_WORKSPACE_ID": "w1", "HERDR_PANE_ID": "p1",
	}), CallerOverrides{Runtime: "codex", Session: "c9"})
	if err != nil {
		t.Fatal(err)
	}
	if caller.Runtime != Codex || caller.Session != "c9" {
		t.Fatalf("caller %+v", caller)
	}
}

func TestDetectCallerReportsEverythingMissing(t *testing.T) {
	_, err := DetectCaller(environment(map[string]string{}), CallerOverrides{})
	if err == nil || !strings.Contains(err.Error(), "runtime not found") || !strings.Contains(err.Error(), "Herdr not found") {
		t.Fatalf("error %v", err)
	}
	_, err = DetectCaller(environment(map[string]string{"HERDR_ENV": "1", "HERDR_WORKSPACE_ID": "w1", "HERDR_PANE_ID": "p1"}), CallerOverrides{Runtime: "other"})
	if err == nil || !strings.Contains(err.Error(), "unknown runtime") {
		t.Fatalf("error %v", err)
	}
}

func TestLaunchArgsAttachEachFolderBeforeBundle(t *testing.T) {
	claude, err := adapterFor(Claude)
	if err != nil {
		t.Fatal(err)
	}
	got := launchArgs(claude, "s1", []string{"/a/.worktrees", "/b/.worktrees"}, []string{"--model", "opus"})
	want := []string{"--resume", "s1", "--fork-session", "--add-dir", "/a/.worktrees", "--add-dir", "/b/.worktrees", "--model", "opus"}
	if !slices.Equal(got, want) {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestArgumentBundleWithoutConfigAddsNothing(t *testing.T) {
	args, err := argumentBundle(home.LocalConfig{}, Claude, "")
	if err != nil || len(args) != 0 {
		t.Fatalf("args %q, err %v", args, err)
	}
	if _, err := argumentBundle(home.LocalConfig{}, Claude, "default"); err == nil || !strings.Contains(err.Error(), "available: none") {
		t.Fatalf("err %v", err)
	}
}

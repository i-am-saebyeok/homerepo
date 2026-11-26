package cli_test

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func (f *fixture) worktreesDir(repository string) string {
	return filepath.Join(f.clonePath(repository), ".worktrees")
}

func TestThreadCreateAttachesWorktreesWhenAttachIsUnsetOrTrue(t *testing.T) {
	cases := map[string]string{
		"no config":         "",
		"no attach":         "runtimes:\n  claude:\n    thread:\n      default: [--model, opus]\n",
		"attach true":       "runtimes:\n  claude:\n    attach: true\n",
		"attach left empty": "runtimes:\n  claude:\n    attach:\n",
	}
	for name, config := range cases {
		t.Run(name, func(t *testing.T) {
			f := newFixture(t)
			f.clone("frontend")
			if config != "" {
				f.writeLocalConfig(config)
			}
			fake := f.useHerdr()
			f.enterClaudeSession()

			f.mustRun("thread", "create", "login-fix")

			args := fake.runtimeArgs(t)
			if !slices.Equal(args[:5], []string{"--resume", "session-1", "--fork-session", "--add-dir", f.worktreesDir("frontend")}) {
				t.Fatalf("runtime args %q", args)
			}
		})
	}
}

func TestThreadCreateWithAttachFalseAddsNoFolders(t *testing.T) {
	f := newFixture(t)
	f.clone("frontend")
	f.writeLocalConfig("runtimes:\n  claude:\n    attach: false\n")
	fake := f.useHerdr()
	f.enterClaudeSession()

	f.mustRun("thread", "create", "login-fix", "--repo", "backend")

	if got := fake.runtimeCommand(t); !slices.Equal(got, []string{"claude", "--resume", "session-1", "--fork-session"}) {
		t.Fatalf("runtime command %q", got)
	}
	for _, repository := range []string{"frontend", "backend"} {
		if _, err := os.Stat(filepath.Join(f.worktreesDir(repository), ".gitignore")); err != nil {
			t.Fatalf("%s worktrees folder was not prepared: %v", repository, err)
		}
	}
	if !exists(f.worktreePath("backend", "login-fix")) {
		t.Fatal("the requested worktree was not created")
	}
}

func TestThreadCreateReadsAttachPerRuntime(t *testing.T) {
	f := newFixture(t)
	f.clone("frontend")
	f.writeLocalConfig("runtimes:\n  claude:\n    attach: false\n  codex:\n    thread:\n      default: [--model, gpt]\n")
	fake := f.useHerdr()
	f.env["CODEX_SESSION_ID"] = "codex-session-9"
	f.enterHerdr()

	f.mustRun("thread", "create", "login-fix")

	want := []string{"codex", "fork", "codex-session-9", "--add-dir", f.worktreesDir("frontend"), "--model", "gpt"}
	if got := fake.runtimeCommand(t); !slices.Equal(got, want) {
		t.Fatalf("runtime command %q, want %q", got, want)
	}
}

func TestThreadCreateWithAttachFalseKeepsFoldersFromBundle(t *testing.T) {
	f := newFixture(t)
	f.clone("frontend")
	f.writeLocalConfig("runtimes:\n  claude:\n    attach: false\n    thread:\n      default: [--add-dir, /work]\n")
	fake := f.useHerdr()
	f.enterClaudeSession()

	f.mustRun("thread", "create", "login-fix")

	if got := fake.runtimeArgs(t); !slices.Equal(got, []string{"--resume", "session-1", "--fork-session", "--add-dir", "/work"}) {
		t.Fatalf("runtime args %q", got)
	}
}

func TestThreadCreateRejectsAttachThatIsNotTrueOrFalse(t *testing.T) {
	cases := map[string]string{
		"word":          "maybe",
		"yes":           "yes",
		"off":           "off",
		"number":        "1",
		"quoted string": `"false"`,
		"list":          "[]",
	}
	for name, value := range cases {
		t.Run(name, func(t *testing.T) {
			f := newFixture(t)
			f.writeLocalConfig("runtimes:\n  claude:\n    attach: " + value + "\n")
			fake := f.useHerdr()
			f.enterClaudeSession()

			got := f.run("thread", "create", "login-fix", "--repo", "frontend")

			if got.code == 0 {
				t.Fatal("the config was accepted")
			}
			for _, expected := range []string{"config.yaml", "line 3", "expected true or false"} {
				if !strings.Contains(got.stderr, expected) {
					t.Fatalf("stderr misses %q:\n%s", expected, got.stderr)
				}
			}
			if len(fake.calls) != 0 || exists(f.clonePath("frontend")) || exists(f.recordPath("login-fix")) {
				t.Fatal("something was created")
			}
		})
	}
}

func TestHooksGetWorktreeFoldersWhenAttachIsFalse(t *testing.T) {
	f := newFixture(t)
	f.clone("frontend")
	f.writeLocalConfig("runtimes:\n  claude:\n    attach: false\n")
	pre := f.recordingHook("pre-thread-create", "0")
	post := f.recordingHook("post-thread-create", "0")
	f.useHerdr()
	f.enterClaudeSession()

	f.mustRun("thread", "create", "login-fix", "--repo", "backend")

	want := f.worktreesDir("frontend") + "\n" + f.worktreesDir("backend")
	for _, recorded := range []string{pre.mustValue("pre-thread-create", "DIRS"), post.mustValue("post-thread-create", "DIRS")} {
		if recorded != want {
			t.Fatalf("HOMEREPO_HOOK_DIRS %q, want %q", recorded, want)
		}
	}
}

package cli_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/i-am-saebyeok/homerepo/internal/gittest"
)

func (f *fixture) pushToRemote(name string, file string) string {
	f.t.Helper()
	writer := filepath.Join(f.t.TempDir(), "writer")
	gittest.Git(f.t, f.root, "clone", "--quiet", f.remotes[name], writer)
	gittest.WriteFile(f.t, filepath.Join(writer, file), file+"\n")
	gittest.Git(f.t, writer, "add", file)
	gittest.Git(f.t, writer, "commit", "--quiet", "--message", "add "+file)
	gittest.Git(f.t, writer, "push", "--quiet", "origin", "main")
	return strings.TrimSpace(gittest.Git(f.t, writer, "rev-parse", "HEAD"))
}

func (f *fixture) worktreePath(repository string, thread string) string {
	return filepath.Join(f.clonePath(repository), ".worktrees", filepath.FromSlash(thread))
}

func exists(path string) bool {
	_, err := os.Lstat(path)
	return err == nil
}

func TestWorkspaceCreateInsideThread(t *testing.T) {
	f := newFixture(t)
	clone := f.clone("frontend")
	remoteHead := f.pushToRemote("frontend", "newer.txt")
	f.env["HOMEREPO_THREAD"] = "login-fix"

	got := f.mustRun("workspace", "create", "frontend")

	path := f.worktreePath("frontend", "login-fix")
	if got.stdout != path+"\n" {
		t.Fatalf("stdout %q, want %q", got.stdout, path)
	}
	if branch := strings.TrimSpace(gittest.Git(t, path, "branch", "--show-current")); branch != "login-fix" {
		t.Fatalf("branch %q", branch)
	}
	if head := strings.TrimSpace(gittest.Git(t, path, "rev-parse", "HEAD")); head != remoteHead {
		t.Fatalf("worktree starts at %s, want the fetched remote branch %s", head, remoteHead)
	}
	if status := gittest.Git(t, clone, "status", "--porcelain", "--untracked-files=all"); status != "" {
		t.Fatalf("clone status is not clean:\n%s", status)
	}
}

func TestWorkspaceCreateTakesThreadFromFlag(t *testing.T) {
	f := newFixture(t)
	f.clone("frontend")
	got := f.mustRun("workspace", "create", "frontend", "--thread", "fix/login")
	if got.stdout != f.worktreePath("frontend", "fix/login")+"\n" {
		t.Fatalf("stdout %q", got.stdout)
	}
}

func TestWorkspaceCreateOutsideThreadCreatesNothing(t *testing.T) {
	f := newFixture(t)
	got := f.run("workspace", "create", "frontend")
	if got.code == 0 || !strings.Contains(got.stderr, "thread name required") {
		t.Fatalf("code %d, stderr:\n%s", got.code, got.stderr)
	}
	if exists(f.clonePath("frontend")) {
		t.Fatal("a clone was created")
	}
}

func TestWorkspaceCreateRejectsUnusableThreadName(t *testing.T) {
	f := newFixture(t)
	f.clone("frontend")
	for _, name := range []string{"../escape", "a..b", "-flag", "a/", "name.lock"} {
		if got := f.run("workspace", "create", "frontend", "--thread", name); got.code == 0 {
			t.Fatalf("thread name %q was accepted", name)
		}
	}
	if exists(filepath.Join(f.clonePath("frontend"), ".worktrees")) {
		t.Fatal("a worktree folder was created")
	}
}

func TestWorkspaceCreateClonesMissingRepository(t *testing.T) {
	f := newFixture(t)
	f.env["HOMEREPO_THREAD"] = "login-fix"
	got := f.mustRun("workspace", "create", "backend")
	if !exists(filepath.Join(f.clonePath("backend"), ".git")) {
		t.Fatal("clone was not created")
	}
	if got.stdout != f.worktreePath("backend", "login-fix")+"\n" || !exists(f.worktreePath("backend", "login-fix")) {
		t.Fatalf("worktree was not created: %q", got.stdout)
	}
}

func TestWorkspaceCreateTwiceReturnsExistingWorktree(t *testing.T) {
	f := newFixture(t)
	f.clone("frontend")
	f.env["HOMEREPO_THREAD"] = "login-fix"
	first := f.mustRun("workspace", "create", "frontend")
	second := f.mustRun("workspace", "create", "frontend")
	if first.stdout != second.stdout {
		t.Fatalf("paths differ: %q and %q", first.stdout, second.stdout)
	}
}

func TestWorkspaceCreateLeavesDirtyOriginalUntouched(t *testing.T) {
	f := newFixture(t)
	clone := f.clone("frontend")
	gittest.Git(t, clone, "switch", "--quiet", "--create", "human-work")
	gittest.WriteFile(t, filepath.Join(clone, "README.md"), "edited by a person\n")
	gittest.WriteFile(t, filepath.Join(clone, "notes.txt"), "draft\n")
	statusBefore := gittest.Git(t, clone, "status", "--porcelain")
	f.pushToRemote("frontend", "newer.txt")
	f.env["HOMEREPO_THREAD"] = "login-fix"

	f.mustRun("workspace", "create", "frontend")

	if !exists(f.worktreePath("frontend", "login-fix")) {
		t.Fatal("worktree was not created")
	}
	if status := gittest.Git(t, clone, "status", "--porcelain"); status != statusBefore {
		t.Fatalf("status changed:\n%s\nwant:\n%s", status, statusBefore)
	}
	if branch := strings.TrimSpace(gittest.Git(t, clone, "branch", "--show-current")); branch != "human-work" {
		t.Fatalf("branch changed to %q", branch)
	}
	content, err := os.ReadFile(filepath.Join(clone, "README.md"))
	if err != nil || string(content) != "edited by a person\n" {
		t.Fatalf("file changed: %q, %v", content, err)
	}
	if exists(filepath.Join(clone, "newer.txt")) {
		t.Fatal("remote changes reached the original working folder")
	}
}

func TestWorkspaceCreateStopsWhenNoRemoteHasTheAddress(t *testing.T) {
	f := newFixture(t)
	clone := f.clone("frontend")
	gittest.Git(t, clone, "remote", "set-url", "origin", f.remotes["backend"])
	f.env["HOMEREPO_THREAD"] = "login-fix"
	got := f.run("workspace", "create", "frontend")
	if got.code == 0 || !strings.Contains(got.stderr, "no remote") {
		t.Fatalf("code %d, stderr:\n%s", got.code, got.stderr)
	}
	if exists(f.worktreePath("frontend", "login-fix")) {
		t.Fatal("worktree was created")
	}
}

func TestWorkspaceCreateFindsRemoteByAddressNotName(t *testing.T) {
	f := newFixture(t)
	clone := f.clone("frontend")
	gittest.Git(t, clone, "remote", "rename", "origin", "upstream")
	gittest.Git(t, clone, "remote", "add", "origin", f.remotes["backend"])
	remoteHead := f.pushToRemote("frontend", "newer.txt")
	f.env["HOMEREPO_THREAD"] = "login-fix"
	f.mustRun("workspace", "create", "frontend")
	if head := strings.TrimSpace(gittest.Git(t, f.worktreePath("frontend", "login-fix"), "rev-parse", "HEAD")); head != remoteHead {
		t.Fatalf("worktree starts at %s, want %s", head, remoteHead)
	}
}

func TestWorkspaceListMarksOutsideWorktrees(t *testing.T) {
	f := newFixture(t)
	clone := f.clone("frontend")
	outside := filepath.Join(f.root, "elsewhere")
	gittest.Git(t, clone, "worktree", "add", "--quiet", "-b", "other", outside)
	f.mustRun("workspace", "create", "frontend", "--thread", "login-fix")
	gittest.WriteFile(t, filepath.Join(f.worktreePath("frontend", "login-fix"), "new.txt"), "new\n")

	got := f.mustRun("workspace", "list")

	rows := map[string]string{}
	for _, line := range strings.Split(strings.TrimSpace(got.stdout), "\n")[1:] {
		rows[strings.Fields(line)[1]] = line
	}
	if len(rows) != 2 {
		t.Fatalf("expected two worktrees:\n%s", got.stdout)
	}
	owned := rows["login-fix"]
	for _, expected := range []string{"frontend", f.worktreePath("frontend", "login-fix"), "login-fix", "changes"} {
		if !strings.Contains(owned, expected) {
			t.Fatalf("row misses %q: %s", expected, owned)
		}
	}
	if !strings.Contains(rows["outside"], "other") || !strings.Contains(rows["outside"], "clean") {
		t.Fatalf("outside row: %q", rows["outside"])
	}

	if released := f.run("workspace", "release", "frontend", "--thread", "other"); released.code == 0 {
		t.Fatal("an outside worktree was released")
	}
	if !exists(outside) {
		t.Fatal("outside worktree was removed")
	}
}

func TestWorkspaceReleaseCleanWorktreeKeepsBranch(t *testing.T) {
	f := newFixture(t)
	clone := f.clone("frontend")
	f.env["HOMEREPO_THREAD"] = "login-fix"
	f.mustRun("workspace", "create", "frontend")
	path := f.worktreePath("frontend", "login-fix")
	gittest.WriteFile(t, filepath.Join(path, "work.txt"), "work\n")
	gittest.Git(t, path, "add", "work.txt")
	gittest.Git(t, path, "commit", "--quiet", "--message", "work")
	commit := strings.TrimSpace(gittest.Git(t, path, "rev-parse", "HEAD"))

	got := f.mustRun("workspace", "release", "frontend")

	if !strings.Contains(got.stdout, "released") {
		t.Fatalf("stdout %q", got.stdout)
	}
	if exists(path) {
		t.Fatal("worktree folder still exists")
	}
	if branchCommit := strings.TrimSpace(gittest.Git(t, clone, "rev-parse", "refs/heads/login-fix")); branchCommit != commit {
		t.Fatalf("branch points at %s, want %s", branchCommit, commit)
	}
}

func TestWorkspaceCreateAfterReleaseReusesBranch(t *testing.T) {
	f := newFixture(t)
	f.clone("frontend")
	f.env["HOMEREPO_THREAD"] = "login-fix"
	f.mustRun("workspace", "create", "frontend")
	f.mustRun("workspace", "release", "frontend")
	f.mustRun("workspace", "create", "frontend")
	if !exists(f.worktreePath("frontend", "login-fix")) {
		t.Fatal("worktree was not created again")
	}
}

func TestWorkspaceReleaseWithChangesRemovesNothing(t *testing.T) {
	cases := map[string]func(t *testing.T, path string){
		"tracked change": func(t *testing.T, path string) {
			gittest.WriteFile(t, filepath.Join(path, "README.md"), "changed\n")
		},
		"untracked file": func(t *testing.T, path string) {
			gittest.WriteFile(t, filepath.Join(path, "untracked.txt"), "new\n")
		},
	}
	for name, change := range cases {
		t.Run(name, func(t *testing.T) {
			f := newFixture(t)
			f.clone("frontend")
			f.env["HOMEREPO_THREAD"] = "login-fix"
			f.mustRun("workspace", "create", "frontend")
			path := f.worktreePath("frontend", "login-fix")
			change(t, path)

			got := f.run("workspace", "release", "frontend")

			if got.code == 0 {
				t.Fatal("release succeeded")
			}
			if !strings.Contains(got.stderr, "README.md") && !strings.Contains(got.stderr, "untracked.txt") {
				t.Fatalf("stderr does not name the changed file:\n%s", got.stderr)
			}
			if !exists(path) {
				t.Fatal("worktree was removed")
			}
		})
	}
}

func TestWorkspaceListShowsDeletedWorktreeAsMissing(t *testing.T) {
	f := newFixture(t)
	f.clone("frontend")
	f.mustRun("workspace", "create", "frontend", "--thread", "login-fix")
	if err := os.RemoveAll(f.worktreePath("frontend", "login-fix")); err != nil {
		t.Fatal(err)
	}
	got := f.mustRun("workspace", "list")
	if !strings.Contains(got.stdout, "missing") {
		t.Fatalf("stdout:\n%s", got.stdout)
	}
}

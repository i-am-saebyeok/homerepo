package cli_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/i-am-saebyeok/homerepo/internal/gittest"
)

func (f *fixture) declaration() string {
	f.t.Helper()
	content, err := os.ReadFile(filepath.Join(f.home, "homerepo.yaml"))
	if err != nil {
		f.t.Fatal(err)
	}
	return string(content)
}

func (f *fixture) writeDeclaration(content string) {
	f.t.Helper()
	gittest.WriteFile(f.t, filepath.Join(f.home, "homerepo.yaml"), content)
}

func (f *fixture) emptyRemote(name string) string {
	f.t.Helper()
	path := filepath.Join(f.root, "remotes", name+".git")
	gittest.Git(f.t, f.root, "init", "--quiet", "--bare", "--initial-branch", "main", path)
	return path
}

func (f *fixture) remoteWithDefaultBranch(name string, branch string) string {
	f.t.Helper()
	seed := filepath.Join(f.t.TempDir(), "seed")
	gittest.Git(f.t, f.root, "init", "--quiet", "--initial-branch", branch, seed)
	gittest.WriteFile(f.t, filepath.Join(seed, "README.md"), "readme\n")
	gittest.Git(f.t, seed, "add", "README.md")
	gittest.Git(f.t, seed, "commit", "--quiet", "--message", "initial")
	path := filepath.Join(f.root, "remotes", name+".git")
	gittest.Git(f.t, f.root, "clone", "--quiet", "--bare", seed, path)
	return path
}

func TestRepoAddAppendsRepository(t *testing.T) {
	f := newFixture(t)
	before := f.declaration()

	got := f.mustRun("repo", "add", "notes", "--remote", "https://example.com/notes.git", "--branch", "main")

	want := before + "  - name: notes\n    remote: https://example.com/notes.git\n    branch: main\n"
	if f.declaration() != want {
		t.Fatalf("declaration:\n%s\nwant:\n%s", f.declaration(), want)
	}
	if !strings.Contains(got.stdout, "added notes") {
		t.Fatalf("stdout %q", got.stdout)
	}
	list := f.mustRun("repo", "list")
	if !strings.Contains(list.stdout, "notes") || !strings.Contains(list.stdout, "https://example.com/notes.git") {
		t.Fatalf("repo list:\n%s", list.stdout)
	}
	if exists(f.clonePath("notes")) {
		t.Fatal("the repository was cloned")
	}
}

func TestRepoAddReplacesEmptyList(t *testing.T) {
	f := newFixture(t)
	f.writeDeclaration("repositories: []\n")

	f.mustRun("repo", "add", "notes", "--remote", "https://example.com/notes.git", "--branch", "main")

	want := "repositories:\n  - name: notes\n    remote: https://example.com/notes.git\n    branch: main\n"
	if f.declaration() != want {
		t.Fatalf("declaration:\n%s\nwant:\n%s", f.declaration(), want)
	}
}

func TestRepoAddReadsDefaultBranchFromRemote(t *testing.T) {
	f := newFixture(t)
	remote := f.remoteWithDefaultBranch("notes", "trunk")

	got := f.mustRun("repo", "add", "notes", "--remote", remote)

	if !strings.Contains(f.declaration(), "  - name: notes\n    remote: "+remote+"\n    branch: trunk\n") {
		t.Fatalf("declaration:\n%s", f.declaration())
	}
	if !strings.Contains(got.stdout, "branch trunk") {
		t.Fatalf("stdout %q", got.stdout)
	}
}

func TestRepoAddWithoutBranchOnEmptyRemoteChangesNothing(t *testing.T) {
	f := newFixture(t)
	remote := f.emptyRemote("notes")
	before := f.declaration()

	got := f.run("repo", "add", "notes", "--remote", remote)

	if got.code == 0 || !strings.Contains(got.stderr, "cannot read the default branch") || !strings.Contains(got.stderr, "--branch") {
		t.Fatalf("code %d, stderr:\n%s", got.code, got.stderr)
	}
	if f.declaration() != before {
		t.Fatal("the declaration changed")
	}
}

func TestRepoAddRejectsDeclaredName(t *testing.T) {
	f := newFixture(t)
	before := f.declaration()

	got := f.run("repo", "add", "frontend", "--remote", "https://example.com/other.git", "--branch", "main")

	if got.code == 0 || !strings.Contains(got.stderr, `repository "frontend" is already declared`) {
		t.Fatalf("code %d, stderr:\n%s", got.code, got.stderr)
	}
	if f.declaration() != before {
		t.Fatal("the declaration changed")
	}
}

func TestRepoAddKeepsLinesWrittenByPeople(t *testing.T) {
	f := newFixture(t)
	f.writeDeclaration("# repositories this Home works on\nrepositories:\n  # the web app\n  - name: web\n    remote: https://example.com/web.git # mirror\n    branch: main\n\n# keep this note\n")

	f.mustRun("repo", "add", "notes", "--remote", "https://example.com/notes.git", "--branch", "main")

	want := "# repositories this Home works on\nrepositories:\n  # the web app\n  - name: web\n    remote: https://example.com/web.git # mirror\n    branch: main\n  - name: notes\n    remote: https://example.com/notes.git\n    branch: main\n\n# keep this note\n"
	if f.declaration() != want {
		t.Fatalf("declaration:\n%s\nwant:\n%s", f.declaration(), want)
	}
}

func TestRepoAddFollowsListIndentation(t *testing.T) {
	f := newFixture(t)
	f.writeDeclaration("repositories:\n- name: web\n  remote: https://example.com/web.git\n  branch: main\n")

	f.mustRun("repo", "add", "notes", "--remote", "https://example.com/notes.git", "--branch", "main")

	want := "repositories:\n- name: web\n  remote: https://example.com/web.git\n  branch: main\n- name: notes\n  remote: https://example.com/notes.git\n  branch: main\n"
	if f.declaration() != want {
		t.Fatalf("declaration:\n%s\nwant:\n%s", f.declaration(), want)
	}
}

func TestRepoAddQuotesValuesThatNeedIt(t *testing.T) {
	f := newFixture(t)

	f.mustRun("repo", "add", "notes", "--remote", "/srv/my notes # copy: 1", "--branch", "main")

	if !strings.Contains(f.mustRun("repo", "list").stdout, "/srv/my notes # copy: 1") {
		t.Fatalf("declaration:\n%s", f.declaration())
	}
}

func TestRepoAddPrintsLinesWhenListCannotBeAppended(t *testing.T) {
	f := newFixture(t)
	original := "repositories: [{name: web, remote: \"https://example.com/web.git\", branch: main}]\n"
	f.writeDeclaration(original)

	got := f.run("repo", "add", "notes", "--remote", "https://example.com/notes.git", "--branch", "main")

	if got.code == 0 || !strings.Contains(got.stderr, "add these lines") || !strings.Contains(got.stderr, "  - name: notes\n    remote: https://example.com/notes.git\n    branch: main") {
		t.Fatalf("code %d, stderr:\n%s", got.code, got.stderr)
	}
	if f.declaration() != original {
		t.Fatal("the declaration changed")
	}
}

func TestRepoAddRejectsInvalidName(t *testing.T) {
	f := newFixture(t)
	before := f.declaration()
	got := f.run("repo", "add", "../notes", "--remote", "https://example.com/notes.git", "--branch", "main")
	if got.code == 0 || f.declaration() != before {
		t.Fatalf("code %d, declaration:\n%s", got.code, f.declaration())
	}
}

func TestRepoAddRequiresRemote(t *testing.T) {
	f := newFixture(t)
	got := f.run("repo", "add", "notes")
	if got.code != 2 || !strings.Contains(got.stderr, "--remote is required") || !strings.Contains(got.stderr, "usage: homerepo repo add") {
		t.Fatalf("code %d, stderr:\n%s", got.code, got.stderr)
	}
}

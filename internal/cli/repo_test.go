package cli_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUnknownCommandPrintsUsage(t *testing.T) {
	f := newFixture(t)
	got := f.run("nothing")
	if got.code != 2 || !strings.Contains(got.stderr, "usage:") {
		t.Fatalf("code %d, stderr:\n%s", got.code, got.stderr)
	}
}

func TestExtraArgumentsAreUsageErrors(t *testing.T) {
	f := newFixture(t)
	got := f.run("repo", "path", "frontend", "backend")
	if got.code != 2 {
		t.Fatalf("code %d, stderr:\n%s", got.code, got.stderr)
	}
}

func TestRepoListShowsDeclaredRepositories(t *testing.T) {
	f := newFixture(t)
	f.clone("frontend")
	got := f.mustRun("repo", "list")
	lines := strings.Split(strings.TrimSpace(got.stdout), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected a header and two rows:\n%s", got.stdout)
	}
	for _, expected := range []string{"frontend", f.remotes["frontend"], f.clonePath("frontend"), "present"} {
		if !strings.Contains(lines[1], expected) {
			t.Fatalf("row misses %q: %s", expected, lines[1])
		}
	}
	if !strings.Contains(lines[2], "backend") || !strings.Contains(lines[2], "missing") {
		t.Fatalf("unexpected row: %s", lines[2])
	}
}

func TestMissingDeclarationExplainsInit(t *testing.T) {
	f := newFixture(t)
	f.cwd = t.TempDir()
	got := f.run("repo", "list")
	if got.code == 0 || !strings.Contains(got.stderr, "homerepo.yaml not found") || !strings.Contains(got.stderr, "homerepo init") {
		t.Fatalf("code %d, stderr:\n%s", got.code, got.stderr)
	}
}

func TestRepoPathDefaultsToSiblingOfHome(t *testing.T) {
	f := newFixture(t)
	got := f.mustRun("repo", "path", "backend")
	if got.stdout != f.clonePath("backend")+"\n" {
		t.Fatalf("stdout %q", got.stdout)
	}
}

func TestRepoPathWorksFromInsideHome(t *testing.T) {
	f := newFixture(t)
	f.cwd = filepath.Join(f.home, "nested")
	if err := os.MkdirAll(f.cwd, 0o755); err != nil {
		t.Fatal(err)
	}
	if got := f.mustRun("repo", "path", "backend"); got.stdout != f.clonePath("backend")+"\n" {
		t.Fatalf("stdout %q", got.stdout)
	}
}

func TestRepoPathUsesOverride(t *testing.T) {
	f := newFixture(t)
	override := filepath.Join(t.TempDir(), "my-frontend")
	f.writeLocalConfig("repositories:\n  overrides:\n    frontend: " + override + "\n")
	if got := f.mustRun("repo", "path", "frontend"); got.stdout != override+"\n" {
		t.Fatalf("stdout %q", got.stdout)
	}
}

func TestRepoPathOfUndeclaredNameListsDeclaredNames(t *testing.T) {
	f := newFixture(t)
	got := f.run("repo", "path", "mobile")
	if got.code == 0 || !strings.Contains(got.stderr, "backend, frontend") {
		t.Fatalf("code %d, stderr:\n%s", got.code, got.stderr)
	}
}

func TestInitInEmptyHome(t *testing.T) {
	f := newFixture(t)
	f.cwd = t.TempDir()
	got := f.mustRun("init")
	for _, expected := range []string{"created homerepo.yaml", "created .homerepo/", "created .agents/skills/homerepo-thread/SKILL.md"} {
		if !strings.Contains(got.stdout, expected) {
			t.Fatalf("output misses %q:\n%s", expected, got.stdout)
		}
	}
	for _, path := range []string{"homerepo.yaml", ".homerepo/.gitignore", ".agents/skills/homerepo-thread/SKILL.md"} {
		if _, err := os.Stat(filepath.Join(f.cwd, path)); err != nil {
			t.Fatalf("%s was not created: %v", path, err)
		}
	}
}

func TestInitAgainSkipsExistingDeclaration(t *testing.T) {
	f := newFixture(t)
	before, err := os.ReadFile(filepath.Join(f.home, "homerepo.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	got := f.mustRun("init")
	if !strings.Contains(got.stdout, "skipped homerepo.yaml (exists)") {
		t.Fatalf("output does not show the skipped file:\n%s", got.stdout)
	}
	after, err := os.ReadFile(filepath.Join(f.home, "homerepo.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(before) {
		t.Fatal("homerepo.yaml changed")
	}
}

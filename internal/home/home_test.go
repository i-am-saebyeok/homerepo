package home_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/i-am-saebyeok/homerepo/internal/gittest"
	"github.com/i-am-saebyeok/homerepo/internal/home"
)

const declared = `repositories:
  - name: frontend
    remote: https://example.invalid/frontend.git
    branch: main
  - name: backend
    remote: https://example.invalid/backend.git
    branch: dev
`

func newHome(t *testing.T, declaration string) string {
	t.Helper()
	root := filepath.Join(t.TempDir(), "home")
	gittest.WriteFile(t, filepath.Join(root, home.DeclarationFile), declaration)
	return root
}

func TestLoadReadsRepositories(t *testing.T) {
	loaded, err := home.Load(newHome(t, declared))
	if err != nil {
		t.Fatal(err)
	}
	repository, err := loaded.Repository("backend")
	if err != nil {
		t.Fatal(err)
	}
	if repository.Remote != "https://example.invalid/backend.git" || repository.Branch != "dev" {
		t.Fatalf("unexpected repository %+v", repository)
	}
}

func TestLoadRejectsInvalidDeclarations(t *testing.T) {
	cases := map[string]string{
		"unknown field":  "repositories:\n  - name: a\n    remote: r\n    branch: main\n    command: make\n",
		"missing remote": "repositories:\n  - name: a\n    branch: main\n",
		"missing branch": "repositories:\n  - name: a\n    remote: r\n",
		"duplicate name": "repositories:\n  - name: a\n    remote: r\n    branch: main\n  - name: a\n    remote: s\n    branch: main\n",
		"path in name":   "repositories:\n  - name: ../a\n    remote: r\n    branch: main\n",
	}
	for name, declaration := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := home.Load(newHome(t, declaration)); err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}

func TestLoadRejectsInvalidOverrides(t *testing.T) {
	cases := map[string]string{
		"undeclared name": "repositories:\n  overrides:\n    other: /somewhere/other\n",
		"relative path":   "repositories:\n  overrides:\n    frontend: somewhere/frontend\n",
	}
	for name, config := range cases {
		t.Run(name, func(t *testing.T) {
			root := newHome(t, declared)
			gittest.WriteFile(t, filepath.Join(root, home.LocalDir, home.LocalConfigFile), config)
			if _, err := home.Load(root); err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}

func TestClonePathDefaultsToSiblingOfHome(t *testing.T) {
	root := newHome(t, declared)
	loaded, err := home.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := loaded.ClonePath("backend"), filepath.Join(filepath.Dir(root), "backend"); got != want {
		t.Fatalf("clone path %s, want %s", got, want)
	}
}

func TestClonePathUsesLocalOverride(t *testing.T) {
	root := newHome(t, declared)
	override := filepath.Join(t.TempDir(), "elsewhere", "frontend")
	gittest.WriteFile(t, filepath.Join(root, home.LocalDir, home.LocalConfigFile), "repositories:\n  overrides:\n    frontend: "+override+"\n")
	loaded, err := home.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	if got := loaded.ClonePath("frontend"); got != override {
		t.Fatalf("clone path %s, want %s", got, override)
	}
}

func TestRepositoryListsDeclaredNamesWhenUnknown(t *testing.T) {
	loaded, err := home.Load(newHome(t, declared))
	if err != nil {
		t.Fatal(err)
	}
	_, err = loaded.Repository("mobile")
	if err == nil || !strings.Contains(err.Error(), "backend, frontend") {
		t.Fatalf("error does not list declared names: %v", err)
	}
}

func TestEnsureIgnoredDirStaysOutOfGitStatus(t *testing.T) {
	gittest.Isolate(t)
	repository := filepath.Join(t.TempDir(), "repository")
	gittest.Repository(t, repository)
	ignored := filepath.Join(repository, home.WorktreesDir)
	created, err := home.EnsureIgnoredDir(ignored)
	if err != nil || !created {
		t.Fatalf("created=%v err=%v", created, err)
	}
	gittest.WriteFile(t, filepath.Join(ignored, "inside", "file.txt"), "content\n")
	if status := gittest.Git(t, repository, "status", "--porcelain", "--untracked-files=all"); status != "" {
		t.Fatalf("ignored folder shows in git status:\n%s", status)
	}
	if gittest.Git(t, repository, "diff", "HEAD", "--name-only") != "" {
		t.Fatal("the repository's tracked files changed")
	}
	created, err = home.EnsureIgnoredDir(ignored)
	if err != nil || created {
		t.Fatalf("second call: created=%v err=%v", created, err)
	}
}

func TestInitCreatesSkeletonAndSkill(t *testing.T) {
	root := t.TempDir()
	result, err := home.Init(root)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"homerepo.yaml", ".homerepo/", ".agents/skills/homerepo-thread/SKILL.md", ".claude/skills/homerepo-thread"}
	if strings.Join(result.Created, "|") != strings.Join(want, "|") {
		t.Fatalf("created %v, want %v", result.Created, want)
	}
	if _, err := home.Load(root); err != nil {
		t.Fatalf("skeleton does not load: %v", err)
	}
	target, err := os.Readlink(filepath.Join(root, ".claude", "skills", "homerepo-thread"))
	if err != nil {
		t.Fatal(err)
	}
	if filepath.IsAbs(target) {
		t.Fatalf("skill link is absolute: %s", target)
	}
	skill, err := os.ReadFile(filepath.Join(root, ".claude", "skills", "homerepo-thread", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{"name: homerepo-thread", "author: homerepo", "Do code work in a worktree."} {
		if !strings.Contains(string(skill), expected) {
			t.Fatalf("skill misses %q", expected)
		}
	}
}

func TestInitKeepsExistingFiles(t *testing.T) {
	root := newHome(t, declared)
	if _, err := home.Init(root); err != nil {
		t.Fatal(err)
	}
	result, err := home.Init(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Created) != 0 {
		t.Fatalf("second init created %v", result.Created)
	}
	if result.Skipped[0] != "homerepo.yaml" {
		t.Fatalf("skipped %v", result.Skipped)
	}
	content, err := os.ReadFile(filepath.Join(root, home.DeclarationFile))
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != declared {
		t.Fatalf("declaration changed:\n%s", content)
	}
}

func TestInitDoesNotLinkWhenClaudeSkillsIsTheSameFolder(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".agents", "skills"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, ".claude"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join("..", ".agents", "skills"), filepath.Join(root, ".claude", "skills")); err != nil {
		t.Fatal(err)
	}
	result, err := home.Init(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range result.Created {
		if strings.HasPrefix(path, ".claude/") {
			t.Fatalf("created a link inside the shared folder: %s", path)
		}
	}
	info, err := os.Lstat(filepath.Join(root, ".agents", "skills", "homerepo-thread"))
	if err != nil || !info.IsDir() {
		t.Fatalf("skill folder is not a real folder: %v", err)
	}
}

func TestInitLocalFolderStaysOutOfGitStatus(t *testing.T) {
	gittest.Isolate(t)
	root := filepath.Join(t.TempDir(), "home")
	gittest.Repository(t, root)
	if _, err := home.Init(root); err != nil {
		t.Fatal(err)
	}
	gittest.WriteFile(t, filepath.Join(root, home.LocalDir, "threads", "a.json"), "{}\n")
	status := gittest.Git(t, root, "status", "--porcelain", "--untracked-files=all")
	if strings.Contains(status, home.LocalDir) {
		t.Fatalf("local folder shows in git status:\n%s", status)
	}
	if !strings.Contains(status, home.DeclarationFile) {
		t.Fatalf("declaration should be visible to commit:\n%s", status)
	}
}

package gittest

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func Isolate(t *testing.T) {
	t.Helper()
	globalConfig := filepath.Join(t.TempDir(), "gitconfig")
	if err := os.WriteFile(globalConfig, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GIT_CONFIG_GLOBAL", globalConfig)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	t.Setenv("GIT_AUTHOR_NAME", "Test")
	t.Setenv("GIT_AUTHOR_EMAIL", "test@example.invalid")
	t.Setenv("GIT_COMMITTER_NAME", "Test")
	t.Setenv("GIT_COMMITTER_EMAIL", "test@example.invalid")
}

func Git(t *testing.T, dir string, args ...string) string {
	t.Helper()
	process := exec.Command("git", args...)
	process.Dir = dir
	output, err := process.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, output)
	}
	return string(output)
}

func Repository(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	Git(t, dir, "init", "--quiet", "--initial-branch", "main")
	WriteFile(t, filepath.Join(dir, "README.md"), "readme\n")
	Git(t, dir, "add", "README.md")
	Git(t, dir, "commit", "--quiet", "--message", "initial")
}

func BareRemote(t *testing.T, dir string) string {
	t.Helper()
	seed := filepath.Join(t.TempDir(), "seed")
	Repository(t, seed)
	Git(t, seed, "clone", "--quiet", "--bare", seed, dir)
	return dir
}

func WriteFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

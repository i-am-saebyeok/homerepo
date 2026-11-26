package cli_test

import (
	"strings"
	"testing"
)

func TestEveryCommandPrintsHelp(t *testing.T) {
	commands := [][]string{
		{"init"}, {"repo", "list"}, {"repo", "path"}, {"repo", "add"},
		{"workspace", "create"}, {"workspace", "list"}, {"workspace", "release"},
		{"thread", "create"}, {"thread", "list"}, {"thread", "close"},
	}
	for _, name := range commands {
		for _, flag := range []string{"--help", "-h"} {
			t.Run(strings.Join(name, " ")+" "+flag, func(t *testing.T) {
				f := newFixture(t)
				got := f.run(append(name, flag)...)
				if got.code != 0 || got.stderr != "" || !strings.HasPrefix(got.stdout, "usage: homerepo "+strings.Join(name, " ")) {
					t.Fatalf("code %d, stdout:\n%s\nstderr:\n%s", got.code, got.stdout, got.stderr)
				}
			})
		}
	}
}

func TestThreadCreateHelpDescribesOptions(t *testing.T) {
	f := newFixture(t)
	got := f.run("thread", "create", "--help")
	if got.code != 0 {
		t.Fatalf("code %d, stderr:\n%s", got.code, got.stderr)
	}
	for _, option := range []string{"--repo <repository>", "--mode <bundle>", "--runtime <runtime>", "--session <id>"} {
		if !strings.Contains(got.stdout, option) {
			t.Fatalf("help misses %q:\n%s", option, got.stdout)
		}
	}
}

func TestGroupCommandsPrintTheirCommands(t *testing.T) {
	cases := map[string][]string{
		"homerepo":  {"homerepo repo add", "homerepo workspace create", "homerepo thread close"},
		"repo":      {"homerepo repo list", "homerepo repo path", "homerepo repo add"},
		"workspace": {"homerepo workspace create", "homerepo workspace list", "homerepo workspace release"},
		"thread":    {"homerepo thread create", "homerepo thread list", "homerepo thread close"},
	}
	for group, expected := range cases {
		for _, flag := range []string{"--help", "-h"} {
			t.Run(group+" "+flag, func(t *testing.T) {
				f := newFixture(t)
				args := []string{flag}
				if group != "homerepo" {
					args = []string{group, flag}
				}
				got := f.run(args...)
				if got.code != 0 || got.stderr != "" {
					t.Fatalf("code %d, stderr:\n%s", got.code, got.stderr)
				}
				for _, line := range expected {
					if !strings.Contains(got.stdout, line) {
						t.Fatalf("stdout misses %q:\n%s", line, got.stdout)
					}
				}
				if group == "repo" && strings.Contains(got.stdout, "thread") {
					t.Fatalf("repo help lists other commands:\n%s", got.stdout)
				}
			})
		}
	}
}

func TestWrongArgumentsPrintUsageToStderr(t *testing.T) {
	f := newFixture(t)
	got := f.run("thread", "create")
	if got.code != 2 || got.stdout != "" {
		t.Fatalf("code %d, stdout %q", got.code, got.stdout)
	}
	for _, expected := range []string{"expected 1 argument(s), got 0", "usage: homerepo thread create <name>"} {
		if !strings.Contains(got.stderr, expected) {
			t.Fatalf("stderr misses %q:\n%s", expected, got.stderr)
		}
	}
}

func TestUnknownGroupCommandPrintsGroupUsage(t *testing.T) {
	f := newFixture(t)
	got := f.run("repo", "remove", "notes")
	if got.code != 2 || !strings.Contains(got.stderr, "unknown command: repo remove notes") || !strings.Contains(got.stderr, "homerepo repo add") {
		t.Fatalf("code %d, stderr:\n%s", got.code, got.stderr)
	}
}

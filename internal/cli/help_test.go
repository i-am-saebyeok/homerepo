package cli_test

import (
	"strings"
	"testing"
)

func TestEveryCommandPrintsHelp(t *testing.T) {
	commands := [][]string{
		{"init"}, {"repo", "list"}, {"repo", "path"}, {"repo", "add"},
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

func TestRepoAddHelpDescribesOptions(t *testing.T) {
	f := newFixture(t)
	got := f.run("repo", "add", "--help")
	if got.code != 0 {
		t.Fatalf("code %d, stderr:\n%s", got.code, got.stderr)
	}
	for _, option := range []string{"--remote <address>", "--branch <branch>"} {
		if !strings.Contains(got.stdout, option) {
			t.Fatalf("help misses %q:\n%s", option, got.stdout)
		}
	}
}

func TestGroupCommandsPrintTheirCommands(t *testing.T) {
	cases := map[string][]string{
		"homerepo": {"homerepo init", "homerepo repo list", "homerepo repo path", "homerepo repo add"},
		"repo":     {"homerepo repo list", "homerepo repo path", "homerepo repo add"},
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
				if group == "repo" && strings.Contains(got.stdout, "homerepo init") {
					t.Fatalf("repo help lists other commands:\n%s", got.stdout)
				}
				for _, removed := range []string{"thread", "workspace"} {
					if strings.Contains(got.stdout, removed) {
						t.Fatalf("help lists the removed command %q:\n%s", removed, got.stdout)
					}
				}
			})
		}
	}
}

func TestWrongArgumentsPrintUsageToStderr(t *testing.T) {
	f := newFixture(t)
	got := f.run("repo", "path")
	if got.code != 2 || got.stdout != "" {
		t.Fatalf("code %d, stdout %q", got.code, got.stdout)
	}
	for _, expected := range []string{"expected 1 argument(s), got 0", "usage: homerepo repo path <repository>"} {
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

func TestRemovedCommandsAreUnknown(t *testing.T) {
	for _, args := range [][]string{{"thread", "create", "login-fix"}, {"workspace", "create", "frontend"}} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			f := newFixture(t)
			got := f.run(args...)
			if got.code != 2 || got.stdout != "" {
				t.Fatalf("code %d, stdout %q", got.code, got.stdout)
			}
			for _, expected := range []string{"unknown command: " + strings.Join(args, " "), "homerepo repo add"} {
				if !strings.Contains(got.stderr, expected) {
					t.Fatalf("stderr misses %q:\n%s", expected, got.stderr)
				}
			}
		})
	}
}

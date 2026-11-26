package workspace

import "testing"

func TestCheckThreadName(t *testing.T) {
	for _, name := range []string{"login-fix", "fix/login", "v1.2", "A_b"} {
		if err := CheckThreadName(name); err != nil {
			t.Errorf("%q rejected: %v", name, err)
		}
	}
	for _, name := range []string{"", "-x", ".hidden", "a..b", "a/", "/a", "a//b", "x.lock", "end.", "with space", "a/../b"} {
		if err := CheckThreadName(name); err == nil {
			t.Errorf("%q accepted", name)
		}
	}
}

func TestSameAddressIgnoresTrailingSlashAndGitSuffix(t *testing.T) {
	if !sameAddress("https://example.invalid/org/app.git", "https://example.invalid/org/app/") {
		t.Fatal("expected the same address")
	}
	if sameAddress("git@example.invalid:org/app.git", "https://example.invalid/org/app.git") {
		t.Fatal("different transports must not match")
	}
}

func TestParseWorktreeList(t *testing.T) {
	output := "worktree /clone\nHEAD 1111\nbranch refs/heads/main\n\nworktree /clone/.worktrees/a\nHEAD 2222\nbranch refs/heads/a\n\nworktree /gone\nHEAD 3333\ndetached\nprunable gitdir file points to non-existent location\n"
	entries := linkedWorktrees(parseWorktreeList(output))
	if len(entries) != 2 {
		t.Fatalf("entries %+v", entries)
	}
	if entries[0].path != "/clone/.worktrees/a" || entries[0].branch != "a" || entries[0].prunable {
		t.Fatalf("first entry %+v", entries[0])
	}
	if entries[1].branch != "(detached)" || !entries[1].prunable {
		t.Fatalf("second entry %+v", entries[1])
	}
}

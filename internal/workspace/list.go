package workspace

import "strings"

type worktreeEntry struct {
	path     string
	branch   string
	prunable bool
}

func parseWorktreeList(output string) []worktreeEntry {
	var entries []worktreeEntry
	for _, block := range strings.Split(strings.TrimSpace(output), "\n\n") {
		var entry worktreeEntry
		for _, line := range strings.Split(block, "\n") {
			key, value, _ := strings.Cut(line, " ")
			switch key {
			case "worktree":
				entry.path = value
			case "branch":
				entry.branch = strings.TrimPrefix(value, "refs/heads/")
			case "detached":
				entry.branch = "(detached)"
			case "prunable":
				entry.prunable = true
			}
		}
		entries = append(entries, entry)
	}
	return entries
}

func linkedWorktrees(entries []worktreeEntry) []worktreeEntry {
	if len(entries) == 0 {
		return nil
	}
	return entries[1:]
}

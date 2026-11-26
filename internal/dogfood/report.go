package dogfood

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/i-am-saebyeok/homerepo/internal/home"
)

func (s session) report(seen observation, checks []checkResult, elapsed time.Duration) {
	out := s.app.Stdout
	fmt.Fprintf(out, "scenario %s with %s\n", s.options.scenario.Name, s.options.runtime)
	for _, check := range checks {
		if check.failure == nil {
			fmt.Fprintf(out, "PASS  %s\n", check.name)
			continue
		}
		fmt.Fprintf(out, "FAIL  %s: %s\n", check.name, check.failure)
	}
	for _, problem := range seen.problems {
		fmt.Fprintf(out, "note: %s\n", problem)
	}
	if seen.timedOut {
		fmt.Fprintf(out, "timed out after %s\n", s.options.timeout)
	}
	fmt.Fprintf(out, "elapsed: %s\n", elapsed.Round(time.Second))
	fmt.Fprintf(out, "home: %s\n", s.layout.Home())
	fmt.Fprintf(out, "coordinator: tab %s, pane %s\n", s.coordinator.ID, s.coordinator.Pane)
	if seen.thread.Name == "" {
		return
	}
	fmt.Fprintf(out, "thread: %s, tab %s, pane %s\n", seen.thread.Name, seen.thread.TabID, paneLabel(seen.threadPane))
	notes := s.notesPath(seen.thread.Name)
	fmt.Fprintf(out, "notes: %s (%s)\n", notes, presence(notes))
}

func (s session) notesPath(threadName string) string {
	loaded, err := home.Load(s.layout.Home())
	worktrees := filepath.Join(s.layout.Root, s.options.scenario.Repository, home.WorktreesDir)
	if err == nil {
		worktrees = loaded.WorktreesPath(s.options.scenario.Repository)
	}
	return filepath.Join(worktrees, filepath.FromSlash(threadName), NotesFile)
}

func paneLabel(pane string) string {
	if pane == "" {
		return "not found"
	}
	return pane
}

func presence(path string) string {
	if isFile(path) {
		return "present"
	}
	return "missing"
}

package dogfood

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/i-am-saebyeok/homerepo/internal/home"
)

type checkResult struct {
	name    string
	failure error
}

func allPassed(checks []checkResult) bool {
	for _, check := range checks {
		if check.failure != nil {
			return false
		}
	}
	return true
}

var errNoThread = errors.New("no thread was created")

func (s session) check(ctx context.Context, seen observation) []checkResult {
	return []checkResult{
		{name: "thread started", failure: threadStartedFailure(seen)},
		{name: "repository declared", failure: s.checkRepositoryDeclared(ctx)},
		{name: "worktree", failure: s.checkWorktree(ctx, seen.thread.Name)},
		{name: "remote", failure: s.checkRemote(ctx, seen.thread.Name)},
	}
}

func threadStartedFailure(seen observation) error {
	if seen.thread.Name == "" {
		return errNoThread
	}
	if !seen.threadStarted {
		return fmt.Errorf("thread %q did not take input", seen.thread.Name)
	}
	return nil
}

func (s session) checkRepositoryDeclared(ctx context.Context) error {
	repository := s.options.scenario.Repository
	output, err := s.homerepo(ctx, "repo", "list")
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(bytes.NewReader(output.Stdout))
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) == 0 || fields[0] != repository {
			continue
		}
		if fields[len(fields)-1] != "present" {
			return fmt.Errorf("%s is declared but its clone is missing", repository)
		}
		return nil
	}
	return fmt.Errorf("%s is not in `homerepo repo list`", repository)
}

func (s session) checkWorktree(ctx context.Context, threadName string) error {
	if threadName == "" {
		return errNoThread
	}
	loaded, err := home.Load(s.layout.Home())
	if err != nil {
		return err
	}
	clone := loaded.ClonePath(s.options.scenario.Repository)
	if !home.IsDir(clone) {
		return fmt.Errorf("no clone at %s", clone)
	}
	output, err := s.git(ctx, clone, "worktree", "list", "--porcelain")
	if err != nil {
		return err
	}
	expected := filepath.Join(loaded.WorktreesPath(s.options.scenario.Repository), filepath.FromSlash(threadName))
	for _, worktree := range parseWorktrees(output.Stdout) {
		if samePath(worktree.path, expected) && worktree.branch == "refs/heads/"+threadName {
			return nil
		}
	}
	return fmt.Errorf("no worktree at %s on branch %s", expected, threadName)
}

type listedWorktree struct {
	path   string
	branch string
}

func parseWorktrees(porcelain []byte) []listedWorktree {
	var worktrees []listedWorktree
	for _, block := range strings.Split(strings.TrimSpace(string(porcelain)), "\n\n") {
		var worktree listedWorktree
		for _, line := range strings.Split(block, "\n") {
			if path, ok := strings.CutPrefix(line, "worktree "); ok {
				worktree.path = path
			}
			if branch, ok := strings.CutPrefix(line, "branch "); ok {
				worktree.branch = branch
			}
		}
		worktrees = append(worktrees, worktree)
	}
	return worktrees
}

func samePath(first string, second string) bool {
	return resolved(first) == resolved(second)
}

func resolved(path string) string {
	followed, err := filepath.EvalSymlinks(path)
	if err != nil {
		return filepath.Clean(path)
	}
	return followed
}

func (s session) checkRemote(ctx context.Context, threadName string) error {
	if threadName == "" {
		return errNoThread
	}
	loaded, err := home.Load(s.layout.Home())
	if err != nil {
		return err
	}
	repository, err := loaded.Repository(s.options.scenario.Repository)
	if err != nil {
		return err
	}
	remote := s.layout.Remote(repository.Name)
	threadBranch := "refs/heads/" + threadName
	baseBranch := "refs/heads/" + repository.Branch
	for _, branch := range []string{threadBranch, baseBranch} {
		if _, err := s.git(ctx, remote, "rev-parse", "--verify", "--quiet", branch); err != nil {
			return fmt.Errorf("no branch %s in %s", strings.TrimPrefix(branch, "refs/heads/"), remote)
		}
	}
	output, err := s.git(ctx, remote, "rev-list", "--count", threadBranch, "^"+baseBranch)
	if err != nil {
		return err
	}
	if strings.TrimSpace(string(output.Stdout)) == "0" {
		return fmt.Errorf("branch %s in %s has no commit that %s lacks", threadName, remote, repository.Branch)
	}
	return nil
}

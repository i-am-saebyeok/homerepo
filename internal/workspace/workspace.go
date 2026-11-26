package workspace

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/i-am-saebyeok/homerepo/internal/command"
	"github.com/i-am-saebyeok/homerepo/internal/home"
)

type ChangeState string

const (
	Clean   ChangeState = "clean"
	Changed ChangeState = "changes"
	Missing ChangeState = "missing"
)

type Worktree struct {
	Repository string
	Path       string
	Branch     string
	Thread     string
	Outside    bool
	Changes    ChangeState
}

type Workspaces struct {
	Home   home.Home
	Runner command.Runner
}

type UncommittedChanges struct {
	Repository string
	Path       string
	Files      []string
}

func (e *UncommittedChanges) Error() string {
	return fmt.Sprintf("worktree %s of %s has uncommitted changes:\n  %s", e.Path, e.Repository, strings.Join(e.Files, "\n  "))
}

var threadNamePart = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`)

func CheckThreadName(name string) error {
	for _, part := range strings.Split(name, "/") {
		if !threadNamePart.MatchString(part) || strings.Contains(part, "..") || strings.HasSuffix(part, ".") || strings.HasSuffix(part, ".lock") {
			return fmt.Errorf("thread name %q cannot be a branch and folder name; use letters, digits, '.', '_' and '-', with '/' between parts", name)
		}
	}
	return nil
}

func (w Workspaces) Create(ctx context.Context, repositoryName string, thread string) (string, error) {
	if err := CheckThreadName(thread); err != nil {
		return "", err
	}
	repository, err := w.Home.Repository(repositoryName)
	if err != nil {
		return "", err
	}
	clone := w.Home.ClonePath(repositoryName)
	path := w.threadPath(repositoryName, thread)
	if home.IsDir(clone) {
		_, found, err := w.find(ctx, repositoryName, thread)
		if err != nil {
			return "", err
		}
		if found {
			return path, nil
		}
	}
	if err := w.requireRemoteBranch(ctx, repository.Remote, repository.Branch); err != nil {
		return "", fmt.Errorf("%s: %w", repositoryName, err)
	}
	if !home.IsDir(clone) {
		if err := w.cloneRepository(ctx, repository, clone); err != nil {
			return "", err
		}
	}
	if _, err := os.Lstat(path); err == nil {
		return "", fmt.Errorf("%s exists but is not a worktree of %s", path, clone)
	}
	remote, err := w.findRemote(ctx, clone, repository.Remote)
	if err != nil {
		return "", fmt.Errorf("%s: %w", repositoryName, err)
	}
	base, err := w.fetch(ctx, clone, remote, repository.Branch)
	if err != nil {
		return "", err
	}
	if _, err := home.EnsureIgnoredDir(w.Home.WorktreesPath(repositoryName)); err != nil {
		return "", err
	}
	return path, w.addWorktree(ctx, clone, path, thread, base)
}

func (w Workspaces) cloneRepository(ctx context.Context, repository home.Repository, clone string) error {
	if err := os.MkdirAll(filepath.Dir(clone), 0o755); err != nil {
		return err
	}
	_, err := w.git(ctx, filepath.Dir(clone), "clone", "--quiet", repository.Remote, clone)
	return err
}

func (w Workspaces) Release(ctx context.Context, repositoryName string, thread string) (string, error) {
	if err := CheckThreadName(thread); err != nil {
		return "", err
	}
	if _, err := w.Home.Repository(repositoryName); err != nil {
		return "", err
	}
	clone := w.Home.ClonePath(repositoryName)
	if !home.IsDir(clone) {
		return "", fmt.Errorf("clone of %s not found at %s", repositoryName, clone)
	}
	worktree, found, err := w.find(ctx, repositoryName, thread)
	if err != nil {
		return "", err
	}
	if !found {
		return "", fmt.Errorf("thread %q has no worktree in %s", thread, repositoryName)
	}
	if err := w.CheckClean(ctx, worktree); err != nil {
		return "", err
	}
	_, err = w.git(ctx, clone, "worktree", "remove", worktree.Path)
	return worktree.Path, err
}

func (w Workspaces) CheckClean(ctx context.Context, worktree Worktree) error {
	files, err := w.changedFiles(ctx, worktree.Path)
	if err != nil {
		return err
	}
	if len(files) > 0 {
		return &UncommittedChanges{Repository: worktree.Repository, Path: worktree.Path, Files: files}
	}
	return nil
}

func (w Workspaces) List(ctx context.Context) ([]Worktree, error) {
	var all []Worktree
	for _, repository := range w.Home.Repositories {
		if !home.IsDir(w.Home.ClonePath(repository.Name)) {
			continue
		}
		worktrees, err := w.listRepository(ctx, repository.Name)
		if err != nil {
			return nil, err
		}
		all = append(all, worktrees...)
	}
	return all, nil
}

func (w Workspaces) OfThread(ctx context.Context, thread string) ([]Worktree, error) {
	all, err := w.List(ctx)
	if err != nil {
		return nil, err
	}
	return OwnedBy(all, thread), nil
}

func OwnedBy(worktrees []Worktree, thread string) []Worktree {
	var owned []Worktree
	for _, worktree := range worktrees {
		if !worktree.Outside && worktree.Thread == thread {
			owned = append(owned, worktree)
		}
	}
	return owned
}

func (w Workspaces) find(ctx context.Context, repositoryName string, thread string) (Worktree, bool, error) {
	worktrees, err := w.listRepository(ctx, repositoryName)
	if err != nil {
		return Worktree{}, false, err
	}
	owned := OwnedBy(worktrees, thread)
	if len(owned) == 0 {
		return Worktree{}, false, nil
	}
	return owned[0], true, nil
}

func (w Workspaces) listRepository(ctx context.Context, repositoryName string) ([]Worktree, error) {
	output, err := w.git(ctx, w.Home.ClonePath(repositoryName), "worktree", "list", "--porcelain")
	if err != nil {
		return nil, err
	}
	worktreesDir := resolve(w.Home.WorktreesPath(repositoryName))
	var worktrees []Worktree
	for _, entry := range linkedWorktrees(parseWorktreeList(output)) {
		worktree := w.describe(repositoryName, worktreesDir, entry)
		state, err := w.changeState(ctx, worktree.Path, entry)
		if err != nil {
			return nil, err
		}
		worktree.Changes = state
		worktrees = append(worktrees, worktree)
	}
	return worktrees, nil
}

func (w Workspaces) describe(repositoryName string, worktreesDir string, entry worktreeEntry) Worktree {
	worktree := Worktree{Repository: repositoryName, Path: entry.path, Branch: entry.branch, Outside: true}
	relative, err := filepath.Rel(worktreesDir, resolve(entry.path))
	if err != nil || relative == "." || strings.HasPrefix(relative, "..") {
		return worktree
	}
	worktree.Outside = false
	worktree.Thread = filepath.ToSlash(relative)
	worktree.Path = w.threadPath(repositoryName, worktree.Thread)
	return worktree
}

func (w Workspaces) changeState(ctx context.Context, path string, entry worktreeEntry) (ChangeState, error) {
	if entry.prunable || !home.IsDir(path) {
		return Missing, nil
	}
	files, err := w.changedFiles(ctx, path)
	if err != nil {
		return "", err
	}
	if len(files) > 0 {
		return Changed, nil
	}
	return Clean, nil
}

func (w Workspaces) threadPath(repositoryName string, thread string) string {
	return filepath.Join(w.Home.WorktreesPath(repositoryName), filepath.FromSlash(thread))
}

func (w Workspaces) changedFiles(ctx context.Context, path string) ([]string, error) {
	output, err := w.git(ctx, path, "status", "--porcelain")
	if err != nil {
		return nil, err
	}
	trimmed := strings.TrimRight(output, "\n")
	if trimmed == "" {
		return nil, nil
	}
	return strings.Split(trimmed, "\n"), nil
}

func (w Workspaces) addWorktree(ctx context.Context, clone string, path string, branch string, base string) error {
	exists, err := w.branchExists(ctx, clone, branch)
	if err != nil {
		return err
	}
	if exists {
		_, err = w.git(ctx, clone, "worktree", "add", "--quiet", path, branch)
		return err
	}
	_, err = w.git(ctx, clone, "worktree", "add", "--quiet", "--no-track", "-b", branch, path, base)
	return err
}

func (w Workspaces) branchExists(ctx context.Context, clone string, branch string) (bool, error) {
	_, err := w.git(ctx, clone, "show-ref", "--verify", "--quiet", "refs/heads/"+branch)
	if isExitCode(err, 1) {
		return false, nil
	}
	return err == nil, err
}

func (w Workspaces) git(ctx context.Context, dir string, args ...string) (string, error) {
	output, err := w.Runner.Run(ctx, command.Command{Name: "git", Args: append([]string{"-C", dir}, args...)})
	return string(output.Stdout), err
}

func resolve(path string) string {
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		return resolved
	}
	parent := filepath.Dir(path)
	if parent == path {
		return path
	}
	return filepath.Join(resolve(parent), filepath.Base(path))
}

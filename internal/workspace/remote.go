package workspace

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/i-am-saebyeok/homerepo/internal/command"
)

func (w Workspaces) findRemote(ctx context.Context, clone string, address string) (string, error) {
	output, err := w.git(ctx, clone, "config", "--get-regexp", `^remote\..*\.url$`)
	if err != nil && !isExitCode(err, 1) {
		return "", err
	}
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		key, url, found := strings.Cut(line, " ")
		if found && sameAddress(url, address) {
			return strings.TrimSuffix(strings.TrimPrefix(key, "remote."), ".url"), nil
		}
	}
	return "", fmt.Errorf("no remote of %s has the address %s", clone, address)
}

func (w Workspaces) fetch(ctx context.Context, clone string, remote string, branch string) (string, error) {
	tracking := "refs/remotes/" + remote + "/" + branch
	_, err := w.git(ctx, clone, "fetch", "--quiet", remote, "+refs/heads/"+branch+":"+tracking)
	return tracking, err
}

func sameAddress(first string, second string) bool {
	return normalizeAddress(first) == normalizeAddress(second)
}

func normalizeAddress(address string) string {
	return strings.TrimSuffix(strings.TrimRight(strings.TrimSpace(address), "/"), ".git")
}

func isExitCode(err error, code int) bool {
	var failure *command.Failure
	return errors.As(err, &failure) && failure.ExitCode == code
}

func DefaultBranch(ctx context.Context, runner command.Runner, remote string) (string, error) {
	output, err := runner.Run(ctx, command.Command{Name: "git", Args: []string{"ls-remote", "--symref", remote, "HEAD"}})
	if err != nil {
		return "", fmt.Errorf("cannot read the default branch of %s: %w; pass --branch <branch>", remote, err)
	}
	for _, line := range strings.Split(string(output.Stdout), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 3 && fields[0] == "ref:" && fields[2] == "HEAD" && strings.HasPrefix(fields[1], "refs/heads/") {
			return strings.TrimPrefix(fields[1], "refs/heads/"), nil
		}
	}
	return "", fmt.Errorf("cannot read the default branch of %s; it may have no commits yet. Pass --branch <branch>", remote)
}

func (w Workspaces) requireRemoteBranch(ctx context.Context, remote string, branch string) error {
	_, err := w.Runner.Run(ctx, command.Command{Name: "git", Args: []string{"ls-remote", "--exit-code", "--heads", remote, "refs/heads/" + branch}})
	if isExitCode(err, 2) {
		return fmt.Errorf("branch %q is not on the remote %s yet; push a first commit to %s there, then run this again", branch, remote, branch)
	}
	return err
}

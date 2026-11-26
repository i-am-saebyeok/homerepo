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

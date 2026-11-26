package herdrtest

import (
	"os/exec"
	"strings"
	"testing"
)

func ShellWords(t *testing.T, typed []string) []string {
	t.Helper()
	output, err := exec.Command("/bin/sh", "-c", `printf '%s\000' `+strings.Join(typed, " ")).Output()
	if err != nil {
		t.Fatalf("sh could not read %q: %v", strings.Join(typed, " "), err)
	}
	words := strings.Split(string(output), "\x00")
	return words[:len(words)-1]
}

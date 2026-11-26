package home

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"go.yaml.in/yaml/v3"
)

var (
	repositoriesBlock = regexp.MustCompile(`^repositories:[ \t]*(#.*)?$`)
	emptyRepositories = regexp.MustCompile(`^repositories:[ \t]*\[[ \t]*\][ \t]*(#.*)?$`)
	listItem          = regexp.MustCompile(`^([ ]*)- `)
)

const defaultItemIndent = "  "

type ManualAddition struct {
	Entry []string
}

func (e *ManualAddition) Error() string {
	return fmt.Sprintf("cannot add the repository to %s automatically; add these lines to its repositories list:\n%s", DeclarationFile, indentLines(e.Entry, "  "))
}

func (h Home) CheckNewRepository(name string) error {
	if _, err := h.Repository(name); err == nil {
		return fmt.Errorf("repository %q is already declared in %s", name, DeclarationFile)
	}
	return checkRepositoryName(name)
}

func (h Home) AddRepository(repository Repository) error {
	if err := h.CheckNewRepository(repository.Name); err != nil {
		return err
	}
	if err := checkRepositories([]Repository{repository}); err != nil {
		return err
	}
	path := filepath.Join(h.Root, DeclarationFile)
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	before, err := parseDeclaration(content)
	if err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	entry, err := repositoryEntry(repository)
	if err != nil {
		return err
	}
	updated, ok := appendRepository(string(content), entry)
	if !ok || !declares(updated, append(slices.Clone(before.Repositories), repository)) {
		return &ManualAddition{Entry: entry}
	}
	return os.WriteFile(path, []byte(updated), info.Mode().Perm())
}

func declares(content string, repositories []Repository) bool {
	parsed, err := parseDeclaration([]byte(content))
	return err == nil && slices.Equal(parsed.Repositories, repositories)
}

func repositoryEntry(repository Repository) ([]string, error) {
	fields := []struct {
		key   string
		value string
	}{{"name", repository.Name}, {"remote", repository.Remote}, {"branch", repository.Branch}}
	entry := make([]string, 0, len(fields))
	for index, field := range fields {
		value, err := scalar(field.value)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", field.key, err)
		}
		prefix := "  "
		if index == 0 {
			prefix = "- "
		}
		entry = append(entry, prefix+field.key+": "+value)
	}
	return entry, nil
}

func scalar(value string) (string, error) {
	encoded, err := yaml.Marshal(value)
	if err != nil {
		return "", err
	}
	line := strings.TrimSuffix(string(encoded), "\n")
	if strings.Contains(line, "\n") {
		return "", fmt.Errorf("%q must be on one line", value)
	}
	return line, nil
}

func appendRepository(content string, entry []string) (string, bool) {
	lines := strings.Split(content, "\n")
	for index, line := range lines {
		if match := emptyRepositories.FindStringSubmatch(line); match != nil {
			replacement := append([]string{strings.TrimSpace("repositories: " + match[1])}, indentEach(entry, defaultItemIndent)...)
			return joinLines(slices.Replace(lines, index, index+1, replacement...)), true
		}
		if repositoriesBlock.MatchString(line) {
			return appendToBlock(lines, index, entry)
		}
		if strings.HasPrefix(line, "repositories:") {
			return "", false
		}
	}
	block := append([]string{"repositories:"}, indentEach(entry, defaultItemIndent)...)
	return joinLines(slices.Insert(lines, endOfContent(lines), block...)), true
}

func appendToBlock(lines []string, key int, entry []string) (string, bool) {
	last := key
	indent := defaultItemIndent
	sawItem := false
	for index := key + 1; index < len(lines) && !startsTopLevel(lines[index]); index++ {
		trimmed := strings.TrimSpace(lines[index])
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if !sawItem {
			match := listItem.FindStringSubmatch(lines[index])
			if match == nil {
				return "", false
			}
			indent = match[1]
			sawItem = true
		}
		last = index
	}
	return joinLines(slices.Insert(lines, last+1, indentEach(entry, indent)...)), true
}

func startsTopLevel(line string) bool {
	return line != "" && !strings.ContainsRune(" \t#-", rune(line[0]))
}

func indentEach(lines []string, indent string) []string {
	indented := make([]string, 0, len(lines))
	for _, line := range lines {
		indented = append(indented, indent+line)
	}
	return indented
}

func indentLines(lines []string, indent string) string {
	return strings.Join(indentEach(lines, indent), "\n")
}

func endOfContent(lines []string) int {
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		return len(lines) - 1
	}
	return len(lines)
}

func joinLines(lines []string) string {
	joined := strings.Join(lines, "\n")
	if strings.HasSuffix(joined, "\n") {
		return joined
	}
	return joined + "\n"
}

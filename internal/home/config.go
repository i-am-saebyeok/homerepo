package home

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"

	"go.yaml.in/yaml/v3"
)

type Repository struct {
	Name   string `yaml:"name"`
	Remote string `yaml:"remote"`
	Branch string `yaml:"branch"`
}

type declaration struct {
	Repositories []Repository `yaml:"repositories"`
}

type LocalConfig struct {
	Repositories RepositoryOverrides      `yaml:"repositories"`
	Runtimes     map[string]RuntimeConfig `yaml:"runtimes"`
}

type RepositoryOverrides struct {
	Overrides map[string]string `yaml:"overrides"`
}

type RuntimeConfig struct {
	Thread map[string][]string `yaml:"thread"`
}

var repositoryName = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`)

func readDeclaration(path string) (declaration, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return declaration{}, err
	}
	var parsed declaration
	if err := decodeStrict(content, &parsed); err != nil {
		return declaration{}, fmt.Errorf("%s: %w", path, err)
	}
	if err := checkRepositories(parsed.Repositories); err != nil {
		return declaration{}, fmt.Errorf("%s: %w", path, err)
	}
	return parsed, nil
}

func readLocalConfig(path string) (LocalConfig, error) {
	content, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return LocalConfig{}, nil
	}
	if err != nil {
		return LocalConfig{}, err
	}
	var parsed LocalConfig
	if err := decodeStrict(content, &parsed); err != nil {
		return LocalConfig{}, fmt.Errorf("%s: %w", path, err)
	}
	return parsed, nil
}

func decodeStrict(content []byte, target any) error {
	decoder := yaml.NewDecoder(bytes.NewReader(content))
	decoder.KnownFields(true)
	err := decoder.Decode(target)
	if errors.Is(err, io.EOF) {
		return nil
	}
	return err
}

func checkRepositories(repositories []Repository) error {
	seen := map[string]bool{}
	for index, repository := range repositories {
		if !repositoryName.MatchString(repository.Name) {
			return fmt.Errorf("repository %d: name %q must start with a letter or digit and use only letters, digits, '.', '_' and '-'", index+1, repository.Name)
		}
		if seen[repository.Name] {
			return fmt.Errorf("repository %q is declared twice", repository.Name)
		}
		seen[repository.Name] = true
		if repository.Remote == "" {
			return fmt.Errorf("repository %q: remote is empty", repository.Name)
		}
		if repository.Branch == "" {
			return fmt.Errorf("repository %q: branch is empty", repository.Name)
		}
	}
	return nil
}

func checkOverrides(local LocalConfig, declared declaration) error {
	names := make([]string, 0, len(local.Repositories.Overrides))
	for name := range local.Repositories.Overrides {
		names = append(names, name)
	}
	sort.Strings(names)
	declaredNames := map[string]bool{}
	for _, repository := range declared.Repositories {
		declaredNames[repository.Name] = true
	}
	for _, name := range names {
		if !declaredNames[name] {
			return fmt.Errorf("%s/%s: override for %q, which is not declared in %s", LocalDir, LocalConfigFile, name, DeclarationFile)
		}
		if !filepath.IsAbs(local.Repositories.Overrides[name]) {
			return fmt.Errorf("%s/%s: override for %q must be an absolute path", LocalDir, LocalConfigFile, name)
		}
	}
	return nil
}

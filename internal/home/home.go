package home

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	LocalDir        = ".homerepo"
	LocalConfigFile = "config.yaml"
)

type Home struct {
	Root         string
	Repositories []Repository
	Local        LocalConfig
}

func Load(root string) (Home, error) {
	declaration, err := readDeclaration(filepath.Join(root, DeclarationFile))
	if err != nil {
		return Home{}, err
	}
	local, err := readLocalConfig(filepath.Join(root, LocalDir, LocalConfigFile))
	if err != nil {
		return Home{}, err
	}
	if err := checkOverrides(local, declaration); err != nil {
		return Home{}, err
	}
	return Home{Root: root, Repositories: declaration.Repositories, Local: local}, nil
}

func (h Home) Repository(name string) (Repository, error) {
	for _, repository := range h.Repositories {
		if repository.Name == name {
			return repository, nil
		}
	}
	return Repository{}, fmt.Errorf("repository %q is not declared in %s; declared: %s", name, DeclarationFile, h.declaredNames())
}

func (h Home) ClonePath(name string) string {
	if override, ok := h.Local.Repositories.Overrides[name]; ok {
		return filepath.Clean(override)
	}
	return filepath.Join(filepath.Dir(h.Root), name)
}

func (h Home) declaredNames() string {
	if len(h.Repositories) == 0 {
		return "none"
	}
	names := make([]string, 0, len(h.Repositories))
	for _, repository := range h.Repositories {
		names = append(names, repository.Name)
	}
	sort.Strings(names)
	return strings.Join(names, ", ")
}

func EnsureIgnoredDir(path string) (bool, error) {
	info, err := os.Stat(path)
	if err == nil && info.IsDir() {
		return false, nil
	}
	if err == nil {
		return false, fmt.Errorf("%s exists and is not a folder", path)
	}
	if !errors.Is(err, fs.ErrNotExist) {
		return false, err
	}
	if err := os.MkdirAll(path, 0o755); err != nil {
		return false, err
	}
	return true, os.WriteFile(filepath.Join(path, ".gitignore"), []byte("*\n"), 0o644)
}

func IsDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

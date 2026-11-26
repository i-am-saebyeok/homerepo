package home

import (
	"fmt"
	"os"
	"path/filepath"
)

const DeclarationFile = "homerepo.yaml"

func Find(start string) (string, error) {
	absolute, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}
	for dir := absolute; ; dir = filepath.Dir(dir) {
		if isFile(filepath.Join(dir, DeclarationFile)) {
			return dir, nil
		}
		if filepath.Dir(dir) == dir {
			return "", fmt.Errorf("%s not found in %s or any parent folder; run `homerepo init` in your Home to create it", DeclarationFile, absolute)
		}
	}
}

func isFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.Mode().IsRegular()
}

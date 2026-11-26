package home_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/i-am-saebyeok/homerepo/internal/home"
)

func TestFindWalksUpToTheDeclaration(t *testing.T) {
	root := filepath.Join(t.TempDir(), "home")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, home.DeclarationFile), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	nested := filepath.Join(root, "a", "b")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	found, err := home.Find(nested)
	if err != nil {
		t.Fatal(err)
	}
	if found != root {
		t.Fatalf("found %s, want %s", found, root)
	}
}

func TestFindWithoutDeclarationExplainsInit(t *testing.T) {
	_, err := home.Find(t.TempDir())
	if err == nil {
		t.Fatal("expected an error")
	}
	if !strings.Contains(err.Error(), home.DeclarationFile) || !strings.Contains(err.Error(), "homerepo init") {
		t.Fatalf("error does not name the file and the fix: %v", err)
	}
}

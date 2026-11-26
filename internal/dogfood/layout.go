package dogfood

import "path/filepath"

const (
	markerDir      = ".dogfood"
	stateFile      = "run.json"
	homeDir        = "home"
	remotesDir     = "remotes"
	binDir         = "bin"
	homerepoBinary = "homerepo"
)

type Layout struct {
	Root string
}

func (l Layout) Marker() string {
	return filepath.Join(l.Root, markerDir)
}

func (l Layout) State() string {
	return filepath.Join(l.Marker(), stateFile)
}

func (l Layout) Home() string {
	return filepath.Join(l.Root, homeDir)
}

func (l Layout) Remote(repository string) string {
	return filepath.Join(l.Root, remotesDir, repository+".git")
}

func (l Layout) Bin() string {
	return filepath.Join(l.Root, binDir)
}

func (l Layout) Homerepo() string {
	return filepath.Join(l.Bin(), homerepoBinary)
}

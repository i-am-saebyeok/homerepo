package dogfood

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/i-am-saebyeok/homerepo/internal/home"
	"go.yaml.in/yaml/v3"
)

type Runtime string

const (
	Claude Runtime = "claude"
	Codex  Runtime = "codex"
)

type sessionSetup struct {
	layout      Layout
	runtime     Runtime
	searchPath  string
	runtimeArgs []string
}

func checkRuntime(name string) (Runtime, error) {
	runtime := Runtime(name)
	if runtime == Claude || runtime == Codex {
		return runtime, nil
	}
	return "", fmt.Errorf("unknown runtime %q; use claude or codex", name)
}

func (s sessionSetup) pathWithBuiltHomerepo() string {
	if s.searchPath == "" {
		return s.layout.Bin()
	}
	return s.layout.Bin() + string(os.PathListSeparator) + s.searchPath
}

func (s sessionSetup) launchArgs() []string {
	if s.runtime == Codex {
		return append(s.codexAccessArgs(), s.runtimeArgs...)
	}
	return append([]string{}, s.runtimeArgs...)
}

func (s sessionSetup) codexAccessArgs() []string {
	return []string{"--add-dir", s.layout.Root, "-c", "shell_environment_policy.set.PATH=" + tomlString(s.pathWithBuiltHomerepo())}
}

func tomlString(value string) string {
	quoted, _ := json.Marshal(value)
	return string(quoted)
}

func (s sessionSetup) write() error {
	if err := s.writeLocalConfig(); err != nil {
		return err
	}
	if s.runtime == Claude {
		return s.writeClaudeSettings()
	}
	return nil
}

func (s sessionSetup) writeLocalConfig() error {
	attach := home.Switch(false)
	runtimeConfig := home.RuntimeConfig{Attach: &attach}
	if args := s.launchArgs(); len(args) > 0 {
		runtimeConfig.Thread = map[string][]string{"default": args}
	}
	content, err := yaml.Marshal(map[string]any{"runtimes": map[Runtime]home.RuntimeConfig{s.runtime: runtimeConfig}})
	if err != nil {
		return err
	}
	return writeFile(filepath.Join(s.layout.Home(), home.LocalDir, home.LocalConfigFile), content)
}

func (s sessionSetup) writeClaudeSettings() error {
	settings := map[string]any{
		"permissions": map[string]any{"additionalDirectories": []string{s.layout.Root}},
		"env":         map[string]string{"PATH": s.pathWithBuiltHomerepo()},
	}
	content, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	return writeFile(filepath.Join(s.layout.Home(), ".claude", "settings.local.json"), append(content, '\n'))
}

func writeFile(path string, content []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, content, 0o644)
}

package thread

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/i-am-saebyeok/homerepo/internal/home"
)

const recordExtension = ".json"

type Record struct {
	Name      string    `json:"name"`
	Runtime   Runtime   `json:"runtime"`
	TabID     string    `json:"tab_id"`
	CreatedAt time.Time `json:"created_at"`
}

type Records struct {
	HomeRoot string
}

func (r Records) dir() string {
	return filepath.Join(r.HomeRoot, home.LocalDir, "threads")
}

func (r Records) path(name string) string {
	return filepath.Join(r.dir(), filepath.FromSlash(name)+recordExtension)
}

func (r Records) Write(record Record) error {
	if _, err := home.EnsureIgnoredDir(filepath.Join(r.HomeRoot, home.LocalDir)); err != nil {
		return err
	}
	path := r.path(record.Name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	content, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(content, '\n'), 0o644)
}

func (r Records) Read(name string) (Record, error) {
	content, err := os.ReadFile(r.path(name))
	if errors.Is(err, fs.ErrNotExist) {
		return Record{}, fmt.Errorf("no thread named %q", name)
	}
	if err != nil {
		return Record{}, err
	}
	var record Record
	if err := json.Unmarshal(content, &record); err != nil {
		return Record{}, fmt.Errorf("%s: %w", r.path(name), err)
	}
	return record, nil
}

func (r Records) List() ([]Record, error) {
	var records []Record
	err := filepath.WalkDir(r.dir(), func(path string, entry fs.DirEntry, err error) error {
		if errors.Is(err, fs.ErrNotExist) {
			return fs.SkipAll
		}
		if err != nil || entry.IsDir() || !strings.HasSuffix(path, recordExtension) {
			return err
		}
		relative, err := filepath.Rel(r.dir(), path)
		if err != nil {
			return err
		}
		record, err := r.Read(filepath.ToSlash(strings.TrimSuffix(relative, recordExtension)))
		records = append(records, record)
		return err
	})
	sort.Slice(records, func(i, j int) bool { return records[i].Name < records[j].Name })
	return records, err
}

func (r Records) Delete(name string) error {
	return os.Remove(r.path(name))
}

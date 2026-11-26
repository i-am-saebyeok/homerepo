package home

import (
	_ "embed"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

const (
	skillName           = "homerepo-thread"
	agentsSkillsDir     = ".agents/skills"
	claudeSkillsDir     = ".claude/skills"
	declarationSkeleton = "repositories: []\n"
)

//go:embed skill/SKILL.md
var skillContent []byte

type InitResult struct {
	Created []string
	Skipped []string
}

func (r *InitResult) record(path string, created bool) {
	if created {
		r.Created = append(r.Created, path)
		return
	}
	r.Skipped = append(r.Skipped, path)
}

func Init(root string) (InitResult, error) {
	var result InitResult
	created, err := writeIfAbsent(filepath.Join(root, DeclarationFile), []byte(declarationSkeleton))
	if err != nil {
		return result, err
	}
	result.record(DeclarationFile, created)

	created, err = EnsureIgnoredDir(filepath.Join(root, LocalDir))
	if err != nil {
		return result, err
	}
	result.record(LocalDir+"/", created)

	skillFile := filepath.Join(agentsSkillsDir, skillName, "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(filepath.Join(root, skillFile)), 0o755); err != nil {
		return result, err
	}
	created, err = writeIfAbsent(filepath.Join(root, skillFile), skillContent)
	if err != nil {
		return result, err
	}
	result.record(filepath.ToSlash(skillFile), created)

	if sameDir(filepath.Join(root, claudeSkillsDir), filepath.Join(root, agentsSkillsDir)) {
		return result, nil
	}
	link := filepath.Join(claudeSkillsDir, skillName)
	created, err = linkIfAbsent(filepath.Join(root, link), filepath.Join("..", "..", agentsSkillsDir, skillName))
	if err != nil {
		return result, err
	}
	result.record(filepath.ToSlash(link), created)
	return result, nil
}

func writeIfAbsent(path string, content []byte) (bool, error) {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if errors.Is(err, fs.ErrExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if _, err := file.Write(content); err != nil {
		file.Close()
		return false, err
	}
	return true, file.Close()
}

func linkIfAbsent(link string, target string) (bool, error) {
	if _, err := os.Lstat(link); err == nil {
		return false, nil
	}
	if err := os.MkdirAll(filepath.Dir(link), 0o755); err != nil {
		return false, err
	}
	return true, os.Symlink(target, link)
}

func sameDir(first string, second string) bool {
	resolvedFirst, err := filepath.EvalSymlinks(first)
	if err != nil {
		return false
	}
	resolvedSecond, err := filepath.EvalSymlinks(second)
	if err != nil {
		return false
	}
	return resolvedFirst == resolvedSecond
}

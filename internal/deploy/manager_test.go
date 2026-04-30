package deploy

import (
	"os"
	"path/filepath"
	"testing"

	"localskillmanager/internal/discovery"
)

func TestManagerApplyInstallsAndUninstallsSkill(t *testing.T) {
	tmp := t.TempDir()
	store := filepath.Join(tmp, "store")
	targetRoot := filepath.Join(tmp, ".claude")
	skillPath := filepath.Join(store, "alpha")
	mustMkdirAll(t, skillPath)
	mustWriteFile(t, filepath.Join(skillPath, "SKILL.md"), "content")
	mustWriteFile(t, filepath.Join(skillPath, "script.sh"), "echo hi")

	manager := NewManager([]ResolvedTarget{{
		Name:       "claude",
		RootPath:   targetRoot,
		DeployPath: filepath.Join(targetRoot, "skills"),
	}})
	skill := discovery.Skill{Name: "alpha", SourcePath: skillPath}

	install := manager.Apply(skill, map[string]bool{targetRoot: true})
	if len(install) != 1 || install[0].Err != nil {
		t.Fatalf("install results = %+v", install)
	}

	linkPath := filepath.Join(targetRoot, "skills", "alpha")
	targetPath, err := os.Readlink(linkPath)
	if err != nil {
		t.Fatal(err)
	}
	if targetPath != skillPath {
		t.Fatalf("symlink target = %q, want %q", targetPath, skillPath)
	}

	uninstall := manager.Apply(skill, map[string]bool{targetRoot: false})
	if len(uninstall) != 1 || uninstall[0].Err != nil {
		t.Fatalf("uninstall results = %+v", uninstall)
	}

	if _, err := os.Stat(filepath.Join(targetRoot, "skills", "alpha")); !os.IsNotExist(err) {
		t.Fatalf("expected skill to be removed, stat err = %v", err)
	}
}

func TestManagerApplyReplacesExistingSymlink(t *testing.T) {
	tmp := t.TempDir()
	store := filepath.Join(tmp, "store")
	targetRoot := filepath.Join(tmp, ".claude")
	firstSkillPath := filepath.Join(store, "alpha-v1")
	secondSkillPath := filepath.Join(store, "alpha-v2")
	mustMkdirAll(t, firstSkillPath)
	mustMkdirAll(t, secondSkillPath)
	mustWriteFile(t, filepath.Join(firstSkillPath, "SKILL.md"), "content")
	mustWriteFile(t, filepath.Join(secondSkillPath, "SKILL.md"), "content")
	mustMkdirAll(t, filepath.Join(targetRoot, "skills"))
	if err := os.Symlink(firstSkillPath, filepath.Join(targetRoot, "skills", "alpha")); err != nil {
		t.Fatal(err)
	}

	manager := NewManager([]ResolvedTarget{{
		Name:       "claude",
		RootPath:   targetRoot,
		DeployPath: filepath.Join(targetRoot, "skills"),
	}})
	skill := discovery.Skill{Name: "alpha", SourcePath: secondSkillPath}

	results := manager.Apply(skill, map[string]bool{targetRoot: true})
	if len(results) != 1 || results[0].Err != nil {
		t.Fatalf("results = %+v", results)
	}

	targetPath, err := os.Readlink(filepath.Join(targetRoot, "skills", "alpha"))
	if err != nil {
		t.Fatal(err)
	}
	if targetPath != secondSkillPath {
		t.Fatalf("symlink target = %q, want %q", targetPath, secondSkillPath)
	}
}

func mustMkdirAll(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
}

func mustWriteFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

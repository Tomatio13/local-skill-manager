package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadNormalizesPathsAndTargets(t *testing.T) {
	tmp := t.TempDir()
	home := filepath.Join(tmp, "home")
	if err := os.MkdirAll(filepath.Join(home, ".open_skills", "skills"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", home)

	configPath := filepath.Join(tmp, "config.json")
	data := []byte(`{
		"store_path": "~/.open_skills",
		"link_targets": ["~/.claude", "~/.claude", "~/.agents"]
	}`)
	if err := os.WriteFile(configPath, data, 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatal(err)
	}

	if got, want := cfg.StoreSkillsPath, filepath.Join(home, ".open_skills", "skills"); got != want {
		t.Fatalf("StoreSkillsPath = %q, want %q", got, want)
	}

	if got, want := len(cfg.LinkTargets), 2; got != want {
		t.Fatalf("len(LinkTargets) = %d, want %d", got, want)
	}
}

func TestLoadRequiresStoreSkillsDirectory(t *testing.T) {
	tmp := t.TempDir()
	home := filepath.Join(tmp, "home")
	if err := os.MkdirAll(home, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", home)

	configPath := filepath.Join(tmp, "config.json")
	data := []byte(`{
		"store_path": "~/.open_skills",
		"link_targets": ["~/.claude"]
	}`)
	if err := os.WriteFile(configPath, data, 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := Load(configPath); err == nil {
		t.Fatal("expected error")
	}
}

package discovery

import (
	"os"
	"path/filepath"
	"testing"
)

func TestListSkillsFiltersBySKILLMarkdown(t *testing.T) {
	tmp := t.TempDir()
	mustMkdirAll(t, filepath.Join(tmp, "alpha"))
	mustWriteFile(t, filepath.Join(tmp, "alpha", "SKILL.md"), "---\ndescription: Alpha skill\n---\n")
	mustMkdirAll(t, filepath.Join(tmp, "beta"))
	mustWriteFile(t, filepath.Join(tmp, "note.txt"), "test")

	skills, err := ListSkills(tmp)
	if err != nil {
		t.Fatal(err)
	}

	if got, want := len(skills), 1; got != want {
		t.Fatalf("len(skills) = %d, want %d", got, want)
	}
	if got, want := skills[0].Name, "alpha"; got != want {
		t.Fatalf("Name = %q, want %q", got, want)
	}
	if got, want := skills[0].Description, "Alpha skill"; got != want {
		t.Fatalf("Description = %q, want %q", got, want)
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

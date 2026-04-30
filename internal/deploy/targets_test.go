package deploy

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveTargetUsesKnownDirectory(t *testing.T) {
	tmp := t.TempDir()
	root := filepath.Join(tmp, ".cursor")
	if err := os.MkdirAll(filepath.Join(root, "skills-cursor"), 0o755); err != nil {
		t.Fatal(err)
	}

	target, err := ResolveTarget(root)
	if err != nil {
		t.Fatal(err)
	}

	if got, want := target.DeployPath, filepath.Join(root, "skills-cursor"); got != want {
		t.Fatalf("DeployPath = %q, want %q", got, want)
	}
}

func TestResolveTargetFallsBackToSkills(t *testing.T) {
	tmp := t.TempDir()
	root := filepath.Join(tmp, ".gemini")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}

	target, err := ResolveTarget(root)
	if err != nil {
		t.Fatal(err)
	}

	if got, want := target.DeployPath, filepath.Join(root, "skills"); got != want {
		t.Fatalf("DeployPath = %q, want %q", got, want)
	}
}

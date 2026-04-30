package deploy

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type ResolvedTarget struct {
	Name       string
	RootPath   string
	DeployPath string
}

func ResolveTargets(targetRoots []string) ([]ResolvedTarget, error) {
	resolved := make([]ResolvedTarget, 0, len(targetRoots))
	for _, root := range targetRoots {
		target, err := ResolveTarget(root)
		if err != nil {
			return nil, err
		}
		resolved = append(resolved, target)
	}
	return resolved, nil
}

func ResolveTarget(root string) (ResolvedTarget, error) {
	if strings.TrimSpace(root) == "" {
		return ResolvedTarget{}, fmt.Errorf("empty target root")
	}

	deployDir := "skills"
	switch {
	case strings.HasSuffix(root, filepath.Clean("/.claude")):
		deployDir = "skills"
	case strings.HasSuffix(root, filepath.Clean("/.agents")):
		deployDir = "skills"
	case strings.HasSuffix(root, filepath.Clean("/.config/opencode")):
		deployDir = "skills"
	case strings.HasSuffix(root, filepath.Clean("/.cursor")):
		if directoryExists(filepath.Join(root, "skills-cursor")) {
			deployDir = "skills-cursor"
		}
	case strings.HasSuffix(root, filepath.Clean("/.codeium/windsurf")):
		deployDir = "skills"
	}

	name := filepath.Base(root)
	if name == "." || name == string(filepath.Separator) {
		name = root
	}
	return ResolvedTarget{
		Name:       name,
		RootPath:   root,
		DeployPath: filepath.Join(root, deployDir),
	}, nil
}

func directoryExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

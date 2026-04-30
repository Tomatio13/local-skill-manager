package deploy

import (
	"fmt"
	"os"
	"path/filepath"

	"localskillmanager/internal/discovery"
)

type Manager struct {
	targets []ResolvedTarget
}

type ApplyResult struct {
	Target  ResolvedTarget
	Enabled bool
	Err     error
}

func NewManager(targets []ResolvedTarget) *Manager {
	return &Manager{targets: targets}
}

func (m *Manager) Targets() []ResolvedTarget {
	out := make([]ResolvedTarget, len(m.targets))
	copy(out, m.targets)
	return out
}

func (m *Manager) CurrentState(skill discovery.Skill) map[string]bool {
	state := make(map[string]bool, len(m.targets))
	for _, target := range m.targets {
		state[target.RootPath] = directoryExists(filepath.Join(target.DeployPath, skill.Name))
	}
	return state
}

func (m *Manager) Apply(skill discovery.Skill, desired map[string]bool) []ApplyResult {
	results := make([]ApplyResult, 0, len(m.targets))
	for _, target := range m.targets {
		wantEnabled, ok := desired[target.RootPath]
		if !ok {
			continue
		}
		var err error
		if wantEnabled {
			err = installSkill(skill, target)
		} else {
			err = uninstallSkill(skill, target)
		}
		results = append(results, ApplyResult{
			Target:  target,
			Enabled: wantEnabled,
			Err:     err,
		})
	}
	return results
}

func installSkill(skill discovery.Skill, target ResolvedTarget) error {
	if err := os.MkdirAll(target.DeployPath, 0o755); err != nil {
		return fmt.Errorf("create deploy dir %q: %w", target.DeployPath, err)
	}

	finalPath := filepath.Join(target.DeployPath, skill.Name)
	if pathExists(finalPath) {
		if err := os.RemoveAll(finalPath); err != nil {
			return fmt.Errorf("remove existing skill %q: %w", finalPath, err)
		}
	}

	if err := os.Symlink(skill.SourcePath, finalPath); err != nil {
		return fmt.Errorf("create symlink %q -> %q: %w", finalPath, skill.SourcePath, err)
	}
	return nil
}

func uninstallSkill(skill discovery.Skill, target ResolvedTarget) error {
	finalPath := filepath.Join(target.DeployPath, skill.Name)
	if !directoryExists(finalPath) {
		return nil
	}
	if err := os.RemoveAll(finalPath); err != nil {
		return fmt.Errorf("remove skill %q: %w", finalPath, err)
	}
	return nil
}

func pathExists(path string) bool {
	_, err := os.Lstat(path)
	return err == nil
}

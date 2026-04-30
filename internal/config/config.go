package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const defaultConfigPath = "./config.json"

type Config struct {
	StorePath       string   `json:"store_path"`
	LinkTargets     []string `json:"link_targets"`
	StoreSkillsPath string   `json:"-"`
	ConfigPath      string   `json:"-"`
}

func Load(path string) (Config, error) {
	resolvedPath, err := expandPath(path)
	if err != nil {
		return Config{}, fmt.Errorf("expand config path: %w", err)
	}

	data, err := os.ReadFile(resolvedPath)
	if err != nil {
		if path == "" || path == defaultConfigPath {
			return Config{}, fmt.Errorf("read config %q: %w", resolvedPath, err)
		}
		return Config{}, fmt.Errorf("read config %q: %w", resolvedPath, err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config %q: %w", resolvedPath, err)
	}

	if strings.TrimSpace(cfg.StorePath) == "" {
		return Config{}, fmt.Errorf("config missing store_path")
	}

	storePath, err := expandPath(cfg.StorePath)
	if err != nil {
		return Config{}, fmt.Errorf("expand store_path: %w", err)
	}

	storeSkillsPath := filepath.Join(storePath, "skills")
	info, err := os.Stat(storeSkillsPath)
	if err != nil {
		return Config{}, fmt.Errorf("store skills path %q: %w", storeSkillsPath, err)
	}
	if !info.IsDir() {
		return Config{}, fmt.Errorf("store skills path %q is not a directory", storeSkillsPath)
	}

	targets, err := normalizeTargets(cfg.LinkTargets)
	if err != nil {
		return Config{}, err
	}

	cfg.ConfigPath = resolvedPath
	cfg.StorePath = storePath
	cfg.StoreSkillsPath = storeSkillsPath
	cfg.LinkTargets = targets
	return cfg, nil
}

func normalizeTargets(targets []string) ([]string, error) {
	seen := make(map[string]struct{}, len(targets))
	result := make([]string, 0, len(targets))
	for _, target := range targets {
		if strings.TrimSpace(target) == "" {
			continue
		}

		resolved, err := expandPath(target)
		if err != nil {
			return nil, fmt.Errorf("expand link target %q: %w", target, err)
		}

		if _, ok := seen[resolved]; ok {
			continue
		}
		seen[resolved] = struct{}{}
		result = append(result, resolved)
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("config missing link_targets")
	}

	return result, nil
}

func expandPath(path string) (string, error) {
	if path == "" {
		path = defaultConfigPath
	}
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		if path == "~" {
			path = home
		} else if strings.HasPrefix(path, "~/") {
			path = filepath.Join(home, path[2:])
		}
	}

	return filepath.Abs(path)
}

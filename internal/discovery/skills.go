package discovery

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Skill struct {
	Name        string
	Description string
	SourcePath  string
}

func ListSkills(storeSkillsPath string) ([]Skill, error) {
	entries, err := os.ReadDir(storeSkillsPath)
	if err != nil {
		return nil, fmt.Errorf("read store skills path %q: %w", storeSkillsPath, err)
	}

	skills := make([]Skill, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		skillPath := filepath.Join(storeSkillsPath, entry.Name())
		skillFilePath := filepath.Join(skillPath, "SKILL.md")
		if _, err := os.Stat(skillFilePath); err != nil {
			continue
		}
		description, err := readDescription(skillFilePath)
		if err != nil {
			return nil, fmt.Errorf("read skill description %q: %w", skillFilePath, err)
		}
		skills = append(skills, Skill{
			Name:        entry.Name(),
			Description: description,
			SourcePath:  skillPath,
		})
	}

	sort.Slice(skills, func(i, j int) bool {
		return strings.ToLower(skills[i].Name) < strings.ToLower(skills[j].Name)
	})
	return skills, nil
}

func readDescription(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(data), "\n")
	inFrontmatter := false
	frontmatterSeen := false

	for _, rawLine := range lines {
		line := strings.TrimSpace(rawLine)
		if line == "---" {
			if !frontmatterSeen {
				inFrontmatter = true
				frontmatterSeen = true
				continue
			}
			if inFrontmatter {
				inFrontmatter = false
				continue
			}
		}
		if inFrontmatter && strings.HasPrefix(line, "description:") {
			value := strings.TrimSpace(strings.TrimPrefix(line, "description:"))
			value = strings.Trim(value, `"'`)
			return value, nil
		}
	}

	for _, rawLine := range lines {
		line := strings.TrimSpace(rawLine)
		if line == "" || line == "---" || strings.HasPrefix(line, "name:") || strings.HasPrefix(line, "description:") {
			continue
		}
		if strings.HasPrefix(line, "#") {
			line = strings.TrimSpace(strings.TrimLeft(line, "#"))
		}
		if line != "" {
			return line, nil
		}
	}

	return "", nil
}

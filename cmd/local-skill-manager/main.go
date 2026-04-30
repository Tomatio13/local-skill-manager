package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"localskillmanager/internal/config"
	"localskillmanager/internal/deploy"
	"localskillmanager/internal/discovery"
	"localskillmanager/internal/tui"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "", "path to config JSON")
	flag.Parse()

	cfg, err := config.Load(configPath)
	if err != nil {
		exitWithError(err)
	}

	targets, err := deploy.ResolveTargets(cfg.LinkTargets)
	if err != nil {
		exitWithError(err)
	}

	skills, err := discovery.ListSkills(cfg.StoreSkillsPath)
	if err != nil {
		exitWithError(err)
	}

	manager := deploy.NewManager(targets)
	model, err := tui.NewModel(cfg, skills, manager)
	if err != nil {
		exitWithError(err)
	}

	program := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := program.Run(); err != nil {
		exitWithError(err)
	}
}

func exitWithError(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

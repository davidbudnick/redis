package main

import (
	"log"
	"log/slog"
	"os"
	"path/filepath"

	"redis/internal/cmd"
	"redis/internal/db"
	"redis/internal/types"
	"redis/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	var logs []string
	logWriter := types.LogWriter{Logs: &logs}
	handler := slog.NewJSONHandler(logWriter, nil)
	slog.SetDefault(slog.New(handler))

	config, err := initConfig()
	if err != nil {
		log.Fatalf("Failed to initialize config: %v", err)
	}
	defer config.Close()

	cmd.Config = config

	m := ui.NewModel()
	m.Logs = &logs

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

func initConfig() (*db.Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = os.TempDir()
	}

	configDir := filepath.Join(homeDir, ".redis")
	if err := os.MkdirAll(configDir, 0750); err != nil {
		return nil, err
	}

	return db.NewConfig(filepath.Join(configDir, "config.json"))
}

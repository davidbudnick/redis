package main

import (
	"log"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/davidbudnick/redis/internal/cmd"
	"github.com/davidbudnick/redis/internal/db"
	"github.com/davidbudnick/redis/internal/types"
	"github.com/davidbudnick/redis/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Minimal setup before starting UI
	var logs []string

	// Start the UI immediately for perceived speed
	m := ui.NewModel()
	m.Logs = &logs

	sendFunc := func(msg tea.Msg) {}
	m.SendFunc = &sendFunc

	// Initialize logger in background (non-blocking)
	logWriter := types.LogWriter{Logs: &logs}
	handler := slog.NewJSONHandler(logWriter, nil)
	slog.SetDefault(slog.New(handler))

	// Load config (this happens quickly, but after UI is ready)
	config, err := initConfig()
	if err != nil {
		log.Fatalf("Failed to initialize config: %v", err)
	}
	defer config.Close()
	cmd.Config = config

	p := tea.NewProgram(m, tea.WithAltScreen())
	*m.SendFunc = p.Send
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

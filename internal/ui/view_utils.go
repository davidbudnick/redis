package ui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/davidbudnick/redis/internal/types"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).MarginBottom(1)
	headerStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15"))
	normalStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("15"))
	selectedStyle = lipgloss.NewStyle().Bold(true).Background(lipgloss.Color("39")).Foreground(lipgloss.Color("0"))
	keyStyle      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
	descStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("15"))
	errorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	successStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	dimStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	helpStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)

func getTypeColor(keyType types.KeyType) lipgloss.Color {
	switch keyType {
	case types.KeyTypeString:
		return lipgloss.Color("2") // Green
	case types.KeyTypeList:
		return lipgloss.Color("3") // Yellow
	case types.KeyTypeSet:
		return lipgloss.Color("4") // Blue
	case types.KeyTypeZSet:
		return lipgloss.Color("5") // Magenta
	case types.KeyTypeHash:
		return lipgloss.Color("6") // Cyan
	case types.KeyTypeStream:
		return lipgloss.Color("13") // Bright Magenta
	default:
		return lipgloss.Color("15") // White
	}
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func formatPossibleJSON(s string) string {
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return s
	}

	// Check if it looks like JSON (starts with { or [)
	if s[0] == '{' || s[0] == '[' {
		var prettyJSON bytes.Buffer
		if err := json.Indent(&prettyJSON, []byte(s), "", "  "); err == nil {
			return prettyJSON.String()
		}
	}
	return s
}

func (m Model) renderModal(content string) string {
	modalWidth := 60
	if m.Width-10 < 60 {
		modalWidth = m.Width - 10
	}
	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("39")).
		Padding(1, 2).
		Width(modalWidth)

	return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, modalStyle.Render(content))
}

func (m Model) renderModalWide(content string) string {
	modalWidth := 90
	if m.Width-10 < 90 {
		modalWidth = m.Width - 10
	}
	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("39")).
		Padding(1, 2).
		Width(modalWidth)

	return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, modalStyle.Render(content))
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

type logEntry struct {
	Time  string
	Level string
	Msg   string
}

func parseLogEntry(logLine string) logEntry {
	entry := logEntry{}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(logLine), &data); err != nil {
		entry.Msg = logLine
		return entry
	}

	if t, ok := data["time"].(string); ok {
		if parsed, err := time.Parse(time.RFC3339Nano, t); err == nil {
			entry.Time = parsed.Format("15:04:05")
		} else {
			entry.Time = t
		}
	}
	if l, ok := data["level"].(string); ok {
		entry.Level = strings.ToUpper(l)
	}
	if m, ok := data["msg"].(string); ok {
		entry.Msg = m
	}

	return entry
}
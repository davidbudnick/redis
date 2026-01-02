package ui

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) viewLogs() string {
	var b strings.Builder

	logCount := 0
	if m.Logs != nil {
		logCount = len(*m.Logs)
	}

	b.WriteString(titleStyle.Render(fmt.Sprintf("Application Logs (%d entries)", logCount)))
	b.WriteString("\n\n")

	if m.Logs == nil || len(*m.Logs) == 0 {
		b.WriteString(dimStyle.Render("No logs yet. Logs will appear as you use the app."))
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render("esc:close"))
		return m.renderModalWide(b.String())
	}

	// If showing detail view for a log entry
	if m.ShowingLogDetail && m.LogCursor < len(*m.Logs) {
		return m.viewLogDetail((*m.Logs)[m.LogCursor])
	}

	// Calculate visible window
	maxVisible := m.Height - 12
	if maxVisible < 5 {
		maxVisible = 5
	}
	startIdx := 0
	if m.LogCursor >= maxVisible {
		startIdx = m.LogCursor - maxVisible + 1
	}
	endIdx := startIdx + maxVisible
	if endIdx > len(*m.Logs) {
		endIdx = len(*m.Logs)
	}

	// Header
	b.WriteString(headerStyle.Render(fmt.Sprintf("%-10s  %-6s  %s", "Time", "Level", "Message")))
	b.WriteString("\n")
	b.WriteString(dimStyle.Render(strings.Repeat("─", 80)))
	b.WriteString("\n")

	for i := startIdx; i < endIdx; i++ {
		logLine := (*m.Logs)[i]
		entry := parseLogEntry(logLine)

		// Format level with color
		var levelStyled string
		switch entry.Level {
		case "ERROR":
			levelStyled = errorStyle.Render(fmt.Sprintf("%-6s", entry.Level))
		case "WARN":
			levelStyled = lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Render(fmt.Sprintf("%-6s", entry.Level))
		case "INFO":
			levelStyled = lipgloss.NewStyle().Foreground(lipgloss.Color("4")).Render(fmt.Sprintf("%-6s", entry.Level))
		default:
			levelStyled = dimStyle.Render(fmt.Sprintf("%-6s", entry.Level))
		}

		msg := entry.Msg
		if len(msg) > 55 {
			msg = msg[:52] + "..."
		}

		timeStr := entry.Time
		if timeStr == "" {
			timeStr = "          "
		}

		if i == m.LogCursor {
			plainLine := fmt.Sprintf("%-10s  %-6s  %s", timeStr, entry.Level, msg)
			b.WriteString(selectedStyle.Render(plainLine))
		} else {
			b.WriteString(fmt.Sprintf("%-10s  %s  %s", timeStr, levelStyled, normalStyle.Render(msg)))
		}
		b.WriteString("\n")
	}

	if len(*m.Logs) > maxVisible {
		b.WriteString("\n")
		b.WriteString(dimStyle.Render(fmt.Sprintf("Showing %d-%d of %d", startIdx+1, endIdx, len(*m.Logs))))
	}

	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("j/k:navigate  enter:details  esc:close"))

	return m.renderModalWide(b.String())
}

func (m Model) viewLogDetail(logLine string) string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Log Entry Details"))
	b.WriteString("\n\n")

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(logLine), &data); err != nil {
		b.WriteString(normalStyle.Render(logLine))
	} else {
		prettyJSON, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			b.WriteString(normalStyle.Render(logLine))
		} else {
			b.WriteString(normalStyle.Render(string(prettyJSON)))
		}
	}

	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("enter/esc:back"))

	return m.renderModalWide(b.String())
}

package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) viewHelp() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Help"))
	b.WriteString("\n\n")

	sections := []struct {
		title    string
		bindings [][2]string
	}{
		{
			title: "Global",
			bindings: [][2]string{
				{"q", "Quit / Go back"},
				{"?", "Show help"},
				{"j/k", "Navigate up/down"},
				{"PgUp/PgDn", "Page up/down"},
				{"g/G", "Top/Bottom"},
			},
		},
		{
			title: "Connections",
			bindings: [][2]string{
				{"a", "Add connection"},
				{"e", "Edit connection"},
				{"d", "Delete connection"},
				{"Ctrl+T", "Test connection"},
			},
		},
		{
			title: "Keys",
			bindings: [][2]string{
				{"enter", "View key detail"},
				{"a", "Add key"},
				{"d", "Delete key"},
				{"r", "Refresh keys"},
				{"l", "Load more keys"},
				{"/", "Filter by pattern"},
				{"s/S", "Sort / Toggle direction"},
				{"v", "Search by value"},
				{"e", "Export to JSON"},
				{"I", "Import from JSON"},
				{"D", "Switch database"},
				{"p", "Pub/Sub publish"},
				{"L", "View slow log"},
				{"E", "Execute Lua script"},
				{"O", "View application logs"},
				{"i", "Server info"},
				{"f", "Flush database"},
			},
		},
		{
			title: "Key Detail",
			bindings: [][2]string{
				{"e", "Edit value (string)"},
				{"a", "Add to collection"},
				{"x", "Remove from collection"},
				{"R", "Rename key"},
				{"c", "Copy key"},
				{"t", "Set TTL"},
			},
		},
	}

	for _, section := range sections {
		b.WriteString(keyStyle.Render(section.title))
		b.WriteString("\n")
		for _, binding := range section.bindings {
			b.WriteString(fmt.Sprintf("  %-10s %s\n", dimStyle.Render(binding[0]), descStyle.Render(binding[1])))
		}
		b.WriteString("\n")
	}

	b.WriteString(helpStyle.Render("Press ? or esc to close"))

	modalWidth := 50
	if m.Width-10 < 50 {
		modalWidth = m.Width - 10
	}
	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("39")).
		Padding(1, 2).
		Width(modalWidth)

	return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, modalStyle.Render(b.String()))
}

func (m Model) viewTestConnection() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Test Connection"))
	b.WriteString("\n\n")

	if m.Loading {
		b.WriteString(dimStyle.Render("Testing connection..."))
	} else if m.TestConnResult != "" {
		if strings.HasPrefix(m.TestConnResult, "Failed") {
			b.WriteString(errorStyle.Render(m.TestConnResult))
		} else {
			b.WriteString(successStyle.Render(m.TestConnResult))
		}
	}

	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("esc:back"))

	return m.renderModal(b.String())
}

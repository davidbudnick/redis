package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) viewConnections() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Redis Connections"))
	b.WriteString("\n\n")

	if len(m.Connections) == 0 {
		b.WriteString(dimStyle.Render("No connections saved."))
		b.WriteString("\n")
		b.WriteString(dimStyle.Render("Press 'a' to add one."))
	} else {
		header := fmt.Sprintf("  %-20s %-25s %-8s %-4s", "Name", "Host", "Port", "DB")
		b.WriteString(headerStyle.Render(header))
		b.WriteString("\n")
		b.WriteString(dimStyle.Render(strings.Repeat("─", 65)))
		b.WriteString("\n")

		for i, conn := range m.Connections {
			name := conn.Name
			if len(name) > 20 {
				name = name[:17] + "..."
			}
			host := conn.Host
			if len(host) > 25 {
				host = host[:22] + "..."
			}

			line := fmt.Sprintf("%-20s %-25s %-8d %-4d", name, host, conn.Port, conn.DB)
			if i == m.SelectedConnIdx {
				b.WriteString(selectedStyle.Render("▶ " + line))
			} else {
				b.WriteString(normalStyle.Render("  " + line))
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("j/k:navigate  enter:connect  a:add  e:edit  d:delete  r:refresh  q:quit"))

	return b.String()
}

func (m Model) viewAddConnection() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Add Connection"))
	b.WriteString("\n\n")

	labels := []string{"Name", "Host", "Port", "Password", "Database"}
	for i, input := range m.ConnInputs {
		b.WriteString(keyStyle.Render(labels[i] + ":"))
		b.WriteString("\n")
		b.WriteString(input.View())
		b.WriteString("\n\n")
	}

	b.WriteString(helpStyle.Render("tab:next  Ctrl+T:test  enter:save  esc:cancel"))

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

func (m Model) viewEditConnection() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Edit Connection"))
	b.WriteString("\n\n")

	labels := []string{"Name", "Host", "Port", "Password", "Database"}
	for i, input := range m.ConnInputs {
		b.WriteString(keyStyle.Render(labels[i] + ":"))
		b.WriteString("\n")
		b.WriteString(input.View())
		b.WriteString("\n\n")
	}

	b.WriteString(helpStyle.Render("tab:next field  enter:save  esc:cancel"))

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
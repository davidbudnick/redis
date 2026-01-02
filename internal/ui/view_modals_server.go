package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/davidbudnick/redis/internal/types"
)

func (m Model) viewConfirmDelete() string {
	var b strings.Builder

	warningStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true)

	b.WriteString(warningStyle.Render("Confirm Delete"))
	b.WriteString("\n\n")

	switch m.ConfirmType {
	case "connection":
		if conn, ok := m.ConfirmData.(types.Connection); ok {
			b.WriteString(normalStyle.Render(fmt.Sprintf("Delete connection '%s'?", conn.Name)))
		}
	case "key":
		if key, ok := m.ConfirmData.(types.RedisKey); ok {
			b.WriteString(normalStyle.Render(fmt.Sprintf("Delete key '%s'?", key.Key)))
		}
	case "flushdb":
		b.WriteString(warningStyle.Render("FLUSH entire database?"))
		b.WriteString("\n")
		b.WriteString(warningStyle.Render("This will delete ALL keys!"))
	}

	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("[y] confirm  [n/esc] cancel"))

	borderColor := lipgloss.Color("1")
	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(1, 2).
		Width(50)

	return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, modalStyle.Render(b.String()))
}

func (m Model) viewServerInfo() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Server Info"))
	b.WriteString("\n\n")

	info := []struct {
		label string
		value string
	}{
		{"Version", m.ServerInfo.Version},
		{"Mode", m.ServerInfo.Mode},
		{"OS", m.ServerInfo.OS},
		{"Memory", m.ServerInfo.UsedMemory},
		{"Clients", m.ServerInfo.Clients},
		{"Keys", m.ServerInfo.TotalKeys},
		{"Uptime", m.ServerInfo.Uptime},
	}

	for _, item := range info {
		if item.value != "" {
			b.WriteString(keyStyle.Render(fmt.Sprintf("%-12s", item.label+":")))
			b.WriteString(normalStyle.Render(item.value))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("r:refresh  esc:back"))

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

func (m Model) viewPubSub() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Publish Message"))
	b.WriteString("\n\n")

	b.WriteString(keyStyle.Render("Channel:"))
	b.WriteString("\n")
	b.WriteString(m.PubSubInput[0].View())
	b.WriteString("\n\n")

	b.WriteString(keyStyle.Render("Message:"))
	b.WriteString("\n")
	b.WriteString(m.PubSubInput[1].View())
	b.WriteString("\n\n")

	b.WriteString(helpStyle.Render("tab:next  enter:publish  esc:cancel"))

	return m.renderModal(b.String())
}

func (m Model) viewSwitchDB() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Switch Database"))
	b.WriteString("\n\n")

	if m.CurrentConn != nil {
		b.WriteString(keyStyle.Render("Current: "))
		b.WriteString(normalStyle.Render(fmt.Sprintf("db%d", m.CurrentConn.DB)))
		b.WriteString("\n\n")
	}

	b.WriteString(keyStyle.Render("New Database (0-15):"))
	b.WriteString("\n")
	b.WriteString(m.DBSwitchInput.View())
	b.WriteString("\n\n")

	b.WriteString(helpStyle.Render("enter:switch  esc:cancel"))

	return m.renderModal(b.String())
}

func (m Model) viewSlowLog() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Slow Log"))
	b.WriteString("\n\n")

	if len(m.SlowLogEntries) == 0 {
		b.WriteString(dimStyle.Render("No slow log entries found."))
	} else {
		for _, entry := range m.SlowLogEntries {
			b.WriteString(keyStyle.Render(fmt.Sprintf("#%d ", entry.ID)))
			b.WriteString(dimStyle.Render(entry.Duration.String()))
			b.WriteString("\n")
			b.WriteString(normalStyle.Render("  " + truncate(entry.Command, 60)))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("r:refresh  esc:back"))

	return m.renderModal(b.String())
}

func (m Model) viewLuaScript() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Execute Lua Script"))
	b.WriteString("\n\n")

	b.WriteString(keyStyle.Render("Script:"))
	b.WriteString("\n")
	b.WriteString(m.LuaScriptInput.View())
	b.WriteString("\n\n")

	if m.LuaResult != "" {
		b.WriteString(keyStyle.Render("Result: "))
		b.WriteString(normalStyle.Render(m.LuaResult))
		b.WriteString("\n\n")
	}

	b.WriteString(helpStyle.Render("enter:execute  esc:back"))

	return m.renderModal(b.String())
}

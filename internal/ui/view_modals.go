package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/davidbudnick/redis/internal/types"
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
				{"a", "Add key"},
				{"d", "Delete key"},
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

func (m Model) viewConfirmDelete() string {
	var b strings.Builder

	warningStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true)

	b.WriteString(warningStyle.Render("⚠ Confirm Delete"))
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

func (m Model) viewTTLEditor() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Set TTL"))
	b.WriteString("\n\n")

	if m.CurrentKey != nil {
		b.WriteString(keyStyle.Render("Key: "))
		b.WriteString(normalStyle.Render(m.CurrentKey.Key))
		b.WriteString("\n\n")
	}

	b.WriteString(keyStyle.Render("TTL (seconds):"))
	b.WriteString("\n")
	b.WriteString(m.TTLInput.View())
	b.WriteString("\n")
	b.WriteString(dimStyle.Render("Enter 0 to remove expiry"))
	b.WriteString("\n\n")

	b.WriteString(helpStyle.Render("enter:save  esc:cancel"))

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

func (m Model) viewEditValue() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Edit Value"))
	b.WriteString("\n\n")

	if m.CurrentKey != nil {
		b.WriteString(keyStyle.Render("Key: "))
		b.WriteString(normalStyle.Render(m.CurrentKey.Key))
		b.WriteString("\n\n")
	}

	b.WriteString(keyStyle.Render("New Value:"))
	b.WriteString("\n")
	b.WriteString(m.EditValueInput.View())
	b.WriteString("\n\n")

	b.WriteString(helpStyle.Render("enter:save  esc:cancel"))

	return m.renderModal(b.String())
}

func (m Model) viewAddToCollection() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Add to Collection"))
	b.WriteString("\n\n")

	if m.CurrentKey != nil {
		b.WriteString(keyStyle.Render("Key: "))
		b.WriteString(normalStyle.Render(m.CurrentKey.Key))
		b.WriteString(" (")
		b.WriteString(lipgloss.NewStyle().Foreground(getTypeColor(m.CurrentKey.Type)).Render(string(m.CurrentKey.Type)))
		b.WriteString(")")
		b.WriteString("\n\n")

		var label1, label2 string
		switch m.CurrentKey.Type {
		case types.KeyTypeList:
			label1, label2 = "Element:", ""
		case types.KeyTypeSet:
			label1, label2 = "Member:", ""
		case types.KeyTypeZSet:
			label1, label2 = "Member:", "Score:"
		case types.KeyTypeHash:
			label1, label2 = "Field:", "Value:"
		case types.KeyTypeStream:
			label1, label2 = "Field:", "Value:"
		}

		b.WriteString(keyStyle.Render(label1))
		b.WriteString("\n")
		b.WriteString(m.AddCollectionInput[0].View())
		b.WriteString("\n\n")

		if label2 != "" {
			b.WriteString(keyStyle.Render(label2))
			b.WriteString("\n")
			b.WriteString(m.AddCollectionInput[1].View())
			b.WriteString("\n\n")
		}
	}

	b.WriteString(helpStyle.Render("tab:next  enter:add  esc:cancel"))

	return m.renderModal(b.String())
}

func (m Model) viewRemoveFromCollection() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Remove from Collection"))
	b.WriteString("\n\n")

	if m.CurrentKey != nil {
		b.WriteString(keyStyle.Render("Key: "))
		b.WriteString(normalStyle.Render(m.CurrentKey.Key))
		b.WriteString("\n")
		b.WriteString(dimStyle.Render("Select item to remove:"))
		b.WriteString("\n\n")

		switch m.CurrentValue.Type {
		case types.KeyTypeList:
			for i, v := range m.CurrentValue.ListValue {
				prefix := "  "
				if i == m.SelectedItemIdx {
					prefix = "▶ "
					b.WriteString(selectedStyle.Render(fmt.Sprintf("%s%d: %s", prefix, i, truncate(v, 50))))
				} else {
					b.WriteString(normalStyle.Render(fmt.Sprintf("%s%d: %s", prefix, i, truncate(v, 50))))
				}
				b.WriteString("\n")
			}
		case types.KeyTypeSet:
			for i, v := range m.CurrentValue.SetValue {
				prefix := "  "
				if i == m.SelectedItemIdx {
					b.WriteString(selectedStyle.Render(fmt.Sprintf("▶ %s", truncate(v, 50))))
				} else {
					b.WriteString(normalStyle.Render(fmt.Sprintf("%s%s", prefix, truncate(v, 50))))
				}
				b.WriteString("\n")
			}
		case types.KeyTypeZSet:
			for i, v := range m.CurrentValue.ZSetValue {
				if i == m.SelectedItemIdx {
					b.WriteString(selectedStyle.Render(fmt.Sprintf("▶ %.2f: %s", v.Score, truncate(v.Member, 45))))
				} else {
					b.WriteString(normalStyle.Render(fmt.Sprintf("  %.2f: %s", v.Score, truncate(v.Member, 45))))
				}
				b.WriteString("\n")
			}
		case types.KeyTypeHash:
			keys := make([]string, 0, len(m.CurrentValue.HashValue))
			for k := range m.CurrentValue.HashValue {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for i, k := range keys {
				v := m.CurrentValue.HashValue[k]
				if i == m.SelectedItemIdx {
					b.WriteString(selectedStyle.Render(fmt.Sprintf("▶ %s: %s", k, truncate(v, 40))))
				} else {
					b.WriteString(normalStyle.Render(fmt.Sprintf("  %s: %s", k, truncate(v, 40))))
				}
				b.WriteString("\n")
			}
		case types.KeyTypeStream:
			for i, e := range m.CurrentValue.StreamValue {
				if i == m.SelectedItemIdx {
					b.WriteString(selectedStyle.Render(fmt.Sprintf("▶ %s", e.ID)))
				} else {
					b.WriteString(normalStyle.Render(fmt.Sprintf("  %s", e.ID)))
				}
				b.WriteString("\n")
			}
		}
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("j/k:navigate  enter/d:delete  esc:cancel"))

	return m.renderModal(b.String())
}

func (m Model) viewRenameKey() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Rename Key"))
	b.WriteString("\n\n")

	if m.CurrentKey != nil {
		b.WriteString(keyStyle.Render("Current: "))
		b.WriteString(normalStyle.Render(m.CurrentKey.Key))
		b.WriteString("\n\n")
	}

	b.WriteString(keyStyle.Render("New Name:"))
	b.WriteString("\n")
	b.WriteString(m.RenameInput.View())
	b.WriteString("\n\n")

	b.WriteString(helpStyle.Render("enter:rename  esc:cancel"))

	return m.renderModal(b.String())
}

func (m Model) viewCopyKey() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Copy Key"))
	b.WriteString("\n\n")

	if m.CurrentKey != nil {
		b.WriteString(keyStyle.Render("Source: "))
		b.WriteString(normalStyle.Render(m.CurrentKey.Key))
		b.WriteString("\n\n")
	}

	b.WriteString(keyStyle.Render("Destination:"))
	b.WriteString("\n")
	b.WriteString(m.CopyInput.View())
	b.WriteString("\n\n")

	b.WriteString(helpStyle.Render("enter:copy  esc:cancel"))

	return m.renderModal(b.String())
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

func (m Model) viewSearchValues() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Search by Value"))
	b.WriteString("\n\n")

	b.WriteString(dimStyle.Render("Find keys containing a specific value"))
	b.WriteString("\n\n")

	b.WriteString(keyStyle.Render("Search:"))
	b.WriteString("\n")
	b.WriteString(m.SearchValueInput.View())
	b.WriteString("\n\n")

	b.WriteString(helpStyle.Render("enter:search  esc:cancel"))

	return m.renderModal(b.String())
}

func (m Model) viewExport() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Export Keys"))
	b.WriteString("\n\n")

	pattern := m.KeyPattern
	if pattern == "" {
		pattern = "*"
	}
	b.WriteString(keyStyle.Render("Pattern: "))
	b.WriteString(normalStyle.Render(pattern))
	b.WriteString("\n\n")

	b.WriteString(keyStyle.Render("Filename:"))
	b.WriteString("\n")
	b.WriteString(m.ExportInput.View())
	b.WriteString("\n\n")

	b.WriteString(helpStyle.Render("enter:export  esc:cancel"))

	return m.renderModal(b.String())
}

func (m Model) viewImport() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Import Keys"))
	b.WriteString("\n\n")

	b.WriteString(dimStyle.Render("Import keys from a JSON file"))
	b.WriteString("\n\n")

	b.WriteString(keyStyle.Render("Filename:"))
	b.WriteString("\n")
	b.WriteString(m.ImportInput.View())
	b.WriteString("\n\n")

	b.WriteString(helpStyle.Render("enter:import  esc:cancel"))

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

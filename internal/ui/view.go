package ui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"redis/internal/types"

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

func (m Model) View() string {
	if m.Width < 50 || m.Height < 15 {
		return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center,
			"Terminal too small.\nResize to at least 50x15.")
	}

	var content string
	switch m.Screen {
	case types.ScreenConnections:
		content = m.viewConnections()
	case types.ScreenAddConnection:
		content = m.viewAddConnection()
	case types.ScreenEditConnection:
		content = m.viewEditConnection()
	case types.ScreenKeys:
		content = m.viewKeys()
	case types.ScreenKeyDetail:
		content = m.viewKeyDetail()
	case types.ScreenAddKey:
		content = m.viewAddKey()
	case types.ScreenHelp:
		content = m.viewHelp()
	case types.ScreenConfirmDelete:
		content = m.viewConfirmDelete()
	case types.ScreenServerInfo:
		content = m.viewServerInfo()
	case types.ScreenTTLEditor:
		content = m.viewTTLEditor()
	case types.ScreenEditValue:
		content = m.viewEditValue()
	case types.ScreenAddToCollection:
		content = m.viewAddToCollection()
	case types.ScreenRemoveFromCollection:
		content = m.viewRemoveFromCollection()
	case types.ScreenRenameKey:
		content = m.viewRenameKey()
	case types.ScreenCopyKey:
		content = m.viewCopyKey()
	case types.ScreenPubSub:
		content = m.viewPubSub()
	case types.ScreenPublishMessage:
		content = m.viewPubSub()
	case types.ScreenSwitchDB:
		content = m.viewSwitchDB()
	case types.ScreenSearchValues:
		content = m.viewSearchValues()
	case types.ScreenExport:
		content = m.viewExport()
	case types.ScreenImport:
		content = m.viewImport()
	case types.ScreenSlowLog:
		content = m.viewSlowLog()
	case types.ScreenLuaScript:
		content = m.viewLuaScript()
	case types.ScreenTestConnection:
		content = m.viewTestConnection()
	case types.ScreenLogs:
		content = m.viewLogs()
	case types.ScreenBulkDelete:
		content = m.viewBulkDelete()
	case types.ScreenBatchTTL:
		content = m.viewBatchTTL()
	case types.ScreenFavorites:
		content = m.viewFavorites()
	case types.ScreenRecentKeys:
		content = m.viewRecentKeys()
	case types.ScreenTreeView:
		content = m.viewTreeView()
	case types.ScreenRegexSearch:
		content = m.viewRegexSearch()
	case types.ScreenFuzzySearch:
		content = m.viewFuzzySearch()
	case types.ScreenClientList:
		content = m.viewClientList()
	case types.ScreenMemoryStats:
		content = m.viewMemoryStats()
	case types.ScreenClusterInfo:
		content = m.viewClusterInfo()
	case types.ScreenCompareKeys:
		content = m.viewCompareKeys()
	case types.ScreenTemplates:
		content = m.viewTemplates()
	case types.ScreenValueHistory:
		content = m.viewValueHistory()
	case types.ScreenKeyspaceEvents:
		content = m.viewKeyspaceEvents()
	case types.ScreenJSONPath:
		content = m.viewJSONPath()
	case types.ScreenExpiringKeys:
		content = m.viewExpiringKeys()
	}

	// Status bar
	var status string
	if m.Loading {
		status = dimStyle.Render("Loading...")
	} else if m.StatusMsg != "" {
		if strings.HasPrefix(m.StatusMsg, "Error") {
			status = errorStyle.Render(m.StatusMsg)
		} else {
			status = successStyle.Render(m.StatusMsg)
		}
	}

	fullContent := content + "\n\n" + status
	return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, fullContent)
}

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

func (m Model) viewKeys() string {
	var b strings.Builder

	connInfo := ""
	if m.CurrentConn != nil {
		connInfo = fmt.Sprintf(" - %s (%s:%d/db%d)", m.CurrentConn.Name, m.CurrentConn.Host, m.CurrentConn.Port, m.CurrentConn.DB)
	}

	b.WriteString(titleStyle.Render("Keys" + connInfo))
	b.WriteString("\n\n")

	// Pattern filter
	b.WriteString(keyStyle.Render("Filter: "))
	if m.PatternInput.Focused() {
		b.WriteString(m.PatternInput.View())
	} else {
		pattern := m.KeyPattern
		if pattern == "" {
			pattern = "*"
		}
		b.WriteString(normalStyle.Render(pattern))
	}
	b.WriteString("\n\n")

	if len(m.Keys) == 0 {
		b.WriteString(dimStyle.Render("No keys found."))
		b.WriteString("\n")
		b.WriteString(dimStyle.Render("Press 'a' to add one."))
	} else {
		header := fmt.Sprintf("  %-40s %-10s %-15s", "Key", "Type", "TTL")
		b.WriteString(headerStyle.Render(header))
		b.WriteString("\n")
		b.WriteString(dimStyle.Render(strings.Repeat("─", 70)))
		b.WriteString("\n")

		// Calculate visible window
		maxVisible := m.Height - 15
		if maxVisible < 5 {
			maxVisible = 5
		}
		startIdx := 0
		if m.SelectedKeyIdx >= maxVisible {
			startIdx = m.SelectedKeyIdx - maxVisible + 1
		}
		endIdx := startIdx + maxVisible
		if endIdx > len(m.Keys) {
			endIdx = len(m.Keys)
		}

		for i := startIdx; i < endIdx; i++ {
			key := m.Keys[i]
			keyName := key.Key
			if len(keyName) > 40 {
				keyName = keyName[:37] + "..."
			}

			ttlStr := "∞"
			var ttlStyle lipgloss.Style
			if key.TTL > 0 {
				ttlStr = key.TTL.String()
				// Warning colors for low TTL
				if key.TTL <= 10*time.Second {
					ttlStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true) // Red
				} else if key.TTL <= 60*time.Second {
					ttlStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("3")) // Yellow
				} else {
					ttlStyle = normalStyle
				}
			} else if key.TTL == -2 {
				ttlStr = "expired"
				ttlStyle = errorStyle
			} else {
				ttlStyle = dimStyle
			}

			typeColor := getTypeColor(key.Type)
			typePart := lipgloss.NewStyle().Foreground(typeColor).Render(fmt.Sprintf("%-10s", key.Type))
			ttlPart := ttlStyle.Render(fmt.Sprintf("%-15s", ttlStr))

			line := fmt.Sprintf("%-40s %s %s", keyName, typePart, ttlPart)
			if i == m.SelectedKeyIdx {
				b.WriteString(selectedStyle.Render("▶ " + fmt.Sprintf("%-40s %-10s %-15s", keyName, key.Type, ttlStr)))
			} else {
				b.WriteString(normalStyle.Render("  ") + line)
			}
			b.WriteString("\n")
		}

		if len(m.Keys) > maxVisible {
			b.WriteString(dimStyle.Render(fmt.Sprintf("\nShowing %d-%d of %d", startIdx+1, endIdx, len(m.Keys))))
		}
		if m.KeyCursor > 0 {
			b.WriteString(dimStyle.Render(" [l:load more]"))
		}
	}

	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("j/k:nav  enter:view  a:add  d:del  /:filter  O:logs  i:info  q:back"))

	return b.String()
}

func (m Model) viewKeyDetail() string {
	var b strings.Builder

	if m.CurrentKey == nil {
		return "No key selected"
	}

	b.WriteString(titleStyle.Render("Key Detail"))
	b.WriteString("\n\n")

	b.WriteString(keyStyle.Render("Key: "))
	b.WriteString(normalStyle.Render(m.CurrentKey.Key))
	b.WriteString("\n")

	b.WriteString(keyStyle.Render("Type: "))
	typeColor := getTypeColor(m.CurrentKey.Type)
	b.WriteString(lipgloss.NewStyle().Foreground(typeColor).Bold(true).Render(string(m.CurrentKey.Type)))
	b.WriteString("\n")

	b.WriteString(keyStyle.Render("TTL: "))
	ttlStr := "No expiry"
	var ttlDetailStyle lipgloss.Style
	if m.CurrentKey.TTL > 0 {
		ttlStr = m.CurrentKey.TTL.String()
		if m.CurrentKey.TTL <= 10*time.Second {
			ttlDetailStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true) // Red - critical
			ttlStr = "⚠ " + ttlStr
		} else if m.CurrentKey.TTL <= 60*time.Second {
			ttlDetailStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("3")) // Yellow - warning
			ttlStr = "⏱ " + ttlStr
		} else {
			ttlDetailStyle = normalStyle
		}
	} else {
		ttlDetailStyle = dimStyle
	}
	b.WriteString(ttlDetailStyle.Render(ttlStr))

	// Show memory usage if available
	if m.MemoryUsage > 0 {
		b.WriteString("  ")
		b.WriteString(keyStyle.Render("Memory: "))
		b.WriteString(normalStyle.Render(formatBytes(m.MemoryUsage)))
	}
	b.WriteString("\n\n")

	b.WriteString(keyStyle.Render("Value:"))
	b.WriteString("\n")

	valueBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(1, 2).
		Width(70)

	var valueContent string
	switch m.CurrentValue.Type {
	case types.KeyTypeString:
		valueContent = formatPossibleJSON(m.CurrentValue.StringValue)
	case types.KeyTypeList:
		if len(m.CurrentValue.ListValue) == 0 {
			valueContent = "(empty list)"
		} else {
			for i, v := range m.CurrentValue.ListValue {
				valueContent += fmt.Sprintf("%d. %s\n", i, formatPossibleJSON(v))
			}
		}
	case types.KeyTypeSet:
		if len(m.CurrentValue.SetValue) == 0 {
			valueContent = "(empty set)"
		} else {
			for _, v := range m.CurrentValue.SetValue {
				valueContent += "• " + formatPossibleJSON(v) + "\n"
			}
		}
	case types.KeyTypeZSet:
		if len(m.CurrentValue.ZSetValue) == 0 {
			valueContent = "(empty sorted set)"
		} else {
			for _, v := range m.CurrentValue.ZSetValue {
				valueContent += fmt.Sprintf("%.2f: %s\n", v.Score, formatPossibleJSON(v.Member))
			}
		}
	case types.KeyTypeHash:
		if len(m.CurrentValue.HashValue) == 0 {
			valueContent = "(empty hash)"
		} else {
			// Sort hash keys for consistent display
			hashKeys := make([]string, 0, len(m.CurrentValue.HashValue))
			for k := range m.CurrentValue.HashValue {
				hashKeys = append(hashKeys, k)
			}
			sort.Strings(hashKeys)
			for _, k := range hashKeys {
				v := m.CurrentValue.HashValue[k]
				formattedValue := formatPossibleJSON(v)
				// Check if value is multi-line JSON
				if strings.Contains(formattedValue, "\n") {
					valueContent += fmt.Sprintf("◆ %s:\n%s\n", k, formattedValue)
				} else {
					valueContent += fmt.Sprintf("◆ %s: %s\n", k, formattedValue)
				}
			}
		}
	case types.KeyTypeStream:
		if len(m.CurrentValue.StreamValue) == 0 {
			valueContent = "(empty stream)"
		} else {
			for _, entry := range m.CurrentValue.StreamValue {
				// Try to format stream fields as JSON
				jsonBytes, err := json.MarshalIndent(entry.Fields, "", "  ")
				if err == nil {
					valueContent += fmt.Sprintf("%s:\n%s\n", entry.ID, string(jsonBytes))
				} else {
					fields := []string{}
					for k, v := range entry.Fields {
						fields = append(fields, fmt.Sprintf("%s=%v", k, v))
					}
					valueContent += fmt.Sprintf("%s: %s\n", entry.ID, strings.Join(fields, ", "))
				}
			}
		}
	}

	b.WriteString(valueBox.Render(strings.TrimSpace(valueContent)))
	b.WriteString("\n\n")

	helpText := "t:TTL  d:del  r:refresh  R:rename  c:copy"
	if m.CurrentKey.Type == types.KeyTypeString {
		helpText += "  e:edit"
	} else {
		helpText += "  a:add  x:remove"
	}
	helpText += "  esc:back"
	b.WriteString(helpStyle.Render(helpText))

	return b.String()
}

// formatPossibleJSON tries to pretty-print JSON, returns original if not valid JSON
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

func (m Model) viewAddKey() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Add Key"))
	b.WriteString("\n\n")

	b.WriteString(keyStyle.Render("Type: "))
	typeColor := getTypeColor(m.AddKeyType)
	b.WriteString(lipgloss.NewStyle().Foreground(typeColor).Bold(true).Render(string(m.AddKeyType)))
	b.WriteString(dimStyle.Render(" (Ctrl+T to change)"))
	b.WriteString("\n\n")

	b.WriteString(keyStyle.Render("Key Name:"))
	b.WriteString("\n")
	b.WriteString(m.AddKeyInputs[0].View())
	b.WriteString("\n\n")

	valueLabel := "Value:"
	switch m.AddKeyType {
	case types.KeyTypeList:
		valueLabel = "Initial Element:"
	case types.KeyTypeSet:
		valueLabel = "Initial Member:"
	case types.KeyTypeZSet:
		valueLabel = "Initial Member:"
	case types.KeyTypeHash:
		valueLabel = "Initial Field Value:"
	case types.KeyTypeStream:
		valueLabel = "Initial Data:"
	}

	b.WriteString(keyStyle.Render(valueLabel))
	b.WriteString("\n")
	b.WriteString(m.AddKeyInputs[1].View())
	b.WriteString("\n\n")

	b.WriteString(helpStyle.Render("tab:next  Ctrl+T:type  enter:save  esc:cancel"))

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
		b.WriteString(normalStyle.Render(m.TestConnResult))
	}

	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("esc:back"))

	return m.renderModal(b.String())
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

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func (m Model) viewLogs() string {
	var b strings.Builder

	logCount := 0
	if m.Logs != nil {
		logCount = len(*m.Logs)
	}

	b.WriteString(titleStyle.Render(fmt.Sprintf("📋 Application Logs (%d entries)", logCount)))
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

	b.WriteString(titleStyle.Render("📄 Log Entry Details"))
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

// New view functions for additional features

func (m Model) viewBulkDelete() string {
	var b strings.Builder

	warningStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true)

	b.WriteString(warningStyle.Render("⚠ Bulk Delete Keys"))
	b.WriteString("\n\n")

	b.WriteString(dimStyle.Render("Delete all keys matching a pattern"))
	b.WriteString("\n\n")

	b.WriteString(keyStyle.Render("Pattern:"))
	b.WriteString("\n")
	b.WriteString(m.BulkDeleteInput.View())
	b.WriteString("\n\n")

	if len(m.BulkDeletePreview) > 0 {
		b.WriteString(keyStyle.Render(fmt.Sprintf("Will delete %d keys:", len(m.BulkDeletePreview))))
		b.WriteString("\n")
		for i, k := range m.BulkDeletePreview {
			if i >= 5 {
				b.WriteString(dimStyle.Render(fmt.Sprintf("  ... and %d more", len(m.BulkDeletePreview)-5)))
				break
			}
			b.WriteString(normalStyle.Render("  • " + k))
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	b.WriteString(helpStyle.Render("enter:delete  esc:cancel"))

	return m.renderModal(b.String())
}

func (m Model) viewBatchTTL() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Batch Set TTL"))
	b.WriteString("\n\n")

	b.WriteString(dimStyle.Render("Set TTL on all keys matching a pattern"))
	b.WriteString("\n\n")

	b.WriteString(keyStyle.Render("TTL (seconds):"))
	b.WriteString("\n")
	b.WriteString(m.BatchTTLInput.View())
	b.WriteString("\n\n")

	b.WriteString(keyStyle.Render("Pattern:"))
	b.WriteString("\n")
	b.WriteString(m.BatchTTLPattern.View())
	b.WriteString("\n\n")

	b.WriteString(helpStyle.Render("tab:next  enter:apply  esc:cancel"))

	return m.renderModal(b.String())
}

func (m Model) viewFavorites() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("⭐ Favorites"))
	b.WriteString("\n\n")

	if len(m.Favorites) == 0 {
		b.WriteString(dimStyle.Render("No favorites yet."))
		b.WriteString("\n")
		b.WriteString(dimStyle.Render("Press 'f' on a key to add it."))
	} else {
		for i, fav := range m.Favorites {
			prefix := "  "
			connName := fav.Connection
			if connName == "" {
				connName = fav.Label
			}
			if i == m.SelectedFavIdx {
				b.WriteString(selectedStyle.Render(fmt.Sprintf("▶ %-40s %s", truncate(fav.Key, 40), connName)))
			} else {
				b.WriteString(normalStyle.Render(fmt.Sprintf("%s%-40s %s", prefix, truncate(fav.Key, 40), dimStyle.Render(connName))))
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("j/k:nav  enter:view  d:remove  esc:back"))

	return m.renderModal(b.String())
}

func (m Model) viewRecentKeys() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("🕒 Recent Keys"))
	b.WriteString("\n\n")

	if len(m.RecentKeys) == 0 {
		b.WriteString(dimStyle.Render("No recent keys."))
	} else {
		for i, recent := range m.RecentKeys {
			if i == m.SelectedRecentIdx {
				b.WriteString(selectedStyle.Render(fmt.Sprintf("▶ %-40s", truncate(recent.Key, 40))))
			} else {
				b.WriteString(normalStyle.Render(fmt.Sprintf("  %-40s", truncate(recent.Key, 40))))
			}
			b.WriteString(dimStyle.Render(" " + recent.AccessedAt.Format("15:04:05")))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("j/k:nav  enter:view  esc:back"))

	return m.renderModal(b.String())
}

func (m Model) viewTreeView() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("🌳 Key Tree View"))
	b.WriteString("\n\n")

	if len(m.TreeNodes) == 0 {
		b.WriteString(dimStyle.Render("No keys found."))
	} else {
		for i, node := range m.TreeNodes {
			indent := strings.Repeat("  ", node.GetDepth())
			prefix := "  "
			icon := "📄"
			if !node.IsKey {
				if m.TreeExpanded[node.FullPath] {
					icon = "📂"
				} else {
					icon = "📁"
				}
			}

			line := fmt.Sprintf("%s%s %s (%d)", indent, icon, node.Name, node.ChildCount)
			if i == m.SelectedTreeIdx {
				b.WriteString(selectedStyle.Render("▶ " + line[2:]))
			} else {
				b.WriteString(normalStyle.Render(prefix + line[2:]))
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("j/k:nav  enter/space:expand  esc:back"))

	return m.renderModalWide(b.String())
}

func (m Model) viewRegexSearch() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("🔍 Regex Search"))
	b.WriteString("\n\n")

	b.WriteString(dimStyle.Render("Search keys using regular expressions"))
	b.WriteString("\n\n")

	b.WriteString(keyStyle.Render("Pattern:"))
	b.WriteString("\n")
	b.WriteString(m.RegexSearchInput.View())
	b.WriteString("\n\n")

	b.WriteString(dimStyle.Render("Examples: user:\\d+  session:[a-f0-9]+"))
	b.WriteString("\n\n")

	b.WriteString(helpStyle.Render("enter:search  esc:cancel"))

	return m.renderModal(b.String())
}

func (m Model) viewFuzzySearch() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("🔎 Fuzzy Search"))
	b.WriteString("\n\n")

	b.WriteString(dimStyle.Render("Search keys with fuzzy matching"))
	b.WriteString("\n\n")

	b.WriteString(keyStyle.Render("Search:"))
	b.WriteString("\n")
	b.WriteString(m.FuzzySearchInput.View())
	b.WriteString("\n\n")

	b.WriteString(helpStyle.Render("enter:search  esc:cancel"))

	return m.renderModal(b.String())
}

func (m Model) viewClientList() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("👥 Connected Clients"))
	b.WriteString("\n\n")

	if len(m.ClientList) == 0 {
		b.WriteString(dimStyle.Render("No clients connected."))
	} else {
		header := fmt.Sprintf("  %-20s %-15s %-10s", "Address", "Name", "Age")
		b.WriteString(headerStyle.Render(header))
		b.WriteString("\n")
		b.WriteString(dimStyle.Render(strings.Repeat("─", 50)))
		b.WriteString("\n")

		for i, client := range m.ClientList {
			name := client.Name
			if name == "" {
				name = "-"
			}
			line := fmt.Sprintf("%-20s %-15s %-10s", client.Addr, name, client.Age)
			if i == m.SelectedClientIdx {
				b.WriteString(selectedStyle.Render("▶ " + line))
			} else {
				b.WriteString(normalStyle.Render("  " + line))
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("j/k:nav  r:refresh  esc:back"))

	return m.renderModalWide(b.String())
}

func (m Model) viewMemoryStats() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("📊 Memory Statistics"))
	b.WriteString("\n\n")

	if m.MemoryStats == nil {
		b.WriteString(dimStyle.Render("Loading memory stats..."))
	} else {
		stats := []struct {
			label string
			value string
		}{
			{"Used Memory", formatBytes(m.MemoryStats.UsedMemory)},
			{"Peak Memory", formatBytes(m.MemoryStats.PeakMemory)},
			{"Fragmentation Ratio", fmt.Sprintf("%.2f", m.MemoryStats.FragRatio)},
			{"RSS", m.MemoryStats.RSS},
			{"Lua Memory", m.MemoryStats.LuaMemory},
		}

		for _, stat := range stats {
			b.WriteString(keyStyle.Render(fmt.Sprintf("%-22s", stat.label+":")))
			b.WriteString(normalStyle.Render(stat.value))
			b.WriteString("\n")
		}

		if len(m.MemoryStats.TopKeys) > 0 {
			b.WriteString("\n")
			b.WriteString(keyStyle.Render("Top Keys by Memory:"))
			b.WriteString("\n")
			for _, key := range m.MemoryStats.TopKeys {
				b.WriteString(fmt.Sprintf("  %s: %s\n", key.Key, formatBytes(key.Memory)))
			}
		}
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("r:refresh  esc:back"))

	return m.renderModal(b.String())
}

func (m Model) viewClusterInfo() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("🌐 Cluster Info"))
	b.WriteString("\n\n")

	if !m.ClusterEnabled {
		b.WriteString(dimStyle.Render("Cluster mode is not enabled on this Redis instance."))
	} else if len(m.ClusterNodes) == 0 {
		b.WriteString(dimStyle.Render("No cluster nodes found."))
	} else {
		header := fmt.Sprintf("  %-20s %-10s %-15s", "Node", "Role", "Slots")
		b.WriteString(headerStyle.Render(header))
		b.WriteString("\n")
		b.WriteString(dimStyle.Render(strings.Repeat("─", 50)))
		b.WriteString("\n")

		for i, node := range m.ClusterNodes {
			slots := node.Slots
			if slots == "" {
				slots = "-"
			}
			nodeID := node.ID
			if len(nodeID) > 8 {
				nodeID = nodeID[:8] + "..."
			}
			line := fmt.Sprintf("%-20s %-10s %-15s", nodeID, node.Role, slots)
			if i == m.SelectedNodeIdx {
				b.WriteString(selectedStyle.Render("▶ " + line))
			} else {
				b.WriteString(normalStyle.Render("  " + line))
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("j/k:nav  r:refresh  esc:back"))

	return m.renderModalWide(b.String())
}

func (m Model) viewCompareKeys() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("⚖️ Compare Keys"))
	b.WriteString("\n\n")

	b.WriteString(keyStyle.Render("Key 1:"))
	b.WriteString("\n")
	b.WriteString(m.CompareKey1Input.View())
	b.WriteString("\n\n")

	b.WriteString(keyStyle.Render("Key 2:"))
	b.WriteString("\n")
	b.WriteString(m.CompareKey2Input.View())
	b.WriteString("\n\n")

	if m.CompareResult != nil {
		if m.CompareResult.Equal {
			b.WriteString(successStyle.Render("✓ Keys are identical"))
		} else {
			b.WriteString(errorStyle.Render("✗ Keys differ"))
			b.WriteString("\n\n")
			if len(m.CompareResult.Differences) > 0 {
				b.WriteString(keyStyle.Render("Differences:"))
				b.WriteString("\n")
				for _, diff := range m.CompareResult.Differences {
					b.WriteString("  • " + diff + "\n")
				}
			}
		}
		b.WriteString("\n")
	}

	b.WriteString(helpStyle.Render("tab:switch  enter:compare  esc:back"))

	return m.renderModal(b.String())
}

func (m Model) viewTemplates() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("📝 Key Templates"))
	b.WriteString("\n\n")

	if len(m.Templates) == 0 {
		b.WriteString(dimStyle.Render("No templates configured."))
	} else {
		for i, tmpl := range m.Templates {
			typeColor := getTypeColor(tmpl.KeyType)
			typePart := lipgloss.NewStyle().Foreground(typeColor).Render(string(tmpl.KeyType))

			if i == m.SelectedTemplateIdx {
				b.WriteString(selectedStyle.Render(fmt.Sprintf("▶ %-25s %-10s", tmpl.Name, tmpl.KeyType)))
			} else {
				b.WriteString(normalStyle.Render(fmt.Sprintf("  %-25s ", tmpl.Name)))
				b.WriteString(typePart)
			}
			b.WriteString("\n")
			b.WriteString(dimStyle.Render(fmt.Sprintf("    %s", tmpl.Pattern)))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("j/k:nav  enter:use  esc:back"))

	return m.renderModal(b.String())
}

func (m Model) viewValueHistory() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("📜 Value History"))
	b.WriteString("\n\n")

	if m.CurrentKey != nil {
		b.WriteString(keyStyle.Render("Key: "))
		b.WriteString(normalStyle.Render(m.CurrentKey.Key))
		b.WriteString("\n\n")
	}

	if len(m.ValueHistory) == 0 {
		b.WriteString(dimStyle.Render("No history available for this key."))
	} else {
		for i, entry := range m.ValueHistory {
			timeStr := entry.Timestamp.Format("2006-01-02 15:04:05")
			value := truncate(entry.Value.StringValue, 40)

			if i == m.SelectedHistoryIdx {
				b.WriteString(selectedStyle.Render(fmt.Sprintf("▶ %s  %s", timeStr, value)))
			} else {
				b.WriteString(dimStyle.Render(timeStr))
				b.WriteString("  ")
				b.WriteString(normalStyle.Render(value))
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("j/k:nav  enter:restore  esc:back"))

	return m.renderModalWide(b.String())
}

func (m Model) viewKeyspaceEvents() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("📡 Keyspace Events"))
	b.WriteString("\n\n")

	if m.KeyspaceSubActive {
		b.WriteString(successStyle.Render("● Listening for events..."))
	} else {
		b.WriteString(dimStyle.Render("○ Not subscribed"))
	}
	b.WriteString("\n\n")

	if len(m.KeyspaceEvents) == 0 {
		b.WriteString(dimStyle.Render("No events received yet."))
	} else {
		header := fmt.Sprintf("%-12s %-10s %-30s", "Time", "Event", "Key")
		b.WriteString(headerStyle.Render(header))
		b.WriteString("\n")
		b.WriteString(dimStyle.Render(strings.Repeat("─", 55)))
		b.WriteString("\n")

		// Show last 15 events
		start := 0
		if len(m.KeyspaceEvents) > 15 {
			start = len(m.KeyspaceEvents) - 15
		}
		for _, event := range m.KeyspaceEvents[start:] {
			b.WriteString(fmt.Sprintf("%-12s %-10s %-30s\n",
				event.Timestamp.Format("15:04:05"),
				event.Event,
				truncate(event.Key, 30)))
		}
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("c:clear  esc:back"))

	return m.renderModalWide(b.String())
}

func (m Model) viewJSONPath() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("🔍 JSON Path Query"))
	b.WriteString("\n\n")

	if m.CurrentKey != nil {
		b.WriteString(keyStyle.Render("Key: "))
		b.WriteString(normalStyle.Render(m.CurrentKey.Key))
		b.WriteString("\n\n")
	}

	b.WriteString(keyStyle.Render("JSON Path:"))
	b.WriteString("\n")
	b.WriteString(m.JSONPathInput.View())
	b.WriteString("\n\n")

	b.WriteString(dimStyle.Render("Examples: $.name  $.users[0]  $.items[*].id"))
	b.WriteString("\n\n")

	if m.JSONPathResult != "" {
		b.WriteString(keyStyle.Render("Result:"))
		b.WriteString("\n")
		b.WriteString(normalStyle.Render(m.JSONPathResult))
		b.WriteString("\n\n")
	}

	b.WriteString(helpStyle.Render("enter:query  esc:back"))

	return m.renderModal(b.String())
}

func (m Model) viewExpiringKeys() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("⏰ Expiring Keys"))
	b.WriteString("\n\n")

	b.WriteString(dimStyle.Render(fmt.Sprintf("Keys expiring within %d seconds", m.ExpiryThreshold)))
	b.WriteString("\n\n")

	if len(m.ExpiringKeys) == 0 {
		b.WriteString(dimStyle.Render("No keys expiring soon."))
	} else {
		header := fmt.Sprintf("  %-40s %-15s", "Key", "TTL")
		b.WriteString(headerStyle.Render(header))
		b.WriteString("\n")
		b.WriteString(dimStyle.Render(strings.Repeat("─", 60)))
		b.WriteString("\n")

		for i, key := range m.ExpiringKeys {
			keyName := truncate(key.Key, 40)
			ttlStr := key.TTL.String()

			var ttlStyle lipgloss.Style
			if key.TTL <= 10*time.Second {
				ttlStyle = errorStyle
			} else if key.TTL <= 60*time.Second {
				ttlStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
			} else {
				ttlStyle = normalStyle
			}

			if i == m.SelectedKeyIdx {
				b.WriteString(selectedStyle.Render(fmt.Sprintf("▶ %-40s %-15s", keyName, ttlStr)))
			} else {
				b.WriteString(normalStyle.Render(fmt.Sprintf("  %-40s ", keyName)))
				b.WriteString(ttlStyle.Render(ttlStr))
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("j/k:nav  enter:view  esc:back"))

	return m.renderModalWide(b.String())
}

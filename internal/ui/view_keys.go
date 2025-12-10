package ui

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/davidbudnick/redis/internal/types"
)

func (m Model) viewKeys() string {
	// Calculate panel widths - left panel for keys, right panel for preview
	totalWidth := m.Width
	if totalWidth < 80 || m.Height < 20 {
		// If terminal is too narrow or short, just show keys list without preview
		return m.viewKeysListOnly()
	}

	// Left panel gets 60%, right panel gets 40%
	leftWidth := (totalWidth * 60) / 100
	rightWidth := totalWidth - leftWidth - 1

	// Ensure minimum widths
	if leftWidth < 30 {
		leftWidth = 30
	}
	if rightWidth < 20 {
		rightWidth = 20
	}

	// Build left panel (keys list)
	leftContent := m.buildKeysListPanel(leftWidth - 2)

	// Build right panel (preview)
	rightContent := m.buildPreviewPanel(rightWidth - 2)

	// Create border styles
	panelHeight := m.Height - 4
	if panelHeight < 10 {
		panelHeight = 10
	}

	leftPanelStyle := lipgloss.NewStyle().
		Width(leftWidth).
		Height(panelHeight).
		MaxHeight(panelHeight).
		Padding(0, 1)

	rightPanelStyle := lipgloss.NewStyle().
		Width(rightWidth).
		Height(panelHeight).
		MaxHeight(panelHeight).
		Padding(0, 1).
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(lipgloss.Color("240"))

	// Normalize both panels to exact height to prevent rendering artifacts
	maxLines := panelHeight - 2
	if maxLines < 5 {
		maxLines = 5
	}

	// Pad/truncate left content
	leftLines := strings.Split(leftContent, "\n")
	if len(leftLines) > maxLines {
		leftLines = leftLines[:maxLines]
	}
	// Pad each line to full width
	padWidth := leftWidth - 2
	if padWidth < 1 {
		padWidth = 1
	}
	padStyle := lipgloss.NewStyle().Width(padWidth)
	for i := range leftLines {
		leftLines[i] = padStyle.Render(leftLines[i])
	}
	// Add empty lines to fill height
	for len(leftLines) < maxLines {
		leftLines = append(leftLines, strings.Repeat(" ", padWidth))
	}
	leftContent = strings.Join(leftLines, "\n")

	// Pad/truncate right content
	rightLines := strings.Split(rightContent, "\n")
	if len(rightLines) > maxLines {
		rightLines = rightLines[:maxLines]
	}
	// Pad each line to full width
	rightPadWidth := rightWidth - 2
	if rightPadWidth < 1 {
		rightPadWidth = 1
	}
	rightPadStyle := lipgloss.NewStyle().Width(rightPadWidth)
	for i := range rightLines {
		rightLines[i] = rightPadStyle.Render(rightLines[i])
	}
	// Add empty lines to fill height
	for len(rightLines) < maxLines {
		rightLines = append(rightLines, strings.Repeat(" ", rightPadWidth))
	}
	rightContent = strings.Join(rightLines, "\n")

	leftPanel := leftPanelStyle.Render(leftContent)
	rightPanel := rightPanelStyle.Render(rightContent)

	// Join panels horizontally
	content := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)

	// Add help text at the bottom
	helpText := helpStyle.Render("j/k:nav  enter:view  a:add  d:del  /:filter  O:logs  i:info  q:back")

	return content + "\n" + helpText
}

func (m Model) viewKeysListOnly() string {
	var b strings.Builder

	connInfo := ""
	if m.CurrentConn != nil {
		connInfo = fmt.Sprintf(" - %s (%s:%d/db%d)", m.CurrentConn.Name, m.CurrentConn.Host, m.CurrentConn.Port, m.CurrentConn.DB)
	}

	titleText := "Keys" + connInfo
	if m.TotalKeys > 0 {
		titleText += fmt.Sprintf("  [Total: %d]", m.TotalKeys)
	}
	b.WriteString(titleStyle.Render(titleText))
	b.WriteString("\n\n")

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
			var ttlStyleLocal lipgloss.Style
			if key.TTL > 0 {
				seconds := int(key.TTL.Seconds() + 0.5)
				ttlStr = fmt.Sprintf("%ds", seconds)
				if seconds <= 10 {
					ttlStyleLocal = lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true)
				} else if seconds <= 60 {
					ttlStyleLocal = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
				} else {
					ttlStyleLocal = normalStyle
				}
			} else if key.TTL == -2 {
				ttlStr = "expired"
				ttlStyleLocal = errorStyle
			} else {
				ttlStyleLocal = dimStyle
			}

			typeColor := getTypeColor(key.Type)
			typePart := lipgloss.NewStyle().Foreground(typeColor).Render(fmt.Sprintf("%-10s", key.Type))
			ttlPart := ttlStyleLocal.Render(fmt.Sprintf("%-15s", ttlStr))

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

func (m Model) buildKeysListPanel(width int) string {
	var b strings.Builder

	// Title with connection info
	connInfo := ""
	if m.CurrentConn != nil {
		connInfo = fmt.Sprintf(" - %s", m.CurrentConn.Name)
	}
	titleText := "Keys" + connInfo
	if m.TotalKeys > 0 {
		titleText += fmt.Sprintf(" [%d]", m.TotalKeys)
	}
	b.WriteString(titleStyle.Render(titleText))
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
		return b.String()
	}

	// Calculate column widths
	typeWidth := 8
	ttlWidth := 6
	keyWidth := width - typeWidth - ttlWidth - 6 // 6 for spacing and cursor
	if keyWidth < 20 {
		keyWidth = 20
	}

	// Header
	header := fmt.Sprintf("  %-*s  %-*s  %s", keyWidth, "Key", typeWidth, "Type", "TTL")
	b.WriteString(headerStyle.Render(header))
	b.WriteString("\n")
	b.WriteString(dimStyle.Render(strings.Repeat("─", width)))
	b.WriteString("\n")

	// Calculate visible range
	maxVisible := m.Height - 12
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

	// Render keys
	for i := startIdx; i < endIdx; i++ {
		key := m.Keys[i]

		// Truncate key name if needed
		keyName := key.Key
		if len(keyName) > keyWidth {
			keyName = keyName[:keyWidth-3] + "..."
		}

		// Format TTL
		ttlStr := "∞"
		if key.TTL > 0 {
			seconds := int(key.TTL.Seconds() + 0.5)
			if seconds < 60 {
				ttlStr = fmt.Sprintf("%ds", seconds)
			} else if seconds < 3600 {
				ttlStr = fmt.Sprintf("%dm", seconds/60)
			} else {
				ttlStr = fmt.Sprintf("%dh", seconds/3600)
			}
		} else if key.TTL == -2 {
			ttlStr = "exp"
		}

		if i == m.SelectedKeyIdx {
			// Selected row - highlight entire row
			typeColor := getTypeColor(key.Type)
			typeStyled := lipgloss.NewStyle().Foreground(typeColor).Bold(true).Render(fmt.Sprintf("%-*s", typeWidth, key.Type))

			cursor := selectedStyle.Render("▶ ")
			keyPart := selectedStyle.Render(fmt.Sprintf("%-*s", keyWidth, keyName))

			b.WriteString(cursor)
			b.WriteString(keyPart)
			b.WriteString("  ")
			b.WriteString(typeStyled)
			b.WriteString("  ")
			b.WriteString(normalStyle.Render(ttlStr))
		} else {
			// Normal row
			typeColor := getTypeColor(key.Type)
			typeStyled := lipgloss.NewStyle().Foreground(typeColor).Render(fmt.Sprintf("%-*s", typeWidth, key.Type))

			b.WriteString("  ")
			b.WriteString(normalStyle.Render(fmt.Sprintf("%-*s", keyWidth, keyName)))
			b.WriteString("  ")
			b.WriteString(typeStyled)
			b.WriteString("  ")
			b.WriteString(dimStyle.Render(ttlStr))
		}
		b.WriteString("\n")
	}

	// Show pagination info
	if len(m.Keys) > maxVisible {
		b.WriteString("\n")
		b.WriteString(dimStyle.Render(fmt.Sprintf("%d-%d of %d", startIdx+1, endIdx, len(m.Keys))))
		if m.KeyCursor > 0 {
			b.WriteString(dimStyle.Render(" • l:more"))
		}
	}

	return b.String()
}

func (m Model) buildPreviewPanel(width int) string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Preview"))
	b.WriteString("\n")
	b.WriteString(dimStyle.Render(strings.Repeat("─", width)))
	b.WriteString("\n")

	// Check if we have keys
	if len(m.Keys) == 0 || m.SelectedKeyIdx >= len(m.Keys) {
		b.WriteString(dimStyle.Render("No key selected"))
		return b.String()
	}

	selectedKey := m.Keys[m.SelectedKeyIdx]

	// Key name (with word wrap for long keys)
	b.WriteString(keyStyle.Render("Key: "))
	keyName := selectedKey.Key
	if len(keyName) > width-6 {
		keyName = keyName[:width-9] + "..."
	}
	b.WriteString(normalStyle.Render(keyName))
	b.WriteString("\n")

	// Type with color
	b.WriteString(keyStyle.Render("Type: "))
	typeColor := getTypeColor(selectedKey.Type)
	b.WriteString(lipgloss.NewStyle().Foreground(typeColor).Bold(true).Render(string(selectedKey.Type)))
	b.WriteString("\n")

	// TTL with visual indicator
	b.WriteString(keyStyle.Render("TTL: "))
	if selectedKey.TTL > 0 {
		seconds := int(selectedKey.TTL.Seconds() + 0.5)
		var ttlStr string
		var ttlColor lipgloss.Color

		if seconds < 60 {
			ttlStr = fmt.Sprintf("%d seconds", seconds)
		} else if seconds < 3600 {
			ttlStr = fmt.Sprintf("%d minutes", seconds/60)
		} else if seconds < 86400 {
			ttlStr = fmt.Sprintf("%d hours", seconds/3600)
		} else {
			ttlStr = fmt.Sprintf("%d days", seconds/86400)
		}

		if seconds <= 10 {
			ttlColor = lipgloss.Color("1") // Red
			ttlStr = "⚠ " + ttlStr
		} else if seconds <= 60 {
			ttlColor = lipgloss.Color("3") // Yellow
		} else {
			ttlColor = lipgloss.Color("2") // Green
		}
		b.WriteString(lipgloss.NewStyle().Foreground(ttlColor).Render(ttlStr))
	} else {
		b.WriteString(dimStyle.Render("No expiry"))
	}
	b.WriteString("\n")

	// Separator before value
	b.WriteString(dimStyle.Render(strings.Repeat("─", width)))
	b.WriteString("\n")

	// Value section
	b.WriteString(keyStyle.Render("Value"))
	b.WriteString("\n")

	// Check if preview is loaded
	if m.PreviewKey != selectedKey.Key {
		b.WriteString(dimStyle.Render("Loading..."))
		return b.String()
	}

	// Render value based on type
	maxLines := m.Height - 20
	if maxLines < 5 {
		maxLines = 5
	}

	var valueContent string
	valueContent = m.formatPreviewValue(width, maxLines)

	b.WriteString(valueContent)

	return b.String()
}

func (m Model) formatPreviewValue(maxWidth, maxLines int) string {
	var lines []string

	switch m.PreviewValue.Type {
	case types.KeyTypeString:
		value := m.PreviewValue.StringValue
		formatted := formatPossibleJSON(value)

		// Split into lines
		valueLines := strings.Split(formatted, "\n")

		var displayLines []string
		if len(valueLines) > maxLines {
			displayLines = valueLines[:maxLines-1]
			displayLines = append(displayLines, dimStyle.Render(fmt.Sprintf("↓ %d more lines", len(valueLines)-(maxLines-1))))
		} else {
			displayLines = valueLines
		}

		for _, line := range displayLines {
			lines = append(lines, normalStyle.Render(line))
		}

	case types.KeyTypeList:
		if len(m.PreviewValue.ListValue) == 0 {
			return dimStyle.Render("(empty list)")
		}
		lines = append(lines, dimStyle.Render(fmt.Sprintf("Length: %d", len(m.PreviewValue.ListValue))))
		lines = append(lines, "")

		for i, v := range m.PreviewValue.ListValue {
			if i >= maxLines-2 {
				lines = append(lines, dimStyle.Render(fmt.Sprintf("... and %d more items", len(m.PreviewValue.ListValue)-i)))
				break
			}
			val, _ := sanitizeBinaryString(v)
			if len(val) > maxWidth-8 {
				val = val[:maxWidth-11] + "..."
			}
			idx := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(fmt.Sprintf("[%d]", i))
			lines = append(lines, fmt.Sprintf("%s %s", idx, normalStyle.Render(val)))
		}

	case types.KeyTypeSet:
		if len(m.PreviewValue.SetValue) == 0 {
			return dimStyle.Render("(empty set)")
		}
		lines = append(lines, dimStyle.Render(fmt.Sprintf("Members: %d", len(m.PreviewValue.SetValue))))
		lines = append(lines, "")

		for i, v := range m.PreviewValue.SetValue {
			if i >= maxLines-2 {
				lines = append(lines, dimStyle.Render(fmt.Sprintf("... and %d more", len(m.PreviewValue.SetValue)-i)))
				break
			}
			val, _ := sanitizeBinaryString(v)
			if len(val) > maxWidth-4 {
				val = val[:maxWidth-7] + "..."
			}
			lines = append(lines, normalStyle.Render("• "+val))
		}

	case types.KeyTypeZSet:
		if len(m.PreviewValue.ZSetValue) == 0 {
			return dimStyle.Render("(empty sorted set)")
		}
		lines = append(lines, dimStyle.Render(fmt.Sprintf("Members: %d", len(m.PreviewValue.ZSetValue))))
		lines = append(lines, "")

		for i, v := range m.PreviewValue.ZSetValue {
			if i >= maxLines-2 {
				lines = append(lines, dimStyle.Render(fmt.Sprintf("... and %d more", len(m.PreviewValue.ZSetValue)-i)))
				break
			}
			member, _ := sanitizeBinaryString(v.Member)
			if len(member) > maxWidth-12 {
				member = member[:maxWidth-15] + "..."
			}
			score := lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Render(fmt.Sprintf("%.1f", v.Score))
			lines = append(lines, fmt.Sprintf("%s → %s", score, normalStyle.Render(member)))
		}

	case types.KeyTypeHash:
		if len(m.PreviewValue.HashValue) == 0 {
			return dimStyle.Render("(empty hash)")
		}

		// Sort keys for consistent display
		hashKeys := make([]string, 0, len(m.PreviewValue.HashValue))
		for k := range m.PreviewValue.HashValue {
			hashKeys = append(hashKeys, k)
		}
		sort.Strings(hashKeys)

		lines = append(lines, dimStyle.Render(fmt.Sprintf("Fields: %d", len(hashKeys))))
		lines = append(lines, "")

		for i, k := range hashKeys {
			if i >= maxLines-2 {
				lines = append(lines, dimStyle.Render(fmt.Sprintf("... and %d more", len(hashKeys)-i)))
				break
			}
			v := m.PreviewValue.HashValue[k]

			// Truncate key and value
			displayKey, _ := sanitizeBinaryString(k)
			if len(displayKey) > 15 {
				displayKey = displayKey[:12] + "..."
			}

			maxValLen := maxWidth - len(displayKey) - 5
			if maxValLen < 10 {
				maxValLen = 10
			}
			displayVal, _ := sanitizeBinaryString(v)
			if len(displayVal) > maxValLen {
				displayVal = displayVal[:maxValLen-3] + "..."
			}

			fieldName := lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Render(displayKey)
			lines = append(lines, fmt.Sprintf("%s: %s", fieldName, normalStyle.Render(displayVal)))
		}

	case types.KeyTypeStream:
		if len(m.PreviewValue.StreamValue) == 0 {
			return dimStyle.Render("(empty stream)")
		}
		lines = append(lines, dimStyle.Render(fmt.Sprintf("Entries: %d", len(m.PreviewValue.StreamValue))))
		lines = append(lines, "")

		for i, entry := range m.PreviewValue.StreamValue {
			if i >= maxLines-2 {
				lines = append(lines, dimStyle.Render(fmt.Sprintf("... and %d more", len(m.PreviewValue.StreamValue)-i)))
				break
			}

			// Format entry ID
			idStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("5"))

			// Format fields
			var fields []string
			for k, v := range entry.Fields {
				fields = append(fields, fmt.Sprintf("%s=%v", k, v))
			}
			fieldStr := strings.Join(fields, ", ")
			if len(fieldStr) > maxWidth-20 {
				fieldStr = fieldStr[:maxWidth-23] + "..."
			}

			lines = append(lines, fmt.Sprintf("%s %s", idStyle.Render(entry.ID), dimStyle.Render(fieldStr)))
		}

	default:
		return dimStyle.Render("(unknown type)")
	}

	return strings.Join(lines, "\n")
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
		seconds := int(m.CurrentKey.TTL.Seconds() + 0.5) // round to nearest second
		ttlStr = fmt.Sprintf("%ds", seconds)
		if seconds <= 10 {
			ttlDetailStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true) // Red - critical
			ttlStr = "⚠ " + ttlStr
		} else if seconds <= 60 {
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
		// Use pre-split lines for performance
		allLines := m.DetailLines
		maxLines := 20
		start := m.DetailScroll
		if start < 0 {
			start = 0
		}
		if start > len(allLines)-maxLines {
			start = len(allLines) - maxLines
		}
		if start < 0 {
			start = 0
		}
		end := start + maxLines
		if end > len(allLines) {
			end = len(allLines)
		}
		var displayLines []string
		if start > 0 {
			displayLines = append(displayLines, dimStyle.Render(fmt.Sprintf("↑ %d more lines", start)))
		}
		displayLines = append(displayLines, allLines[start:end]...)
		if end < len(allLines) {
			displayLines = append(displayLines, dimStyle.Render(fmt.Sprintf("↓ %d more lines", len(allLines)-end)))
		}
		valueContent = strings.Join(displayLines, "\n")
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
		helpText += "  e:edit  j/k:scroll"
	} else {
		helpText += "  a:add  x:remove"
	}
	helpText += "  esc:back"
	b.WriteString(helpStyle.Render(helpText))

	return b.String()
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

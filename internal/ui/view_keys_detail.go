package ui

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/davidbudnick/redis/internal/types"
)

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

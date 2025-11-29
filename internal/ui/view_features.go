package ui

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

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

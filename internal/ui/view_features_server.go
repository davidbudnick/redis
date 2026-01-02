package ui

import (
	"fmt"
	"strings"
)

func (m Model) viewClientList() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Connected Clients"))
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

	b.WriteString(titleStyle.Render("Memory Statistics"))
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

	b.WriteString(titleStyle.Render("Cluster Info"))
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

func (m Model) viewKeyspaceEvents() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Keyspace Events"))
	b.WriteString("\n\n")

	if m.KeyspaceSubActive {
		b.WriteString(successStyle.Render("* Listening for events..."))
	} else {
		b.WriteString(dimStyle.Render("o Not subscribed"))
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

package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// viewLiveMetrics renders the live metrics dashboard with ASCII charts
func (m Model) viewLiveMetrics() string {
	var b strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	b.WriteString(titleStyle.Render("📊 Live Metrics Dashboard"))
	b.WriteString("\n")

	connInfo := ""
	if m.CurrentConn != nil {
		connInfo = fmt.Sprintf("%s (%s:%d)", m.CurrentConn.Name, m.CurrentConn.Host, m.CurrentConn.Port)
	}
	b.WriteString(dimStyle.Render(connInfo))
	b.WriteString("\n")

	separatorWidth := m.Width - 10
	if separatorWidth < 20 {
		separatorWidth = 20
	}
	if separatorWidth > 60 {
		separatorWidth = 60
	}
	b.WriteString(dimStyle.Render(strings.Repeat("─", separatorWidth)))
	b.WriteString("\n\n")

	if m.LiveMetrics == nil || len(m.LiveMetrics.DataPoints) == 0 {
		b.WriteString(dimStyle.Render("Collecting metrics..."))
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render("Press q/esc to go back"))
		return b.String()
	}

	// Calculate available width for charts
	chartWidth := m.Width - 20
	if chartWidth < 30 {
		chartWidth = 30
	}
	if chartWidth > 100 {
		chartWidth = 100
	}

	// Get latest data point for current values
	latest := m.LiveMetrics.DataPoints[len(m.LiveMetrics.DataPoints)-1]

	// Current stats in a nice grid layout
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	valueStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))

	// Calculate hit rate
	hitRate := float64(0)
	if latest.KeyspaceHits+latest.KeyspaceMisses > 0 {
		hitRate = float64(latest.KeyspaceHits) / float64(latest.KeyspaceHits+latest.KeyspaceMisses) * 100
	}

	// Stats grid - 3 columns
	col1 := fmt.Sprintf("%s %s\n%s %s\n%s %s",
		labelStyle.Render("Ops/sec:"),
		valueStyle.Render(fmt.Sprintf("%6.0f", latest.OpsPerSec)),
		labelStyle.Render("Clients:"),
		valueStyle.Render(fmt.Sprintf("%6d", latest.ConnectedClients)),
		labelStyle.Render("Hit Rate:"),
		valueStyle.Render(fmt.Sprintf("%5.1f%%", hitRate)),
	)

	col2 := fmt.Sprintf("%s %s\n%s %s\n%s %s",
		labelStyle.Render("Memory:"),
		valueStyle.Render(fmt.Sprintf("%8s", formatBytes(latest.UsedMemoryBytes))),
		labelStyle.Render("Net In:"),
		valueStyle.Render(fmt.Sprintf("%6.2f KB/s", latest.InputKbps)),
		labelStyle.Render("Expired:"),
		valueStyle.Render(fmt.Sprintf("%8d", latest.ExpiredKeys)),
	)

	col3 := fmt.Sprintf("%s %s\n%s %s\n%s %s",
		labelStyle.Render("Blocked:"),
		valueStyle.Render(fmt.Sprintf("%6d", latest.BlockedClients)),
		labelStyle.Render("Net Out:"),
		valueStyle.Render(fmt.Sprintf("%6.2f KB/s", latest.OutputKbps)),
		labelStyle.Render("Evicted:"),
		valueStyle.Render(fmt.Sprintf("%8d", latest.EvictedKeys)),
	)

	colStyle := lipgloss.NewStyle().Width(22)
	statsRow := lipgloss.JoinHorizontal(lipgloss.Top,
		colStyle.Render(col1),
		colStyle.Render(col2),
		colStyle.Render(col3),
	)
	b.WriteString(statsRow)
	b.WriteString("\n\n")

	// Charts section
	b.WriteString(dimStyle.Render(strings.Repeat("─", 60)))
	b.WriteString("\n\n")

	// Ops/sec line chart
	opsData := make([]float64, len(m.LiveMetrics.DataPoints))
	for i, dp := range m.LiveMetrics.DataPoints {
		opsData[i] = dp.OpsPerSec
	}
	b.WriteString(renderLineChart("Ops/sec", opsData, chartWidth, 6, lipgloss.Color("39")))

	// Memory line chart
	memData := make([]float64, len(m.LiveMetrics.DataPoints))
	for i, dp := range m.LiveMetrics.DataPoints {
		memData[i] = float64(dp.UsedMemoryBytes) / 1024 / 1024 // Convert to MB
	}
	b.WriteString(renderLineChart("Memory (MB)", memData, chartWidth, 6, lipgloss.Color("35")))

	// Network I/O line chart
	netData := make([]float64, len(m.LiveMetrics.DataPoints))
	for i, dp := range m.LiveMetrics.DataPoints {
		netData[i] = dp.InputKbps + dp.OutputKbps
	}
	b.WriteString(renderLineChart("Network KB/s", netData, chartWidth, 5, lipgloss.Color("33")))

	// Clients line chart
	clientsData := make([]float64, len(m.LiveMetrics.DataPoints))
	for i, dp := range m.LiveMetrics.DataPoints {
		clientsData[i] = float64(dp.ConnectedClients)
	}
	b.WriteString(renderLineChart("Clients", clientsData, chartWidth, 5, lipgloss.Color("32")))

	b.WriteString(helpStyle.Render("Auto-refreshing • c:clear • q/esc:back"))

	return b.String()
}

// renderLineChart creates a bar chart using block characters
func renderLineChart(title string, data []float64, width, height int, color lipgloss.Color) string {
	if len(data) == 0 {
		return ""
	}

	var b strings.Builder

	// Find min/max for scaling
	minVal, maxVal := data[0], data[0]
	for _, v := range data {
		if v < minVal {
			minVal = v
		}
		if v > maxVal {
			maxVal = v
		}
	}

	// Ensure we have a range
	if maxVal == minVal {
		maxVal = minVal + 1
	}
	rangeVal := maxVal - minVal

	// Current value
	current := data[len(data)-1]

	// Title with current/max values
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(color)
	infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))

	b.WriteString(titleStyle.Render(title))
	b.WriteString(infoStyle.Render(fmt.Sprintf("  %.1f", current)))
	b.WriteString(dimStyle.Render(fmt.Sprintf(" (max: %.1f)", maxVal)))
	b.WriteString("\n")

	// Resample data to fit width
	chartData := resampleData(data, width)

	// Block characters for bar heights (from empty to full)
	blocks := []rune{' ', '▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}
	chartStyle := lipgloss.NewStyle().Foreground(color)

	// Render the chart row by row from top to bottom
	for row := height - 1; row >= 0; row-- {
		for _, val := range chartData {
			// Normalize value to 0-1
			normalized := (val - minVal) / rangeVal
			// Total height in "sub-rows" (each character has 8 levels)
			totalSubRows := normalized * float64(height) * 8.0
			// How many full rows below this row
			fullRowsBelow := float64(row) * 8.0

			if totalSubRows >= fullRowsBelow+8 {
				// This row is fully filled
				b.WriteString(chartStyle.Render("█"))
			} else if totalSubRows > fullRowsBelow {
				// This row is partially filled
				partialFill := int(totalSubRows - fullRowsBelow)
				if partialFill > 7 {
					partialFill = 7
				}
				b.WriteString(chartStyle.Render(string(blocks[partialFill])))
			} else {
				// This row is empty
				b.WriteString(" ")
			}
		}
		b.WriteString("\n")
	}

	// Bottom axis
	b.WriteString(dimStyle.Render(strings.Repeat("─", width)))
	b.WriteString("\n")

	return b.String()
}

// resampleData resamples data to fit the target width
func resampleData(data []float64, targetWidth int) []float64 {
	if len(data) == 0 {
		return data
	}
	if len(data) <= targetWidth {
		// Pad with the same values
		result := make([]float64, targetWidth)
		for i := range result {
			idx := i * len(data) / targetWidth
			if idx >= len(data) {
				idx = len(data) - 1
			}
			result[i] = data[idx]
		}
		return result
	}

	// Downsample
	result := make([]float64, targetWidth)
	for i := range result {
		startIdx := i * len(data) / targetWidth
		endIdx := (i + 1) * len(data) / targetWidth
		if endIdx > len(data) {
			endIdx = len(data)
		}
		if startIdx >= endIdx {
			startIdx = endIdx - 1
		}
		if startIdx < 0 {
			startIdx = 0
		}

		// Average the values in this bucket
		sum := 0.0
		count := 0
		for j := startIdx; j < endIdx; j++ {
			sum += data[j]
			count++
		}
		if count > 0 {
			result[i] = sum / float64(count)
		}
	}
	return result
}

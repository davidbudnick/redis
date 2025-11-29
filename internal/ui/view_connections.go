package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Logo style
	logoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)

	// Accent colors
	accentStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("39"))

	// Stats box style
	statsBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(0, 1)

	// Connection card style
	connCardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(0, 1).
			MarginBottom(0)

	connCardSelectedStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("39")).
				Padding(0, 1).
				MarginBottom(0)
)

func (m Model) viewConnections() string {
	var b strings.Builder

	// ASCII Art Logo
	logo := `
 ██████╗ ███████╗██████╗ ██╗███████╗
 ██╔══██╗██╔════╝██╔══██╗██║██╔════╝
 ██████╔╝█████╗  ██║  ██║██║███████╗
 ██╔══██╗██╔══╝  ██║  ██║██║╚════██║
 ██║  ██║███████╗██████╔╝██║███████║
 ╚═╝  ╚═╝╚══════╝╚═════╝ ╚═╝╚══════╝`

	b.WriteString(logoStyle.Render(logo))
	b.WriteString("\n\n")

	// Stats bar
	statsContent := m.buildStatsBar()
	b.WriteString(statsContent)
	b.WriteString("\n\n")

	// Connection error display (prominent error box)
	if m.ConnectionError != "" {
		errorBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("196")).
			Foreground(lipgloss.Color("196")).
			Padding(0, 2).
			Width(55).
			Render(fmt.Sprintf("Connection Failed\n%s", dimStyle.Render(m.ConnectionError)))
		b.WriteString(errorBox)
		b.WriteString("\n\n")
	}

	// Section title
	connCount := len(m.Connections)
	sectionTitle := fmt.Sprintf("╭─ Saved Connections (%d) ", connCount)
	sectionTitle += strings.Repeat("─", 50-len(sectionTitle)) + "╮"
	b.WriteString(accentStyle.Render(sectionTitle))
	b.WriteString("\n")

	if len(m.Connections) == 0 {
		b.WriteString("\n")
		emptyBox := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Padding(1, 2).
			Render("  No connections saved.\n\n  Press 'a' to add your first Redis connection.")
		b.WriteString(emptyBox)
		b.WriteString("\n")
	} else {
		b.WriteString("\n")

		// Calculate visible range for scrolling
		maxVisible := (m.Height - 20) / 3
		if maxVisible < 3 {
			maxVisible = 3
		}
		startIdx := 0
		if m.SelectedConnIdx >= maxVisible {
			startIdx = m.SelectedConnIdx - maxVisible + 1
		}
		endIdx := startIdx + maxVisible
		if endIdx > len(m.Connections) {
			endIdx = len(m.Connections)
		}

		for i := startIdx; i < endIdx; i++ {
			conn := m.Connections[i]
			isSelected := i == m.SelectedConnIdx

			// Build connection card content
			var card strings.Builder

			// Connection name with icon
			icon := "○"
			if isSelected {
				icon = "●"
			}

			nameStyle := normalStyle
			if isSelected {
				nameStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true)
			}
			card.WriteString(fmt.Sprintf(" %s %s", icon, nameStyle.Render(conn.Name)))
			card.WriteString("\n")

			// Connection details
			hostPort := fmt.Sprintf("   %s:%d", conn.Host, conn.Port)
			card.WriteString(dimStyle.Render(hostPort))

			// Database badge
			dbBadge := lipgloss.NewStyle().
				Background(lipgloss.Color("236")).
				Foreground(lipgloss.Color("245")).
				Padding(0, 1).
				Render(fmt.Sprintf("db%d", conn.DB))
			card.WriteString("  ")
			card.WriteString(dbBadge)

			// TLS indicator
			if conn.UseTLS {
				tlsBadge := lipgloss.NewStyle().
					Background(lipgloss.Color("22")).
					Foreground(lipgloss.Color("46")).
					Padding(0, 1).
					Render("TLS")
				card.WriteString(" ")
				card.WriteString(tlsBadge)
			}

			// Render the card with appropriate style
			cardStyle := connCardStyle
			if isSelected {
				cardStyle = connCardSelectedStyle
			}

			// Set card width
			cardWidth := 55
			if m.Width-10 < cardWidth {
				cardWidth = m.Width - 10
			}
			cardStyle = cardStyle.Width(cardWidth)

			b.WriteString(cardStyle.Render(card.String()))
			b.WriteString("\n")
		}

		// Scroll indicator
		if len(m.Connections) > maxVisible {
			scrollInfo := fmt.Sprintf("  ↕ %d-%d of %d connections", startIdx+1, endIdx, len(m.Connections))
			b.WriteString(dimStyle.Render(scrollInfo))
			b.WriteString("\n")
		}
	}

	// Bottom section line
	sectionBottom := "╰" + strings.Repeat("─", 54) + "╯"
	b.WriteString(accentStyle.Render(sectionBottom))
	b.WriteString("\n\n")

	// Keybindings footer
	keybindings := []struct {
		key  string
		desc string
	}{
		{"↑/↓", "navigate"},
		{"enter", "connect"},
		{"a", "add"},
		{"e", "edit"},
		{"d", "delete"},
		{"q", "quit"},
	}

	var keyHelp strings.Builder
	for i, kb := range keybindings {
		keyStyle := lipgloss.NewStyle().
			Background(lipgloss.Color("236")).
			Foreground(lipgloss.Color("255")).
			Padding(0, 1).
			Render(kb.key)
		keyHelp.WriteString(keyStyle)
		keyHelp.WriteString(" ")
		keyHelp.WriteString(dimStyle.Render(kb.desc))
		if i < len(keybindings)-1 {
			keyHelp.WriteString("  ")
		}
	}
	b.WriteString(keyHelp.String())

	return b.String()
}

func (m Model) buildStatsBar() string {
	// Create stats boxes
	boxes := []struct {
		label string
		value string
		color string
	}{
		{"Connections", fmt.Sprintf("%d saved", len(m.Connections)), "39"},
		{"Time", time.Now().Format("15:04:05"), "245"},
	}

	var statsBoxes []string
	for _, box := range boxes {
		content := fmt.Sprintf("%s\n%s",
			dimStyle.Render(box.label),
			lipgloss.NewStyle().Foreground(lipgloss.Color(box.color)).Bold(true).Render(box.value),
		)
		styled := statsBoxStyle.Width(18).Render(content)
		statsBoxes = append(statsBoxes, styled)
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, statsBoxes...)
}

func (m Model) viewAddConnection() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Add Connection"))
	b.WriteString("\n\n")

	labels := []string{"Name", "Host", "Port", "Password", "Database"}

	for i, input := range m.ConnInputs {
		labelStyle := keyStyle
		if m.ConnFocusIdx == i {
			labelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true)
		}
		b.WriteString(labelStyle.Render(labels[i] + ":"))
		b.WriteString("\n")
		b.WriteString(input.View())
		b.WriteString("\n\n")
	}

	// Action buttons hint
	actions := lipgloss.NewStyle().
		Background(lipgloss.Color("22")).
		Foreground(lipgloss.Color("46")).
		Padding(0, 1).
		Render("Ctrl+T: Test")
	b.WriteString(actions)
	b.WriteString("\n\n")

	b.WriteString(helpStyle.Render("tab:next  enter:save  esc:cancel"))

	modalWidth := 55
	if m.Width-10 < 55 {
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
		labelStyle := keyStyle
		if m.ConnFocusIdx == i {
			labelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true)
		}
		b.WriteString(labelStyle.Render(labels[i] + ":"))
		b.WriteString("\n")
		b.WriteString(input.View())
		b.WriteString("\n\n")
	}

	b.WriteString(helpStyle.Render("tab:next  enter:save  esc:cancel"))

	modalWidth := 55
	if m.Width-10 < 55 {
		modalWidth = m.Width - 10
	}
	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("39")).
		Padding(1, 2).
		Width(modalWidth)

	return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, modalStyle.Render(b.String()))
}

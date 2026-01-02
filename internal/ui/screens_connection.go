package ui

import (
	"github.com/davidbudnick/redis/internal/cmd"
	"github.com/davidbudnick/redis/internal/types"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) handleConnectionsScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.SelectedConnIdx > 0 {
			m.SelectedConnIdx--
			m.ConnectionError = "" // Clear error on navigation
		}
	case "down", "j":
		if m.SelectedConnIdx < len(m.Connections)-1 {
			m.SelectedConnIdx++
			m.ConnectionError = "" // Clear error on navigation
		}
	case "enter":
		if len(m.Connections) > 0 && m.SelectedConnIdx < len(m.Connections) {
			conn := m.Connections[m.SelectedConnIdx]
			m.CurrentConn = &conn
			m.Loading = true
			m.StatusMsg = "Connecting..."
			m.ConnectionError = "" // Clear any previous connection error
			return m, cmd.ConnectCmd(conn.Host, conn.Port, conn.Password, conn.DB)
		}
	case "a", "n":
		m.Screen = types.ScreenAddConnection
		m.resetConnInputs()
	case "e":
		if len(m.Connections) > 0 && m.SelectedConnIdx < len(m.Connections) {
			conn := m.Connections[m.SelectedConnIdx]
			m.EditingConnection = &conn
			m.populateConnInputs(conn)
			m.Screen = types.ScreenEditConnection
		}
	case "d", "delete", "backspace":
		if len(m.Connections) > 0 && m.SelectedConnIdx < len(m.Connections) {
			m.ConfirmType = "connection"
			m.ConfirmData = m.Connections[m.SelectedConnIdx]
			m.Screen = types.ScreenConfirmDelete
		}
	case "r":
		m.Loading = true
		return m, cmd.LoadConnectionsCmd()
	}
	return m, nil
}

func (m Model) handleAddConnectionScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "tab", "down":
		m.ConnInputs[m.ConnFocusIdx].Blur()
		m.ConnFocusIdx = (m.ConnFocusIdx + 1) % len(m.ConnInputs)
		m.ConnInputs[m.ConnFocusIdx].Focus()
	case "shift+tab", "up":
		m.ConnInputs[m.ConnFocusIdx].Blur()
		m.ConnFocusIdx--
		if m.ConnFocusIdx < 0 {
			m.ConnFocusIdx = len(m.ConnInputs) - 1
		}
		m.ConnInputs[m.ConnFocusIdx].Focus()
	case "enter":
		if m.ConnInputs[0].Value() != "" && m.ConnInputs[1].Value() != "" {
			m.Loading = true
			return m, cmd.AddConnectionCmd(
				m.ConnInputs[0].Value(),
				m.ConnInputs[1].Value(),
				m.getPort(),
				m.ConnInputs[3].Value(),
				m.getDB(),
			)
		}
	case "ctrl+t":
		m.Loading = true
		m.Screen = types.ScreenTestConnection
		return m, cmd.TestConnectionCmd(
			m.ConnInputs[1].Value(),
			m.getPort(),
			m.ConnInputs[3].Value(),
			m.getDB(),
		)
	case "esc":
		m.Screen = types.ScreenConnections
		m.resetConnInputs()
	default:
		var cmds []tea.Cmd
		for i := range m.ConnInputs {
			var inputCmd tea.Cmd
			m.ConnInputs[i], inputCmd = m.ConnInputs[i].Update(msg)
			cmds = append(cmds, inputCmd)
		}
		return m, tea.Batch(cmds...)
	}
	return m, nil
}

func (m Model) handleEditConnectionScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "tab", "down":
		m.ConnInputs[m.ConnFocusIdx].Blur()
		m.ConnFocusIdx = (m.ConnFocusIdx + 1) % len(m.ConnInputs)
		m.ConnInputs[m.ConnFocusIdx].Focus()
	case "shift+tab", "up":
		m.ConnInputs[m.ConnFocusIdx].Blur()
		m.ConnFocusIdx--
		if m.ConnFocusIdx < 0 {
			m.ConnFocusIdx = len(m.ConnInputs) - 1
		}
		m.ConnInputs[m.ConnFocusIdx].Focus()
	case "enter":
		if m.EditingConnection != nil && m.ConnInputs[0].Value() != "" && m.ConnInputs[1].Value() != "" {
			m.Loading = true
			return m, cmd.UpdateConnectionCmd(
				m.EditingConnection.ID,
				m.ConnInputs[0].Value(),
				m.ConnInputs[1].Value(),
				m.getPort(),
				m.ConnInputs[3].Value(),
				m.getDB(),
			)
		}
	case "esc":
		m.Screen = types.ScreenConnections
		m.EditingConnection = nil
		m.resetConnInputs()
	default:
		var cmds []tea.Cmd
		for i := range m.ConnInputs {
			var inputCmd tea.Cmd
			m.ConnInputs[i], inputCmd = m.ConnInputs[i].Update(msg)
			cmds = append(cmds, inputCmd)
		}
		return m, tea.Batch(cmds...)
	}
	return m, nil
}

func (m Model) handleTestConnectionScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "enter":
		m.Screen = types.ScreenAddConnection
	}
	return m, nil
}

package ui

import (
	"strconv"

	"github.com/davidbudnick/redis-tui/internal/cmd"
	"github.com/davidbudnick/redis-tui/internal/types"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) handleHelpScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "enter", "?":
		if m.CurrentConn != nil {
			m.Screen = types.ScreenKeys
		} else {
			m.Screen = types.ScreenConnections
		}
	}
	return m, nil
}

func (m Model) handleServerInfoScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "enter":
		m.Screen = types.ScreenKeys
	case "r":
		m.Loading = true
		return m, cmd.LoadServerInfoCmd()
	}
	return m, nil
}

func (m Model) handlePubSubScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "tab":
		m.PubSubInput[m.PubSubFocusIdx].Blur()
		m.PubSubFocusIdx = (m.PubSubFocusIdx + 1) % len(m.PubSubInput)
		m.PubSubInput[m.PubSubFocusIdx].Focus()
	case "enter":
		if m.PubSubInput[0].Value() != "" && m.PubSubInput[1].Value() != "" {
			m.Loading = true
			return m, cmd.PublishMessageCmd(m.PubSubInput[0].Value(), m.PubSubInput[1].Value())
		}
	case "esc":
		m.Screen = types.ScreenKeys
		m.resetPubSubInputs()
	default:
		var cmds []tea.Cmd
		for i := range m.PubSubInput {
			var inputCmd tea.Cmd
			m.PubSubInput[i], inputCmd = m.PubSubInput[i].Update(msg)
			cmds = append(cmds, inputCmd)
		}
		return m, tea.Batch(cmds...)
	}
	return m, nil
}

func (m Model) handlePublishMessageScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m.handlePubSubScreen(msg)
}

func (m Model) handleSwitchDBScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		dbNum, err := strconv.Atoi(m.DBSwitchInput.Value())
		if err == nil && dbNum >= 0 && dbNum <= 15 {
			m.Loading = true
			return m, cmd.SwitchDBCmd(dbNum)
		} else {
			m.StatusMsg = "Invalid database number (0-15)"
		}
	case "esc":
		m.Screen = types.ScreenKeys
		m.DBSwitchInput.Blur()
	default:
		var inputCmd tea.Cmd
		m.DBSwitchInput, inputCmd = m.DBSwitchInput.Update(msg)
		return m, inputCmd
	}
	return m, nil
}

func (m Model) handleExportScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		if m.ExportInput.Value() != "" {
			m.Loading = true
			pattern := m.KeyPattern
			if pattern == "" {
				pattern = "*"
			}
			return m, cmd.ExportKeysCmd(pattern, m.ExportInput.Value())
		}
	case "esc":
		m.Screen = types.ScreenKeys
		m.ExportInput.Blur()
	default:
		var inputCmd tea.Cmd
		m.ExportInput, inputCmd = m.ExportInput.Update(msg)
		return m, inputCmd
	}
	return m, nil
}

func (m Model) handleImportScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		if m.ImportInput.Value() != "" {
			m.Loading = true
			return m, cmd.ImportKeysCmd(m.ImportInput.Value())
		}
	case "esc":
		m.Screen = types.ScreenKeys
		m.ImportInput.Blur()
	default:
		var inputCmd tea.Cmd
		m.ImportInput, inputCmd = m.ImportInput.Update(msg)
		return m, inputCmd
	}
	return m, nil
}

func (m Model) handleSlowLogScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "enter":
		m.Screen = types.ScreenKeys
	case "r":
		m.Loading = true
		return m, cmd.GetSlowLogCmd(20)
	}
	return m, nil
}

func (m Model) handleLuaScriptScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		if m.LuaScriptInput.Value() != "" {
			m.Loading = true
			return m, cmd.EvalLuaScriptCmd(m.LuaScriptInput.Value(), []string{})
		}
	case "esc":
		m.Screen = types.ScreenKeys
		m.LuaScriptInput.Blur()
	default:
		var inputCmd tea.Cmd
		m.LuaScriptInput, inputCmd = m.LuaScriptInput.Update(msg)
		return m, inputCmd
	}
	return m, nil
}

func (m Model) handleLogsScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.ShowingLogDetail {
		switch msg.String() {
		case "esc", "enter":
			m.ShowingLogDetail = false
		}
		return m, nil
	}

	logCount := 0
	if m.Logs != nil {
		logCount = len(*m.Logs)
	}

	switch msg.String() {
	case "esc":
		m.Screen = types.ScreenKeys
	case "up", "k":
		if m.LogCursor > 0 {
			m.LogCursor--
		}
	case "down", "j":
		if m.LogCursor < logCount-1 {
			m.LogCursor++
		}
	case "enter":
		if logCount > 0 {
			m.ShowingLogDetail = true
		}
	case "g":
		m.LogCursor = 0
	case "G":
		if logCount > 0 {
			m.LogCursor = logCount - 1
		}
	}
	return m, nil
}

func (m Model) handleClientListScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.SelectedClientIdx > 0 {
			m.SelectedClientIdx--
		}
	case "down", "j":
		if m.SelectedClientIdx < len(m.ClientList)-1 {
			m.SelectedClientIdx++
		}
	case "r":
		m.Loading = true
		return m, cmd.GetClientListCmd()
	case "esc":
		m.Screen = types.ScreenKeys
	}
	return m, nil
}

func (m Model) handleMemoryStatsScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "r":
		m.Loading = true
		return m, cmd.GetMemoryStatsCmd()
	case "esc":
		m.Screen = types.ScreenKeys
	}
	return m, nil
}

func (m Model) handleClusterInfoScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.SelectedNodeIdx > 0 {
			m.SelectedNodeIdx--
		}
	case "down", "j":
		if m.SelectedNodeIdx < len(m.ClusterNodes)-1 {
			m.SelectedNodeIdx++
		}
	case "r":
		m.Loading = true
		return m, cmd.GetClusterInfoCmd()
	case "esc":
		m.Screen = types.ScreenKeys
	}
	return m, nil
}

func (m Model) handleLiveMetricsScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "c":
		if m.LiveMetrics != nil {
			m.LiveMetrics.DataPoints = nil
		}
	case "q", "esc":
		m.LiveMetricsActive = false
		m.Screen = types.ScreenKeys
	}
	return m, nil
}

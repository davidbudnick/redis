package ui

import (
	"github.com/davidbudnick/redis/internal/cmd"
	"github.com/davidbudnick/redis/internal/types"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) handleSearchValuesScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		if m.SearchValueInput.Value() != "" {
			m.Loading = true
			pattern := m.KeyPattern
			if pattern == "" {
				pattern = "*"
			}
			m.Screen = types.ScreenKeys
			return m, cmd.SearchByValueCmd(pattern, m.SearchValueInput.Value(), 100)
		}
	case "esc":
		m.Screen = types.ScreenKeys
		m.SearchValueInput.Blur()
	default:
		var inputCmd tea.Cmd
		m.SearchValueInput, inputCmd = m.SearchValueInput.Update(msg)
		return m, inputCmd
	}
	return m, nil
}

func (m Model) handleRegexSearchScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		if m.RegexSearchInput.Value() != "" {
			m.Loading = true
			return m, cmd.RegexSearchCmd(m.RegexSearchInput.Value(), 100)
		}
	case "esc":
		m.Screen = types.ScreenKeys
		m.RegexSearchInput.Blur()
	default:
		var inputCmd tea.Cmd
		m.RegexSearchInput, inputCmd = m.RegexSearchInput.Update(msg)
		return m, inputCmd
	}
	return m, nil
}

func (m Model) handleFuzzySearchScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		if m.FuzzySearchInput.Value() != "" {
			m.Loading = true
			return m, cmd.FuzzySearchCmd(m.FuzzySearchInput.Value(), 100)
		}
	case "esc":
		m.Screen = types.ScreenKeys
		m.FuzzySearchInput.Blur()
	default:
		var inputCmd tea.Cmd
		m.FuzzySearchInput, inputCmd = m.FuzzySearchInput.Update(msg)
		return m, inputCmd
	}
	return m, nil
}

func (m Model) handleCompareKeysScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "tab":
		if m.CompareFocusIdx == 0 {
			m.CompareKey1Input.Blur()
			m.CompareKey2Input.Focus()
			m.CompareFocusIdx = 1
		} else {
			m.CompareKey2Input.Blur()
			m.CompareKey1Input.Focus()
			m.CompareFocusIdx = 0
		}
	case "enter":
		if m.CompareKey1Input.Value() != "" && m.CompareKey2Input.Value() != "" {
			m.Loading = true
			return m, cmd.CompareKeysCmd(m.CompareKey1Input.Value(), m.CompareKey2Input.Value())
		}
	case "esc":
		m.Screen = types.ScreenKeys
		m.CompareKey1Input.Blur()
		m.CompareKey2Input.Blur()
		m.CompareResult = nil
	default:
		if m.CompareFocusIdx == 0 {
			var inputCmd tea.Cmd
			m.CompareKey1Input, inputCmd = m.CompareKey1Input.Update(msg)
			return m, inputCmd
		}
		var inputCmd tea.Cmd
		m.CompareKey2Input, inputCmd = m.CompareKey2Input.Update(msg)
		return m, inputCmd
	}
	return m, nil
}

func (m Model) handleJSONPathScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		// TODO: Implement JSON path query
		if m.JSONPathInput.Value() != "" {
			m.JSONPathResult = "JSON path queries not yet implemented"
		}
	case "esc":
		m.Screen = types.ScreenKeyDetail
		m.JSONPathInput.Blur()
	default:
		var inputCmd tea.Cmd
		m.JSONPathInput, inputCmd = m.JSONPathInput.Update(msg)
		return m, inputCmd
	}
	return m, nil
}

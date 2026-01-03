package ui

import (
	"sort"
	"strconv"
	"time"

	"github.com/davidbudnick/redis-tui/internal/cmd"
	"github.com/davidbudnick/redis-tui/internal/types"
	"github.com/kujtimiihoxha/vimtea"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) handleAddKeyScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "tab":
		m.AddKeyInputs[m.AddKeyFocusIdx].Blur()
		m.AddKeyFocusIdx = (m.AddKeyFocusIdx + 1) % len(m.AddKeyInputs)
		m.AddKeyInputs[m.AddKeyFocusIdx].Focus()
	case "shift+tab":
		m.AddKeyInputs[m.AddKeyFocusIdx].Blur()
		m.AddKeyFocusIdx--
		if m.AddKeyFocusIdx < 0 {
			m.AddKeyFocusIdx = len(m.AddKeyInputs) - 1
		}
		m.AddKeyInputs[m.AddKeyFocusIdx].Focus()
	case "ctrl+t":
		typeOrder := []types.KeyType{
			types.KeyTypeString, types.KeyTypeList, types.KeyTypeSet,
			types.KeyTypeZSet, types.KeyTypeHash, types.KeyTypeStream,
		}
		for i, t := range typeOrder {
			if t == m.AddKeyType {
				m.AddKeyType = typeOrder[(i+1)%len(typeOrder)]
				break
			}
		}
	case "enter":
		if m.AddKeyInputs[0].Value() != "" {
			m.Loading = true
			return m, cmd.CreateKeyCmd(
				m.AddKeyInputs[0].Value(),
				m.AddKeyType,
				m.AddKeyInputs[1].Value(),
				0,
			)
		}
	case "esc":
		m.Screen = types.ScreenKeys
		m.resetAddKeyInputs()
	default:
		var cmds []tea.Cmd
		for i := range m.AddKeyInputs {
			var inputCmd tea.Cmd
			m.AddKeyInputs[i], inputCmd = m.AddKeyInputs[i].Update(msg)
			cmds = append(cmds, inputCmd)
		}
		return m, tea.Batch(cmds...)
	}
	return m, nil
}

func (m Model) handleConfirmDeleteScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y", "enter":
		m.Loading = true
		switch m.ConfirmType {
		case "connection":
			if conn, ok := m.ConfirmData.(types.Connection); ok {
				return m, cmd.DeleteConnectionCmd(conn.ID)
			}
		case "key":
			if key, ok := m.ConfirmData.(types.RedisKey); ok {
				return m, cmd.DeleteKeyCmd(key.Key)
			}
		case "flushdb":
			return m, cmd.FlushDBCmd()
		}
	case "n", "N", "esc":
		switch m.ConfirmType {
		case "connection":
			m.Screen = types.ScreenConnections
		case "key":
			m.Screen = types.ScreenKeyDetail
		case "flushdb":
			m.Screen = types.ScreenKeys
		}
	}
	return m, nil
}

func (m Model) handleTTLEditorScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		if m.CurrentKey != nil {
			ttlSecs, _ := strconv.Atoi(m.TTLInput.Value())
			ttl := time.Duration(ttlSecs) * time.Second
			m.Loading = true
			return m, cmd.SetTTLCmd(m.CurrentKey.Key, ttl)
		}
	case "esc":
		m.Screen = types.ScreenKeyDetail
		m.TTLInput.Blur()
	default:
		var inputCmd tea.Cmd
		m.TTLInput, inputCmd = m.TTLInput.Update(msg)
		return m, inputCmd
	}
	return m, nil
}

func (m Model) handleEditValueScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle save (ctrl+s) and quit (ctrl+q) globally
	switch msg.String() {
	case "ctrl+s":
		if m.CurrentKey != nil {
			m.Loading = true
			content := m.VimEditor.GetBuffer().Text()
			return m, cmd.EditStringValueCmd(m.CurrentKey.Key, content)
		}
	case "ctrl+q":
		m.Screen = types.ScreenKeyDetail
		return m, nil
	}

	// Delegate everything else to vimtea
	if m.VimEditor != nil {
		updated, editorCmd := m.VimEditor.Update(msg)
		m.VimEditor = updated.(vimtea.Editor)
		return m, editorCmd
	}
	return m, nil
}

func (m Model) handleAddToCollectionScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "tab":
		m.AddCollectionInput[m.AddCollFocusIdx].Blur()
		m.AddCollFocusIdx = (m.AddCollFocusIdx + 1) % len(m.AddCollectionInput)
		m.AddCollectionInput[m.AddCollFocusIdx].Focus()
	case "shift+tab":
		m.AddCollectionInput[m.AddCollFocusIdx].Blur()
		m.AddCollFocusIdx--
		if m.AddCollFocusIdx < 0 {
			m.AddCollFocusIdx = len(m.AddCollectionInput) - 1
		}
		m.AddCollectionInput[m.AddCollFocusIdx].Focus()
	case "enter":
		if m.CurrentKey != nil && m.AddCollectionInput[0].Value() != "" {
			m.Loading = true
			value := m.AddCollectionInput[0].Value()
			extra := m.AddCollectionInput[1].Value()

			switch m.CurrentKey.Type {
			case types.KeyTypeList:
				return m, cmd.AddToListCmd(m.CurrentKey.Key, value)
			case types.KeyTypeSet:
				return m, cmd.AddToSetCmd(m.CurrentKey.Key, value)
			case types.KeyTypeZSet:
				score := 0.0
				if extra != "" {
					score, _ = strconv.ParseFloat(extra, 64)
				}
				return m, cmd.AddToZSetCmd(m.CurrentKey.Key, score, value)
			case types.KeyTypeHash:
				if extra == "" {
					extra = "value"
				}
				return m, cmd.AddToHashCmd(m.CurrentKey.Key, value, extra)
			case types.KeyTypeStream:
				fields := map[string]interface{}{value: extra}
				return m, cmd.AddToStreamCmd(m.CurrentKey.Key, fields)
			}
		}
	case "esc":
		m.Screen = types.ScreenKeyDetail
		m.resetAddCollectionInputs()
	default:
		var cmds []tea.Cmd
		for i := range m.AddCollectionInput {
			var inputCmd tea.Cmd
			m.AddCollectionInput[i], inputCmd = m.AddCollectionInput[i].Update(msg)
			cmds = append(cmds, inputCmd)
		}
		return m, tea.Batch(cmds...)
	}
	return m, nil
}

func (m Model) handleRemoveFromCollectionScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.SelectedItemIdx > 0 {
			m.SelectedItemIdx--
		}
	case "down", "j":
		maxIdx := m.getCollectionLength() - 1
		if m.SelectedItemIdx < maxIdx {
			m.SelectedItemIdx++
		}
	case "enter", "d":
		if m.CurrentKey != nil {
			m.Loading = true
			switch m.CurrentKey.Type {
			case types.KeyTypeList:
				if m.SelectedItemIdx < len(m.CurrentValue.ListValue) {
					return m, cmd.RemoveFromListCmd(m.CurrentKey.Key, m.CurrentValue.ListValue[m.SelectedItemIdx])
				}
			case types.KeyTypeSet:
				if m.SelectedItemIdx < len(m.CurrentValue.SetValue) {
					return m, cmd.RemoveFromSetCmd(m.CurrentKey.Key, m.CurrentValue.SetValue[m.SelectedItemIdx])
				}
			case types.KeyTypeZSet:
				if m.SelectedItemIdx < len(m.CurrentValue.ZSetValue) {
					return m, cmd.RemoveFromZSetCmd(m.CurrentKey.Key, m.CurrentValue.ZSetValue[m.SelectedItemIdx].Member)
				}
			case types.KeyTypeHash:
				keys := make([]string, 0, len(m.CurrentValue.HashValue))
				for k := range m.CurrentValue.HashValue {
					keys = append(keys, k)
				}
				sort.Strings(keys)
				if m.SelectedItemIdx < len(keys) {
					return m, cmd.RemoveFromHashCmd(m.CurrentKey.Key, keys[m.SelectedItemIdx])
				}
			case types.KeyTypeStream:
				if m.SelectedItemIdx < len(m.CurrentValue.StreamValue) {
					return m, cmd.RemoveFromStreamCmd(m.CurrentKey.Key, m.CurrentValue.StreamValue[m.SelectedItemIdx].ID)
				}
			}
		}
	case "esc":
		m.Screen = types.ScreenKeyDetail
	}
	return m, nil
}

func (m Model) handleRenameKeyScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		if m.CurrentKey != nil && m.RenameInput.Value() != "" && m.RenameInput.Value() != m.CurrentKey.Key {
			m.Loading = true
			return m, cmd.RenameKeyCmd(m.CurrentKey.Key, m.RenameInput.Value())
		}
	case "esc":
		m.Screen = types.ScreenKeyDetail
		m.RenameInput.Blur()
	default:
		var inputCmd tea.Cmd
		m.RenameInput, inputCmd = m.RenameInput.Update(msg)
		return m, inputCmd
	}
	return m, nil
}

func (m Model) handleCopyKeyScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		if m.CurrentKey != nil && m.CopyInput.Value() != "" {
			m.Loading = true
			return m, cmd.CopyKeyCmd(m.CurrentKey.Key, m.CopyInput.Value(), false)
		}
	case "esc":
		m.Screen = types.ScreenKeyDetail
		m.CopyInput.Blur()
	default:
		var inputCmd tea.Cmd
		m.CopyInput, inputCmd = m.CopyInput.Update(msg)
		return m, inputCmd
	}
	return m, nil
}

func (m Model) handleBulkDeleteScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		if m.BulkDeleteInput.Value() != "" {
			m.Loading = true
			return m, cmd.BulkDeleteCmd(m.BulkDeleteInput.Value())
		}
	case "esc":
		m.Screen = types.ScreenKeys
		m.BulkDeleteInput.Blur()
	default:
		var inputCmd tea.Cmd
		m.BulkDeleteInput, inputCmd = m.BulkDeleteInput.Update(msg)
		return m, inputCmd
	}
	return m, nil
}

func (m Model) handleBatchTTLScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "tab":
		if m.BatchTTLInput.Focused() {
			m.BatchTTLInput.Blur()
			m.BatchTTLPattern.Focus()
		} else {
			m.BatchTTLPattern.Blur()
			m.BatchTTLInput.Focus()
		}
	case "enter":
		if m.BatchTTLInput.Value() != "" && m.BatchTTLPattern.Value() != "" {
			ttlSecs, err := strconv.Atoi(m.BatchTTLInput.Value())
			if err == nil {
				m.Loading = true
				ttl := time.Duration(ttlSecs) * time.Second
				return m, cmd.BatchSetTTLCmd(m.BatchTTLPattern.Value(), ttl)
			}
		}
	case "esc":
		m.Screen = types.ScreenKeys
		m.BatchTTLInput.Blur()
		m.BatchTTLPattern.Blur()
	default:
		if m.BatchTTLInput.Focused() {
			var inputCmd tea.Cmd
			m.BatchTTLInput, inputCmd = m.BatchTTLInput.Update(msg)
			return m, inputCmd
		}
		var inputCmd tea.Cmd
		m.BatchTTLPattern, inputCmd = m.BatchTTLPattern.Update(msg)
		return m, inputCmd
	}
	return m, nil
}

package ui

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"

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

func (m Model) handleKeysScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.PatternInput.Focused() {
		switch msg.String() {
		case "enter":
			pattern := m.PatternInput.Value()
			// Auto-wrap with wildcards if no glob characters present for partial matching
			if pattern != "" && !strings.ContainsAny(pattern, "*?[]") {
				pattern = "*" + pattern + "*"
			}
			m.KeyPattern = pattern
			m.PatternInput.Blur()
			m.KeyCursor = 0
			m.Loading = true
			return m, cmd.LoadKeysCmd(m.KeyPattern, 0, 1000)
		case "esc":
			m.PatternInput.Blur()
			m.PatternInput.SetValue(m.KeyPattern) // Restore previous value on cancel
		default:
			var inputCmd tea.Cmd
			m.PatternInput, inputCmd = m.PatternInput.Update(msg)
			return m, inputCmd
		}
		return m, nil
	}

	switch msg.String() {
	case "up", "k":
		if m.SelectedKeyIdx > 0 {
			m.SelectedKeyIdx--
			// Load preview for newly selected key
			if len(m.Keys) > 0 && m.SelectedKeyIdx < len(m.Keys) {
				return m, cmd.LoadKeyPreviewCmd(m.Keys[m.SelectedKeyIdx].Key)
			}
		}
	case "down", "j":
		if m.SelectedKeyIdx < len(m.Keys)-1 {
			m.SelectedKeyIdx++
			// Load preview for newly selected key
			if len(m.Keys) > 0 && m.SelectedKeyIdx < len(m.Keys) {
				return m, cmd.LoadKeyPreviewCmd(m.Keys[m.SelectedKeyIdx].Key)
			}
		}
	case "pgup", "ctrl+u":
		m.SelectedKeyIdx -= 10
		if m.SelectedKeyIdx < 0 {
			m.SelectedKeyIdx = 0
		}
		// Load preview for newly selected key
		if len(m.Keys) > 0 && m.SelectedKeyIdx < len(m.Keys) {
			return m, cmd.LoadKeyPreviewCmd(m.Keys[m.SelectedKeyIdx].Key)
		}
	case "pgdown", "ctrl+d":
		m.SelectedKeyIdx += 10
		if m.SelectedKeyIdx >= len(m.Keys) {
			m.SelectedKeyIdx = len(m.Keys) - 1
		}
		if m.SelectedKeyIdx < 0 {
			m.SelectedKeyIdx = 0
		}
		// Load preview for newly selected key
		if len(m.Keys) > 0 && m.SelectedKeyIdx < len(m.Keys) {
			return m, cmd.LoadKeyPreviewCmd(m.Keys[m.SelectedKeyIdx].Key)
		}
	case "home", "g":
		m.SelectedKeyIdx = 0
		// Load preview for newly selected key
		if len(m.Keys) > 0 {
			return m, cmd.LoadKeyPreviewCmd(m.Keys[m.SelectedKeyIdx].Key)
		}
	case "end", "G":
		if len(m.Keys) > 0 {
			m.SelectedKeyIdx = len(m.Keys) - 1
			// Load preview for newly selected key
			return m, cmd.LoadKeyPreviewCmd(m.Keys[m.SelectedKeyIdx].Key)
		}
	case "enter":
		if len(m.Keys) > 0 && m.SelectedKeyIdx < len(m.Keys) {
			key := m.Keys[m.SelectedKeyIdx]
			m.CurrentKey = &key
			m.Loading = true
			m.SelectedItemIdx = 0
			return m, tea.Batch(cmd.LoadKeyValueCmd(key.Key), cmd.GetMemoryUsageCmd(key.Key))
		}
	case "a", "n":
		m.Screen = types.ScreenAddKey
		m.resetAddKeyInputs()
	case "d", "delete", "backspace":
		if len(m.Keys) > 0 && m.SelectedKeyIdx < len(m.Keys) {
			m.ConfirmType = "key"
			m.ConfirmData = m.Keys[m.SelectedKeyIdx]
			m.Screen = types.ScreenConfirmDelete
		}
	case "r":
		m.Loading = true
		m.KeyCursor = 0
		return m, cmd.LoadKeysCmd(m.KeyPattern, 0, 1000)
	case "l":
		if m.KeyCursor > 0 {
			m.Loading = true
			return m, cmd.LoadKeysCmd(m.KeyPattern, m.KeyCursor, 1000)
		}
	case "i":
		return m, cmd.LoadServerInfoCmd()
	case "/":
		m.PatternInput.Focus()
	case "f":
		m.ConfirmType = "flushdb"
		m.Screen = types.ScreenConfirmDelete
	case "s":
		m.sortKeys()
	case "S":
		m.SortAsc = !m.SortAsc
		m.sortKeys()
	case "v":
		m.SearchValueInput.SetValue("")
		m.SearchValueInput.Focus()
		m.Screen = types.ScreenSearchValues
	case "e":
		m.Screen = types.ScreenExport
		m.ExportInput.Focus()
	case "I":
		m.Screen = types.ScreenImport
		m.ImportInput.Focus()
	case "p":
		m.Screen = types.ScreenPubSub
		m.resetPubSubInputs()
	case "L":
		m.Loading = true
		return m, cmd.GetSlowLogCmd(20)
	case "E":
		m.LuaScriptInput.SetValue("")
		m.LuaScriptInput.Focus()
		m.LuaResult = ""
		m.Screen = types.ScreenLuaScript
	case "D":
		m.DBSwitchInput.SetValue("")
		m.DBSwitchInput.Focus()
		m.Screen = types.ScreenSwitchDB
	case "O":
		m.LogCursor = 0
		m.ShowingLogDetail = false
		m.Screen = types.ScreenLogs
	case "B":
		m.BulkDeleteInput.SetValue("")
		m.BulkDeleteInput.Focus()
		m.BulkDeletePreview = nil
		m.Screen = types.ScreenBulkDelete
	case "T":
		m.BatchTTLInput.SetValue("")
		m.BatchTTLPattern.SetValue("")
		m.BatchTTLInput.Focus()
		m.Screen = types.ScreenBatchTTL
	case "F":
		connID := int64(0)
		if m.CurrentConn != nil {
			connID = m.CurrentConn.ID
		}
		m.Screen = types.ScreenFavorites
		return m, cmd.LoadFavoritesCmd(connID)
	case "ctrl+r":
		m.RegexSearchInput.SetValue("")
		m.RegexSearchInput.Focus()
		m.Screen = types.ScreenRegexSearch
	case "ctrl+f":
		m.FuzzySearchInput.SetValue("")
		m.FuzzySearchInput.Focus()
		m.Screen = types.ScreenFuzzySearch
	case "ctrl+l":
		m.Loading = true
		return m, cmd.GetClientListCmd()
	case "m":
		// Live metrics dashboard
		m.LiveMetricsActive = true
		m.Screen = types.ScreenLiveMetrics
		if m.LiveMetrics == nil {
			m.LiveMetrics = &types.LiveMetrics{
				MaxDataPoints:   60, // 1 minute of history
				RefreshInterval: time.Second,
			}
		}
		return m, tea.Batch(cmd.GetLiveMetricsCmd(), cmd.LiveMetricsTickCmd())
	case "M":
		m.Loading = true
		return m, cmd.GetMemoryStatsCmd()
	case "C":
		m.Loading = true
		return m, cmd.GetClusterInfoCmd()
	case "ctrl+k":
		m.CompareKey1Input.SetValue("")
		m.CompareKey2Input.SetValue("")
		m.CompareKey1Input.Focus()
		m.CompareFocusIdx = 0
		m.Screen = types.ScreenCompareKeys
	case "P":
		return m, cmd.LoadTemplatesCmd()
	case "ctrl+h":
		connID := int64(0)
		if m.CurrentConn != nil {
			connID = m.CurrentConn.ID
		}
		m.Screen = types.ScreenRecentKeys
		return m, cmd.LoadRecentKeysCmd(connID)
	case "ctrl+e":
		m.KeyspaceSubActive = !m.KeyspaceSubActive
		if m.KeyspaceSubActive {
			var sendFunc func(tea.Msg)
			if m.SendFunc != nil {
				sendFunc = *m.SendFunc
			}
			return m, cmd.SubscribeKeyspaceCmd("*", sendFunc)
		}
		m.StatusMsg = "Keyspace events disabled"
	case "W":
		m.TreeSeparator = ":"
		m.Screen = types.ScreenTreeView
		m.Loading = true
		return m, cmd.LoadKeyPrefixesCmd(m.TreeSeparator, 3)
	case "ctrl+x":
		// Show expiring keys (TTL < threshold)
		var expiring []types.RedisKey
		for _, k := range m.Keys {
			if k.TTL > 0 && k.TTL.Seconds() < float64(m.ExpiryThreshold) {
				expiring = append(expiring, k)
			}
		}
		m.ExpiringKeys = expiring
		m.Screen = types.ScreenExpiringKeys
	case "esc":
		m.Screen = types.ScreenConnections
		m.KeyPattern = ""
		m.PatternInput.SetValue("")
	}
	return m, nil
}

func (m *Model) sortKeys() {
	switch m.SortBy {
	case "name":
		m.SortBy = "type"
	case "type":
		m.SortBy = "ttl"
	case "ttl":
		m.SortBy = "name"
	default:
		m.SortBy = "name"
	}

	sort.Slice(m.Keys, func(i, j int) bool {
		var less bool
		switch m.SortBy {
		case "name":
			less = m.Keys[i].Key < m.Keys[j].Key
		case "type":
			less = string(m.Keys[i].Type) < string(m.Keys[j].Type)
		case "ttl":
			less = m.Keys[i].TTL < m.Keys[j].TTL
		}
		if m.SortAsc {
			return less
		}
		return !less
	})
}

func (m Model) handleKeyDetailScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "d", "delete":
		if m.CurrentKey != nil {
			m.ConfirmType = "key"
			m.ConfirmData = *m.CurrentKey
			m.Screen = types.ScreenConfirmDelete
		}
	case "t":
		if m.CurrentKey != nil {
			m.TTLInput.SetValue("")
			m.TTLInput.Focus()
			m.Screen = types.ScreenTTLEditor
		}
	case "r":
		if m.CurrentKey != nil {
			m.Loading = true
			return m, tea.Batch(cmd.LoadKeyValueCmd(m.CurrentKey.Key), cmd.GetMemoryUsageCmd(m.CurrentKey.Key))
		}
	case "e":
		if m.CurrentKey != nil && m.CurrentKey.Type == types.KeyTypeString {
			tempFile, err := os.CreateTemp("", "redis_edit_*.json")
			if err != nil {
				m.StatusMsg = "Error creating temp file: " + err.Error()
				return m, nil
			}
			tempFile.Close() // Close it so vim can open it

			value := m.CurrentValue.StringValue
			// Auto-format as JSON if it looks like JSON
			if strings.TrimSpace(value) != "" && (strings.HasPrefix(strings.TrimSpace(value), "{") || strings.HasPrefix(strings.TrimSpace(value), "[")) {
				var prettyJSON bytes.Buffer
				if err := json.Indent(&prettyJSON, []byte(value), "", "  "); err == nil {
					value = prettyJSON.String()
				}
			}

			if err := os.WriteFile(tempFile.Name(), []byte(value), 0644); err != nil {
				m.StatusMsg = "Error writing temp file: " + err.Error()
				os.Remove(tempFile.Name())
				return m, nil
			}

			// Store the temp file name for cleanup
			m.StatusMsg = "Opening in Vim..."
			return m, tea.ExecProcess(exec.Command("vim", tempFile.Name()), func(err error) tea.Msg {
				return types.VimEditDoneMsg{TempFile: tempFile.Name(), Err: err}
			})
		}
	case "a":
		if m.CurrentKey != nil && m.CurrentKey.Type != types.KeyTypeString {
			m.resetAddCollectionInputs()
			m.Screen = types.ScreenAddToCollection
		}
	case "x":
		if m.CurrentKey != nil && m.CurrentKey.Type != types.KeyTypeString {
			m.SelectedItemIdx = 0
			m.Screen = types.ScreenRemoveFromCollection
		}
	case "R":
		if m.CurrentKey != nil {
			m.RenameInput.SetValue(m.CurrentKey.Key)
			m.RenameInput.Focus()
			m.Screen = types.ScreenRenameKey
		}
	case "c":
		if m.CurrentKey != nil {
			m.CopyInput.SetValue(m.CurrentKey.Key + "_copy")
			m.CopyInput.Focus()
			m.Screen = types.ScreenCopyKey
		}
	case "f":
		// Toggle favorite
		if m.CurrentKey != nil {
			connID := int64(0)
			if m.CurrentConn != nil {
				connID = m.CurrentConn.ID
			}
			if m.CurrentKey.IsFavorite {
				return m, cmd.RemoveFavoriteCmd(connID, m.CurrentKey.Key)
			}
			return m, cmd.AddFavoriteCmd(connID, m.CurrentKey.Key, m.CurrentConn.Name)
		}
	case "w":
		// Watch key for changes
		if m.CurrentKey != nil {
			if m.WatchActive && m.WatchKey == m.CurrentKey.Key {
				m.WatchActive = false
				m.StatusMsg = "Watch stopped"
			} else {
				m.WatchActive = true
				m.WatchKey = m.CurrentKey.Key
				m.WatchValue = m.CurrentValue.StringValue
				m.WatchLastUpdate = time.Now()
				m.StatusMsg = "Watching key for changes..."
				return m, cmd.WatchKeyTickCmd()
			}
		}
	case "h":
		// View value history
		if m.CurrentKey != nil {
			return m, cmd.LoadValueHistoryCmd(m.CurrentKey.Key)
		}
	case "y":
		// Copy value to clipboard
		if m.CurrentKey != nil {
			return m, cmd.CopyToClipboardCmd(m.CurrentValue.StringValue)
		}
	case "ctrl+j":
		// JSON path query (for string keys with JSON)
		if m.CurrentKey != nil && m.CurrentKey.Type == types.KeyTypeString {
			m.JSONPathInput.SetValue("")
			m.JSONPathInput.Focus()
			m.Screen = types.ScreenJSONPath
		}
	case "up", "k":
		if m.CurrentKey != nil && m.CurrentKey.Type == types.KeyTypeString {
			if m.DetailScroll > 0 {
				m.DetailScroll--
			}
		} else {
			if m.SelectedItemIdx > 0 {
				m.SelectedItemIdx--
			}
		}
	case "down", "j":
		if m.CurrentKey != nil && m.CurrentKey.Type == types.KeyTypeString {
			maxScroll := len(m.DetailLines) - 20 // 20 is maxLines in viewKeyDetail
			if maxScroll < 0 {
				maxScroll = 0
			}
			if m.DetailScroll < maxScroll {
				m.DetailScroll++
			}
		} else {
			maxIdx := m.getCollectionLength() - 1
			if m.SelectedItemIdx < maxIdx {
				m.SelectedItemIdx++
			}
		}
	case "esc", "backspace":
		m.Screen = types.ScreenKeys
		m.CurrentKey = nil
		m.SelectedItemIdx = 0
		m.WatchActive = false
	}
	return m, nil
}

func (m Model) getCollectionLength() int {
	switch m.CurrentValue.Type {
	case types.KeyTypeList:
		return len(m.CurrentValue.ListValue)
	case types.KeyTypeSet:
		return len(m.CurrentValue.SetValue)
	case types.KeyTypeZSet:
		return len(m.CurrentValue.ZSetValue)
	case types.KeyTypeHash:
		return len(m.CurrentValue.HashValue)
	case types.KeyTypeStream:
		return len(m.CurrentValue.StreamValue)
	default:
		return 0
	}
}

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

func (m Model) handleEditValueScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		if m.CurrentKey != nil {
			value := m.EditValueInput.Value()
			// Validate JSON if it looks like JSON
			if strings.TrimSpace(value) != "" && (strings.HasPrefix(strings.TrimSpace(value), "{") || strings.HasPrefix(strings.TrimSpace(value), "[")) {
				var js interface{}
				if err := json.Unmarshal([]byte(value), &js); err != nil {
					m.StatusMsg = "Error: Invalid JSON - " + err.Error()
					return m, nil
				}
			}
			m.Loading = true
			return m, cmd.EditStringValueCmd(m.CurrentKey.Key, value)
		}
	case "esc":
		m.Screen = types.ScreenKeyDetail
		m.EditValueInput.Blur()
	default:
		var inputCmd tea.Cmd
		m.EditValueInput, inputCmd = m.EditValueInput.Update(msg)
		return m, inputCmd
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

func (m Model) handleTestConnectionScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "enter":
		m.Screen = types.ScreenAddConnection
		m.TestConnResult = ""
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

// New screen handlers for additional features

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

func (m Model) handleFavoritesScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.SelectedFavIdx > 0 {
			m.SelectedFavIdx--
		}
	case "down", "j":
		if m.SelectedFavIdx < len(m.Favorites)-1 {
			m.SelectedFavIdx++
		}
	case "enter":
		if len(m.Favorites) > 0 && m.SelectedFavIdx < len(m.Favorites) {
			key := m.Favorites[m.SelectedFavIdx].Key
			for i, k := range m.Keys {
				if k.Key == key {
					m.SelectedKeyIdx = i
					m.CurrentKey = &m.Keys[i]
					m.Screen = types.ScreenKeyDetail
					return m, cmd.LoadKeyValueCmd(key)
				}
			}
		}
	case "d":
		if len(m.Favorites) > 0 && m.SelectedFavIdx < len(m.Favorites) {
			return m, cmd.RemoveFavoriteCmd(m.Favorites[m.SelectedFavIdx].ConnectionID, m.Favorites[m.SelectedFavIdx].Key)
		}
	case "esc":
		m.Screen = types.ScreenKeys
	}
	return m, nil
}

func (m Model) handleRecentKeysScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.SelectedRecentIdx > 0 {
			m.SelectedRecentIdx--
		}
	case "down", "j":
		if m.SelectedRecentIdx < len(m.RecentKeys)-1 {
			m.SelectedRecentIdx++
		}
	case "enter":
		if len(m.RecentKeys) > 0 && m.SelectedRecentIdx < len(m.RecentKeys) {
			key := m.RecentKeys[m.SelectedRecentIdx].Key
			for i, k := range m.Keys {
				if k.Key == key {
					m.SelectedKeyIdx = i
					m.CurrentKey = &m.Keys[i]
					m.Screen = types.ScreenKeyDetail
					return m, cmd.LoadKeyValueCmd(key)
				}
			}
		}
	case "esc":
		m.Screen = types.ScreenKeys
	}
	return m, nil
}

func (m Model) handleTreeViewScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.SelectedTreeIdx > 0 {
			m.SelectedTreeIdx--
		}
	case "down", "j":
		if m.SelectedTreeIdx < len(m.TreeNodes)-1 {
			m.SelectedTreeIdx++
		}
	case "enter", " ":
		if len(m.TreeNodes) > 0 && m.SelectedTreeIdx < len(m.TreeNodes) {
			node := m.TreeNodes[m.SelectedTreeIdx]
			if !node.IsKey {
				m.TreeExpanded[node.FullPath] = !m.TreeExpanded[node.FullPath]
			} else {
				// Navigate to key
				for i, k := range m.Keys {
					if k.Key == node.FullPath {
						m.SelectedKeyIdx = i
						m.CurrentKey = &m.Keys[i]
						m.Screen = types.ScreenKeyDetail
						return m, cmd.LoadKeyValueCmd(node.FullPath)
					}
				}
			}
		}
	case "esc":
		m.Screen = types.ScreenKeys
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

func (m Model) handleLiveMetricsScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		m.LiveMetricsActive = false
		m.Screen = types.ScreenKeys
	case "c":
		// Clear metrics history
		if m.LiveMetrics != nil {
			m.LiveMetrics.DataPoints = nil
		}
		m.StatusMsg = "Metrics cleared"
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

func (m Model) handleTemplatesScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.SelectedTemplateIdx > 0 {
			m.SelectedTemplateIdx--
		}
	case "down", "j":
		if m.SelectedTemplateIdx < len(m.Templates)-1 {
			m.SelectedTemplateIdx++
		}
	case "enter":
		// Use template to create a key
		if len(m.Templates) > 0 && m.SelectedTemplateIdx < len(m.Templates) {
			template := m.Templates[m.SelectedTemplateIdx]
			m.AddKeyInputs[0].SetValue(template.KeyPattern)
			m.AddKeyInputs[1].SetValue(template.DefaultValue)
			m.AddKeyType = template.Type
			m.Screen = types.ScreenAddKey
		}
	case "esc":
		m.Screen = types.ScreenKeys
	}
	return m, nil
}

func (m Model) handleValueHistoryScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.SelectedHistoryIdx > 0 {
			m.SelectedHistoryIdx--
		}
	case "down", "j":
		if m.SelectedHistoryIdx < len(m.ValueHistory)-1 {
			m.SelectedHistoryIdx++
		}
	case "enter":
		// Restore this value
		if m.CurrentKey != nil && len(m.ValueHistory) > 0 && m.SelectedHistoryIdx < len(m.ValueHistory) {
			entry := m.ValueHistory[m.SelectedHistoryIdx]
			m.Loading = true
			return m, cmd.EditStringValueCmd(m.CurrentKey.Key, entry.Value.StringValue)
		}
	case "esc":
		m.Screen = types.ScreenKeyDetail
	}
	return m, nil
}

func (m Model) handleKeyspaceEventsScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "c":
		m.KeyspaceEvents = nil
	case "esc":
		m.Screen = types.ScreenKeys
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

func (m Model) handleWatchKeyScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.WatchActive = false
		m.Screen = types.ScreenKeyDetail
	}
	return m, nil
}

func (m Model) handleConnectionGroupsScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.SelectedGroupIdx > 0 {
			m.SelectedGroupIdx--
		}
	case "down", "j":
		if m.SelectedGroupIdx < len(m.ConnectionGroups)-1 {
			m.SelectedGroupIdx++
		}
	case "esc":
		m.Screen = types.ScreenConnections
	}
	return m, nil
}

func (m Model) handleExpiringKeysScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.SelectedKeyIdx > 0 {
			m.SelectedKeyIdx--
		}
	case "down", "j":
		if m.SelectedKeyIdx < len(m.ExpiringKeys)-1 {
			m.SelectedKeyIdx++
		}
	case "enter":
		if len(m.ExpiringKeys) > 0 && m.SelectedKeyIdx < len(m.ExpiringKeys) {
			key := m.ExpiringKeys[m.SelectedKeyIdx]
			m.CurrentKey = &key
			m.Screen = types.ScreenKeyDetail
			return m, cmd.LoadKeyValueCmd(key.Key)
		}
	case "esc":
		m.Screen = types.ScreenKeys
	}
	return m, nil
}

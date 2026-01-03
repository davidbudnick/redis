package ui

import (
	"sort"
	"strings"
	"time"

	"github.com/davidbudnick/redis-tui/internal/cmd"
	"github.com/davidbudnick/redis-tui/internal/types"
	"github.com/kujtimiihoxha/vimtea"

	tea "github.com/charmbracelet/bubbletea"
)

func createVimEditor(content string, width, height int) vimtea.Editor {
	editor := vimtea.NewEditor(
		vimtea.WithContent(content),
		vimtea.WithEnableStatusBar(true),
		vimtea.WithEnableModeCommand(true),
	)

	// Add :w command to save
	editor.AddCommand("w", func(buf vimtea.Buffer, args []string) tea.Cmd {
		return func() tea.Msg {
			return types.EditorSaveMsg{Content: buf.Text()}
		}
	})

	// Add :q command to quit
	editor.AddCommand("q", func(buf vimtea.Buffer, args []string) tea.Cmd {
		return func() tea.Msg {
			return types.EditorQuitMsg{}
		}
	})

	// Add :wq command to save and quit
	editor.AddCommand("wq", func(buf vimtea.Buffer, args []string) tea.Cmd {
		return func() tea.Msg {
			return types.EditorSaveMsg{Content: buf.Text()}
		}
	})

	// Set size after creation
	sized, _ := editor.SetSize(width, height)
	return sized.(vimtea.Editor)
}

func (m Model) handleKeysScreen(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.PatternInput.Focused() {
		switch msg.String() {
		case "enter":
			pattern := m.PatternInput.Value()
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
			m.PatternInput.SetValue(m.KeyPattern)
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
			if len(m.Keys) > 0 && m.SelectedKeyIdx < len(m.Keys) {
				return m, cmd.LoadKeyPreviewCmd(m.Keys[m.SelectedKeyIdx].Key)
			}
		}
	case "down", "j":
		if m.SelectedKeyIdx < len(m.Keys)-1 {
			m.SelectedKeyIdx++
			if len(m.Keys) > 0 && m.SelectedKeyIdx < len(m.Keys) {
				return m, cmd.LoadKeyPreviewCmd(m.Keys[m.SelectedKeyIdx].Key)
			}
		}
	case "pgup", "ctrl+u":
		m.SelectedKeyIdx -= 10
		if m.SelectedKeyIdx < 0 {
			m.SelectedKeyIdx = 0
		}
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
		if len(m.Keys) > 0 && m.SelectedKeyIdx < len(m.Keys) {
			return m, cmd.LoadKeyPreviewCmd(m.Keys[m.SelectedKeyIdx].Key)
		}
	case "home", "g":
		m.SelectedKeyIdx = 0
		if len(m.Keys) > 0 {
			return m, cmd.LoadKeyPreviewCmd(m.Keys[m.SelectedKeyIdx].Key)
		}
	case "end", "G":
		if len(m.Keys) > 0 {
			m.SelectedKeyIdx = len(m.Keys) - 1
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
		m.LiveMetricsActive = true
		m.Loading = true
		return m, cmd.LoadLiveMetricsCmd()
	case "M":
		m.Loading = true
		return m, cmd.GetMemoryStatsCmd()
	case "C":
		m.Loading = true
		return m, cmd.GetClusterInfoCmd()
	case "K":
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
			m.VimEditor = createVimEditor(m.CurrentValue.StringValue, m.Width-4, m.Height-10)
			m.Screen = types.ScreenEditValue
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
		if m.CurrentKey != nil {
			return m, cmd.LoadValueHistoryCmd(m.CurrentKey.Key)
		}
	case "y":
		if m.CurrentKey != nil {
			return m, cmd.CopyToClipboardCmd(m.CurrentValue.StringValue)
		}
	case "J":
		if m.CurrentKey != nil && m.CurrentKey.Type == types.KeyTypeString {
			m.JSONPathInput.SetValue("")
			m.JSONPathInput.Focus()
			m.Screen = types.ScreenJSONPath
		}
	case "up", "k":
		if m.SelectedItemIdx > 0 {
			m.SelectedItemIdx--
		}
	case "down", "j":
		maxIdx := m.getCollectionLength() - 1
		if m.SelectedItemIdx < maxIdx {
			m.SelectedItemIdx++
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

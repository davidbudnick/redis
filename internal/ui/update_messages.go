package ui

import (
	"fmt"
	"log/slog"
	"strconv"

	"github.com/davidbudnick/redis/internal/cmd"
	"github.com/davidbudnick/redis/internal/types"

	tea "github.com/charmbracelet/bubbletea"
)

// Connection message handlers
func (m Model) handleConnectionsLoadedMsg(msg types.ConnectionsLoadedMsg) (tea.Model, tea.Cmd) {
	m.Loading = false
	if msg.Err != nil {
		slog.Error("Failed to load connections", "error", msg.Err)
		m.Err = msg.Err
		m.StatusMsg = "Error: " + msg.Err.Error()
	} else {
		m.Connections = msg.Connections
		m.StatusMsg = ""
	}
	return m, nil
}

func (m Model) handleConnectionAddedMsg(msg types.ConnectionAddedMsg) (tea.Model, tea.Cmd) {
	m.Loading = false
	if msg.Err != nil {
		slog.Error("Failed to add connection", "error", msg.Err)
		m.StatusMsg = "Error: " + msg.Err.Error()
	} else {
		m.Connections = append(m.Connections, msg.Connection)
		m.Screen = types.ScreenConnections
		m.resetConnInputs()
		m.StatusMsg = "Connection added"
	}
	return m, nil
}

func (m Model) handleConnectionUpdatedMsg(msg types.ConnectionUpdatedMsg) (tea.Model, tea.Cmd) {
	m.Loading = false
	if msg.Err != nil {
		m.StatusMsg = "Error: " + msg.Err.Error()
	} else {
		for i, c := range m.Connections {
			if c.ID == msg.Connection.ID {
				m.Connections[i] = msg.Connection
				break
			}
		}
		m.Screen = types.ScreenConnections
		m.EditingConnection = nil
		m.resetConnInputs()
		m.StatusMsg = "Connection updated"
	}
	return m, nil
}

func (m Model) handleConnectionDeletedMsg(msg types.ConnectionDeletedMsg) (tea.Model, tea.Cmd) {
	m.Loading = false
	if msg.Err == nil {
		for i, c := range m.Connections {
			if c.ID == msg.ID {
				m.Connections = append(m.Connections[:i], m.Connections[i+1:]...)
				break
			}
		}
		if m.SelectedConnIdx >= len(m.Connections) && m.SelectedConnIdx > 0 {
			m.SelectedConnIdx--
		}
		m.StatusMsg = "Connection deleted"
	}
	m.Screen = types.ScreenConnections
	return m, nil
}

func (m Model) handleConnectedMsg(msg types.ConnectedMsg) (tea.Model, tea.Cmd) {
	m.Loading = false
	if msg.Err != nil {
		slog.Error("Failed to connect", "error", msg.Err)
		m.ConnectionError = msg.Err.Error()
		m.StatusMsg = "Connection failed"
		return m, nil
	}
	// Initialize inputs lazily on first connection
	m.ensureInputsInitialized()
	m.ConnectionError = ""
	m.Screen = types.ScreenKeys
	m.StatusMsg = "Connected"
	var sendFunc func(tea.Msg)
	if m.SendFunc != nil {
		sendFunc = *m.SendFunc
	}
	return m, tea.Batch(cmd.LoadKeysCmd(m.KeyPattern, 0, 1000), cmd.SubscribeKeyspaceCmd("*", sendFunc), tickCmd())
}

func (m Model) handleDisconnectedMsg() (tea.Model, tea.Cmd) {
	m.CurrentConn = nil
	m.Keys = nil
	m.CurrentKey = nil
	m.Screen = types.ScreenConnections
	m.StatusMsg = "Disconnected"
	return m, cmd.UnsubscribeKeyspaceCmd()
}

func (m Model) handleConnectionTestMsg(msg types.ConnectionTestMsg) (tea.Model, tea.Cmd) {
	m.Loading = false
	if msg.Err != nil {
		m.TestConnResult = "Failed: " + msg.Err.Error()
	} else {
		m.TestConnResult = "Connected in " + msg.Latency.String()
	}
	return m, nil
}

func (m Model) handleGroupsLoadedMsg(msg types.GroupsLoadedMsg) (tea.Model, tea.Cmd) {
	m.Loading = false
	if msg.Err == nil {
		m.ConnectionGroups = msg.Groups
	}
	return m, nil
}

// Key message handlers
func (m Model) handleKeysLoadedMsg(msg types.KeysLoadedMsg) (tea.Model, tea.Cmd) {
	m.Loading = false
	if msg.Err != nil {
		slog.Error("Failed to load keys", "error", msg.Err)
		m.StatusMsg = "Error: " + msg.Err.Error()
	} else {
		if m.KeyCursor == 0 {
			m.Keys = msg.Keys
			m.SelectedKeyIdx = 0
		} else {
			m.Keys = append(m.Keys, msg.Keys...)
		}
		m.KeyCursor = msg.Cursor
		m.TotalKeys = msg.TotalKeys
		// Load preview for selected key
		if len(m.Keys) > 0 && m.SelectedKeyIdx < len(m.Keys) {
			return m, cmd.LoadKeyPreviewCmd(m.Keys[m.SelectedKeyIdx].Key)
		}
	}
	return m, nil
}

func (m Model) handleKeyValueLoadedMsg(msg types.KeyValueLoadedMsg) (tea.Model, tea.Cmd) {
	m.Loading = false
	if msg.Err != nil {
		slog.Error("Failed to load key value", "key", msg.Key, "error", msg.Err)
		m.StatusMsg = "Error: " + msg.Err.Error()
	} else {
		m.CurrentValue = msg.Value
		m.Screen = types.ScreenKeyDetail
	}
	return m, nil
}

func (m Model) handleKeyPreviewLoadedMsg(msg types.KeyPreviewLoadedMsg) (tea.Model, tea.Cmd) {
	if msg.Err != nil {
		return m, nil
	}
	// Only update if the key matches the currently selected key
	if len(m.Keys) > 0 && m.SelectedKeyIdx < len(m.Keys) && m.Keys[m.SelectedKeyIdx].Key == msg.Key {
		m.PreviewKey = msg.Key
		m.PreviewValue = msg.Value
	}
	return m, nil
}

func (m Model) handleKeyDeletedMsg(msg types.KeyDeletedMsg) (tea.Model, tea.Cmd) {
	m.Loading = false
	if msg.Err == nil {
		for i, k := range m.Keys {
			if k.Key == msg.Key {
				m.Keys = append(m.Keys[:i], m.Keys[i+1:]...)
				break
			}
		}
		if m.SelectedKeyIdx >= len(m.Keys) && m.SelectedKeyIdx > 0 {
			m.SelectedKeyIdx--
		}
		m.CurrentKey = nil
		m.StatusMsg = "Key deleted"
		m.Screen = types.ScreenKeys
	} else {
		slog.Error("Failed to delete key", "key", msg.Key, "error", msg.Err)
		m.StatusMsg = "Error: " + msg.Err.Error()
	}
	return m, nil
}

func (m Model) handleKeySetMsg(msg types.KeySetMsg) (tea.Model, tea.Cmd) {
	m.Loading = false
	if msg.Err != nil {
		slog.Error("Failed to set key", "key", msg.Key, "error", msg.Err)
		m.StatusMsg = "Error: " + msg.Err.Error()
		return m, nil
	}
	m.StatusMsg = "Key saved"
	m.Screen = types.ScreenKeys
	m.resetAddKeyInputs()
	return m, cmd.LoadKeysCmd(m.KeyPattern, 0, 1000)
}

func (m Model) handleKeyRenamedMsg(msg types.KeyRenamedMsg) (tea.Model, tea.Cmd) {
	m.Loading = false
	if msg.Err != nil {
		m.StatusMsg = "Error: " + msg.Err.Error()
	} else {
		m.StatusMsg = "Key renamed to " + msg.NewKey
		if m.CurrentKey != nil {
			m.CurrentKey.Key = msg.NewKey
			for i, k := range m.Keys {
				if k.Key == msg.OldKey {
					m.Keys[i].Key = msg.NewKey
					break
				}
			}
		}
		m.Screen = types.ScreenKeyDetail
	}
	return m, nil
}

func (m Model) handleKeyCopiedMsg(msg types.KeyCopiedMsg) (tea.Model, tea.Cmd) {
	m.Loading = false
	if msg.Err != nil {
		m.StatusMsg = "Error: " + msg.Err.Error()
		return m, nil
	}
	m.StatusMsg = "Key copied to " + msg.DestKey
	m.Screen = types.ScreenKeyDetail
	m.KeyCursor = 0
	return m, cmd.LoadKeysCmd(m.KeyPattern, 0, 1000)
}

// Value message handlers
func (m Model) handleValueEditedMsg(msg types.ValueEditedMsg) (tea.Model, tea.Cmd) {
	m.Loading = false
	if msg.Err != nil {
		m.StatusMsg = "Error: " + msg.Err.Error()
		return m, nil
	}
	m.StatusMsg = "Value updated"
	m.Screen = types.ScreenKeyDetail
	if m.CurrentKey != nil {
		return m, cmd.LoadKeyValueCmd(m.CurrentKey.Key)
	}
	return m, nil
}

func (m Model) handleItemAddedToCollectionMsg(msg types.ItemAddedToCollectionMsg) (tea.Model, tea.Cmd) {
	m.Loading = false
	if msg.Err != nil {
		m.StatusMsg = "Error: " + msg.Err.Error()
		return m, nil
	}
	m.StatusMsg = "Item added"
	m.Screen = types.ScreenKeyDetail
	m.resetAddCollectionInputs()
	if m.CurrentKey != nil {
		return m, cmd.LoadKeyValueCmd(m.CurrentKey.Key)
	}
	return m, nil
}

func (m Model) handleItemRemovedFromCollectionMsg(msg types.ItemRemovedFromCollectionMsg) (tea.Model, tea.Cmd) {
	m.Loading = false
	if msg.Err != nil {
		m.StatusMsg = "Error: " + msg.Err.Error()
		return m, nil
	}
	m.StatusMsg = "Item removed"
	if m.CurrentKey != nil {
		return m, cmd.LoadKeyValueCmd(m.CurrentKey.Key)
	}
	return m, nil
}

// TTL message handlers
func (m Model) handleTTLSetMsg(msg types.TTLSetMsg) (tea.Model, tea.Cmd) {
	m.Loading = false
	if msg.Err != nil {
		m.StatusMsg = "Error: " + msg.Err.Error()
		return m, nil
	}
	m.StatusMsg = "TTL updated"
	if m.CurrentKey != nil {
		m.CurrentKey.TTL = msg.TTL
		for i, k := range m.Keys {
			if k.Key == m.CurrentKey.Key {
				m.Keys[i].TTL = msg.TTL
				break
			}
		}
	}
	m.Screen = types.ScreenKeyDetail
	return m, nil
}

func (m Model) handleBatchTTLSetMsg(msg types.BatchTTLSetMsg) (tea.Model, tea.Cmd) {
	m.Loading = false
	if msg.Err != nil {
		m.StatusMsg = "Batch TTL error: " + msg.Err.Error()
		return m, nil
	}
	m.StatusMsg = "Set TTL on " + strconv.Itoa(msg.Count) + " keys"
	m.Screen = types.ScreenKeys
	m.KeyCursor = 0
	return m, cmd.LoadKeysCmd(m.KeyPattern, 0, 100)
}

// Server message handlers
func (m Model) handleServerInfoLoadedMsg(msg types.ServerInfoLoadedMsg) (tea.Model, tea.Cmd) {
	m.Loading = false
	if msg.Err == nil {
		m.ServerInfo = msg.Info
		m.Screen = types.ScreenServerInfo
	}
	return m, nil
}

func (m Model) handleDBSwitchedMsg(msg types.DBSwitchedMsg) (tea.Model, tea.Cmd) {
	m.Loading = false
	if msg.Err != nil {
		m.StatusMsg = "Error: " + msg.Err.Error()
		return m, nil
	}
	if m.CurrentConn != nil {
		m.CurrentConn.DB = msg.DB
	}
	m.StatusMsg = "Switched to database " + strconv.Itoa(msg.DB)
	m.Screen = types.ScreenKeys
	m.KeyCursor = 0
	m.Keys = []types.RedisKey{}
	return m, cmd.LoadKeysCmd(m.KeyPattern, 0, 100)
}

func (m Model) handleFlushDBMsg(msg types.FlushDBMsg) (tea.Model, tea.Cmd) {
	m.Loading = false
	if msg.Err == nil {
		m.Keys = []types.RedisKey{}
		m.StatusMsg = "Database flushed"
	}
	m.Screen = types.ScreenKeys
	return m, nil
}

func (m Model) handleSlowLogLoadedMsg(msg types.SlowLogLoadedMsg) (tea.Model, tea.Cmd) {
	m.Loading = false
	if msg.Err != nil {
		m.StatusMsg = "Error: " + msg.Err.Error()
	} else {
		m.SlowLogEntries = msg.Entries
		m.Screen = types.ScreenSlowLog
	}
	return m, nil
}

func (m Model) handleClientListLoadedMsg(msg types.ClientListLoadedMsg) (tea.Model, tea.Cmd) {
	m.Loading = false
	if msg.Err == nil {
		m.ClientList = msg.Clients
		m.Screen = types.ScreenClientList
	}
	return m, nil
}

func (m Model) handleMemoryStatsLoadedMsg(msg types.MemoryStatsLoadedMsg) (tea.Model, tea.Cmd) {
	m.Loading = false
	if msg.Err == nil {
		m.MemoryStats = &msg.Stats
		m.Screen = types.ScreenMemoryStats
	}
	return m, nil
}

func (m Model) handleClusterInfoLoadedMsg(msg types.ClusterInfoLoadedMsg) (tea.Model, tea.Cmd) {
	m.Loading = false
	if msg.Err == nil {
		m.ClusterNodes = msg.Nodes
		m.ClusterEnabled = len(msg.Nodes) > 0
		m.Screen = types.ScreenClusterInfo
	}
	return m, nil
}

func (m Model) handleMemoryUsageMsg(msg types.MemoryUsageMsg) (tea.Model, tea.Cmd) {
	m.Loading = false
	if msg.Err == nil {
		m.MemoryUsage = msg.Bytes
	}
	return m, nil
}

// Script and Pub/Sub handlers
func (m Model) handleLuaScriptResultMsg(msg types.LuaScriptResultMsg) (tea.Model, tea.Cmd) {
	m.Loading = false
	if msg.Err != nil {
		m.LuaResult = "Error: " + msg.Err.Error()
	} else {
		switch v := msg.Result.(type) {
		case string:
			m.LuaResult = v
		case int64:
			m.LuaResult = strconv.FormatInt(v, 10)
		case []interface{}:
			m.LuaResult = "Array result (length: " + strconv.Itoa(len(v)) + ")"
		default:
			m.LuaResult = "OK"
		}
	}
	return m, nil
}

func (m Model) handlePublishResultMsg(msg types.PublishResultMsg) (tea.Model, tea.Cmd) {
	m.Loading = false
	if msg.Err != nil {
		m.StatusMsg = "Publish failed: " + msg.Err.Error()
	} else {
		m.StatusMsg = "Message sent to " + strconv.FormatInt(msg.Receivers, 10) + " subscribers"
	}
	return m, nil
}

func (m Model) handleKeyspaceEventMsg(msg types.KeyspaceEventMsg) (tea.Model, tea.Cmd) {
	m.KeyspaceEvents = append(m.KeyspaceEvents, msg.Event)
	if len(m.KeyspaceEvents) > 100 {
		// Create new slice to allow GC of old backing array (prevents memory leak)
		newEvents := make([]types.KeyspaceEvent, 99)
		copy(newEvents, m.KeyspaceEvents[1:])
		m.KeyspaceEvents = newEvents
	}
	// Refresh keys if a key was set or deleted
	if msg.Event.Event == "set" || msg.Event.Event == "del" {
		m.StatusMsg = fmt.Sprintf("Key %s: %s", msg.Event.Event, msg.Event.Key)
		return m, cmd.LoadKeysCmd(m.KeyPattern, 0, 1000)
	}
	return m, nil
}

// Import/Export handlers
func (m Model) handleExportCompleteMsg(msg types.ExportCompleteMsg) (tea.Model, tea.Cmd) {
	m.Loading = false
	if msg.Err != nil {
		m.StatusMsg = "Export failed: " + msg.Err.Error()
	} else {
		m.StatusMsg = "Exported " + strconv.Itoa(msg.KeyCount) + " keys to " + msg.Filename
		m.Screen = types.ScreenKeys
	}
	return m, nil
}

func (m Model) handleImportCompleteMsg(msg types.ImportCompleteMsg) (tea.Model, tea.Cmd) {
	m.Loading = false
	if msg.Err != nil {
		m.StatusMsg = "Import failed: " + msg.Err.Error()
		return m, nil
	}
	m.StatusMsg = "Imported " + strconv.Itoa(msg.KeyCount) + " keys from " + msg.Filename
	m.Screen = types.ScreenKeys
	m.KeyCursor = 0
	return m, cmd.LoadKeysCmd(m.KeyPattern, 0, 100)
}

// Feature message handlers
func (m Model) handleBulkDeleteMsg(msg types.BulkDeleteMsg) (tea.Model, tea.Cmd) {
	m.Loading = false
	if msg.Err != nil {
		m.StatusMsg = "Bulk delete error: " + msg.Err.Error()
		return m, nil
	}
	m.StatusMsg = "Deleted " + strconv.Itoa(msg.Deleted) + " keys"
	m.Screen = types.ScreenKeys
	m.KeyCursor = 0
	return m, cmd.LoadKeysCmd(m.KeyPattern, 0, 100)
}

func (m Model) handleFavoritesLoadedMsg(msg types.FavoritesLoadedMsg) (tea.Model, tea.Cmd) {
	m.Loading = false
	if msg.Err == nil {
		m.Favorites = msg.Favorites
	}
	return m, nil
}

func (m Model) handleFavoriteAddedMsg(msg types.FavoriteAddedMsg) (tea.Model, tea.Cmd) {
	m.Loading = false
	if msg.Err == nil {
		m.StatusMsg = "Added to favorites"
		for i := range m.Keys {
			if m.Keys[i].Key == msg.Favorite.Key {
				m.Keys[i].IsFavorite = true
				break
			}
		}
		if m.CurrentKey != nil && m.CurrentKey.Key == msg.Favorite.Key {
			m.CurrentKey.IsFavorite = true
		}
	}
	return m, nil
}

func (m Model) handleFavoriteRemovedMsg(msg types.FavoriteRemovedMsg) (tea.Model, tea.Cmd) {
	m.Loading = false
	if msg.Err == nil {
		m.StatusMsg = "Removed from favorites"
		for i := range m.Keys {
			if m.Keys[i].Key == msg.Key {
				m.Keys[i].IsFavorite = false
				break
			}
		}
		if m.CurrentKey != nil && m.CurrentKey.Key == msg.Key {
			m.CurrentKey.IsFavorite = false
		}
	}
	return m, nil
}

func (m Model) handleRecentKeysLoadedMsg(msg types.RecentKeysLoadedMsg) (tea.Model, tea.Cmd) {
	m.Loading = false
	if msg.Err == nil {
		m.RecentKeys = msg.Keys
	}
	return m, nil
}

func (m Model) handleTemplatesLoadedMsg(msg types.TemplatesLoadedMsg) (tea.Model, tea.Cmd) {
	m.Loading = false
	if msg.Err == nil {
		m.Templates = msg.Templates
		m.Screen = types.ScreenTemplates
	}
	return m, nil
}

func (m Model) handleValueHistoryMsg(msg types.ValueHistoryMsg) (tea.Model, tea.Cmd) {
	m.Loading = false
	if msg.Err == nil {
		m.ValueHistory = msg.History
		m.Screen = types.ScreenValueHistory
	}
	return m, nil
}

// Search message handlers
func (m Model) handleRegexSearchResultMsg(msg types.RegexSearchResultMsg) (tea.Model, tea.Cmd) {
	m.Loading = false
	if msg.Err != nil {
		m.StatusMsg = "Regex search error: " + msg.Err.Error()
	} else {
		m.Keys = msg.Keys
		m.Screen = types.ScreenKeys
		m.StatusMsg = "Found " + strconv.Itoa(len(msg.Keys)) + " keys"
	}
	return m, nil
}

func (m Model) handleFuzzySearchResultMsg(msg types.FuzzySearchResultMsg) (tea.Model, tea.Cmd) {
	m.Loading = false
	if msg.Err != nil {
		m.StatusMsg = "Fuzzy search error: " + msg.Err.Error()
	} else {
		m.Keys = msg.Keys
		m.Screen = types.ScreenKeys
		m.StatusMsg = "Found " + strconv.Itoa(len(msg.Keys)) + " keys"
	}
	return m, nil
}

func (m Model) handleCompareKeysResultMsg(msg types.CompareKeysResultMsg) (tea.Model, tea.Cmd) {
	m.Loading = false
	if msg.Err != nil {
		m.StatusMsg = "Compare error: " + msg.Err.Error()
	} else {
		equal := msg.Key1Value.StringValue == msg.Key2Value.StringValue
		m.CompareResult = &types.KeyComparison{
			Equal:       equal,
			Differences: []string{msg.Diff},
		}
	}
	return m, nil
}

func (m Model) handleClipboardCopiedMsg(msg types.ClipboardCopiedMsg) (tea.Model, tea.Cmd) {
	m.Loading = false
	if msg.Err != nil {
		m.StatusMsg = "Copy failed: " + msg.Err.Error()
	} else {
		m.StatusMsg = "Copied to clipboard"
	}
	return m, nil
}

// Live metrics handlers
func (m Model) handleLiveMetricsMsg(msg types.LiveMetricsMsg) (tea.Model, tea.Cmd) {
	if msg.Err != nil {
		m.StatusMsg = "Metrics error: " + msg.Err.Error()
		return m, nil
	}

	if m.LiveMetrics == nil {
		m.LiveMetrics = &types.LiveMetrics{
			MaxDataPoints: 60, // 1 minute of history
		}
	}

	// Add new data point
	m.LiveMetrics.DataPoints = append(m.LiveMetrics.DataPoints, msg.Data)

	// Keep only last MaxDataPoints
	if len(m.LiveMetrics.DataPoints) > m.LiveMetrics.MaxDataPoints {
		m.LiveMetrics.DataPoints = m.LiveMetrics.DataPoints[1:]
	}

	return m, nil
}

func (m Model) handleLiveMetricsTickMsg() (tea.Model, tea.Cmd) {
	// Only continue refreshing if we're on the live metrics screen
	if m.Screen == types.ScreenLiveMetrics && m.LiveMetricsActive {
		return m, tea.Batch(cmd.GetLiveMetricsCmd(), cmd.LiveMetricsTickCmd())
	}
	return m, nil
}

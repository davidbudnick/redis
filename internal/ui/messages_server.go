package ui

import (
	"fmt"
	"strconv"
	"time"

	"github.com/davidbudnick/redis-tui/internal/cmd"
	"github.com/davidbudnick/redis-tui/internal/types"

	tea "github.com/charmbracelet/bubbletea"
)

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

func (m Model) handleLiveMetricsMsg(msg types.LiveMetricsMsg) (tea.Model, tea.Cmd) {
	m.Loading = false
	if msg.Err != nil {
		m.StatusMsg = "Error: " + msg.Err.Error()
		return m, nil
	}
	if m.LiveMetrics == nil {
		m.LiveMetrics = &types.LiveMetrics{
			MaxDataPoints:   60,
			RefreshInterval: time.Second,
		}
	}
	m.LiveMetrics.DataPoints = append(m.LiveMetrics.DataPoints, msg.Data)
	if len(m.LiveMetrics.DataPoints) > m.LiveMetrics.MaxDataPoints {
		m.LiveMetrics.DataPoints = m.LiveMetrics.DataPoints[1:]
	}
	m.LiveMetricsActive = true
	m.Screen = types.ScreenLiveMetrics
	return m, nil
}

func (m Model) handleLiveMetricsTickMsg() (tea.Model, tea.Cmd) {
	if !m.LiveMetricsActive {
		return m, nil
	}
	return m, cmd.LoadLiveMetricsCmd()
}

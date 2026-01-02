package cmd

import (
	"log/slog"

	"github.com/davidbudnick/redis/internal/redis"
	"github.com/davidbudnick/redis/internal/types"

	tea "github.com/charmbracelet/bubbletea"
)

func LoadConnectionsCmd() tea.Cmd {
	return func() tea.Msg {
		if Config == nil {
			return types.ConnectionsLoadedMsg{Err: nil}
		}
		connections, err := Config.ListConnections()
		if err != nil {
			slog.Error("Failed to load connections", "error", err)
		}
		return types.ConnectionsLoadedMsg{Connections: connections, Err: err}
	}
}

func AddConnectionCmd(name, host string, port int, password string, dbNum int) tea.Cmd {
	return func() tea.Msg {
		if Config == nil {
			return types.ConnectionAddedMsg{Err: nil}
		}
		conn, err := Config.AddConnection(name, host, port, password, dbNum)
		if err != nil {
			slog.Error("Failed to add connection", "error", err)
		}
		return types.ConnectionAddedMsg{Connection: conn, Err: err}
	}
}

func UpdateConnectionCmd(id int64, name, host string, port int, password string, dbNum int) tea.Cmd {
	return func() tea.Msg {
		if Config == nil {
			return types.ConnectionUpdatedMsg{Err: nil}
		}
		conn, err := Config.UpdateConnection(id, name, host, port, password, dbNum)
		if err != nil {
			slog.Error("Failed to update connection", "error", err)
		}
		return types.ConnectionUpdatedMsg{Connection: conn, Err: err}
	}
}

func DeleteConnectionCmd(id int64) tea.Cmd {
	return func() tea.Msg {
		if Config == nil {
			return types.ConnectionDeletedMsg{Err: nil}
		}
		err := Config.DeleteConnection(id)
		return types.ConnectionDeletedMsg{ID: id, Err: err}
	}
}

func ConnectCmd(host string, port int, password string, dbNum int) tea.Cmd {
	return func() tea.Msg {
		if RedisClient == nil {
			RedisClient = redis.NewClient()
		}
		err := RedisClient.Connect(host, port, password, dbNum)
		if err != nil {
			slog.Error("Failed to connect", "error", err)
		}
		return types.ConnectedMsg{Err: err}
	}
}

func DisconnectCmd() tea.Cmd {
	return func() tea.Msg {
		if RedisClient != nil {
			_ = RedisClient.Disconnect()
		}
		return types.DisconnectedMsg{}
	}
}

func TestConnectionCmd(host string, port int, password string, db int) tea.Cmd {
	return func() tea.Msg {
		if RedisClient == nil {
			return types.ConnectionTestMsg{Success: false, Err: nil}
		}
		latency, err := RedisClient.TestConnection(host, port, password, db)
		return types.ConnectionTestMsg{Success: err == nil, Latency: latency, Err: err}
	}
}

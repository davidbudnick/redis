package cmd

import (
	"github.com/davidbudnick/redis-tui/internal/types"

	tea "github.com/charmbracelet/bubbletea"
)

// Add to collection commands

func AddToListCmd(key string, values ...string) tea.Cmd {
	return func() tea.Msg {
		if RedisClient == nil {
			return types.ItemAddedToCollectionMsg{Key: key, Err: nil}
		}
		err := RedisClient.RPush(key, values...)
		return types.ItemAddedToCollectionMsg{Key: key, Err: err}
	}
}

func AddToSetCmd(key string, members ...string) tea.Cmd {
	return func() tea.Msg {
		if RedisClient == nil {
			return types.ItemAddedToCollectionMsg{Key: key, Err: nil}
		}
		err := RedisClient.SAdd(key, members...)
		return types.ItemAddedToCollectionMsg{Key: key, Err: err}
	}
}

func AddToZSetCmd(key string, score float64, member string) tea.Cmd {
	return func() tea.Msg {
		if RedisClient == nil {
			return types.ItemAddedToCollectionMsg{Key: key, Err: nil}
		}
		err := RedisClient.ZAdd(key, score, member)
		return types.ItemAddedToCollectionMsg{Key: key, Err: err}
	}
}

func AddToHashCmd(key, field, value string) tea.Cmd {
	return func() tea.Msg {
		if RedisClient == nil {
			return types.ItemAddedToCollectionMsg{Key: key, Err: nil}
		}
		err := RedisClient.HSet(key, field, value)
		return types.ItemAddedToCollectionMsg{Key: key, Err: err}
	}
}

func AddToStreamCmd(key string, fields map[string]interface{}) tea.Cmd {
	return func() tea.Msg {
		if RedisClient == nil {
			return types.ItemAddedToCollectionMsg{Key: key, Err: nil}
		}
		_, err := RedisClient.XAdd(key, fields)
		return types.ItemAddedToCollectionMsg{Key: key, Err: err}
	}
}

// Remove from collection commands

func RemoveFromListCmd(key string, value string) tea.Cmd {
	return func() tea.Msg {
		if RedisClient == nil {
			return types.ItemRemovedFromCollectionMsg{Key: key, Err: nil}
		}
		err := RedisClient.LRem(key, 1, value)
		return types.ItemRemovedFromCollectionMsg{Key: key, Err: err}
	}
}

func RemoveFromSetCmd(key string, members ...string) tea.Cmd {
	return func() tea.Msg {
		if RedisClient == nil {
			return types.ItemRemovedFromCollectionMsg{Key: key, Err: nil}
		}
		err := RedisClient.SRem(key, members...)
		return types.ItemRemovedFromCollectionMsg{Key: key, Err: err}
	}
}

func RemoveFromZSetCmd(key string, members ...string) tea.Cmd {
	return func() tea.Msg {
		if RedisClient == nil {
			return types.ItemRemovedFromCollectionMsg{Key: key, Err: nil}
		}
		err := RedisClient.ZRem(key, members...)
		return types.ItemRemovedFromCollectionMsg{Key: key, Err: err}
	}
}

func RemoveFromHashCmd(key string, fields ...string) tea.Cmd {
	return func() tea.Msg {
		if RedisClient == nil {
			return types.ItemRemovedFromCollectionMsg{Key: key, Err: nil}
		}
		err := RedisClient.HDel(key, fields...)
		return types.ItemRemovedFromCollectionMsg{Key: key, Err: err}
	}
}

func RemoveFromStreamCmd(key string, ids ...string) tea.Cmd {
	return func() tea.Msg {
		if RedisClient == nil {
			return types.ItemRemovedFromCollectionMsg{Key: key, Err: nil}
		}
		err := RedisClient.XDel(key, ids...)
		return types.ItemRemovedFromCollectionMsg{Key: key, Err: err}
	}
}

package cmd

import (
	"log/slog"
	"strconv"
	"time"

	"github.com/davidbudnick/redis/internal/types"

	tea "github.com/charmbracelet/bubbletea"
)

func LoadKeysCmd(pattern string, cursor uint64, count int64) tea.Cmd {
	return func() tea.Msg {
		if RedisClient == nil {
			return types.KeysLoadedMsg{Err: nil}
		}
		keys, nextCursor, err := RedisClient.ScanKeys(pattern, cursor, count)
		totalKeys := RedisClient.GetTotalKeys()
		return types.KeysLoadedMsg{Keys: keys, Cursor: nextCursor, TotalKeys: totalKeys, Err: err}
	}
}

func LoadKeyValueCmd(key string) tea.Cmd {
	return func() tea.Msg {
		if RedisClient == nil {
			return types.KeyValueLoadedMsg{Err: nil}
		}
		value, err := RedisClient.GetValue(key)
		return types.KeyValueLoadedMsg{Key: key, Value: value, Err: err}
	}
}

func LoadKeyPreviewCmd(key string) tea.Cmd {
	return func() tea.Msg {
		if RedisClient == nil {
			return types.KeyPreviewLoadedMsg{Err: nil}
		}
		value, err := RedisClient.GetValue(key)
		return types.KeyPreviewLoadedMsg{Key: key, Value: value, Err: err}
	}
}

func DeleteKeyCmd(key string) tea.Cmd {
	return func() tea.Msg {
		if RedisClient == nil {
			return types.KeyDeletedMsg{Key: key, Err: nil}
		}
		err := RedisClient.DeleteKey(key)
		return types.KeyDeletedMsg{Key: key, Err: err}
	}
}

func SetTTLCmd(key string, ttl time.Duration) tea.Cmd {
	return func() tea.Msg {
		if RedisClient == nil {
			return types.TTLSetMsg{Key: key, Err: nil}
		}
		err := RedisClient.SetTTL(key, ttl)
		return types.TTLSetMsg{Key: key, TTL: ttl, Err: err}
	}
}

func CreateKeyCmd(key string, keyType types.KeyType, value string, ttl time.Duration) tea.Cmd {
	return func() tea.Msg {
		if RedisClient == nil {
			return types.KeySetMsg{Key: key, Err: nil}
		}
		var err error
		switch keyType {
		case types.KeyTypeString:
			err = RedisClient.SetString(key, value, ttl)
		case types.KeyTypeList:
			err = RedisClient.RPush(key, value)
		case types.KeyTypeSet:
			err = RedisClient.SAdd(key, value)
		case types.KeyTypeZSet:
			score := 0.0
			if s, parseErr := strconv.ParseFloat(value, 64); parseErr == nil {
				score = s
				value = "member"
			}
			err = RedisClient.ZAdd(key, score, value)
		case types.KeyTypeHash:
			err = RedisClient.HSet(key, "field", value)
		case types.KeyTypeStream:
			fields := map[string]interface{}{"data": value}
			_, err = RedisClient.XAdd(key, fields)
		}
		return types.KeySetMsg{Key: key, Err: err}
	}
}

func EditStringValueCmd(key, value string) tea.Cmd {
	return func() tea.Msg {
		if RedisClient == nil {
			return types.ValueEditedMsg{Key: key, Err: nil}
		}
		err := RedisClient.SetString(key, value, 0)
		return types.ValueEditedMsg{Key: key, Err: err}
	}
}

func EditListElementCmd(key string, index int64, value string) tea.Cmd {
	return func() tea.Msg {
		if RedisClient == nil {
			return types.ValueEditedMsg{Key: key, Err: nil}
		}
		err := RedisClient.LSet(key, index, value)
		return types.ValueEditedMsg{Key: key, Err: err}
	}
}

func EditHashFieldCmd(key, field, value string) tea.Cmd {
	return func() tea.Msg {
		if RedisClient == nil {
			return types.ValueEditedMsg{Key: key, Err: nil}
		}
		err := RedisClient.HSet(key, field, value)
		return types.ValueEditedMsg{Key: key, Err: err}
	}
}

func RenameKeyCmd(oldKey, newKey string) tea.Cmd {
	return func() tea.Msg {
		if RedisClient == nil {
			return types.KeyRenamedMsg{OldKey: oldKey, NewKey: newKey, Err: nil}
		}
		err := RedisClient.Rename(oldKey, newKey)
		return types.KeyRenamedMsg{OldKey: oldKey, NewKey: newKey, Err: err}
	}
}

func CopyKeyCmd(src, dst string, replace bool) tea.Cmd {
	return func() tea.Msg {
		if RedisClient == nil {
			return types.KeyCopiedMsg{SourceKey: src, DestKey: dst, Err: nil}
		}
		err := RedisClient.Copy(src, dst, replace)
		return types.KeyCopiedMsg{SourceKey: src, DestKey: dst, Err: err}
	}
}

func GetMemoryUsageCmd(key string) tea.Cmd {
	return func() tea.Msg {
		if RedisClient == nil {
			return types.MemoryUsageMsg{Key: key, Err: nil}
		}
		bytes, err := RedisClient.MemoryUsage(key)
		return types.MemoryUsageMsg{Key: key, Bytes: bytes, Err: err}
	}
}

func BulkDeleteCmd(pattern string) tea.Cmd {
	return func() tea.Msg {
		if RedisClient == nil {
			return types.BulkDeleteMsg{Pattern: pattern, Err: nil}
		}
		deleted, err := RedisClient.BulkDelete(pattern)
		return types.BulkDeleteMsg{Pattern: pattern, Deleted: deleted, Err: err}
	}
}

func BatchSetTTLCmd(pattern string, ttl time.Duration) tea.Cmd {
	return func() tea.Msg {
		if RedisClient == nil {
			return types.BatchTTLSetMsg{Pattern: pattern, Err: nil}
		}
		count, err := RedisClient.BatchSetTTL(pattern, ttl)
		return types.BatchTTLSetMsg{Pattern: pattern, Count: count, TTL: ttl, Err: err}
	}
}

func WatchKeyTickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return types.WatchTickMsg{}
	})
}

func LoadValueHistoryCmd(key string) tea.Cmd {
	return func() tea.Msg {
		if Config == nil {
			return types.ValueHistoryMsg{Err: nil}
		}
		history := Config.GetValueHistory(key)
		return types.ValueHistoryMsg{History: history, Err: nil}
	}
}

func SaveValueHistoryCmd(key string, value types.RedisValue, action string) tea.Cmd {
	return func() tea.Msg {
		if Config != nil {
			Config.AddValueHistory(key, value, action)
		}
		return nil
	}
}

// Keyspace events

func SubscribeKeyspaceCmd(pattern string, sendFunc func(tea.Msg)) tea.Cmd {
	return func() tea.Msg {
		if RedisClient == nil {
			return types.KeyspaceSubscribedMsg{Err: nil}
		}
		err := RedisClient.SubscribeKeyspace(pattern, func(event types.KeyspaceEvent) {
			if sendFunc != nil {
				sendFunc(types.KeyspaceEventMsg{Event: event})
			}
		})
		return types.KeyspaceSubscribedMsg{Pattern: pattern, Err: err}
	}
}

func UnsubscribeKeyspaceCmd() tea.Cmd {
	return func() tea.Msg {
		if RedisClient != nil {
			_ = RedisClient.UnsubscribeKeyspace()
		}
		return nil
	}
}

func LoadKeyPrefixesCmd(separator string, maxDepth int) tea.Cmd {
	return func() tea.Msg {
		if RedisClient == nil {
			return types.TreeNodeExpandedMsg{Err: nil}
		}
		prefixes, err := RedisClient.GetKeyPrefixes(separator, maxDepth)
		return types.TreeNodeExpandedMsg{Children: prefixes, Err: err}
	}
}

// slog is used for error logging
var _ = slog.Error

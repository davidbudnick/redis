package cmd

import (
	"encoding/json"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/davidbudnick/redis/internal/db"
	"github.com/davidbudnick/redis/internal/redis"
	"github.com/davidbudnick/redis/internal/types"

	tea "github.com/charmbracelet/bubbletea"
)

var (
	Config      *db.Config
	RedisClient *redis.Client
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

// LoadKeyPreviewCmd loads a key value for preview in the keys list
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

func LoadServerInfoCmd() tea.Cmd {
	return func() tea.Msg {
		if RedisClient == nil {
			return types.ServerInfoLoadedMsg{Err: nil}
		}
		info, err := RedisClient.GetServerInfo()
		return types.ServerInfoLoadedMsg{Info: info, Err: err}
	}
}

func FlushDBCmd() tea.Cmd {
	return func() tea.Msg {
		if RedisClient == nil {
			return types.FlushDBMsg{Err: nil}
		}
		err := RedisClient.FlushDB()
		return types.FlushDBMsg{Err: err}
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

// Edit value commands
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

// Rename and Copy
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

// Switch database
func SwitchDBCmd(dbNum int) tea.Cmd {
	return func() tea.Msg {
		if RedisClient == nil {
			return types.DBSwitchedMsg{DB: dbNum, Err: nil}
		}
		err := RedisClient.SelectDB(dbNum)
		return types.DBSwitchedMsg{DB: dbNum, Err: err}
	}
}

// Search by value
func SearchByValueCmd(pattern, valueSearch string, maxKeys int) tea.Cmd {
	return func() tea.Msg {
		if RedisClient == nil {
			return types.KeysLoadedMsg{Err: nil}
		}
		keys, err := RedisClient.SearchByValue(pattern, valueSearch, maxKeys)
		return types.KeysLoadedMsg{Keys: keys, Cursor: 0, Err: err}
	}
}

// Memory usage
func GetMemoryUsageCmd(key string) tea.Cmd {
	return func() tea.Msg {
		if RedisClient == nil {
			return types.MemoryUsageMsg{Key: key, Err: nil}
		}
		bytes, err := RedisClient.MemoryUsage(key)
		return types.MemoryUsageMsg{Key: key, Bytes: bytes, Err: err}
	}
}

// Slow log
func GetSlowLogCmd(count int64) tea.Cmd {
	return func() tea.Msg {
		if RedisClient == nil {
			return types.SlowLogLoadedMsg{Err: nil}
		}
		entries, err := RedisClient.SlowLogGet(count)
		return types.SlowLogLoadedMsg{Entries: entries, Err: err}
	}
}

// Lua script
func EvalLuaScriptCmd(script string, keys []string, args ...interface{}) tea.Cmd {
	return func() tea.Msg {
		if RedisClient == nil {
			return types.LuaScriptResultMsg{Err: nil}
		}
		result, err := RedisClient.Eval(script, keys, args...)
		return types.LuaScriptResultMsg{Result: result, Err: err}
	}
}

// Pub/Sub
func PublishMessageCmd(channel, message string) tea.Cmd {
	return func() tea.Msg {
		if RedisClient == nil {
			return types.PublishResultMsg{Channel: channel, Err: nil}
		}
		receivers, err := RedisClient.Publish(channel, message)
		return types.PublishResultMsg{Channel: channel, Receivers: receivers, Err: err}
	}
}

func GetPubSubChannelsCmd(pattern string) tea.Cmd {
	return func() tea.Msg {
		if RedisClient == nil {
			return types.ServerInfoLoadedMsg{Err: nil}
		}
		channels, _ := RedisClient.PubSubChannels(pattern)
		info := types.ServerInfo{}
		for _, ch := range channels {
			if info.ClusterInfo != "" {
				info.ClusterInfo += ", "
			}
			info.ClusterInfo += ch
		}
		return types.ServerInfoLoadedMsg{Info: info, Err: nil}
	}
}

// Export/Import
func ExportKeysCmd(pattern, filename string) tea.Cmd {
	return func() tea.Msg {
		if RedisClient == nil {
			return types.ExportCompleteMsg{Filename: filename, Err: nil}
		}
		data, err := RedisClient.ExportKeys(pattern)
		if err != nil {
			return types.ExportCompleteMsg{Filename: filename, Err: err}
		}

		jsonData, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return types.ExportCompleteMsg{Filename: filename, Err: err}
		}

		err = os.WriteFile(filename, jsonData, 0600)
		return types.ExportCompleteMsg{Filename: filename, KeyCount: len(data), Err: err}
	}
}

func ImportKeysCmd(filename string) tea.Cmd {
	return func() tea.Msg {
		if RedisClient == nil {
			return types.ImportCompleteMsg{Filename: filename, Err: nil}
		}

		// Clean the file path to prevent directory traversal
		cleanPath := filepath.Clean(filename)
		jsonData, err := os.ReadFile(cleanPath) // #nosec G304 - user-provided import path is intentional
		if err != nil {
			return types.ImportCompleteMsg{Filename: filename, Err: err}
		}

		var data map[string]interface{}
		if err := json.Unmarshal(jsonData, &data); err != nil {
			return types.ImportCompleteMsg{Filename: filename, Err: err}
		}

		count, err := RedisClient.ImportKeys(data)
		return types.ImportCompleteMsg{Filename: filename, KeyCount: count, Err: err}
	}
}

// Test connection
func TestConnectionCmd(host string, port int, password string, db int) tea.Cmd {
	return func() tea.Msg {
		if RedisClient == nil {
			return types.ConnectionTestMsg{Success: false, Err: nil}
		}
		latency, err := RedisClient.TestConnection(host, port, password, db)
		return types.ConnectionTestMsg{Success: err == nil, Latency: latency, Err: err}
	}
}

// ============== NEW FEATURE COMMANDS ==============

// Bulk delete
func BulkDeleteCmd(pattern string) tea.Cmd {
	return func() tea.Msg {
		if RedisClient == nil {
			return types.BulkDeleteMsg{Pattern: pattern, Err: nil}
		}
		deleted, err := RedisClient.BulkDelete(pattern)
		return types.BulkDeleteMsg{Pattern: pattern, Deleted: deleted, Err: err}
	}
}

// Batch TTL
func BatchSetTTLCmd(pattern string, ttl time.Duration) tea.Cmd {
	return func() tea.Msg {
		if RedisClient == nil {
			return types.BatchTTLSetMsg{Pattern: pattern, Err: nil}
		}
		count, err := RedisClient.BatchSetTTL(pattern, ttl)
		return types.BatchTTLSetMsg{Pattern: pattern, Count: count, TTL: ttl, Err: err}
	}
}

// Favorites
func LoadFavoritesCmd(connID int64) tea.Cmd {
	return func() tea.Msg {
		if Config == nil {
			return types.FavoritesLoadedMsg{Err: nil}
		}
		favorites := Config.ListFavorites(connID)
		return types.FavoritesLoadedMsg{Favorites: favorites, Err: nil}
	}
}

func AddFavoriteCmd(connID int64, key, label string) tea.Cmd {
	return func() tea.Msg {
		if Config == nil {
			return types.FavoriteAddedMsg{Err: nil}
		}
		fav, err := Config.AddFavorite(connID, key, label)
		return types.FavoriteAddedMsg{Favorite: fav, Err: err}
	}
}

func RemoveFavoriteCmd(connID int64, key string) tea.Cmd {
	return func() tea.Msg {
		if Config == nil {
			return types.FavoriteRemovedMsg{Err: nil}
		}
		err := Config.RemoveFavorite(connID, key)
		return types.FavoriteRemovedMsg{Key: key, Err: err}
	}
}

// Recent keys
func LoadRecentKeysCmd(connID int64) tea.Cmd {
	return func() tea.Msg {
		if Config == nil {
			return types.RecentKeysLoadedMsg{Err: nil}
		}
		keys := Config.ListRecentKeys(connID)
		return types.RecentKeysLoadedMsg{Keys: keys, Err: nil}
	}
}

func AddRecentKeyCmd(connID int64, key string, keyType types.KeyType) tea.Cmd {
	return func() tea.Msg {
		if Config != nil {
			Config.AddRecentKey(connID, key, keyType)
		}
		return nil
	}
}

// Regex search
func RegexSearchCmd(pattern string, maxKeys int) tea.Cmd {
	return func() tea.Msg {
		if RedisClient == nil {
			return types.RegexSearchResultMsg{Err: nil}
		}
		keys, err := RedisClient.ScanKeysWithRegex(pattern, maxKeys)
		return types.RegexSearchResultMsg{Keys: keys, Err: err}
	}
}

// Fuzzy search
func FuzzySearchCmd(term string, maxKeys int) tea.Cmd {
	return func() tea.Msg {
		if RedisClient == nil {
			return types.FuzzySearchResultMsg{Err: nil}
		}
		keys, err := RedisClient.FuzzySearchKeys(term, maxKeys)
		return types.FuzzySearchResultMsg{Keys: keys, Err: err}
	}
}

// Client list
func GetClientListCmd() tea.Cmd {
	return func() tea.Msg {
		if RedisClient == nil {
			return types.ClientListLoadedMsg{Err: nil}
		}
		clients, err := RedisClient.ClientList()
		return types.ClientListLoadedMsg{Clients: clients, Err: err}
	}
}

// Memory stats
func GetMemoryStatsCmd() tea.Cmd {
	return func() tea.Msg {
		if RedisClient == nil {
			return types.MemoryStatsLoadedMsg{Err: nil}
		}
		stats, err := RedisClient.GetMemoryStats()
		return types.MemoryStatsLoadedMsg{Stats: stats, Err: err}
	}
}

// Cluster info
func GetClusterInfoCmd() tea.Cmd {
	return func() tea.Msg {
		if RedisClient == nil {
			return types.ClusterInfoLoadedMsg{Err: nil}
		}
		nodes, err := RedisClient.ClusterNodes()
		info, _ := RedisClient.ClusterInfo()
		return types.ClusterInfoLoadedMsg{Nodes: nodes, Info: info, Err: err}
	}
}

// Compare keys
func CompareKeysCmd(key1, key2 string) tea.Cmd {
	return func() tea.Msg {
		if RedisClient == nil {
			return types.CompareKeysResultMsg{Err: nil}
		}
		val1, val2, err := RedisClient.CompareKeys(key1, key2)
		return types.CompareKeysResultMsg{Key1Value: val1, Key2Value: val2, Err: err}
	}
}

// Templates
func LoadTemplatesCmd() tea.Cmd {
	return func() tea.Msg {
		if Config == nil {
			return types.TemplatesLoadedMsg{Err: nil}
		}
		templates := Config.ListTemplates()
		return types.TemplatesLoadedMsg{Templates: templates, Err: nil}
	}
}

// Value history
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

// Watch key (returns a tick command)
func WatchKeyTickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return types.WatchTickMsg{}
	})
}

// Keyspace events
func SubscribeKeyspaceCmd(pattern string) tea.Cmd {
	return func() tea.Msg {
		if RedisClient == nil {
			return types.KeyspaceSubscribedMsg{Pattern: pattern, Err: nil}
		}
		// The handler will send events through a channel
		err := RedisClient.SubscribeKeyspace(pattern, func(event types.KeyspaceEvent) {
			// Events are handled in the client
		})
		return types.KeyspaceSubscribedMsg{Pattern: pattern, Err: err}
	}
}

// Tree view prefixes
func LoadKeyPrefixesCmd(separator string, maxDepth int) tea.Cmd {
	return func() tea.Msg {
		if RedisClient == nil {
			return types.TreeNodeExpandedMsg{Err: nil}
		}
		prefixes, err := RedisClient.GetKeyPrefixes(separator, maxDepth)
		return types.TreeNodeExpandedMsg{Children: prefixes, Err: err}
	}
}

// Clipboard (platform-specific, using exec)
func CopyToClipboardCmd(content string) tea.Cmd {
	return func() tea.Msg {
		// Use pbcopy on macOS
		cmd := exec.Command("pbcopy")
		cmd.Stdin = strings.NewReader(content)
		err := cmd.Run()
		return types.ClipboardCopiedMsg{Content: content, Err: err}
	}
}

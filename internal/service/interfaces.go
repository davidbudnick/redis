// Package service provides interfaces for dependency injection and testability.
package service

import (
	"crypto/tls"
	"time"

	"github.com/davidbudnick/redis-tui/internal/types"
	"github.com/redis/go-redis/v9"
)

// ConfigService defines the interface for configuration management.
type ConfigService interface {
	// Connection management
	ListConnections() ([]types.Connection, error)
	AddConnection(name, host string, port int, password string, db int) (types.Connection, error)
	UpdateConnection(id int64, name, host string, port int, password string, db int) (types.Connection, error)
	DeleteConnection(id int64) error

	// Favorites management
	AddFavorite(connID int64, key, label string) (types.Favorite, error)
	RemoveFavorite(connID int64, key string) error
	ListFavorites(connID int64) []types.Favorite
	IsFavorite(connID int64, key string) bool

	// Recent keys management
	AddRecentKey(connID int64, key string, keyType types.KeyType)
	ListRecentKeys(connID int64) []types.RecentKey
	ClearRecentKeys(connID int64)

	// Value history management
	AddValueHistory(key string, value types.RedisValue, action string)
	GetValueHistory(key string) []types.ValueHistoryEntry
	ClearValueHistory()

	// Templates management
	ListTemplates() []types.KeyTemplate
	AddTemplate(t types.KeyTemplate) error
	DeleteTemplate(name string) error

	// Groups management
	ListGroups() []types.ConnectionGroup
	AddGroup(name, color string) error
	AddConnectionToGroup(groupName string, connID int64) error
	RemoveConnectionFromGroup(groupName string, connID int64) error

	// Settings
	GetKeyBindings() types.KeyBindings
	SetKeyBindings(kb types.KeyBindings) error
	ResetKeyBindings() error
	GetTreeSeparator() string
	SetTreeSeparator(sep string) error
	GetWatchInterval() time.Duration

	// Lifecycle
	Close() error
}

// RedisService defines the interface for Redis operations.
type RedisService interface {
	// Connection management
	Connect(host string, port int, password string, db int) error
	ConnectWithTLS(host string, port int, password string, db int, tlsConfig *tls.Config) error
	ConnectCluster(addrs []string, password string) error
	Disconnect() error
	IsCluster() bool
	TestConnection(host string, port int, password string, db int) (time.Duration, error)

	// Key operations
	GetTotalKeys() int64
	ScanKeys(pattern string, cursor uint64, count int64) ([]types.RedisKey, uint64, error)
	ScanKeysWithRegex(regexPattern string, maxKeys int) ([]types.RedisKey, error)
	FuzzySearchKeys(searchTerm string, maxKeys int) ([]types.RedisKey, error)
	GetValue(key string) (types.RedisValue, error)
	DeleteKey(key string) error
	DeleteKeys(keys ...string) (int64, error)
	BulkDelete(pattern string) (int, error)
	Rename(oldKey, newKey string) error
	Copy(src, dst string, replace bool) error
	SearchByValue(pattern string, valueSearch string, maxKeys int) ([]types.RedisKey, error)
	CompareKeys(key1, key2 string) (types.RedisValue, types.RedisValue, error)
	GetKeyPrefixes(separator string, maxDepth int) ([]string, error)

	// String operations
	SetString(key, value string, ttl time.Duration) error

	// TTL operations
	SetTTL(key string, ttl time.Duration) error
	BatchSetTTL(pattern string, ttl time.Duration) (int, error)

	// List operations
	RPush(key string, values ...string) error
	LSet(key string, index int64, value string) error
	LRem(key string, count int64, value string) error

	// Set operations
	SAdd(key string, members ...string) error
	SRem(key string, members ...string) error

	// Sorted set operations
	ZAdd(key string, score float64, member string) error
	ZRem(key string, members ...string) error

	// Hash operations
	HSet(key, field, value string) error
	HDel(key string, fields ...string) error

	// Stream operations
	XAdd(key string, fields map[string]interface{}) (string, error)
	XDel(key string, ids ...string) error

	// Database operations
	SelectDB(db int) error
	FlushDB() error

	// Server info and monitoring
	GetServerInfo() (types.ServerInfo, error)
	GetMemoryStats() (types.MemoryStats, error)
	MemoryUsage(key string) (int64, error)
	SlowLogGet(count int64) ([]types.SlowLogEntry, error)
	ClientList() ([]types.ClientInfo, error)

	// Cluster operations
	ClusterNodes() ([]types.ClusterNode, error)
	ClusterInfo() (string, error)

	// Scripting
	Eval(script string, keys []string, args ...interface{}) (interface{}, error)

	// Pub/Sub
	Publish(channel, message string) (int64, error)
	Subscribe(channel string) *redis.PubSub
	PubSubChannels(pattern string) ([]string, error)

	// Keyspace events
	SubscribeKeyspace(pattern string, handler func(types.KeyspaceEvent)) error
	UnsubscribeKeyspace() error

	// Import/Export
	ExportKeys(pattern string) (map[string]interface{}, error)
	ImportKeys(data map[string]interface{}) (int, error)
}

package types

import "time"

// KeyType represents Redis data types
type KeyType string

const (
	KeyTypeString KeyType = "string"
	KeyTypeList   KeyType = "list"
	KeyTypeSet    KeyType = "set"
	KeyTypeZSet   KeyType = "zset"
	KeyTypeHash   KeyType = "hash"
	KeyTypeStream KeyType = "stream"
)

// RedisKey represents a key with metadata
type RedisKey struct {
	Key        string
	Type       KeyType
	TTL        time.Duration
	MemorySize int64
	IsFavorite bool
}

// RedisValue holds the value for any Redis type
type RedisValue struct {
	Type        KeyType
	StringValue string
	ListValue   []string
	SetValue    []string
	ZSetValue   []ZSetMember
	HashValue   map[string]string
	StreamValue []StreamEntry
}

// ZSetMember represents a sorted set member with score
type ZSetMember struct {
	Member string
	Score  float64
}

// StreamEntry represents a stream entry
type StreamEntry struct {
	ID     string
	Fields map[string]interface{}
}

// ServerInfo holds Redis server information
type ServerInfo struct {
	Version       string
	Mode          string
	OS            string
	UsedMemory    string
	PeakMemory    string
	Clients       string
	TotalKeys     string
	Uptime        string
	ConnectedDB   int
	ClusterMode   bool
	ClusterInfo   string
	ReplicaInfo   string
	RDBLastSave   string
	AOFEnabled    bool
	MemFragRatio  string
	CPUUsage      string
	TotalCommands string
}

// SlowLogEntry represents a slow query log entry
type SlowLogEntry struct {
	ID         int64
	Timestamp  time.Time
	Duration   time.Duration
	Command    string
	ClientAddr string
	ClientName string
}

// PubSubChannel represents a pub/sub channel
type PubSubChannel struct {
	Name        string
	Subscribers int64
}

// PubSubMessage represents a received pub/sub message
type PubSubMessage struct {
	Channel string
	Message string
	Time    time.Time
}

// ClientInfo represents connected client information
type ClientInfo struct {
	ID       int64
	Addr     string
	Name     string
	Age      time.Duration
	Idle     time.Duration
	Flags    string
	DB       int
	Cmd      string
	SubCount int
}

// ClusterNode represents a Redis cluster node
type ClusterNode struct {
	ID         string
	Addr       string
	Flags      string
	Role       string
	Master     string
	PingSent   int64
	PongRecv   int64
	ConfigEpoc int64
	LinkState  string
	Slots      string
	SlotStart  int
	SlotEnd    int
}

// MemoryStats holds memory statistics
type MemoryStats struct {
	TotalMemory        int64
	UsedMemory         int64
	PeakMemory         int64
	FragmentedBytes    int64
	FragRatio          float64
	FragmentationRatio float64
	RSS                string
	LuaMemory          string
	ByType             map[KeyType]int64
	TopKeys            []KeyMemory
}

// KeyMemory holds memory info for a specific key
type KeyMemory struct {
	Key    string
	Type   KeyType
	Memory int64
	Bytes  int64
}

// ValueHistoryEntry stores a previous value for undo
type ValueHistoryEntry struct {
	Key       string
	Value     RedisValue
	Timestamp time.Time
	Action    string // "set", "delete", "modify"
}

// KeyspaceEvent represents a keyspace notification
type KeyspaceEvent struct {
	Timestamp time.Time
	DB        int
	Event     string
	Key       string
}

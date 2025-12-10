package redis

import (
	"context"
	"crypto/tls"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/davidbudnick/redis/internal/types"

	"github.com/redis/go-redis/v9"
)

// silentLogger discards all log output from the Redis client
type silentLogger struct{}

func (l *silentLogger) Printf(ctx context.Context, format string, v ...interface{}) {}

func init() {
	// Disable go-redis internal logging to prevent noisy connection pool messages
	redis.SetLogger(&silentLogger{})
}

// Client wraps the Redis client with additional functionality
type Client struct {
	client  *redis.Client
	cluster *redis.ClusterClient
	ctx     context.Context

	host     string
	port     int
	password string
	db       int

	isCluster     bool
	pubsub        *redis.PubSub
	keyspacePS    *redis.PubSub
	eventHandlers []func(types.KeyspaceEvent)
}

// NewClient creates a new Redis client wrapper
func NewClient() *Client {
	return &Client{
		ctx:           context.Background(),
		eventHandlers: []func(types.KeyspaceEvent){},
	}
}

// Connect establishes a connection to Redis
func (c *Client) Connect(host string, port int, password string, db int) error {
	c.host = host
	c.port = port
	c.password = password
	c.db = db

	c.client = redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", host, port),
		Password:     password,
		DB:           db,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	ctx, cancel := context.WithTimeout(c.ctx, 5*time.Second)
	defer cancel()

	_, err := c.client.Ping(ctx).Result()
	return err
}

// ConnectWithTLS establishes a TLS connection to Redis
func (c *Client) ConnectWithTLS(host string, port int, password string, db int, tlsConfig *tls.Config) error {
	c.host = host
	c.port = port
	c.password = password
	c.db = db

	c.client = redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", host, port),
		Password:     password,
		DB:           db,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		TLSConfig:    tlsConfig,
	})

	ctx, cancel := context.WithTimeout(c.ctx, 5*time.Second)
	defer cancel()

	_, err := c.client.Ping(ctx).Result()
	return err
}

// ConnectCluster establishes a connection to a Redis cluster
func (c *Client) ConnectCluster(addrs []string, password string) error {
	c.isCluster = true

	c.cluster = redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:        addrs,
		Password:     password,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	ctx, cancel := context.WithTimeout(c.ctx, 5*time.Second)
	defer cancel()

	_, err := c.cluster.Ping(ctx).Result()
	return err
}

// Disconnect closes the Redis connection
func (c *Client) Disconnect() error {
	if c.pubsub != nil {
		_ = c.pubsub.Close()
	}
	if c.keyspacePS != nil {
		_ = c.keyspacePS.Close()
	}
	if c.cluster != nil {
		return c.cluster.Close()
	}
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

// IsCluster returns whether connected to a cluster
func (c *Client) IsCluster() bool {
	return c.isCluster
}

// GetTotalKeys returns the total number of keys in the current database
func (c *Client) GetTotalKeys() int64 {
	ctx, cancel := context.WithTimeout(c.ctx, 5*time.Second)
	defer cancel()
	count, err := c.client.DBSize(ctx).Result()
	if err != nil {
		return 0
	}
	return count
}

// ScanKeys scans keys matching a pattern
func (c *Client) ScanKeys(pattern string, cursor uint64, count int64) ([]types.RedisKey, uint64, error) {
	if pattern == "" {
		pattern = "*"
	}

	ctx, cancel := context.WithTimeout(c.ctx, 10*time.Second)
	defer cancel()

	keys, nextCursor, err := c.client.Scan(ctx, cursor, pattern, count).Result()
	if err != nil {
		return nil, 0, err
	}

	var result []types.RedisKey
	for _, key := range keys {
		keyType, _ := c.client.Type(ctx, key).Result()
		ttl, _ := c.client.TTL(ctx, key).Result()
		result = append(result, types.RedisKey{
			Key:  key,
			Type: types.KeyType(keyType),
			TTL:  ttl,
		})
	}

	return result, nextCursor, nil
}

// ScanKeysWithRegex scans keys using regex pattern
func (c *Client) ScanKeysWithRegex(regexPattern string, maxKeys int) ([]types.RedisKey, error) {
	re, err := regexp.Compile(regexPattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex: %w", err)
	}

	var result []types.RedisKey
	var cursor uint64
	count := 0

	for count < maxKeys {
		keys, nextCursor, err := c.client.Scan(c.ctx, cursor, "*", 100).Result()
		if err != nil {
			return result, err
		}

		for _, key := range keys {
			if count >= maxKeys {
				break
			}
			if re.MatchString(key) {
				keyType, _ := c.client.Type(c.ctx, key).Result()
				ttl, _ := c.client.TTL(c.ctx, key).Result()
				result = append(result, types.RedisKey{
					Key:  key,
					Type: types.KeyType(keyType),
					TTL:  ttl,
				})
				count++
			}
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return result, nil
}

// FuzzySearchKeys performs fuzzy matching on key names
func (c *Client) FuzzySearchKeys(searchTerm string, maxKeys int) ([]types.RedisKey, error) {
	searchLower := strings.ToLower(searchTerm)

	type scoredKey struct {
		key   types.RedisKey
		score int
	}
	var scoredKeys []scoredKey
	var cursor uint64
	count := 0

	for count < maxKeys*10 {
		keys, nextCursor, err := c.client.Scan(c.ctx, cursor, "*", 100).Result()
		if err != nil {
			break
		}

		for _, key := range keys {
			keyLower := strings.ToLower(key)
			score := fuzzyScore(keyLower, searchLower)
			if score > 0 {
				keyType, _ := c.client.Type(c.ctx, key).Result()
				ttl, _ := c.client.TTL(c.ctx, key).Result()
				scoredKeys = append(scoredKeys, scoredKey{
					key: types.RedisKey{
						Key:  key,
						Type: types.KeyType(keyType),
						TTL:  ttl,
					},
					score: score,
				})
			}
			count++
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	sort.Slice(scoredKeys, func(i, j int) bool {
		return scoredKeys[i].score > scoredKeys[j].score
	})

	var result []types.RedisKey
	limit := maxKeys
	if len(scoredKeys) < limit {
		limit = len(scoredKeys)
	}
	for i := 0; i < limit; i++ {
		result = append(result, scoredKeys[i].key)
	}

	return result, nil
}

func fuzzyScore(str, pattern string) int {
	if strings.Contains(str, pattern) {
		return 100 + (len(str) - len(pattern))
	}

	score := 0
	patternIdx := 0

	for i := 0; i < len(str) && patternIdx < len(pattern); i++ {
		if str[i] == pattern[patternIdx] {
			score += 10
			if i > 0 && (str[i-1] == ':' || str[i-1] == '_' || str[i-1] == '-') {
				score += 5
			}
			patternIdx++
		}
	}

	if patternIdx == len(pattern) {
		return score
	}
	return 0
}

// GetValue retrieves the value for a key
func (c *Client) GetValue(key string) (types.RedisValue, error) {
	keyType, err := c.client.Type(c.ctx, key).Result()
	if err != nil {
		return types.RedisValue{}, err
	}

	var value types.RedisValue
	value.Type = types.KeyType(keyType)

	switch keyType {
	case "string":
		val, err := c.client.Get(c.ctx, key).Result()
		if err != nil {
			return value, err
		}
		value.StringValue = val

	case "list":
		vals, err := c.client.LRange(c.ctx, key, 0, -1).Result()
		if err != nil {
			return value, err
		}
		value.ListValue = vals

	case "set":
		vals, err := c.client.SMembers(c.ctx, key).Result()
		if err != nil {
			return value, err
		}
		value.SetValue = vals

	case "zset":
		vals, err := c.client.ZRangeWithScores(c.ctx, key, 0, -1).Result()
		if err != nil {
			return value, err
		}
		for _, z := range vals {
			value.ZSetValue = append(value.ZSetValue, types.ZSetMember{
				Member: z.Member.(string),
				Score:  z.Score,
			})
		}

	case "hash":
		vals, err := c.client.HGetAll(c.ctx, key).Result()
		if err != nil {
			return value, err
		}
		value.HashValue = vals

	case "stream":
		entries, err := c.client.XRange(c.ctx, key, "-", "+").Result()
		if err != nil {
			return value, err
		}
		for _, entry := range entries {
			value.StreamValue = append(value.StreamValue, types.StreamEntry{
				ID:     entry.ID,
				Fields: entry.Values,
			})
		}
	}

	return value, nil
}

// DeleteKey deletes a single key
func (c *Client) DeleteKey(key string) error {
	return c.client.Del(c.ctx, key).Err()
}

// DeleteKeys deletes multiple keys
func (c *Client) DeleteKeys(keys ...string) (int64, error) {
	return c.client.Del(c.ctx, keys...).Result()
}

// BulkDelete deletes all keys matching a pattern
func (c *Client) BulkDelete(pattern string) (int, error) {
	var deleted int
	var cursor uint64

	for {
		keys, nextCursor, err := c.client.Scan(c.ctx, cursor, pattern, 100).Result()
		if err != nil {
			return deleted, err
		}

		if len(keys) > 0 {
			count, err := c.client.Del(c.ctx, keys...).Result()
			if err != nil {
				return deleted, err
			}
			deleted += int(count)
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return deleted, nil
}

// SetString sets a string value
func (c *Client) SetString(key, value string, ttl time.Duration) error {
	return c.client.Set(c.ctx, key, value, ttl).Err()
}

// SetTTL sets or removes TTL on a key
func (c *Client) SetTTL(key string, ttl time.Duration) error {
	if ttl <= 0 {
		return c.client.Persist(c.ctx, key).Err()
	}
	return c.client.Expire(c.ctx, key, ttl).Err()
}

// BatchSetTTL sets TTL on all keys matching a pattern
func (c *Client) BatchSetTTL(pattern string, ttl time.Duration) (int, error) {
	var count int
	var cursor uint64

	for {
		keys, nextCursor, err := c.client.Scan(c.ctx, cursor, pattern, 100).Result()
		if err != nil {
			return count, err
		}

		for _, key := range keys {
			var err error
			if ttl <= 0 {
				err = c.client.Persist(c.ctx, key).Err()
			} else {
				err = c.client.Expire(c.ctx, key, ttl).Err()
			}
			if err == nil {
				count++
			}
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return count, nil
}

// RPush appends values to a list
func (c *Client) RPush(key string, values ...string) error {
	args := make([]interface{}, len(values))
	for i, v := range values {
		args[i] = v
	}
	return c.client.RPush(c.ctx, key, args...).Err()
}

// SAdd adds members to a set
func (c *Client) SAdd(key string, members ...string) error {
	args := make([]interface{}, len(members))
	for i, v := range members {
		args[i] = v
	}
	return c.client.SAdd(c.ctx, key, args...).Err()
}

// ZAdd adds a member to a sorted set
func (c *Client) ZAdd(key string, score float64, member string) error {
	return c.client.ZAdd(c.ctx, key, redis.Z{Score: score, Member: member}).Err()
}

// HSet sets a hash field
func (c *Client) HSet(key, field, value string) error {
	return c.client.HSet(c.ctx, key, field, value).Err()
}

// XAdd adds an entry to a stream
func (c *Client) XAdd(key string, fields map[string]interface{}) (string, error) {
	return c.client.XAdd(c.ctx, &redis.XAddArgs{
		Stream: key,
		Values: fields,
	}).Result()
}

// LSet sets a list element by index
func (c *Client) LSet(key string, index int64, value string) error {
	return c.client.LSet(c.ctx, key, index, value).Err()
}

// LRem removes list elements
func (c *Client) LRem(key string, count int64, value string) error {
	return c.client.LRem(c.ctx, key, count, value).Err()
}

// SRem removes set members
func (c *Client) SRem(key string, members ...string) error {
	args := make([]interface{}, len(members))
	for i, v := range members {
		args[i] = v
	}
	return c.client.SRem(c.ctx, key, args...).Err()
}

// ZRem removes sorted set members
func (c *Client) ZRem(key string, members ...string) error {
	args := make([]interface{}, len(members))
	for i, v := range members {
		args[i] = v
	}
	return c.client.ZRem(c.ctx, key, args...).Err()
}

// HDel removes hash fields
func (c *Client) HDel(key string, fields ...string) error {
	return c.client.HDel(c.ctx, key, fields...).Err()
}

// XDel removes stream entries
func (c *Client) XDel(key string, ids ...string) error {
	return c.client.XDel(c.ctx, key, ids...).Err()
}

// Rename renames a key
func (c *Client) Rename(oldKey, newKey string) error {
	return c.client.Rename(c.ctx, oldKey, newKey).Err()
}

// Copy copies a key
func (c *Client) Copy(src, dst string, replace bool) error {
	return c.client.Copy(c.ctx, src, dst, 0, replace).Err()
}

// SelectDB switches the database
func (c *Client) SelectDB(db int) error {
	return c.client.Do(c.ctx, "SELECT", db).Err()
}

// MemoryUsage returns memory usage for a key
func (c *Client) MemoryUsage(key string) (int64, error) {
	return c.client.MemoryUsage(c.ctx, key).Result()
}

// GetServerInfo returns server information
func (c *Client) GetServerInfo() (types.ServerInfo, error) {
	info, err := c.client.Info(c.ctx).Result()
	if err != nil {
		return types.ServerInfo{}, err
	}

	var serverInfo types.ServerInfo
	lines := strings.Split(info, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key, value := parts[0], parts[1]
		switch key {
		case "redis_version":
			serverInfo.Version = value
		case "redis_mode":
			serverInfo.Mode = value
		case "os":
			serverInfo.OS = value
		case "used_memory_human":
			serverInfo.UsedMemory = value
		case "used_memory_peak_human":
			serverInfo.PeakMemory = value
		case "connected_clients":
			serverInfo.Clients = value
		case "uptime_in_seconds":
			if secs, err := strconv.Atoi(value); err == nil {
				serverInfo.Uptime = (time.Duration(secs) * time.Second).String()
			}
		case "cluster_enabled":
			serverInfo.ClusterMode = value == "1"
		case "mem_fragmentation_ratio":
			serverInfo.MemFragRatio = value
		case "total_commands_processed":
			serverInfo.TotalCommands = value
		case "aof_enabled":
			serverInfo.AOFEnabled = value == "1"
		}
	}

	dbInfo, err := c.client.DBSize(c.ctx).Result()
	if err == nil {
		serverInfo.TotalKeys = strconv.FormatInt(dbInfo, 10)
	}

	return serverInfo, nil
}

// GetMemoryStats returns memory statistics
func (c *Client) GetMemoryStats() (types.MemoryStats, error) {
	var stats types.MemoryStats
	stats.ByType = make(map[types.KeyType]int64)

	info, err := c.client.Info(c.ctx, "memory").Result()
	if err != nil {
		return stats, err
	}

	for _, line := range strings.Split(info, "\n") {
		parts := strings.SplitN(strings.TrimSpace(line), ":", 2)
		if len(parts) != 2 {
			continue
		}

		switch parts[0] {
		case "used_memory":
			stats.UsedMemory, _ = strconv.ParseInt(parts[1], 10, 64)
		case "used_memory_peak":
			stats.PeakMemory, _ = strconv.ParseInt(parts[1], 10, 64)
		case "mem_fragmentation_bytes":
			stats.FragmentedBytes, _ = strconv.ParseInt(parts[1], 10, 64)
		case "mem_fragmentation_ratio":
			stats.FragRatio, _ = strconv.ParseFloat(parts[1], 64)
		case "total_system_memory":
			stats.TotalMemory, _ = strconv.ParseInt(parts[1], 10, 64)
		}
	}

	stats.TopKeys = c.getTopKeysByMemory(20)

	return stats, nil
}

func (c *Client) getTopKeysByMemory(limit int) []types.KeyMemory {
	type keyMem struct {
		key    string
		typ    types.KeyType
		memory int64
	}

	// Pre-allocate with reasonable capacity to reduce allocations
	maxSamples := limit * 5
	allKeys := make([]keyMem, 0, maxSamples)
	var cursor uint64

	for {
		keys, nextCursor, err := c.client.Scan(c.ctx, cursor, "*", 100).Result()
		if err != nil {
			break
		}

		for _, key := range keys {
			mem, err := c.client.MemoryUsage(c.ctx, key).Result()
			if err != nil {
				continue
			}
			keyType, _ := c.client.Type(c.ctx, key).Result()
			allKeys = append(allKeys, keyMem{
				key:    key,
				typ:    types.KeyType(keyType),
				memory: mem,
			})
			// Break early if we have enough samples
			if len(allKeys) >= maxSamples {
				break
			}
		}

		cursor = nextCursor
		if cursor == 0 || len(allKeys) >= maxSamples {
			break
		}
	}

	sort.Slice(allKeys, func(i, j int) bool {
		return allKeys[i].memory > allKeys[j].memory
	})

	result := make([]types.KeyMemory, 0, limit)
	for i := 0; i < len(allKeys) && i < limit; i++ {
		result = append(result, types.KeyMemory{
			Key:    allKeys[i].key,
			Type:   allKeys[i].typ,
			Memory: allKeys[i].memory,
		})
	}

	return result
}

// GetLiveMetrics fetches real-time metrics from Redis INFO command
func (c *Client) GetLiveMetrics() (types.LiveMetricsData, error) {
	var data types.LiveMetricsData
	data.Timestamp = time.Now()

	// Get stats info
	info, err := c.client.Info(c.ctx, "stats", "memory", "clients", "cpu").Result()
	if err != nil {
		return data, err
	}

	for _, line := range strings.Split(info, "\n") {
		parts := strings.SplitN(strings.TrimSpace(line), ":", 2)
		if len(parts) != 2 {
			continue
		}

		key, value := parts[0], parts[1]
		switch key {
		case "instantaneous_ops_per_sec":
			data.OpsPerSec, _ = strconv.ParseFloat(value, 64)
		case "used_memory":
			data.UsedMemoryBytes, _ = strconv.ParseInt(value, 10, 64)
		case "connected_clients":
			data.ConnectedClients, _ = strconv.ParseInt(value, 10, 64)
		case "blocked_clients":
			data.BlockedClients, _ = strconv.ParseInt(value, 10, 64)
		case "keyspace_hits":
			data.KeyspaceHits, _ = strconv.ParseInt(value, 10, 64)
		case "keyspace_misses":
			data.KeyspaceMisses, _ = strconv.ParseInt(value, 10, 64)
		case "expired_keys":
			data.ExpiredKeys, _ = strconv.ParseInt(value, 10, 64)
		case "evicted_keys":
			data.EvictedKeys, _ = strconv.ParseInt(value, 10, 64)
		case "instantaneous_input_kbps":
			data.InputKbps, _ = strconv.ParseFloat(value, 64)
		case "instantaneous_output_kbps":
			data.OutputKbps, _ = strconv.ParseFloat(value, 64)
		case "used_cpu_sys":
			data.UsedCPUSys, _ = strconv.ParseFloat(value, 64)
		case "used_cpu_user":
			data.UsedCPUUser, _ = strconv.ParseFloat(value, 64)
		case "total_connections_received":
			data.TotalConnections, _ = strconv.ParseInt(value, 10, 64)
		case "rejected_connections":
			data.RejectedConns, _ = strconv.ParseInt(value, 10, 64)
		}
	}

	return data, nil
}

// FlushDB flushes the current database
func (c *Client) FlushDB() error {
	return c.client.FlushDB(c.ctx).Err()
}

// SlowLogGet returns slow log entries
func (c *Client) SlowLogGet(count int64) ([]types.SlowLogEntry, error) {
	result, err := c.client.SlowLogGet(c.ctx, count).Result()
	if err != nil {
		return nil, err
	}

	entries := make([]types.SlowLogEntry, len(result))
	for i, log := range result {
		entries[i] = types.SlowLogEntry{
			ID:         log.ID,
			Timestamp:  log.Time,
			Duration:   log.Duration,
			Command:    strings.Join(log.Args, " "),
			ClientAddr: log.ClientAddr,
			ClientName: log.ClientName,
		}
	}

	return entries, nil
}

// Eval executes a Lua script
func (c *Client) Eval(script string, keys []string, args ...interface{}) (interface{}, error) {
	return c.client.Eval(c.ctx, script, keys, args...).Result()
}

// Publish publishes a message to a channel
func (c *Client) Publish(channel, message string) (int64, error) {
	return c.client.Publish(c.ctx, channel, message).Result()
}

// Subscribe subscribes to channels
func (c *Client) Subscribe(channel string) *redis.PubSub {
	return c.client.Subscribe(c.ctx, channel)
}

// PubSubChannels lists active channels
func (c *Client) PubSubChannels(pattern string) ([]string, error) {
	return c.client.PubSubChannels(c.ctx, pattern).Result()
}

// ClientList returns connected clients
func (c *Client) ClientList() ([]types.ClientInfo, error) {
	result, err := c.client.ClientList(c.ctx).Result()
	if err != nil {
		return nil, err
	}

	var clients []types.ClientInfo
	lines := strings.Split(result, "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}

		client := types.ClientInfo{}
		fields := strings.Fields(line)

		for _, field := range fields {
			parts := strings.SplitN(field, "=", 2)
			if len(parts) != 2 {
				continue
			}

			switch parts[0] {
			case "id":
				client.ID, _ = strconv.ParseInt(parts[1], 10, 64)
			case "addr":
				client.Addr = parts[1]
			case "name":
				client.Name = parts[1]
			case "age":
				age, _ := strconv.ParseInt(parts[1], 10, 64)
				client.Age = time.Duration(age) * time.Second
			case "idle":
				idle, _ := strconv.ParseInt(parts[1], 10, 64)
				client.Idle = time.Duration(idle) * time.Second
			case "flags":
				client.Flags = parts[1]
			case "db":
				client.DB, _ = strconv.Atoi(parts[1])
			case "cmd":
				client.Cmd = parts[1]
			case "sub":
				client.SubCount, _ = strconv.Atoi(parts[1])
			}
		}

		clients = append(clients, client)
	}

	return clients, nil
}

// ClusterNodes returns cluster node information
func (c *Client) ClusterNodes() ([]types.ClusterNode, error) {
	var result string
	var err error

	if c.isCluster {
		result, err = c.cluster.ClusterNodes(c.ctx).Result()
	} else {
		result, err = c.client.ClusterNodes(c.ctx).Result()
	}

	if err != nil {
		return nil, err
	}

	return parseClusterNodes(result), nil
}

func parseClusterNodes(result string) []types.ClusterNode {
	var nodes []types.ClusterNode
	lines := strings.Split(result, "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 8 {
			continue
		}

		node := types.ClusterNode{
			ID:        fields[0],
			Addr:      fields[1],
			Flags:     fields[2],
			Master:    fields[3],
			LinkState: fields[7],
		}

		if len(fields) > 8 {
			node.Slots = strings.Join(fields[8:], " ")
		}

		nodes = append(nodes, node)
	}

	return nodes
}

// ClusterInfo returns cluster information string
func (c *Client) ClusterInfo() (string, error) {
	if c.isCluster {
		return c.cluster.ClusterInfo(c.ctx).Result()
	}
	return c.client.ClusterInfo(c.ctx).Result()
}

// SearchByValue searches for keys containing a value
func (c *Client) SearchByValue(pattern string, valueSearch string, maxKeys int) ([]types.RedisKey, error) {
	var result []types.RedisKey
	var cursor uint64
	count := 0

	for count < maxKeys {
		keys, nextCursor, err := c.client.Scan(c.ctx, cursor, pattern, 100).Result()
		if err != nil {
			return result, err
		}

		for _, key := range keys {
			if count >= maxKeys {
				break
			}

			keyType, _ := c.client.Type(c.ctx, key).Result()
			ttl, _ := c.client.TTL(c.ctx, key).Result()

			found := false
			switch keyType {
			case "string":
				val, _ := c.client.Get(c.ctx, key).Result()
				if strings.Contains(val, valueSearch) {
					found = true
				}
			case "hash":
				vals, _ := c.client.HGetAll(c.ctx, key).Result()
				for _, v := range vals {
					if strings.Contains(v, valueSearch) {
						found = true
						break
					}
				}
			case "list":
				vals, _ := c.client.LRange(c.ctx, key, 0, -1).Result()
				for _, v := range vals {
					if strings.Contains(v, valueSearch) {
						found = true
						break
					}
				}
			case "set":
				vals, _ := c.client.SMembers(c.ctx, key).Result()
				for _, v := range vals {
					if strings.Contains(v, valueSearch) {
						found = true
						break
					}
				}
			}

			if found {
				result = append(result, types.RedisKey{
					Key:  key,
					Type: types.KeyType(keyType),
					TTL:  ttl,
				})
				count++
			}
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return result, nil
}

// ExportKeys exports keys matching a pattern to a map
func (c *Client) ExportKeys(pattern string) (map[string]interface{}, error) {
	export := make(map[string]interface{})
	var cursor uint64

	for {
		keys, nextCursor, err := c.client.Scan(c.ctx, cursor, pattern, 100).Result()
		if err != nil {
			return export, err
		}

		for _, key := range keys {
			value, err := c.GetValue(key)
			if err != nil {
				continue
			}
			ttl, _ := c.client.TTL(c.ctx, key).Result()

			keyData := map[string]interface{}{
				"type": string(value.Type),
				"ttl":  ttl.Seconds(),
			}

			switch value.Type {
			case types.KeyTypeString:
				keyData["value"] = value.StringValue
			case types.KeyTypeList:
				keyData["value"] = value.ListValue
			case types.KeyTypeSet:
				keyData["value"] = value.SetValue
			case types.KeyTypeZSet:
				members := make([]map[string]interface{}, len(value.ZSetValue))
				for i, m := range value.ZSetValue {
					members[i] = map[string]interface{}{"member": m.Member, "score": m.Score}
				}
				keyData["value"] = members
			case types.KeyTypeHash:
				keyData["value"] = value.HashValue
			case types.KeyTypeStream:
				entries := make([]map[string]interface{}, len(value.StreamValue))
				for i, e := range value.StreamValue {
					entries[i] = map[string]interface{}{"id": e.ID, "fields": e.Fields}
				}
				keyData["value"] = entries
			}

			export[key] = keyData
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return export, nil
}

// ImportKeys imports keys from a map
func (c *Client) ImportKeys(data map[string]interface{}) (int, error) {
	count := 0

	for key, keyDataRaw := range data {
		keyData, ok := keyDataRaw.(map[string]interface{})
		if !ok {
			continue
		}

		keyType, _ := keyData["type"].(string)
		ttlSecs, _ := keyData["ttl"].(float64)
		ttl := time.Duration(ttlSecs) * time.Second

		switch keyType {
		case "string":
			if val, ok := keyData["value"].(string); ok {
				_ = c.SetString(key, val, ttl)
				count++
			}
		case "list":
			if vals, ok := keyData["value"].([]interface{}); ok {
				for _, v := range vals {
					if s, ok := v.(string); ok {
						_ = c.RPush(key, s)
					}
				}
				if ttl > 0 {
					_ = c.SetTTL(key, ttl)
				}
				count++
			}
		case "set":
			if vals, ok := keyData["value"].([]interface{}); ok {
				for _, v := range vals {
					if s, ok := v.(string); ok {
						_ = c.SAdd(key, s)
					}
				}
				if ttl > 0 {
					_ = c.SetTTL(key, ttl)
				}
				count++
			}
		case "zset":
			if vals, ok := keyData["value"].([]interface{}); ok {
				for _, v := range vals {
					if m, ok := v.(map[string]interface{}); ok {
						member, _ := m["member"].(string)
						score, _ := m["score"].(float64)
						_ = c.ZAdd(key, score, member)
					}
				}
				if ttl > 0 {
					_ = c.SetTTL(key, ttl)
				}
				count++
			}
		case "hash":
			if vals, ok := keyData["value"].(map[string]interface{}); ok {
				for field, val := range vals {
					if s, ok := val.(string); ok {
						_ = c.HSet(key, field, s)
					}
				}
				if ttl > 0 {
					_ = c.SetTTL(key, ttl)
				}
				count++
			}
		}
	}

	return count, nil
}

// TestConnection tests a connection
func (c *Client) TestConnection(host string, port int, password string, db int) (time.Duration, error) {
	testClient := redis.NewClient(&redis.Options{
		Addr:        fmt.Sprintf("%s:%d", host, port),
		Password:    password,
		DB:          db,
		DialTimeout: 5 * time.Second,
	})
	defer testClient.Close()

	start := time.Now()
	ctx, cancel := context.WithTimeout(c.ctx, 5*time.Second)
	defer cancel()

	_, err := testClient.Ping(ctx).Result()
	return time.Since(start), err
}

// GetKeyPrefixes returns all unique key prefixes (for tree view)
func (c *Client) GetKeyPrefixes(separator string, maxDepth int) ([]string, error) {
	prefixes := make(map[string]bool)
	var cursor uint64

	for {
		keys, nextCursor, err := c.client.Scan(c.ctx, cursor, "*", 500).Result()
		if err != nil {
			return nil, err
		}

		for _, key := range keys {
			parts := strings.Split(key, separator)
			for i := 1; i <= len(parts) && i <= maxDepth; i++ {
				prefix := strings.Join(parts[:i], separator)
				prefixes[prefix] = true
			}
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	result := make([]string, 0, len(prefixes))
	for p := range prefixes {
		result = append(result, p)
	}
	sort.Strings(result)

	return result, nil
}

// SubscribeKeyspace subscribes to keyspace notifications
func (c *Client) SubscribeKeyspace(pattern string, handler func(types.KeyspaceEvent)) error {
	// Enable keyspace notifications (may fail on managed Redis, but we try)
	_ = c.client.ConfigSet(c.ctx, "notify-keyspace-events", "KEA").Err()

	// Close existing subscription if any to prevent leaks
	if c.keyspacePS != nil {
		_ = c.keyspacePS.Close()
		c.keyspacePS = nil
	}

	// Clear old handlers to prevent memory leak and duplicate events
	c.eventHandlers = []func(types.KeyspaceEvent){handler}

	channel := "__keyspace@" + strconv.Itoa(c.db) + "__:" + pattern
	c.keyspacePS = c.client.PSubscribe(c.ctx, channel)

	go func() {
		ch := c.keyspacePS.Channel()
		for msg := range ch {
			event := types.KeyspaceEvent{
				Timestamp: time.Now(),
				DB:        c.db,
				Event:     msg.Payload,
				Key:       strings.TrimPrefix(msg.Channel, "__keyspace@"+strconv.Itoa(c.db)+"__:"),
			}
			for _, h := range c.eventHandlers {
				h(event)
			}
		}
	}()

	return nil
}

// UnsubscribeKeyspace unsubscribes from keyspace notifications
func (c *Client) UnsubscribeKeyspace() error {
	if c.keyspacePS != nil {
		return c.keyspacePS.Close()
	}
	return nil
}

// CompareKeys compares two keys and returns their values
func (c *Client) CompareKeys(key1, key2 string) (types.RedisValue, types.RedisValue, error) {
	val1, err := c.GetValue(key1)
	if err != nil {
		return types.RedisValue{}, types.RedisValue{}, fmt.Errorf("error getting key1: %w", err)
	}

	val2, err := c.GetValue(key2)
	if err != nil {
		return val1, types.RedisValue{}, fmt.Errorf("error getting key2: %w", err)
	}

	return val1, val2, nil
}

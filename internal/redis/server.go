package redis

import (
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/davidbudnick/redis/internal/types"
	"github.com/redis/go-redis/v9"
)

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

		if len(keys) == 0 {
			cursor = nextCursor
			if cursor == 0 {
				break
			}
			continue
		}

		// Use pipeline to batch MemoryUsage and Type calls
		pipe := c.client.Pipeline()
		memCmds := make([]*redis.IntCmd, len(keys))
		typeCmds := make([]*redis.StatusCmd, len(keys))

		for i, key := range keys {
			memCmds[i] = pipe.MemoryUsage(c.ctx, key)
			typeCmds[i] = pipe.Type(c.ctx, key)
		}

		_, _ = pipe.Exec(c.ctx)

		for i, key := range keys {
			mem, err := memCmds[i].Result()
			if err != nil {
				continue
			}
			keyType, _ := typeCmds[i].Result()
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

// GetLiveMetrics returns real-time server metrics
func (c *Client) GetLiveMetrics() (types.LiveMetricsData, error) {
	info, err := c.client.Info(c.ctx, "stats", "memory", "clients", "cpu").Result()
	if err != nil {
		return types.LiveMetricsData{}, err
	}

	data := types.LiveMetricsData{
		Timestamp: time.Now(),
	}

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

package redis

import (
	"regexp"
	"sort"
	"strings"

	"github.com/davidbudnick/redis-tui/internal/types"
	"github.com/redis/go-redis/v9"
)

// GetTotalKeys returns the total number of keys in the current database
func (c *Client) GetTotalKeys() int64 {
	count, err := c.client.DBSize(c.ctx).Result()
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

	keys, nextCursor, err := c.client.Scan(c.ctx, cursor, pattern, count).Result()
	if err != nil {
		return nil, 0, err
	}

	if len(keys) == 0 {
		return []types.RedisKey{}, nextCursor, nil
	}

	// Use pipeline to batch Type and TTL calls (fixes N+1 query pattern)
	pipe := c.client.Pipeline()
	typeCmds := make([]*redis.StatusCmd, len(keys))
	ttlCmds := make([]*redis.DurationCmd, len(keys))

	for i, key := range keys {
		typeCmds[i] = pipe.Type(c.ctx, key)
		ttlCmds[i] = pipe.TTL(c.ctx, key)
	}

	_, err = pipe.Exec(c.ctx)
	if err != nil && err != redis.Nil {
		return nil, 0, err
	}

	result := make([]types.RedisKey, len(keys))
	for i, key := range keys {
		keyType, _ := typeCmds[i].Result()
		ttl, _ := ttlCmds[i].Result()
		result[i] = types.RedisKey{
			Key:  key,
			Type: types.KeyType(keyType),
			TTL:  ttl,
		}
	}

	return result, nextCursor, nil
}

// ScanKeysWithRegex scans keys using regex pattern
func (c *Client) ScanKeysWithRegex(regexPattern string, maxKeys int) ([]types.RedisKey, error) {
	re, err := regexp.Compile(regexPattern)
	if err != nil {
		return nil, errInvalidRegex(err)
	}

	result := make([]types.RedisKey, 0, maxKeys)
	var cursor uint64

	for len(result) < maxKeys {
		keys, nextCursor, err := c.client.Scan(c.ctx, cursor, "*", 100).Result()
		if err != nil {
			return result, err
		}

		// Filter matching keys first
		matchingKeys := make([]string, 0, len(keys))
		for _, key := range keys {
			if re.MatchString(key) {
				matchingKeys = append(matchingKeys, key)
				if len(result)+len(matchingKeys) >= maxKeys {
					break
				}
			}
		}

		if len(matchingKeys) > 0 {
			// Use pipeline to batch Type and TTL calls
			pipe := c.client.Pipeline()
			typeCmds := make([]*redis.StatusCmd, len(matchingKeys))
			ttlCmds := make([]*redis.DurationCmd, len(matchingKeys))

			for i, key := range matchingKeys {
				typeCmds[i] = pipe.Type(c.ctx, key)
				ttlCmds[i] = pipe.TTL(c.ctx, key)
			}

			_, _ = pipe.Exec(c.ctx)

			for i, key := range matchingKeys {
				if len(result) >= maxKeys {
					break
				}
				keyType, _ := typeCmds[i].Result()
				ttl, _ := ttlCmds[i].Result()
				result = append(result, types.RedisKey{
					Key:  key,
					Type: types.KeyType(keyType),
					TTL:  ttl,
				})
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
		key   string
		score int
	}
	scoredKeys := make([]scoredKey, 0, maxKeys*2)
	var cursor uint64
	count := 0

	// First pass: find all matching keys with scores (no Redis calls for type/ttl)
	for count < maxKeys*10 {
		keys, nextCursor, err := c.client.Scan(c.ctx, cursor, "*", 100).Result()
		if err != nil {
			break
		}

		for _, key := range keys {
			keyLower := strings.ToLower(key)
			score := fuzzyScore(keyLower, searchLower)
			if score > 0 {
				scoredKeys = append(scoredKeys, scoredKey{key: key, score: score})
			}
			count++
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	// Sort by score descending
	sort.Slice(scoredKeys, func(i, j int) bool {
		return scoredKeys[i].score > scoredKeys[j].score
	})

	// Limit to maxKeys
	limit := maxKeys
	if len(scoredKeys) < limit {
		limit = len(scoredKeys)
	}
	scoredKeys = scoredKeys[:limit]

	if len(scoredKeys) == 0 {
		return []types.RedisKey{}, nil
	}

	// Use pipeline to batch Type and TTL calls for top results only
	pipe := c.client.Pipeline()
	typeCmds := make([]*redis.StatusCmd, len(scoredKeys))
	ttlCmds := make([]*redis.DurationCmd, len(scoredKeys))

	for i, sk := range scoredKeys {
		typeCmds[i] = pipe.Type(c.ctx, sk.key)
		ttlCmds[i] = pipe.TTL(c.ctx, sk.key)
	}

	_, _ = pipe.Exec(c.ctx)

	result := make([]types.RedisKey, len(scoredKeys))
	for i, sk := range scoredKeys {
		keyType, _ := typeCmds[i].Result()
		ttl, _ := ttlCmds[i].Result()
		result[i] = types.RedisKey{
			Key:  sk.key,
			Type: types.KeyType(keyType),
			TTL:  ttl,
		}
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

// SearchByValue searches for keys containing a value
func (c *Client) SearchByValue(pattern string, valueSearch string, maxKeys int) ([]types.RedisKey, error) {
	result := make([]types.RedisKey, 0, maxKeys)
	var cursor uint64

	for len(result) < maxKeys {
		keys, nextCursor, err := c.client.Scan(c.ctx, cursor, pattern, 100).Result()
		if err != nil {
			return result, err
		}

		if len(keys) == 0 {
			cursor = nextCursor
			if cursor == 0 {
				break
			}
			continue
		}

		// First pipeline: get types for all keys
		typePipe := c.client.Pipeline()
		typeCmds := make([]*redis.StatusCmd, len(keys))
		for i, key := range keys {
			typeCmds[i] = typePipe.Type(c.ctx, key)
		}
		_, _ = typePipe.Exec(c.ctx)

		// Build type map and group keys by type for value fetching
		keyTypes := make([]string, len(keys))
		for i := range keys {
			keyTypes[i], _ = typeCmds[i].Result()
		}

		// Second pipeline: get values based on type
		valuePipe := c.client.Pipeline()
		type valueCmd struct {
			idx     int
			keyType string
			strCmd  *redis.StringCmd
			hashCmd *redis.MapStringStringCmd
			listCmd *redis.StringSliceCmd
			setCmd  *redis.StringSliceCmd
		}
		valueCmds := make([]valueCmd, 0, len(keys))

		for i, key := range keys {
			kt := keyTypes[i]
			vc := valueCmd{idx: i, keyType: kt}
			switch kt {
			case "string":
				vc.strCmd = valuePipe.Get(c.ctx, key)
			case "hash":
				vc.hashCmd = valuePipe.HGetAll(c.ctx, key)
			case "list":
				vc.listCmd = valuePipe.LRange(c.ctx, key, 0, -1)
			case "set":
				vc.setCmd = valuePipe.SMembers(c.ctx, key)
			default:
				continue
			}
			valueCmds = append(valueCmds, vc)
		}
		_, _ = valuePipe.Exec(c.ctx)

		// Third pipeline: get TTL only for matching keys
		matchingIndices := make([]int, 0)
		for _, vc := range valueCmds {
			found := false
			switch vc.keyType {
			case "string":
				val, _ := vc.strCmd.Result()
				found = strings.Contains(val, valueSearch)
			case "hash":
				vals, _ := vc.hashCmd.Result()
				for _, v := range vals {
					if strings.Contains(v, valueSearch) {
						found = true
						break
					}
				}
			case "list":
				vals, _ := vc.listCmd.Result()
				for _, v := range vals {
					if strings.Contains(v, valueSearch) {
						found = true
						break
					}
				}
			case "set":
				vals, _ := vc.setCmd.Result()
				for _, v := range vals {
					if strings.Contains(v, valueSearch) {
						found = true
						break
					}
				}
			}
			if found {
				matchingIndices = append(matchingIndices, vc.idx)
			}
		}

		if len(matchingIndices) > 0 {
			ttlPipe := c.client.Pipeline()
			ttlCmds := make([]*redis.DurationCmd, len(matchingIndices))
			for i, idx := range matchingIndices {
				ttlCmds[i] = ttlPipe.TTL(c.ctx, keys[idx])
			}
			_, _ = ttlPipe.Exec(c.ctx)

			for i, idx := range matchingIndices {
				if len(result) >= maxKeys {
					break
				}
				ttl, _ := ttlCmds[i].Result()
				result = append(result, types.RedisKey{
					Key:  keys[idx],
					Type: types.KeyType(keyTypes[idx]),
					TTL:  ttl,
				})
			}
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return result, nil
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

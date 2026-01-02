package redis

import (
	"time"

	"github.com/davidbudnick/redis/internal/types"
	"github.com/redis/go-redis/v9"
)

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

		if len(keys) > 0 {
			// Use pipeline to batch TTL operations
			pipe := c.client.Pipeline()
			cmds := make([]*redis.BoolCmd, len(keys))

			for i, key := range keys {
				if ttl <= 0 {
					cmds[i] = pipe.Persist(c.ctx, key)
				} else {
					cmds[i] = pipe.Expire(c.ctx, key, ttl)
				}
			}

			_, _ = pipe.Exec(c.ctx)

			for _, cmd := range cmds {
				if cmd.Err() == nil {
					count++
				}
			}
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return count, nil
}

// MemoryUsage returns memory usage for a key
func (c *Client) MemoryUsage(key string) (int64, error) {
	return c.client.MemoryUsage(c.ctx, key).Result()
}

// Rename renames a key
func (c *Client) Rename(oldKey, newKey string) error {
	return c.client.Rename(c.ctx, oldKey, newKey).Err()
}

// Copy copies a key
func (c *Client) Copy(src, dst string, replace bool) error {
	return c.client.Copy(c.ctx, src, dst, 0, replace).Err()
}

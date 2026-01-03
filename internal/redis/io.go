package redis

import (
	"fmt"
	"time"

	"github.com/davidbudnick/redis-tui/internal/types"
)

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

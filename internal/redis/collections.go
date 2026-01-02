package redis

import "github.com/redis/go-redis/v9"

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

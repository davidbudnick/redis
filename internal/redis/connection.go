package redis

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

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

// SelectDB switches the database
func (c *Client) SelectDB(db int) error {
	return c.client.Do(c.ctx, "SELECT", db).Err()
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

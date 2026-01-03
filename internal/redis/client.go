package redis

import (
	"context"

	"github.com/davidbudnick/redis-tui/internal/types"
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

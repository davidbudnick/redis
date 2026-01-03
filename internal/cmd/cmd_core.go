package cmd

import (
	"github.com/davidbudnick/redis-tui/internal/db"
	"github.com/davidbudnick/redis-tui/internal/redis"
)

var (
	Config      *db.Config
	RedisClient *redis.Client
)

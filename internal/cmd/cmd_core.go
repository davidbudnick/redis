package cmd

import (
	"github.com/davidbudnick/redis/internal/db"
	"github.com/davidbudnick/redis/internal/redis"
)

var (
	Config      *db.Config
	RedisClient *redis.Client
)

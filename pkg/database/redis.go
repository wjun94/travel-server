package database

import (
	"context"
	"log"
	"strconv"

	"github.com/go-redis/redis/v8"

	"travel-server/pkg/config"
)

var RedisClient *redis.Client

// InitRedis 初始化 Redis 连接
func InitRedis() {
	cfg := config.AppConfig
	// 将 REDIS_DB 字符串转为整数
	dbNum, err := strconv.Atoi(cfg.REDIS_DB)
	if err != nil {
		dbNum = 0
	}
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     cfg.REDIS_HOST + ":" + cfg.REDIS_PORT,
		Password: cfg.REDIS_PASSWORD,
		DB:       dbNum,
	})
	if err := RedisClient.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("Redis连接失败: %v", err)
	}
}

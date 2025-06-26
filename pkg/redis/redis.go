package redis

import (
	"github.com/hedeqiang/skeleton/internal/config"
	"context"

	"github.com/redis/go-redis/v9"
)

// NewRedis 根据提供的配置初始化 Redis 客户端
func NewRedis(cfg *config.Redis) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// 使用 Ping 命令检查连接是否正常
	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}

	return rdb, nil
}

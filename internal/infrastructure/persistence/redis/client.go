package redis

import (
	"context"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

func NewRedisClient(redisURL string) (*goredis.Client, error) {
	opt, err := goredis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}

	rdb := goredis.NewClient(opt)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return rdb, nil
}

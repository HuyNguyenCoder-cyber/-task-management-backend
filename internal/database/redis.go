package database

import (
	"context"

	"github.com/redis/go-redis/v9"
)

func ConnectRedis(addr string) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	ctx := context.Background()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return client, nil
}
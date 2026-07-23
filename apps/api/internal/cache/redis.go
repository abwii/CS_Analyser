package cache

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

func Connect(redisURL string) (*redis.Client, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}
	return redis.NewClient(opts), nil
}

func GetJSON[T any](ctx context.Context, client *redis.Client, key string) (T, bool, error) {
	var value T

	raw, err := client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return value, false, nil
		}
		return value, false, err
	}

	if err := json.Unmarshal(raw, &value); err != nil {
		return value, false, err
	}

	return value, true, nil
}

func SetJSON(ctx context.Context, client *redis.Client, key string, value interface{}, ttl time.Duration) error {
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return client.Set(ctx, key, raw, ttl).Err()
}

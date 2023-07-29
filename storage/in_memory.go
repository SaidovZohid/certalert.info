package storage

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type InMemoryStorageI interface {
	Set(key, value string, exp time.Duration) error
	Get(key string) (string, error)
}

type storageRedis struct {
	client *redis.Client
}

func NewInMemoryStorage(rdb *redis.Client) InMemoryStorageI {
	return &storageRedis{
		client: rdb,
	}
}

func (rd *storageRedis) Set(key, value string, exp time.Duration) error {
	err := rd.client.Set(context.Background(), key, value, exp).Err()
	if err != nil {
		return err
	}
	return nil
}

func (rd *storageRedis) Get(key string) (string, error) {
	val, err := rd.client.Get(context.Background(), key).Result()
	if err != nil {
		return "", err
	}
	return val, nil
}

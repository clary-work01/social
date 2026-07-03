package cache

import "github.com/redis/go-redis/v9"

func NewRedisClient(addr, pwd string, db int) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: pwd,
		DB:       db,
	})
}

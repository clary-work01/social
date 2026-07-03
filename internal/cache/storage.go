package cache

import (
	"context"

	"github.com/chainflow/chainflow-api/internal/store"
	"github.com/redis/go-redis/v9"
)

type Storage struct {
	Users interface {
		GetByID(context.Context, int64) (*store.User, error)
		Set(context.Context, *store.User) error
	}
}

func NewRedisStorage(rdb *redis.Client) Storage {
	return Storage{
		Users: &userStore{rdb: rdb},
	}
}

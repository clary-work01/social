package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/chainflow/chainflow-api/internal/store"
	"github.com/redis/go-redis/v9"
)

const UserExpTime = time.Minute

type userStore struct {
	rdb *redis.Client
}

func (s *userStore) GetByID(ctx context.Context, userID int64) (*store.User, error) {
	cacheKey := fmt.Sprintf("user-%d", userID)

	res, err := s.rdb.Get(ctx, cacheKey).Result()
	if err == redis.Nil { // redis:nil
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	user := &store.User{}
	if res != "" {
		err := json.Unmarshal([]byte(res), user)
		if err != nil {
			return nil, err
		}
	}

	return user, nil
}

func (s *userStore) Set(ctx context.Context, user *store.User) error {
	cacheKey := fmt.Sprintf("user-%d", user.ID)

	data, err := json.Marshal(user)
	if err != nil {
		return err
	}
	return s.rdb.Set(ctx, cacheKey, data, UserExpTime).Err()
}

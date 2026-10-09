package cache

import (
	"SocialMedia/internal/store"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type UserStore struct {
	rdb *redis.Client
}

const TTL = 24 * time.Hour // Set a default TTL for cached users

func (s *UserStore) Get(ctx context.Context, id int64) (*store.User, error) {
	if s == nil || s.rdb == nil {
		return nil, nil
	}

	cacheKey := fmt.Sprintf("user:%d", id)
	data, err := s.rdb.Get(ctx, cacheKey).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var user store.User
	if err := json.Unmarshal([]byte(data), &user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *UserStore) Set(ctx context.Context, user *store.User) error {
	if s == nil || s.rdb == nil || user == nil {
		return nil
	}

	cacheKey := fmt.Sprintf("user:%d", user.ID)
	data, err := json.Marshal(user)
	if err != nil {
		return err
	}
	return s.rdb.SetEX(ctx, cacheKey, data, TTL).Err()
}

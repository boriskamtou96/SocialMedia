package cache

import (
	"SocialMedia/internal/store"
	"context"

	"github.com/go-redis/redis/v8"
)

type Storage struct {
	Users interface {
		Get(ctx context.Context, id int64) (*store.User, error)
		Set(ctx context.Context, user *store.User) error
	}
}

func NewRedisStorage(rdb *redis.Client) Storage {
	return Storage{
		Users: &UserStore{rdb: rdb},
	}
}

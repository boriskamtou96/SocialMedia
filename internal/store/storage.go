package store

import (
	"context"
	"database/sql"
)

type Storage struct {
	Posts interface {
		Create(ctx context.Context, post *Post) error
		GetById(ctx context.Context, id int64) (*Post, error)
	}
	Users interface {
		Create(ctx context.Context, user *User) error
		GetUsers(ctx context.Context) ([]User, error)
	}
}

func NewStorage(db *sql.DB) Storage {
	return Storage{
		Posts: &PostsStore{
			db: db,
		},
		Users: &UsersStore{
			db: db,
		},
	}
}

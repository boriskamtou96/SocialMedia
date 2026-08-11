package store

import (
	"context"
	"database/sql"
)

type Storage struct {
	Posts interface {
		Create(ctx context.Context, post *Post) error
		GetById(ctx context.Context, id int64) (*Post, error)
		DeletePostById(ctx context.Context, id int64) error
		UpdatePostById(ctx context.Context, post *Post) error
	}
	Users interface {
		Create(ctx context.Context, user *User) error
		GetUsers(ctx context.Context) ([]User, error)
		GetById(ctx context.Context, id int64) (*User, error)
	}
	Comments interface {
		GetByPostID(ctx context.Context, postID int64) ([]Comment, error)
		Create(ctx context.Context, user *Comment) error
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
		Comments: &CommentsStore{
			db: db,
		},
	}
}

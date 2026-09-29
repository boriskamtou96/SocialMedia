package store

import (
	"context"
	"database/sql"
	"time"
)

type Storage struct {
	Posts interface {
		Create(ctx context.Context, post *Post) error
		GetById(ctx context.Context, id int64) (*Post, error)
		DeletePostById(ctx context.Context, id int64) error
		UpdatePostById(ctx context.Context, post *Post) error
		GetUserFeed(ctx context.Context, userID int64, query PaginatedFeedQuery) ([]*PostWithMetaData, error)
	}
	Users interface {
		Create(ctx context.Context, tx *sql.Tx, user *User) error
		GetUsers(ctx context.Context) ([]User, error)
		GetById(ctx context.Context, id int64) (*User, error)
		CreateAndInvite(ctx context.Context, user *User, token string, invitationExp time.Duration) error
	}
	Comments interface {
		GetByPostID(ctx context.Context, postID int64) ([]Comment, error)
		Create(ctx context.Context, user *Comment) error
	}
	Followers interface {
		Follow(ctx context.Context, userID, followerID int64) error
		UnFollow(ctx context.Context, userID, followerID int64) error
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
		Followers: &FollowerStore{
			db: db,
		},
	}
}

func withTx(ctx context.Context, db *sql.DB, fn func(tx *sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

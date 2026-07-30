package store

import (
	"context"
	"database/sql"
)

type Comment struct {
	ID        int64  `json:"id"`
	UserID    int64  `json:"userId"`
	PostID    int64  `json:"postId"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type CommentsStore struct {
	db *sql.DB
}

func (c CommentsStore) Create(ctx context.Context, comment *Comment) error {
	return nil
}

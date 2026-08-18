package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"
)

type Follower struct {
	UserID     int64  `json:"user_id"`
	FollowerID int64  `json:"follower_id"`
	CreatedAt  string `json:"created_at"`
}

type FollowerStore struct {
	db *sql.DB
}

func (s *FollowerStore) Follow(ctx context.Context, userID, followerID int64) error {
	query := `
	INSERT INTO followers (user_id, follower_id) VALUES ($1, $2)
`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeOut)
	defer cancel()

	_, err := s.db.ExecContext(ctx, query, userID, followerID)
	if err != nil {
		if pqErr, ok := errors.AsType[*pq.Error](err); ok && pqErr.Code == "23505" {
			return fmt.Errorf("follower already exists")
		}
		return err
	}

	return nil
}

func (s *FollowerStore) UnFollow(ctx context.Context, userID, followerID int64) error {
	query := `
	DELETE FROM followers WHERE user_id = $1 AND follower_id = $2
`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeOut)
	defer cancel()

	_, err := s.db.ExecContext(ctx, query, userID, followerID)

	if err != nil {
		return err
	}

	return nil
}

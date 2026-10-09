package store

import (
	"context"
	"database/sql"
	"errors"
)

type Role struct {
	ID          int64  `db:"id"`
	Name        string `db:"name"`
	Description string `db:"description"`
	Level       int    `db:"level"`
}

type RoleStore struct {
	db *sql.DB
}

func (s *RoleStore) GetByName(ctx context.Context, slug string) (*Role, error) {
	query := `
		SELECT id, name, description, level
		FROM roles
		WHERE name = $1
	`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeOut)
	defer cancel()
	row := s.db.QueryRowContext(ctx, query, slug)
	role := &Role{}
	err := row.Scan(&role.ID, &role.Name, &role.Description, &role.Level)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return role, nil
}

package store

import (
	"context"
	"database/sql"
	"errors"
)

type User struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Password  string `json:"-"`
	CreatedAt string `json:"created_at"`
}

type UsersStore struct {
	db *sql.DB
}

func (s *UsersStore) Create(ctx context.Context, user *User) error {
	query := `
		INSERT INTO users (username,email, password)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeOut)
	defer cancel()

	err := s.db.QueryRowContext(
		ctx,
		query,
		user.Username,
		user.Email,
		user.Password,
	).Scan(&user.ID, &user.CreatedAt)
	if err != nil {
		return err
	}

	return nil
}

func (s *UsersStore) GetUsers(ctx context.Context) ([]User, error) {
	query := `
	SELECT id, username, email, created_at
	FROM users
`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeOut)
	defer cancel()
	stmt, err := s.db.PrepareContext(ctx, query)
	if err != nil {
		return nil, errors.New("error preparing query")
	}
	defer stmt.Close()

	rows, err := stmt.Query()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("no user in database")
		}
		return nil, errors.New("error executing query")
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt)
		if err != nil {
			return nil, errors.New("error scanning row")
		}
		users = append(users, user)
	}
	return users, nil
}

func (s *UsersStore) GetById(ctx context.Context, id int64) (*User, error) {
	query := `
		SELECT id, username, email, created_at
		FROM users
		WHERE id = $1
`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeOut)
	defer cancel()
	row := s.db.QueryRowContext(ctx, query, id)
	user := &User{}
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, errors.New("error scanning row")
	}
	return user, nil
}

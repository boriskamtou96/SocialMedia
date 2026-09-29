package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        int64    `json:"id"`
	Username  string   `json:"username"`
	Email     string   `json:"email"`
	Password  Password `json:"-"`
	CreatedAt string   `json:"created_at"`
}

type Password struct {
	text *string
	hash []byte
}

func (s *Password) Set(text string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(text), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	s.text = &text
	s.hash = hash
	return nil
}

type UsersStore struct {
	db *sql.DB
}

func (s *UsersStore) Create(ctx context.Context, tx *sql.Tx, user *User) error {
	query := `
		INSERT INTO users (username, email, password)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeOut)
	defer cancel()

	err := tx.QueryRowContext(
		ctx,
		query,
		user.Username,
		user.Email,
		user.Password.hash,
	).Scan(&user.ID, &user.CreatedAt)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return errors.New("no rows returned")
		case err.Error() == "pq: duplicate key value violates unique constraint \"users_username_key\"":
			return errors.New("username already exists")
		case err.Error() == "pq: duplicate key value violates unique constraint \"users_email_key\"":
			return errors.New("email already exists")
		default:
			return err
		}
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

func (s *UsersStore) CreateAndInvite(ctx context.Context, user *User, token string, invitationExp time.Duration) error {

	return withTx(ctx, s.db, func(tx *sql.Tx) error {
		if err := s.Create(ctx, tx, user); err != nil {
			return err
		}

		err := s.createUserInvitation(ctx, tx, token, invitationExp, user.ID)
		if err != nil {
			return err
		}
		return nil
	})
}

func (s *UsersStore) createUserInvitation(ctx context.Context, tx *sql.Tx, token string, invitationExp time.Duration, userID int64) error {
	// Implementation for creating user invitation
	query := `
		INSERT INTO invitations (token, user_id, expiry)
		VALUES ($1, $2, $3)
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeOut)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, token, userID, time.Now().Add(invitationExp))
	if err != nil {
		return err
	}
	return nil
}

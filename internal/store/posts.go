package store

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/lib/pq"
)

var (
	ErrNotFound  = errors.New("resource not found")
	QueryTimeOut = 5 * time.Second
)

type Post struct {
	ID        int64     `json:"id"`
	Content   string    `json:"content"`
	Title     string    `json:"title"`
	UserID    int64     `json:"userId"`
	Tags      []string  `json:"tags"`
	CreatedAt string    `json:"createdAt"`
	UpdatedAt string    `json:"updatedAt"`
	Version   int       `json:"version"`
	Comments  []Comment `json:"comments"`
	User      User      `json:"user"`
}

type PostWithMetaData struct {
	Post
	CommentCount int `json:"comments_count"`
}

type PostsStore struct {
	db *sql.DB
}

func (s *PostsStore) Create(ctx context.Context, post *Post) error {
	query := `
		INSERT INTO posts (content, title, user_id, tags)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeOut)
	defer cancel()

	err := s.db.QueryRowContext(
		ctx,
		query,
		post.Content,
		post.Title,
		post.UserID,
		pq.Array(post.Tags),
	).Scan(&post.ID, &post.CreatedAt, &post.UpdatedAt)
	if err != nil {
		return err
	}

	return nil
}

func (s *PostsStore) GetById(ctx context.Context, id int64) (*Post, error) {
	query := `
		SELECT id, content, title, user_id, tags, created_at, updated_at, version
		FROM posts
		WHERE id = $1
`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeOut)
	defer cancel()

	row := s.db.QueryRowContext(ctx, query, id)

	post := &Post{}
	err := row.Scan(
		&post.ID,
		&post.Content,
		&post.Title, &post.UserID,
		pq.Array(&post.Tags),
		&post.CreatedAt,
		&post.UpdatedAt,
		&post.Version,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
	}

	return post, nil
}

func (s *PostsStore) DeletePostById(ctx context.Context, id int64) error {
	query := `
		DELETE FROM posts
		WHERE id = $1
`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeOut)
	defer cancel()

	res, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}
	return nil

}

func (s *PostsStore) UpdatePostById(ctx context.Context, post *Post) error {
	query := `
	UPDATE posts
	SET title=$1, content=$2, version = version + 1
	WHERE id = $3 AND version = $4
	RETURNING version
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeOut)
	defer cancel()

	err := s.db.QueryRowContext(
		ctx,
		query,
		post.Title,
		post.Content,
		post.ID,
		post.Version,
	).Scan(&post.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrNotFound
		default:
			return err
		}
	}
	return nil
}

func (s *PostsStore) GetUserFeed(ctx context.Context, userID int64, fq PaginatedFeedQuery) ([]*PostWithMetaData, error) {
	// 1. Sécuriser et normaliser la direction du tri
	sortOrder := "DESC"
	if strings.ToLower(fq.Sort) == "asc" {
		sortOrder = "ASC"
	}

	// 2. Requête SQL (avec injection propre du mot-clé ASC/DESC)
	query := `
       SELECT p.id, p.content, p.title, p.user_id, p.tags, p.created_at,
              COUNT(c.id) AS comments_count
       FROM posts p
       LEFT JOIN comments c  ON p.id = c.post_id
       LEFT JOIN followers f ON f.follower_id = p.user_id
       WHERE 
        f.user_id = $1 AND
		(p.title ILIKE '%' || $4 || '%' OR p.content ILIKE '%' || $4 || '%') 
       GROUP BY p.id
       ORDER BY p.created_at ` + sortOrder + `
       LIMIT $2 OFFSET $3
    `

	ctx, cancel := context.WithTimeout(ctx, QueryTimeOut)
	defer cancel()

	// 3. Passage exact des 3 paramètres correspondant à $1, $2 et $3
	rows, err := s.db.QueryContext(ctx, query, userID, fq.Limit, fq.Offset, fq.Search)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*PostWithMetaData
	for rows.Next() {
		post := &PostWithMetaData{}
		err := rows.Scan(
			&post.ID,
			&post.Content,
			&post.Title,
			&post.UserID,
			pq.Array(&post.Tags),
			&post.CreatedAt,
			&post.CommentCount,
		)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}

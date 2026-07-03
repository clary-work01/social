package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
)

// model
type Post struct {
	ID        int64     `json:"id"`
	Content   string    `json:"content"`
	Title     string    `json:"title"`
	UserID    int64     `json:"user_id"`
	Tags      []string  `json:"tags"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
	Version   int       `json:"version"`
	Comments  []Comment `json:"comments"`
	User      User      `json:"user"`
}

type postStore struct {
	db *sql.DB
}

func (s *postStore) Create(ctx context.Context, post *Post) error {
	query := `
		INSERT INTO posts (content, title, user_id, tags)
		VALUES ($1, $2, $3, $4) 
		RETURNING id, created_at, updated_at
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err := s.db.QueryRowContext(
		ctx,
		query,
		post.Content,
		post.Title,
		post.UserID,
		pq.Array(post.Tags),
	).Scan(
		&post.ID,
		&post.CreatedAt,
		&post.UpdatedAt,
	)
	if err != nil {
		return err
	}
	return nil
}

func (s *postStore) GetByID(ctx context.Context, id int64) (*Post, error) {
	query := `
		SELECT id, content, user_id, title, tags, version, created_at, updated_at
		FROM posts
		WHERE id = $1 
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	var post Post
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&post.ID,
		&post.Content,
		&post.UserID,
		&post.Title,
		pq.Array(&post.Tags),
		&post.Version,
		&post.CreatedAt,
		&post.UpdatedAt,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return &post, nil
}

func (s *postStore) DeleteByID(ctx context.Context, id int64) error {
	query := `
		DELETE FROM posts
 		WHERE posts.id = $1
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
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

func (s *postStore) Update(ctx context.Context, post *Post) error {
	query := `
		UPDATE posts SET title=$1, content=$2, tags=$3, version = version + 1
		WHERE id = $4 AND version = $5
		RETURNING version
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err := s.db.QueryRowContext(
		ctx,
		query,
		post.Title,
		post.Content,
		pq.Array(post.Tags),
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

type UserFeed struct {
	Post
	Comment_count int64 `json:"comment_count"`
}

type FeedQueryParam struct {
	Offset int        `json:"offset" validate:"min=0"`
	Limit  int        `json:"limit" validate:"min=1,max=10"`
	Sort   string     `json:"sort" validate:"oneof=asc desc"`
	Search string     `json:"search"`
	Tags   []string   `json:"tags"`
	Since  *time.Time `json:"since"`
	Until  *time.Time `json:"until"`
}

func (s *postStore) GetUserFeed(ctx context.Context, id int64, fq FeedQueryParam) ([]UserFeed, error) {
	args := []any{id, fq.Offset, fq.Limit}
	argIdx := 4

	searchFilter := ""
	if fq.Search != "" {
		searchFilter = fmt.Sprintf("AND (p.title ILIKE $%d OR p.content ILIKE $%d)", argIdx, argIdx)
		args = append(args, "%"+fq.Search+"%")
		argIdx++
	}

	tagsFilter := ""
	if len(fq.Tags) > 0 {
		tagsFilter = fmt.Sprintf("AND p.tags @> $%d", argIdx)
		args = append(args, pq.Array(fq.Tags))

		argIdx++
	}

	sinceFilter := ""
	if fq.Since != nil {
		sinceFilter = fmt.Sprintf("AND p.created_at >= $%d", argIdx)
		args = append(args, fq.Since)

		argIdx++
	}

	untilFilter := ""
	if fq.Until != nil {
		untilFilter = fmt.Sprintf("AND p.created_at <= $%d", argIdx)
		args = append(args, fq.Until)

		argIdx++
	}

	query := fmt.Sprintf(`
        SELECT p.id, p.user_id, p.title, p.content, p.created_at, p.version, p.tags,
               COUNT(c.id) AS comment_count, u.username
        FROM posts p
        LEFT JOIN comments c ON c.post_id = p.id
        JOIN users u ON u.id = p.user_id
        WHERE (p.user_id = $1 OR p.user_id IN (
            SELECT user_id FROM followers WHERE follower_id = $1
        ))
       	%s
        %s
		%s
		%s
        GROUP BY p.id, u.username
        ORDER BY p.created_at %s
        OFFSET $2 LIMIT $3
    `, tagsFilter, searchFilter, sinceFilter, untilFilter, fq.Sort)

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []UserFeed
	for rows.Next() {
		var post UserFeed
		err := rows.Scan(
			&post.ID,
			&post.UserID,
			&post.Title,
			&post.Content,
			&post.CreatedAt,
			&post.Version,
			pq.Array(&post.Tags),
			&post.Comment_count,
			&post.User.UserName,
		)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	return posts, nil
}

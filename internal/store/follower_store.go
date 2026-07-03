package store

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"github.com/lib/pq"
)

// model
type Follower struct {
	FollowerID int64  `json:"follower_id"`
	UserID     int64  `json:"user_id"`
	CreatedAt  string `json:"created_at"`
}

type followerStore struct {
	db *sql.DB
}

func (s *followerStore) Follow(ctx context.Context, followerID, userID int64) error {
	query := `
		INSERT INTO followers (follower_id, user_id)
		VALUES ($1, $2)
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := s.db.ExecContext(ctx, query, followerID, userID)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return ErrConflict
		}
		return err
	}

	return nil
}

func (s *followerStore) UnFollow(ctx context.Context, followerID, userID int64) error {
	log.Println("id:", followerID, userID)
	query := `
		DELETE FROM followers
		WHERE follower_id = $1 AND user_id = $2
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := s.db.ExecContext(ctx, query, followerID, userID)
	return err
}

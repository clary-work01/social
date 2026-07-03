package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrDuplicateEmail    = errors.New("duplicate email")
	ErrDuplicateUsername = errors.New("duplicate username")
)

// model
type User struct {
	ID        int64    `json:"id"`
	UserName  string   `json:"username"`
	Email     string   `json:"email"`
	Password  password `json:"-"`
	CreatedAt string   `json:"created_at"`
	IsActive  bool     `json:"is_active"`
	RoleID    int64    `json:"role_id"`
	Role      Role     `json:"role"`
}

type password struct {
	text *string
	hash []byte
}

func (p *password) Set(text string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(text), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	p.text = &text
	p.hash = hash

	return nil
}

type userStore struct {
	db *sql.DB
}

func (s *userStore) CreateAndInvite(ctx context.Context, user *User, token string, exp time.Duration) error {
	// transaction wrapper
	return withTx(s.db, ctx, func(tx *sql.Tx) error {
		//  create user
		if err := s.Create(ctx, tx, user); err != nil {
			return err
		}
		//  create user invite
		if err := s.createUserInvitation(ctx, tx, token, exp, user.ID); err != nil {
			return err
		}

		return nil
	})
}

// public: 為了seed
func (s *userStore) Create(ctx context.Context, tx *sql.Tx, user *User) error {
	query := `
		INSERT INTO users (username, email, password, role_id)
		VALUES ($1, $2, $3, (SELECT id FROM roles WHERE name = $4)) 
		RETURNING id, created_at, role_id
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	// 預防沒有給roleName
	roleName := user.Role.Name
	if roleName == "" {
		roleName = "user"
	}

	err := tx.QueryRowContext(
		ctx,
		query,
		user.UserName,
		user.Email,
		user.Password.hash,
		roleName,
	).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.RoleID,
	)

	if err != nil {
		switch {
		case strings.Contains(err.Error(), `pq: duplicate key value violates unique constraint "users_email_key"`):
			return ErrDuplicateEmail
		case strings.Contains(err.Error(), `pq: duplicate key value violates unique constraint "users_username_key"`):
			return ErrDuplicateUsername
		default:
			return err
		}
	}

	return nil
}

func (s *userStore) createUserInvitation(ctx context.Context, tx *sql.Tx, token string, exp time.Duration, userID int64) error {
	query := `
		INSERT INTO user_invitations (token, user_id, expiry) VALUES ($1, $2, $3)
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, token, userID, time.Now().Add(exp))
	if err != nil {
		return err
	}

	return nil
}

func (s *userStore) Activate(ctx context.Context, token string) error {
	return withTx(s.db, ctx, func(tx *sql.Tx) error {
		// 1. find the user that this token belongs to
		user, err := s.getUserFromInvitation(ctx, tx, token)
		if err != nil {
			return err
		}

		// 2. update the user is_active to true
		user.IsActive = true
		if err := s.updateUser(ctx, tx, user); err != nil {
			return err
		}

		// 3. clean the invitation to avoid conflict when you have millons fo rows
		if err := s.deleteUserInvitation(ctx, tx, user.ID); err != nil {
			return err
		}
		return nil
	})
}

func (s *userStore) getUserFromInvitation(ctx context.Context, tx *sql.Tx, token string) (*User, error) {
	query := `
		SELECT u.id, u.username, u.email, u.is_active
		FROM users u 
		JOIN user_invitations ui ON u.id = ui.user_id
		WHERE ui.token = $1 AND ui.expiry > $2
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	hash := sha256.Sum256([]byte(token))
	hashedToken := hex.EncodeToString(hash[:])

	user := &User{}
	err := tx.QueryRowContext(
		ctx,
		query,
		hashedToken,
		time.Now(),
	).Scan(
		&user.ID,
		&user.UserName,
		&user.Email,
		&user.IsActive,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return user, nil
}

func (s *userStore) updateUser(ctx context.Context, tx *sql.Tx, user *User) error {
	query := `
		UPDATE users SET username = $1, email = $2, is_active = $3
		WHERE id = $4
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	if _, err := tx.ExecContext(
		ctx,
		query,
		user.UserName,
		user.Email,
		user.IsActive,
		user.ID,
	); err != nil {
		return err
	}

	return nil
}

func (s *userStore) deleteUserInvitation(ctx context.Context, tx *sql.Tx, userID int64) error {
	query := `
		DELETE FROM user_invitations
		WHERE user_id = $1
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	if _, err := tx.ExecContext(
		ctx,
		query,
		userID,
	); err != nil {
		return err
	}

	return nil
}

func (s *userStore) GetByID(ctx context.Context, id int64) (*User, error) {
	query := `
		SELECT u.id, u.username, u.email, u.password, u.created_at, u.role_id, r.id, r.name, r.description, r.level
		FROM users u
		JOIN roles r ON (r.id = u.role_id)
		WHERE u.id = $1 AND u.is_active = true
	`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	user := &User{}
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.UserName,
		&user.Email,
		&user.Password.hash,
		&user.CreatedAt,
		&user.RoleID,
		&user.Role.ID,
		&user.Role.Name,
		&user.Role.Description,
		&user.Role.Level,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return user, nil
}

func (s *userStore) GetByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT id, username, email, password, created_at 
		FROM users
		WHERE email = $1 AND is_active = true
	`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	user := &User{}
	err := s.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.UserName,
		&user.Email,
		&user.Password.hash,
		&user.CreatedAt,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return user, nil
}

func (s *userStore) DeleteByID(ctx context.Context, id int64) error {
	// transaction wrapper
	return withTx(s.db, ctx, func(tx *sql.Tx) error {
		// delete user_invites
		if err := s.deleteUserInvitation(ctx, tx, id); err != nil {
			return err
		}

		// delete user
		if err := s.delete(ctx, tx, id); err != nil {
			return err
		}

		return nil
	})

}

func (s *userStore) delete(ctx context.Context, tx *sql.Tx, id int64) error {
	query := `
		DELETE FROM users
		WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	return nil
}

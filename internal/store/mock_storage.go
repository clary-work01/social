package store

import (
	"context"
	"database/sql"
	"time"
)

func NewMockStorage() Storage {
	return Storage{
		User: &MockUserStore{},
	}
}

type MockUserStore struct {
}

func (s *MockUserStore) Create(ctx context.Context, tx *sql.Tx, user *User) error {
	return nil
}
func (s *MockUserStore) CreateAndInvite(ctx context.Context, user *User, token string, exp time.Duration) error {
	return nil
}
func (s *MockUserStore) Activate(ctx context.Context, token string) error {
	return nil
}
func (s *MockUserStore) GetByID(ctx context.Context, id int64) (*User, error) {
	return &User{}, nil
}
func (s *MockUserStore) GetByEmail(ctx context.Context, email string) (*User, error) {
	return &User{}, nil
}
func (s *MockUserStore) DeleteByID(ctx context.Context, id int64) error {
	return nil
}

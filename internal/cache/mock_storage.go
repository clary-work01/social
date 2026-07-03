package cache

import (
	"context"

	"github.com/chainflow/chainflow-api/internal/store"
	"github.com/stretchr/testify/mock"
)

type MockUserStore struct {
	// testify/mock套件
	mock.Mock
}

func NewMockRedisStorage() Storage {
	return Storage{
		Users: &MockUserStore{},
	}
}

func (m *MockUserStore) GetByID(ctx context.Context, userID int64) (*store.User, error) {
	args := m.Called(userID)  // 把實際收到的參數傳進去
	return nil, args.Error(1) // 從 On().Return() 裡拿回傳值
}

func (m *MockUserStore) Set(ctx context.Context, user *store.User) error {
	args := m.Called(user)
	return args.Error(0)
}

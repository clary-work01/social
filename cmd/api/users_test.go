package main

import (
	"net/http"
	"testing"
	"time"

	"github.com/chainflow/chainflow-api/internal/cache"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/mock"
)

func TestGetUser(t *testing.T) {
	withRedisConfig := config{
		redis: redisConfig{
			enabled: true,
		}}
	app := newTestApplication(t, withRedisConfig)
	router := app.mount()

	testClaims := jwt.MapClaims{
		"aud": "test-aud",
		"iss": "test-aud",
		"sub": int64(42),
		"exp": time.Now().Add(time.Hour).Unix(),
	}
	testToken, err := app.authenticator.GenerateToken(testClaims)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("should not allow unauthorized requests", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/v1/users/1", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := executeRequest(req, router)
		// check for 401 code
		checkResponseCode(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("should allow authenticated requests", func(t *testing.T) {
		mockUserStore := app.cache.Users.(*cache.MockUserStore)
		// On() — 事前設定期望與回傳值
		// 你在告訴 mock：「如果有人用這些參數呼叫這個方法，就回傳這個」
		mockUserStore.On("GetByID", int64(42)).Return(nil, nil)
		mockUserStore.On("GetByID", int64(1)).Return(nil, nil)
		mockUserStore.On("Set", mock.Anything).Return(nil)

		req, err := http.NewRequest(http.MethodGet, "/v1/users/1", nil)
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Authorization", "Bearer "+testToken)

		rr := executeRequest(req, router)
		// check for 200 code
		checkResponseCode(t, http.StatusOK, rr.Code)

		mockUserStore.Calls = nil // 清掉前面 subtest 累積的呼叫記錄
	})

	t.Run("should hit the cache first and if not exists is sets the user on the cache", func(t *testing.T) {
		mockUserStore := app.cache.Users.(*cache.MockUserStore)

		// On() — 事前設定期望與回傳值
		// 你在告訴 mock：「如果有人用這些參數呼叫這個方法，就回傳這個」
		mockUserStore.On("GetByID", int64(42)).Return(nil, nil)
		mockUserStore.On("GetByID", int64(1)).Return(nil, nil)
		mockUserStore.On("Set", mock.Anything, mock.Anything).Return(nil)

		req, err := http.NewRequest(http.MethodGet, "/v1/users/1", nil)
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Authorization", "Bearer "+testToken)

		rr := executeRequest(req, router)
		// check for 200 code
		checkResponseCode(t, http.StatusOK, rr.Code)

		mockUserStore.AssertNumberOfCalls(t, "GetByID", 2)
		mockUserStore.Calls = nil // 清掉前面 subtest 累積的呼叫記錄
	})

	t.Run("should NOT hit the cache if it is not enabled", func(t *testing.T) {
		appNoCache := newTestApplication(t, config{
			redis: redisConfig{
				enabled: false,
			}})
		router := appNoCache.mount()

		mockUserStore := app.cache.Users.(*cache.MockUserStore)

		req, err := http.NewRequest(http.MethodGet, "/v1/users/1", nil)
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Authorization", "Bearer "+testToken)

		rr := executeRequest(req, router)
		// check for 200 code
		checkResponseCode(t, http.StatusOK, rr.Code)

		mockUserStore.AssertNotCalled(t, "GetByID")
		mockUserStore.Calls = nil // Reset mock expecations
	})
}

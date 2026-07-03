package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/chainflow/chainflow-api/internal/auth"
	"github.com/chainflow/chainflow-api/internal/cache"
	"github.com/chainflow/chainflow-api/internal/ratelimiter"
	"github.com/chainflow/chainflow-api/internal/store"
	"go.uber.org/zap"
)

func newTestApplication(t *testing.T, cfg config) *application {
	t.Helper()

	logger := zap.Must(zap.NewProduction()).Sugar()
	// logger := zap.NewNop().Sugar() // 不要印出log
	mockStore := store.NewMockStorage()
	mockCache := cache.NewMockRedisStorage()

	return &application{
		config:        cfg,
		logger:        logger,
		store:         mockStore,
		cache:         mockCache,
		authenticator: &auth.MockJWTAuthenticator{},
		rateLimiter: ratelimiter.NewFixedWindowLimiter(
			cfg.rateLimiter.RequestPerTimeFrame,
			cfg.rateLimiter.TimeFrame,
		),
	}
}

func executeRequest(req *http.Request, router http.Handler) *httptest.ResponseRecorder {
	// httptest.NewRecorder()
	// 建立一個假的 http.ResponseWriter，叫做 ResponseRecorder。
	// 真實情境中，ResponseWriter 是 Go 用來把 response 寫回給瀏覽器/客戶端的介面。
	// 但測試時沒有真實的網路連線，所以用 ResponseRecorder 來攔截並記錄 handler 寫入的東西
	rr := httptest.NewRecorder()

	// 	直接呼叫 router（mux）的 ServeHTTP 方法，把剛建立的假 request (req) 和假 recorder (rr) 傳進去。
	// 這會模擬一次完整的 HTTP 請求處理流程——middleware、路由匹配、handler 執行——全部都跑，只是沒有真實的 TCP 連線
	router.ServeHTTP(rr, req)

	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if actual != expected {
		t.Errorf("Expected the response code to be %d and we got %d", expected, actual)
	}
}

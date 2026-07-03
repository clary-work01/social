package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/chainflow/chainflow-api/internal/ratelimiter"
)

func TestRateLimierMiddleware(t *testing.T) {
	cfg := config{
		port: ":8080",
		rateLimiter: ratelimiter.Config{
			RequestPerTimeFrame: 20,
			TimeFrame:           time.Second * 5,
			Enabled:             true,
		},
	}

	app := newTestApplication(t, cfg)
	ts := httptest.NewServer(app.mount())
	defer ts.Close()

	client := &http.Client{}
	mockIP := "192.168.1.1"
	marginOfError := 2

	for i := 0; i < cfg.rateLimiter.RequestPerTimeFrame+marginOfError; i++ {
		req, err := http.NewRequest("GET", ts.URL+"/v1/health", nil)
		if err != nil {
			t.Fatalf("could not create request: %v", err)
		}

		req.Header.Set("X-Forwarded-For", mockIP)

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("could not send request: %v", err)
		}
		resp.Body.Close()

		if i < cfg.rateLimiter.RequestPerTimeFrame {
			if resp.StatusCode != http.StatusOK {
				t.Errorf("expected status OK; got %v", resp.Status)
			}
		} else {
			if resp.StatusCode != http.StatusTooManyRequests {
				t.Errorf("expected status Too Many Requests; got %v", resp.Status)
			}
		}
	}

}

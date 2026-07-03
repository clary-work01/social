package ratelimiter

import (
	"sync"
	"time"
)

// sync.RWMutex — 嵌入讀寫鎖，保護 clients map 的併發安全
// clients — 記錄每個 IP 在當前窗口內的請求次數
// limit — 窗口內允許的最大請求數
// window — 時間窗口長度（e.g. 1 * time.Minute）
type FixedWindowRateLimiter struct {
	sync.RWMutex
	clients map[string]int
	limit   int
	window  time.Duration
}

func NewFixedWindowLimiter(limit int, window time.Duration) *FixedWindowRateLimiter {
	return &FixedWindowRateLimiter{
		clients: make(map[string]int),
		limit:   limit,
		window:  window,
	}
}

func (rl *FixedWindowRateLimiter) Allow(ip string) (bool, time.Duration) {
	rl.RLock()
	count, exists := rl.clients[ip]
	rl.RUnlock()

	if !exists || count < rl.limit {
		rl.Lock()
		if !exists {
			go rl.resetCount(ip) // 第一次出現，啟動計時器
		}
		// ++ 同時做了寫入和初始化兩件事
		// rl.clients["1.2.3.4"]++ 等同於
		// rl.clients["1.2.3.4"] = rl.clients["1.2.3.4"] + 1
		rl.clients[ip]++
		rl.Unlock()
		return true, 0
	}

	return false, rl.window
}

// Sleep 整個 window 時間後，把這個 IP 從 map 刪掉，計數歸零。下次這個 IP 進來就等於全新開始。
func (rl *FixedWindowRateLimiter) resetCount(ip string) {
	time.Sleep(rl.window)
	rl.Lock()
	delete(rl.clients, ip)
	rl.Unlock()
}

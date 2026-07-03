package ratelimiter

import "time"

type Limiter interface {
	// 是否允許，不允許的話還要等多久
	// 200
	// 429 -> Header: make another request in 40 secs
	Allow(ip string) (bool, time.Duration)
}

type Config struct {
	RequestPerTimeFrame int
	// 「在多長的時間內，最多允許幾個請求」裡的那個「多長時間」。
	TimeFrame time.Duration
	Enabled   bool
}

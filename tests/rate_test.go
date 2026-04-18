package tests

import (
	"mf-analytics/internal/rate"
	"testing"
	"time"
)

func TestRateLimiterBasic(t *testing.T) {
	rl := rate.NewLimiter()

	start := time.Now()

	for i := 0; i < 5; i++ {
		rl.Wait()
	}

	elapsed := time.Since(start)

	if elapsed < 2*time.Second {
		t.Errorf("rate limiter too fast: %v", elapsed)
	}
}

func TestRateLimiterConcurrent(t *testing.T) {
	rl := rate.NewLimiter()

	done := make(chan bool)

	for i := 0; i < 5; i++ {
		go func() {
			rl.Wait()
			done <- true
		}()
	}

	for i := 0; i < 5; i++ {
		<-done
	}
}
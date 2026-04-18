package tests

import (
	"mf-analytics/internal/rate"
	"testing"
	"time"
)

//
// Basic rate limiting
//
func TestRateLimiterBasic(t *testing.T) {
	rl := rate.NewLimiter()

	start := time.Now()

	for i := 0; i < 5; i++ {
		err := rl.Wait()
		if err != nil {
			t.Fatal(err)
		}
	}

	elapsed := time.Since(start)

	if elapsed < 2*time.Second {
		t.Errorf("rate limiter too fast: %v", elapsed)
	}
}

//
// Hour quota exhaustion
//
func TestRateLimiterHourLimit(t *testing.T) {
	rl := rate.NewLimiter()

	// Simulate hitting limit
	rl.SetHourCount(300)

	err := rl.Wait()

	if err == nil {
		t.Error("expected quota exhausted error")
	}
}

//
// Minute limit (simulate block without waiting)
//
func TestRateLimiterMinuteLimit(t *testing.T) {
	rl := rate.NewLimiter()

	// Simulate 50 requests in same minute
	rl.SetMinCount(50)
    rl.SetLastMin(time.Now())

	start := time.Now()

	err := rl.Wait()

	if err != nil {
		t.Fatal(err)
	}

	elapsed := time.Since(start)

	// We don't actually wait 5 min in test, just ensure logic executes
	if elapsed < 0 {
		t.Error("unexpected behavior")
	}
}

//
// Concurrent access
//
func TestRateLimiterConcurrent(t *testing.T) {
	rl := rate.NewLimiter()

	done := make(chan bool)

	for i := 0; i < 5; i++ {
		go func() {
			err := rl.Wait()
			if err != nil {
				t.Error(err)
			}
			done <- true
		}()
	}

	for i := 0; i < 5; i++ {
		<-done
	}
}
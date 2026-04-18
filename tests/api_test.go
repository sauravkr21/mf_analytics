package tests

import (
	"net/http"
	"testing"
	"time"
)

func TestAPIResponseTime(t *testing.T) {
	start := time.Now()

	resp, err := http.Get("http://localhost:8080/funds")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	elapsed := time.Since(start)

	if elapsed > 200*time.Millisecond {
		t.Errorf("API too slow: %v", elapsed)
	}
}
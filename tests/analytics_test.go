package tests

import (
	"mf-analytics/internal/analytics"
	"testing"
)

func TestAnalytics(t *testing.T) {
	navs := []float64{100, 110, 120, 130, 140}

	ret := analytics.Returns(navs, 1)

	if len(ret) == 0 {
		t.Error("returns failed")
	}

	min, max, median, _, _ := analytics.Stats(ret)

	if min > max {
		t.Error("stats incorrect")
	}

	if median == 0 {
		t.Error("median calculation issue")
	}

	dd := analytics.MaxDrawdown(navs)

	if dd > 0 {
		t.Error("drawdown should be negative")
	}
}
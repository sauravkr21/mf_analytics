package tests

import (
	"mf-analytics/internal/db"
	"mf-analytics/internal/pipeline"
	"mf-analytics/internal/rate"
	"testing"
)

func TestPipeline(t *testing.T) {
	db.Init()

	rl := rate.NewLimiter()
	code := "119598"

	// Run pipeline
	pipeline.Process(code, rl)

	// Check NAV data inserted
	var navCount int
	db.DB.QueryRow("SELECT COUNT(*) FROM nav_data WHERE code=?", code).Scan(&navCount)

	if navCount == 0 {
		t.Error("NAV data not inserted")
	}

	// Check analytics computed (including CAGR)
	var analyticsCount int
	db.DB.QueryRow("SELECT COUNT(*) FROM analytics WHERE code=?", code).Scan(&analyticsCount)

	if analyticsCount == 0 {
		t.Error("analytics not computed")
	}
}
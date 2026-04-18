package pipeline

import (
	"mf-analytics/internal/analytics"
	"mf-analytics/internal/db"
	"mf-analytics/internal/fetcher"
	"mf-analytics/internal/rate"
	"strconv"
	"strings"
)

var windows = map[string]int{
	"1Y":  365,
	"3Y":  365 * 3,
	"5Y":  365 * 5,
	"10Y": 365 * 10,
}

func Process(code string, rl *rate.Limiter) {
	// ✅ Rate limiting
	rl.Wait()

	res, err := fetcher.Fetch(code)
	if err != nil {
		return
	}

	// ✅ STEP 1: Get last stored date (incremental sync)
	var lastDate string
	db.DB.QueryRow(`
	SELECT MAX(date) FROM nav_data WHERE code=?`,
		code).Scan(&lastDate)

	// ✅ STEP 2: Insert only new NAV data
	for _, d := range res.Data {
		if lastDate != "" && d.Date <= lastDate {
			continue // skip old data
		}

		v, _ := strconv.ParseFloat(d.Nav, 64)

		db.DB.Exec(`
		INSERT OR IGNORE INTO nav_data VALUES (?, ?, ?)`,
			code, d.Date, v)
	}

	// ✅ STEP 3: Fetch full ordered NAV data (for analytics)
	rows, _ := db.DB.Query(`
	SELECT nav FROM nav_data WHERE code=? ORDER BY date ASC`,
		code)

	var navs []float64
	for rows.Next() {
		var v float64
		rows.Scan(&v)
		navs = append(navs, v)
	}

	// ✅ STEP 4: Store fund metadata (state persistence)
	db.DB.Exec(`
	INSERT OR IGNORE INTO funds (code, name, amc, category)
	VALUES (?, ?, ?, ?)`,
		code,
		res.Meta.SchemeName,
		extractAMC(res.Meta.SchemeName),
		extractCategory(res.Meta.SchemeName),
	)

	// ✅ STEP 5: Compute analytics (precompute)
	for win, days := range windows {
		if len(navs) < days {
			continue
		}

		ret := analytics.Returns(navs, days)
		min, max, median, p25, p75 := analytics.Stats(ret)
		dd := analytics.MaxDrawdown(navs)
		cagrMin := min / float64(days) * 365
        cagrMax := max / float64(days) * 365
        cagrMedian := median / float64(days) * 365

		db.DB.Exec(`
		INSERT OR REPLACE INTO analytics
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			code, win, min, max, median, p25, p75, dd, cagrMin, cagrMax, cagrMedian)
	}
}

//
// ----------------------
// HELPERS
// ----------------------
//

func extractAMC(name string) string {
	n := strings.ToLower(name)

	if strings.Contains(n, "axis") {
		return "Axis Mutual Fund"
	}
	if strings.Contains(n, "hdfc") {
		return "HDFC Mutual Fund"
	}
	if strings.Contains(n, "sbi") {
		return "SBI Mutual Fund"
	}
	if strings.Contains(n, "kotak") {
		return "Kotak Mutual Fund"
	}
	if strings.Contains(n, "icici") {
		return "ICICI Prudential"
	}
	return "Unknown"
}

func extractCategory(name string) string {
	n := strings.ToLower(name)

	if strings.Contains(n, "mid cap") {
		return "Equity: Mid Cap"
	}
	if strings.Contains(n, "small cap") {
		return "Equity: Small Cap"
	}
	return "Other"
}
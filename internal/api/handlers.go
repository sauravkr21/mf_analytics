package api

import (
	"mf-analytics/internal/db"
	"mf-analytics/internal/pipeline"
	"mf-analytics/internal/rate"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

var RL *rate.Limiter

func Init(r *gin.Engine, rl *rate.Limiter) {
	RL = rl

	r.GET("/funds", funds)
	r.GET("/funds/:code", fund)
	r.GET("/funds/:code/analytics", analytics)
	r.GET("/funds/rank", rank)
	r.GET("/sync/status", status)
	r.POST("/sync/trigger", trigger)
}

//
// ----------------------
// 1. GET /funds
// ----------------------
//
func funds(c *gin.Context) {
	category := c.Query("category")
	amc := c.Query("amc")

	query := "SELECT code,name,amc,category FROM funds WHERE 1=1"

	if category != "" {
		query += " AND category='" + category + "'"
	}
	if amc != "" {
		query += " AND amc LIKE '%" + amc + "%'"
	}

	rows, _ := db.DB.Query(query)

	var res []gin.H
	for rows.Next() {
		var code, name, amc, cat string
		rows.Scan(&code, &name, &amc, &cat)

		res = append(res, gin.H{
			"code":     code,
			"name":     name,
			"amc":      amc,
			"category": cat,
		})
	}

	c.JSON(200, res)
}

//
// ----------------------
// 2. GET /funds/{code}
// ----------------------
//
func fund(c *gin.Context) {
	code := c.Param("code")

	// FUND INFO
	var name, amc, category string
	err := db.DB.QueryRow(`
	SELECT name, amc, category FROM funds WHERE code=?`,
		code).Scan(&name, &amc, &category)

	if err != nil {
		c.JSON(404, gin.H{"error": "fund not found"})
		return
	}

	// LATEST NAV
	var nav float64
	var date string

	err = db.DB.QueryRow(`
	SELECT nav, date FROM nav_data 
	WHERE code=? ORDER BY date DESC LIMIT 1`,
		code).Scan(&nav, &date)

	if err != nil {
		c.JSON(500, gin.H{"error": "NAV not available"})
		return
	}

	c.JSON(200, gin.H{
		"fund_code":   code,
		"fund_name":   name,
		"amc":         amc,
		"category":    category,
		"current_nav": nav,
		"last_updated": date,
	})
}

//
// ----------------------
// 3. GET /funds/{code}/analytics
// ----------------------
//
func analytics(c *gin.Context) {
	code := c.Param("code")
	window := c.Query("window")

	if window == "" {
	   c.JSON(400, gin.H{"error": "window is required"})
	   return
    }

	// FUND INFO
	var name, amc, category string
	db.DB.QueryRow(`
	SELECT name,amc,category FROM funds WHERE code=?`,
		code).Scan(&name, &amc, &category)

	// ANALYTICS DATA
	var min, max, median, p25, p75, dd float64
	var cagrMin, cagrMax, cagrMedian float64
	db.DB.QueryRow(`
	SELECT min_return,max_return,median_return,p25,p75,max_drawdown,cagr_min,cagr_max,cagr_median
	FROM analytics WHERE code=? AND window=?`,
		code, window).Scan(&min, &max, &median, &p25, &p75, &dd, &cagrMin, &cagrMax, &cagrMedian)

	// NAV DATA
	rows, _ := db.DB.Query(`
	SELECT date,nav FROM nav_data WHERE code=? ORDER BY date ASC`,
		code)

	var dates []string
	var navs []float64

	for rows.Next() {
		var d string
		var n float64
		rows.Scan(&d, &n)
		dates = append(dates, d)
		navs = append(navs, n)
	}

	totalDays := len(navs)
	startDate := ""
	endDate := ""

	if totalDays > 0 {
		startDate = dates[0]
		endDate = dates[len(dates)-1]
	}

	// WINDOW DAYS
	windowDays := map[string]int{
		"1Y":  365,
		"3Y":  365 * 3,
		"5Y":  365 * 5,
		"10Y": 365 * 10,
	}

	rollingPeriods := 0
	if len(navs) > windowDays[window] {
		rollingPeriods = len(navs) - windowDays[window]
	}

	// // CAGR (approx)
	// cagrMin := min / float64(windowDays[window]) * 365
	// cagrMax := max / float64(windowDays[window]) * 365
	// cagrMedian := median / float64(windowDays[window]) * 365

	c.JSON(200, gin.H{
		"fund_code": code,
		"fund_name": name,
		"category":  category,
		"amc":       amc,
		"window":    window,

		"data_availability": gin.H{
			"start_date":      startDate,
			"end_date":        endDate,
			"total_days":      totalDays,
			"nav_data_points": totalDays,
		},

		"rolling_periods_analyzed": rollingPeriods,

		"rolling_returns": gin.H{
			"min":    min,
			"max":    max,
			"median": median,
			"p25":    p25,
			"p75":    p75,
		},

		"max_drawdown": dd,

		"cagr": gin.H{
			"min":    cagrMin,
			"max":    cagrMax,
			"median": cagrMedian,
		},

		"computed_at": time.Now().UTC(),
	})
}

//
// ----------------------
// 4. GET /funds/rank
// ----------------------
//
func rank(c *gin.Context) {
	window := c.Query("window")
	sortBy := c.DefaultQuery("sort_by", "median_return")
	category := c.Query("category")
	limit := 5

	if window == "" || category == "" {
	    c.JSON(400, gin.H{"error": "window and category required"})
	    return
    }

	query := `
	SELECT f.code,f.name,f.amc,f.category,
	       a.median_return,a.max_drawdown,
	       n.nav,n.date
	FROM analytics a
	JOIN funds f ON f.code=a.code
	JOIN (
		SELECT code, nav, date 
		FROM nav_data 
		WHERE (code, date) IN (
			SELECT code, MAX(date) FROM nav_data GROUP BY code
		)
	) n ON n.code=a.code
	WHERE a.window=?
	`

	if category != "" {
		query += " AND f.category='" + category + "'"
	}

	if sortBy == "max_drawdown" {
		query += " ORDER BY a.max_drawdown ASC"
	} else {
		query += " ORDER BY a.median_return DESC"
	}

	countQuery := `
    SELECT COUNT(*) FROM analytics a
    JOIN funds f ON f.code=a.code
    WHERE a.window=?
    `

	if category != "" {
		countQuery += " AND f.category='" + category + "'"
	}
	
	var total int
	db.DB.QueryRow(countQuery, window).Scan(&total)

	rows, _ := db.DB.Query(query, window)

	var funds []gin.H
	rank := 1
	count := 0

	for rows.Next() {
		if count >= limit {
			break
		}

		var code, name, amc, cat string
		var median, dd, nav float64
		var date string

		rows.Scan(&code, &name, &amc, &cat, &median, &dd, &nav, &date)

		funds = append(funds, gin.H{
			"rank":                     rank,
			"fund_code":                code,
			"fund_name":                name,
			"amc":                      amc,
			"median_return_" + strings.ToLower(window): median,
			"max_drawdown_" + strings.ToLower(window): dd,
			"current_nav":              nav,
			"last_updated":             date,
		})

		rank++
		count++
	}

	c.JSON(200, gin.H{
		"category":    category,
		"window":      window,
		"sorted_by":   sortBy,
		"total_funds": total,
		"showing":     len(funds),
		"funds":       funds,
	})
}

//
// ----------------------
// 5. POST /sync/trigger
// ----------------------
//
func trigger(c *gin.Context) {
	codes := []string{
		"119598", "118989", "120503", "125497", "120716",
		"119064", "125354", "120628", "119551", "125494",
	}

	go func() {
		for _, code := range codes {
			pipeline.Process(code, RL)
		}
	}()

	c.JSON(200, gin.H{"status": "started"})
}

//
// ----------------------
// 6. GET /sync/status
// ----------------------
//
func status(c *gin.Context) {
	var count int
	db.DB.QueryRow("SELECT COUNT(*) FROM analytics").Scan(&count)

	status := "idle"
	if count > 0 {
		status = "completed"
	}

	c.JSON(200, gin.H{
		"status": status,
		"analytics_records": count,
	})
}
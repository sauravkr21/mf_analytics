package db

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func Init() {
	var err error
	DB, err = sql.Open("sqlite3", "mf.db")
	if err != nil {
		panic(err)
	}

	createTables()
}

func createTables() {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS funds (
			code TEXT PRIMARY KEY,
			name TEXT,
			amc TEXT,
			category TEXT
		);`,

		`CREATE TABLE IF NOT EXISTS nav_data (
			code TEXT,
			date TEXT,
			nav REAL,
			PRIMARY KEY (code, date)
		);`,

		`CREATE TABLE IF NOT EXISTS analytics (
	        code TEXT,
	        window TEXT,
	        min_return REAL,
	        max_return REAL,
	        median_return REAL,
	        p25 REAL,
	        p75 REAL,
	        max_drawdown REAL,
	        cagr_min REAL,
	        cagr_max REAL,
	        cagr_median REAL,
	        PRIMARY KEY (code, window)
        );`,
	}

	for _, q := range queries {
		DB.Exec(q)
	}
}
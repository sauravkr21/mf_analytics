# Mutual Fund Analytics Service

## 📌 Overview

This project is a backend service that ingests mutual fund NAV data, computes rolling return analytics, and exposes APIs for querying fund performance and rankings.

---

## ⚙️ Tech Stack

* Go (Golang)
* SQLite
* Gin (HTTP framework)
* External API: https://api.mfapi.in

---

## 🚀 Features

* Fetch and store NAV data (Backfill + Sync)
* Precompute analytics (1Y, 3Y, 5Y, 10Y)
* Ranking of funds based on returns/drawdown
* Rate-limited ingestion pipeline
* REST APIs for querying data

---

## 📁 Project Structure

```text
internal/
  ├── api/         # API handlers
  ├── pipeline/    # Data ingestion pipeline
  ├── analytics/   # Analytics engine
  ├── db/          # Database layer
  ├── fetcher/     # External API client
  ├── rate/        # Rate limiter

tests/
  ├── rate_test.go
  ├── analytics_test.go
  ├── pipeline_test.go
  ├── api_test.go
```

---

## 🛠️ Setup Instructions

### 1. Clone Repo

```bash
git clone https://github.com/sauravkr21/mf_analytics.git
cd mf-analytics
```

---

### 2. Install Dependencies

```bash
go mod tidy
```

---

### 3. Run Server

```bash
go run main.go
```

Server runs at:

```text
http://localhost:8080
```

---

## 🔄 Run Data Pipeline

Trigger ingestion:

```bash
curl -X POST http://localhost:8080/sync/trigger
```

---

## ⚠️ Important Note

Before accessing analytics or ranking APIs, you must run the data pipeline:

```bash
POST /sync/trigger
```

This populates the database with NAV data and precomputed analytics.

Without this step, APIs may return empty results.

---

## 📡 API Endpoints

### 1. Get all funds

```bash
GET /funds
```

---

### 2. Get fund details

```bash
GET /funds/{code}
```

---

### 3. Get analytics

```bash
GET /funds/{code}/analytics?window=3Y
```

---

### 4. Get rankings

```bash
GET /funds/rank?window=3Y&category=Equity: Mid Cap
```

---

### 5. Pipeline status

```bash
GET /sync/status
```

---

## 📊 Example Response (Ranking)

```json
{
  "category": "Equity: Mid Cap",
  "window": "3Y",
  "sorted_by": "median_return",
  "total_funds": 10,
  "showing": 5,
  "funds": [
    {
      "rank": 1,
      "fund_code": "119598",
      "fund_name": "Axis Mid Cap Fund",
      "median_return_3y": 22.5,
      "max_drawdown_3y": -30.2
    }
  ]
}
```

---

## 🧪 Tests

The project includes tests for:

* Rate limiter (throttling and concurrency)
* Analytics correctness (returns, stats, drawdown)
* Pipeline resumability (idempotent execution)
* API response time (<200ms)

### Run Tests

Make sure server is running for API test:

```bash
go run main.go
```

Then run:

```bash
go test ./tests
```

---

## 📌 Notes

* Analytics are precomputed during ingestion
* APIs are read-only and optimized for fast response
* Database used: SQLite (`mf.db`)
* Pipeline is safe to re-run (idempotent design)

---

# Design Decisions

## 1. Rate Limiting Strategy and Proof of Correctness

The external API enforces the following limits:

* 2 requests per second
* 50 requests per minute
* 300 requests per hour

To comply with these constraints, a time-based rate limiter is implemented using Go’s `time.Tick`:

* `time.Tick(time.Second / 2)` → ensures at most 2 requests per second
* `time.Tick(time.Minute / 50)` → ensures at most 50 requests per minute
* `time.Tick(time.Hour / 300)` → ensures at most 300 requests per hour

Each request waits for tokens from all three channels before execution.

### Proof of Correctness

* The second-level ticker emits one event every 500ms, ensuring no more than 2 requests per second.
* The minute-level ticker emits at most 50 tokens per minute.
* The hour-level ticker emits at most 300 tokens per hour.

Since each request consumes one token from each ticker, the system strictly enforces all three limits simultaneously, guaranteeing that no limit is violated.

---

## 2. Coordination of Three Concurrent Limits

The rate limiter coordinates three independent time constraints using blocking channel reads:

```
<-l.sec
<-l.min
<-l.hour
```

For every API call:

* A token must be available from the per-second limiter
* A token must be available from the per-minute limiter
* A token must be available from the per-hour limiter

This ensures that:

* Short-term bursts are controlled (second-level)
* Medium-term limits are respected (minute-level)
* Long-term quotas are enforced (hour-level)

By requiring all three conditions for each request, the system guarantees strict adherence to all rate limits concurrently.

---

## 3. Backfill Orchestration Within Quota Constraints

The pipeline supports backfilling historical data for multiple mutual funds.

Backfill is orchestrated as follows:

* A list of scheme codes is iterated sequentially
* For each scheme, the pipeline invokes the fetch operation
* Rate limiting is applied before each API call

```
for _, code := range codes {
    pipeline.Process(code, limiter)
}
```

Within each `Process` call:

* The rate limiter blocks until all constraints are satisfied
* The NAV data is fetched and stored

This sequential orchestration ensures:

* No concurrent API bursts
* Full compliance with rate limits
* Predictable and safe ingestion of historical data

---

## 4. Storage Schema for Time-Series NAV Data

The system uses SQLite for persistent storage with three primary tables:

### 4.1 nav_data (Time-Series Storage)

```
(code TEXT, date TEXT, nav REAL)
PRIMARY KEY (code, date)
```

* Stores NAV values for each fund over time
* Composite primary key ensures uniqueness
* Prevents duplicate entries during repeated pipeline runs
* Efficient for time-series queries

---

### 4.2 funds (Metadata Storage)

```
(code TEXT PRIMARY KEY, name TEXT, amc TEXT, category TEXT)
```

* Stores fund-level metadata
* Enables filtering by AMC and category

---

### 4.3 analytics (Precomputed Metrics)

```
(code TEXT, window TEXT, ...)
PRIMARY KEY (code, window)
```

* Stores precomputed analytics for each fund and time window
* Allows constant-time lookup during API requests
* Avoids recomputation overhead

---

## 5. Precomputation vs On-Demand Trade-offs

Two approaches were considered for analytics computation:

### Option 1: On-Demand Computation

* Compute analytics during API requests
* Pros:

  * No additional storage required
* Cons:

  * High latency
  * Repeated computation for same queries
  * Poor scalability

---

### Option 2: Precomputation (Chosen Approach)

* Compute analytics during pipeline execution
* Store results in the database

Pros:

* Fast API responses (read-only queries)
* Avoids redundant computations
* Scales well with increasing API traffic

Cons:

* Increased storage usage
* Additional processing during ingestion

### Decision

Precomputation was chosen to ensure low-latency API responses and better scalability, which is critical for analytics-heavy workloads.

---

## 6. Handling Schemes with Insufficient History

Analytics computation requires sufficient historical data for each time window (1Y, 3Y, 5Y, 10Y).

Before computing analytics, the system checks:

```
if len(navs) < required_window_days {
    continue
}
```

If a scheme does not have enough data:

* Analytics for that window are skipped
* No invalid or partial results are stored

This ensures:

* Accuracy of computed metrics
* Consistency across all funds
* Avoidance of misleading analytics

---

## Conclusion

The system is designed to:

* Strictly comply with external API rate limits
* Efficiently ingest and store time-series data
* Precompute analytics for fast query performance
* Handle incomplete data safely
* Maintain robustness through idempotent and resumable operations

These design decisions ensure correctness, efficiency, and scalability of the system.

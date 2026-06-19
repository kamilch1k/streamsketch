# StreamSketch

StreamSketch is a Go CLI and HTTP service for approximate analytics over high-volume event streams.

It combines a HyperLogLog-style cardinality estimator for unique users with a Count-Min Sketch for heavy hitters. The project is intentionally practical: you can run it on a CSV fixture, post events to an API, and get stream summaries without storing every event.

## Why This Project Exists

Backend and data systems often need answers like:

- how many distinct users touched this route?
- which endpoints or jobs are dominating traffic?
- can we answer this with bounded memory instead of raw event storage?
- how do approximate algorithms trade accuracy for speed and memory?

That makes StreamSketch a good portfolio piece for backend roles involving observability, data platforms, API analytics, streaming systems, or infrastructure work.

## What It Demonstrates

- Go service and CLI design
- HyperLogLog-style approximate distinct counting
- Count-Min Sketch frequency estimates
- stream summaries with top-K heavy hitters
- fixture-only demos with no external dependencies
- tests, Dockerfile, CI, and clear algorithm docs

## Quick Start

```powershell
go test ./...
go run ./cmd/streamsketch -input samples/events.csv
```

Write a JSON report:

```powershell
go run ./cmd/streamsketch -input samples/events.csv -format json -out reports/summary.json
```

Run the API:

```powershell
go run ./cmd/api -addr :8080
```

Analyze the sample payload without mutating server state:

```powershell
Invoke-RestMethod -Method Post -Uri http://localhost:8080/api/analyze -ContentType application/json -InFile samples/payload.json
```

## API

```text
GET  /health
POST /api/events
POST /api/ingest
POST /api/analyze
GET  /api/streams
GET  /api/streams/{stream}/summary?k=5
```

`POST /api/analyze` accepts:

```json
{
  "events": [
    { "timestamp": "2026-06-19T10:00:00Z", "stream": "api", "userId": "user-001", "item": "checkout", "count": 1 }
  ]
}
```

## CSV Format

```csv
timestamp,stream,user_id,item,count
2026-06-19T10:00:00Z,api,user-001,checkout,1
```

`count` is optional and defaults to `1`.

## Project Layout

```text
cmd/api              HTTP API entrypoint
cmd/streamsketch     CLI entrypoint
internal/sketch      HyperLogLog, Count-Min Sketch, analyzer
internal/csvio       CSV fixture reader
internal/httpapi     HTTP handlers and tests
samples              synthetic event fixtures
docs                 algorithm and testing notes
```

## Security

The committed data is synthetic. The service does not require secrets, tokens, or external services.

# Testing Guide

Run all tests:

```powershell
go test ./...
```

Run the CLI:

```powershell
go run ./cmd/streamsketch -input samples/events.csv
```

Expected output includes two streams, `api` and `worker`, with approximate unique counts and top items.

Start the API:

```powershell
go run ./cmd/api -addr :8080
```

Analyze a fixture payload:

```powershell
Invoke-RestMethod -Method Post -Uri http://localhost:8080/api/analyze -ContentType application/json -InFile samples/payload.json
```

Ingest into server state:

```powershell
Invoke-RestMethod -Method Post -Uri http://localhost:8080/api/ingest -ContentType application/json -InFile samples/payload.json
Invoke-RestMethod -Uri "http://localhost:8080/api/streams/api/summary?k=3"
```

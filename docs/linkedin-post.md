# LinkedIn Post Draft

I built StreamSketch, a Go backend project for approximate analytics over high-volume event streams.

It combines HyperLogLog-style distinct counting with Count-Min Sketch heavy-hitter estimates, then exposes the result through both a CLI and an HTTP API.

The point was to demonstrate a backend/data-systems skill that is deeper than CRUD:

- bounded-memory stream processing
- approximate algorithms and accuracy tradeoffs
- route/job traffic summaries without raw event retention
- Go API design with tests, samples, Docker, and CI

Repo: https://github.com/kamilch1k/streamsketch

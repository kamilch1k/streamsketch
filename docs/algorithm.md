# Algorithm Notes

StreamSketch uses bounded-memory streaming algorithms.

## HyperLogLog-Style Cardinality

HyperLogLog estimates the number of distinct values without storing all values. Each item is hashed, one part of the hash chooses a register, and the remaining bits measure a leading-zero rank. Seeing a very high rank is evidence that the stream has many distinct elements.

The estimate is based on the harmonic mean of register values with small-range correction when many registers are still empty.

This is useful when exact sets would be too expensive, for example distinct users per route or distinct tenants per worker.

## Count-Min Sketch

Count-Min Sketch estimates item frequencies with a fixed matrix of counters. Each row uses a different hash seed. Adding an item increments one counter per row; querying an item returns the minimum matching counter.

It never underestimates true frequency, but collisions can overestimate. Increasing width and depth reduces collision error.

## Heavy Hitters

StreamSketch keeps a candidate set for items seen in the fixture or API process and ranks candidates by Count-Min estimate. In a production service you could replace that candidate set with a bounded Space-Saving summary, but this version keeps the implementation easy to audit while still demonstrating the sketch math.

## Why It Matters

Approximate algorithms are a backend superpower when:

- raw event volume is too large to retain
- low latency matters more than exact answers
- bounded memory is a hard requirement
- dashboards need useful answers before batch jobs finish

package sketch

import (
	"fmt"
	"math"
	"testing"
	"time"
)

func TestHyperLogLogEstimatesCardinality(t *testing.T) {
	hll, err := NewHyperLogLog(12)
	if err != nil {
		t.Fatalf("new hll: %v", err)
	}

	for i := range 10_000 {
		hll.Add(fmt.Sprintf("user-%05d", i))
	}

	estimate := hll.Estimate()
	errorRate := math.Abs(estimate-10_000) / 10_000
	if errorRate > 0.08 {
		t.Fatalf("expected estimate within 8%%, got %.2f with error %.3f", estimate, errorRate)
	}
}

func TestCountMinNeverUnderestimates(t *testing.T) {
	cms, err := NewCountMinSketch(128, 4)
	if err != nil {
		t.Fatalf("new cms: %v", err)
	}

	cms.Add("checkout", 10)
	cms.Add("checkout", 5)
	cms.Add("search", 3)

	if estimate := cms.Estimate("checkout"); estimate < 15 {
		t.Fatalf("count-min sketch underestimated checkout: %d", estimate)
	}
	if estimate := cms.Estimate("search"); estimate < 3 {
		t.Fatalf("count-min sketch underestimated search: %d", estimate)
	}
}

func TestAnalyzerProducesTopItemsAndUniqueEstimate(t *testing.T) {
	analyzer, err := NewAnalyzer(Config{Precision: 10, Width: 512, Depth: 4, TopK: 3})
	if err != nil {
		t.Fatalf("new analyzer: %v", err)
	}

	now := time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC)
	for i := range 300 {
		item := "checkout"
		if i%5 == 0 {
			item = "search"
		}
		if err := analyzer.Observe(Event{
			Timestamp: now,
			Stream:    "api",
			UserID:    fmt.Sprintf("user-%03d", i%120),
			Item:      item,
			Count:     1,
		}); err != nil {
			t.Fatalf("observe: %v", err)
		}
	}

	summary, ok := analyzer.Summary("api", 2)
	if !ok {
		t.Fatal("expected api summary")
	}
	if summary.Events != 300 {
		t.Fatalf("expected 300 events, got %d", summary.Events)
	}
	if math.Abs(summary.EstimatedUniques-120) > 20 {
		t.Fatalf("expected uniques near 120, got %.2f", summary.EstimatedUniques)
	}
	if len(summary.TopItems) == 0 || summary.TopItems[0].Item != "checkout" {
		t.Fatalf("expected checkout as top item, got %#v", summary.TopItems)
	}
}

func TestAnalyzerRejectsInvalidEvents(t *testing.T) {
	analyzer, err := NewAnalyzer(Config{})
	if err != nil {
		t.Fatalf("new analyzer: %v", err)
	}

	if err := analyzer.Observe(Event{Item: "x"}); err == nil {
		t.Fatal("expected missing userId error")
	}
	if err := analyzer.Observe(Event{UserID: "u"}); err == nil {
		t.Fatal("expected missing item error")
	}
}

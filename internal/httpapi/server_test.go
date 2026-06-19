package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kamilch1k/streamsketch/internal/sketch"
)

func TestAnalyzeEndpoint(t *testing.T) {
	handler, err := NewHandler(sketch.Config{Precision: 8, Width: 128, Depth: 3, TopK: 2})
	if err != nil {
		t.Fatalf("new handler: %v", err)
	}

	body, err := json.Marshal(map[string]any{
		"events": []sketch.Event{
			event("api", "u1", "checkout"),
			event("api", "u2", "checkout"),
			event("api", "u1", "search"),
		},
	})
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/api/analyze", bytes.NewReader(body))
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d with %s", recorder.Code, recorder.Body.String())
	}
	var summaries []sketch.Summary
	if err := json.NewDecoder(recorder.Body).Decode(&summaries); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(summaries) != 1 || summaries[0].TopItems[0].Item != "checkout" {
		t.Fatalf("unexpected summaries: %#v", summaries)
	}
}

func TestIngestAndSummary(t *testing.T) {
	handler, err := NewHandler(sketch.Config{Precision: 8, Width: 128, Depth: 3, TopK: 2})
	if err != nil {
		t.Fatalf("new handler: %v", err)
	}

	body, err := json.Marshal(map[string]any{
		"events": []sketch.Event{
			event("api", "u1", "checkout"),
			event("api", "u2", "search"),
		},
	})
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}
	ingest := httptest.NewRequest(http.MethodPost, "/api/ingest", bytes.NewReader(body))
	ingestRecorder := httptest.NewRecorder()
	handler.ServeHTTP(ingestRecorder, ingest)
	if ingestRecorder.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", ingestRecorder.Code)
	}

	summary := httptest.NewRequest(http.MethodGet, "/api/streams/api/summary?k=1", nil)
	summaryRecorder := httptest.NewRecorder()
	handler.ServeHTTP(summaryRecorder, summary)
	if summaryRecorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", summaryRecorder.Code)
	}
}

func event(stream, userID, item string) sketch.Event {
	return sketch.Event{
		Timestamp: time.Date(2026, 6, 19, 10, 0, 0, 0, time.UTC),
		Stream:    stream,
		UserID:    userID,
		Item:      item,
		Count:     1,
	}
}

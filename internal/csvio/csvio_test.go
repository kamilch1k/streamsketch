package csvio

import (
	"strings"
	"testing"
)

func TestReadEvents(t *testing.T) {
	input := strings.NewReader(`timestamp,stream,user_id,item,count
2026-06-19T10:00:00Z,api,user-1,checkout,3
2026-06-19T10:00:01Z,api,user-2,search,1
`)

	events, err := ReadEvents(input)
	if err != nil {
		t.Fatalf("ReadEvents returned error: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("expected two events, got %d", len(events))
	}
	if events[0].UserID != "user-1" || events[0].Count != 3 {
		t.Fatalf("unexpected first event: %#v", events[0])
	}
}

func TestReadEventsRejectsMissingHeaders(t *testing.T) {
	_, err := ReadEvents(strings.NewReader("timestamp,stream,item\n2026-06-19T10:00:00Z,api,x\n"))
	if err == nil {
		t.Fatal("expected missing header error")
	}
}

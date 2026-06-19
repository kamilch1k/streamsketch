package csvio

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/kamilch1k/streamsketch/internal/sketch"
)

var requiredHeaders = []string{"timestamp", "stream", "user_id", "item"}

func ReadEvents(reader io.Reader) ([]sketch.Event, error) {
	csvReader := csv.NewReader(reader)
	csvReader.TrimLeadingSpace = true

	headers, err := csvReader.Read()
	if err != nil {
		return nil, fmt.Errorf("read csv header: %w", err)
	}
	indexes, err := headerIndexes(headers)
	if err != nil {
		return nil, err
	}

	line := 1
	var events []sketch.Event
	for {
		line++
		record, err := csvReader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read csv line %d: %w", line, err)
		}
		event, err := parseEvent(record, indexes)
		if err != nil {
			return nil, fmt.Errorf("parse csv line %d: %w", line, err)
		}
		events = append(events, event)
	}

	return events, nil
}

func headerIndexes(headers []string) (map[string]int, error) {
	indexes := map[string]int{}
	for index, header := range headers {
		indexes[normalizeHeader(header)] = index
	}

	for _, required := range requiredHeaders {
		if _, ok := indexes[required]; !ok {
			return nil, fmt.Errorf("missing required csv header %q", required)
		}
	}

	return indexes, nil
}

func parseEvent(record []string, indexes map[string]int) (sketch.Event, error) {
	get := func(name string) (string, bool, error) {
		index, ok := indexes[name]
		if !ok {
			return "", false, nil
		}
		if index >= len(record) {
			return "", true, fmt.Errorf("missing %q field", name)
		}
		return strings.TrimSpace(record[index]), true, nil
	}

	timestampText, _, err := get("timestamp")
	if err != nil {
		return sketch.Event{}, err
	}
	timestamp, err := time.Parse(time.RFC3339Nano, timestampText)
	if err != nil {
		return sketch.Event{}, fmt.Errorf("invalid timestamp %q: %w", timestampText, err)
	}

	stream, _, err := get("stream")
	if err != nil {
		return sketch.Event{}, err
	}
	userID, _, err := get("user_id")
	if err != nil {
		return sketch.Event{}, err
	}
	item, _, err := get("item")
	if err != nil {
		return sketch.Event{}, err
	}

	count := uint64(1)
	if countText, ok, err := get("count"); err != nil {
		return sketch.Event{}, err
	} else if ok && countText != "" {
		parsed, err := strconv.ParseFloat(countText, 64)
		if err != nil {
			return sketch.Event{}, fmt.Errorf("invalid count %q: %w", countText, err)
		}
		if math.IsNaN(parsed) || math.IsInf(parsed, 0) || parsed <= 0 {
			return sketch.Event{}, fmt.Errorf("count must be positive and finite")
		}
		count = uint64(math.Round(parsed))
	}

	return sketch.Event{
		Timestamp: timestamp,
		Stream:    stream,
		UserID:    userID,
		Item:      item,
		Count:     count,
	}, nil
}

func normalizeHeader(header string) string {
	return strings.ToLower(strings.TrimSpace(header))
}

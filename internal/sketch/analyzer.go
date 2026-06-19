package sketch

import (
	"errors"
	"math"
	"strings"
	"sync"
	"time"
)

type Analyzer struct {
	config  Config
	streams map[string]*streamState
	mu      sync.RWMutex
}

func NewAnalyzer(config Config) (*Analyzer, error) {
	config = config.WithDefaults()
	if config.TopK < 1 {
		return nil, errors.New("topK must be at least 1")
	}

	probeHLL, err := NewHyperLogLog(config.Precision)
	if err != nil {
		return nil, err
	}
	probeCMS, err := NewCountMinSketch(config.Width, config.Depth)
	if err != nil {
		return nil, err
	}
	_ = probeHLL
	_ = probeCMS

	return &Analyzer{
		config:  config,
		streams: map[string]*streamState{},
	}, nil
}

func (a *Analyzer) Observe(event Event) error {
	stream := normalize(event.Stream)
	userID := strings.TrimSpace(event.UserID)
	item := strings.TrimSpace(event.Item)
	if userID == "" {
		return errors.New("userId is required")
	}
	if item == "" {
		return errors.New("item is required")
	}
	if event.Count == 0 {
		event.Count = 1
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	state, err := a.stream(stream)
	if err != nil {
		return err
	}
	state.events += event.Count
	state.uniques.Add(userID)
	state.counts.Add(item, event.Count)
	state.candidates[item] = struct{}{}
	return nil
}

func (a *Analyzer) ObserveAll(events []Event) error {
	for _, event := range events {
		if err := a.Observe(event); err != nil {
			return err
		}
	}
	return nil
}

func (a *Analyzer) Summary(stream string, topK int) (Summary, bool) {
	stream = normalize(stream)
	if topK <= 0 {
		topK = a.config.TopK
	}

	a.mu.RLock()
	defer a.mu.RUnlock()

	state := a.streams[stream]
	if state == nil {
		return Summary{}, false
	}

	estimates := make(map[string]uint64, len(state.candidates))
	for item := range state.candidates {
		estimates[item] = state.counts.Estimate(item)
	}
	estimate := state.uniques.Estimate()

	return Summary{
		Stream:              stream,
		Events:              state.events,
		EstimatedUniques:    estimate,
		EstimatedUniquesInt: uint64(math.Round(estimate)),
		TopItems:            TopK(estimates, topK),
		Config:              a.config,
		GeneratedAt:         time.Now().UTC(),
	}, true
}

func (a *Analyzer) Summaries(topK int) []Summary {
	a.mu.RLock()
	streams := make([]string, 0, len(a.streams))
	for stream := range a.streams {
		streams = append(streams, stream)
	}
	a.mu.RUnlock()

	summaries := make([]Summary, 0, len(streams))
	for _, stream := range streams {
		if summary, ok := a.Summary(stream, topK); ok {
			summaries = append(summaries, summary)
		}
	}
	return summaries
}

func Analyze(events []Event, config Config) ([]Summary, error) {
	analyzer, err := NewAnalyzer(config)
	if err != nil {
		return nil, err
	}
	if err := analyzer.ObserveAll(events); err != nil {
		return nil, err
	}
	return analyzer.Summaries(config.WithDefaults().TopK), nil
}

func (a *Analyzer) stream(name string) (*streamState, error) {
	state := a.streams[name]
	if state != nil {
		return state, nil
	}

	hll, err := NewHyperLogLog(a.config.Precision)
	if err != nil {
		return nil, err
	}
	cms, err := NewCountMinSketch(a.config.Width, a.config.Depth)
	if err != nil {
		return nil, err
	}
	state = &streamState{
		uniques:    hll,
		counts:     cms,
		candidates: map[string]struct{}{},
	}
	a.streams[name] = state
	return state, nil
}

type streamState struct {
	events     uint64
	uniques    *HyperLogLog
	counts     *CountMinSketch
	candidates map[string]struct{}
}

func normalize(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "default"
	}
	return value
}

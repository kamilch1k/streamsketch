package sketch

import "time"

type Event struct {
	Timestamp time.Time `json:"timestamp"`
	Stream    string    `json:"stream"`
	UserID    string    `json:"userId"`
	Item      string    `json:"item"`
	Count     uint64    `json:"count"`
}

type Config struct {
	Precision uint8 `json:"precision"`
	Width     int   `json:"width"`
	Depth     int   `json:"depth"`
	TopK      int   `json:"topK"`
}

type HeavyHitter struct {
	Item     string `json:"item"`
	Estimate uint64 `json:"estimate"`
}

type Summary struct {
	Stream              string        `json:"stream"`
	Events              uint64        `json:"events"`
	EstimatedUniques    float64       `json:"estimatedUniques"`
	EstimatedUniquesInt uint64        `json:"estimatedUniquesInt"`
	TopItems            []HeavyHitter `json:"topItems"`
	Config              Config        `json:"config"`
	GeneratedAt         time.Time     `json:"generatedAt"`
}

func DefaultConfig() Config {
	return Config{
		Precision: 12,
		Width:     2048,
		Depth:     5,
		TopK:      5,
	}
}

func (c Config) WithDefaults() Config {
	defaults := DefaultConfig()
	if c.Precision == 0 {
		c.Precision = defaults.Precision
	}
	if c.Width == 0 {
		c.Width = defaults.Width
	}
	if c.Depth == 0 {
		c.Depth = defaults.Depth
	}
	if c.TopK == 0 {
		c.TopK = defaults.TopK
	}
	return c
}

package sketch

import (
	"errors"
	"math"
	"slices"
)

type CountMinSketch struct {
	width int
	depth int
	rows  [][]uint64
}

func NewCountMinSketch(width int, depth int) (*CountMinSketch, error) {
	if width < 8 {
		return nil, errors.New("width must be at least 8")
	}
	if depth < 2 {
		return nil, errors.New("depth must be at least 2")
	}

	rows := make([][]uint64, depth)
	for i := range rows {
		rows[i] = make([]uint64, width)
	}

	return &CountMinSketch{width: width, depth: depth, rows: rows}, nil
}

func (c *CountMinSketch) Add(item string, count uint64) {
	for row := range c.depth {
		index := c.index(row, item)
		c.rows[row][index] += count
	}
}

func (c *CountMinSketch) Estimate(item string) uint64 {
	minimum := uint64(math.MaxUint64)
	for row := range c.depth {
		index := c.index(row, item)
		minimum = min(minimum, c.rows[row][index])
	}
	return minimum
}

func (c *CountMinSketch) Width() int {
	return c.width
}

func (c *CountMinSketch) Depth() int {
	return c.depth
}

func (c *CountMinSketch) index(row int, item string) int {
	seed := 0xd6e8feb86659fd93 + uint64(row)*0x9e3779b97f4a7c15
	return int(hash64(seed, item) % uint64(c.width))
}

func TopK(estimates map[string]uint64, k int) []HeavyHitter {
	if k <= 0 {
		return nil
	}

	items := make([]HeavyHitter, 0, len(estimates))
	for item, estimate := range estimates {
		items = append(items, HeavyHitter{Item: item, Estimate: estimate})
	}

	slices.SortFunc(items, func(a, b HeavyHitter) int {
		if a.Estimate != b.Estimate {
			if a.Estimate > b.Estimate {
				return -1
			}
			return 1
		}
		if a.Item < b.Item {
			return -1
		}
		if a.Item > b.Item {
			return 1
		}
		return 0
	})

	return items[:min(k, len(items))]
}

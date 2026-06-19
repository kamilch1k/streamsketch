package sketch

import (
	"errors"
	"math"
	"math/bits"
)

type HyperLogLog struct {
	precision uint8
	registers []uint8
}

func NewHyperLogLog(precision uint8) (*HyperLogLog, error) {
	if precision < 4 || precision > 18 {
		return nil, errors.New("precision must be between 4 and 18")
	}

	return &HyperLogLog{
		precision: precision,
		registers: make([]uint8, 1<<precision),
	}, nil
}

func (h *HyperLogLog) Add(value string) {
	hash := hash64(0x9e3779b97f4a7c15, value)
	indexMask := uint64(len(h.registers) - 1)
	index := int(hash & indexMask)
	remaining := hash >> h.precision

	var rank uint8
	if remaining == 0 {
		rank = 64 - h.precision + 1
	} else {
		rank = uint8(bits.LeadingZeros64(remaining) - int(h.precision) + 1)
	}

	if rank > h.registers[index] {
		h.registers[index] = rank
	}
}

func (h *HyperLogLog) Estimate() float64 {
	m := float64(len(h.registers))
	alpha := alphaFor(m)

	sum := 0.0
	zeros := 0
	for _, register := range h.registers {
		sum += math.Pow(2, -float64(register))
		if register == 0 {
			zeros++
		}
	}

	raw := alpha * m * m / sum
	if raw <= 2.5*m && zeros > 0 {
		return m * math.Log(m/float64(zeros))
	}

	return raw
}

func (h *HyperLogLog) Precision() uint8 {
	return h.precision
}

func (h *HyperLogLog) RegisterCount() int {
	return len(h.registers)
}

func alphaFor(m float64) float64 {
	switch int(m) {
	case 16:
		return 0.673
	case 32:
		return 0.697
	case 64:
		return 0.709
	default:
		return 0.7213 / (1 + 1.079/m)
	}
}

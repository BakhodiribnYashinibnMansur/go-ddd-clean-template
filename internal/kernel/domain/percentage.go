package domain

import "fmt"

// Percentage represents a validated percentage in the range [0, 100].
type Percentage struct {
	value int
}

// NewPercentage constructs a Percentage after validating the range [0, 100].
func NewPercentage(v int) (Percentage, error) {
	if v < 0 || v > 100 {
		return Percentage{}, fmt.Errorf("percentage out of range [0, 100]: %d", v)
	}
	return Percentage{value: v}, nil
}

// Int returns the underlying integer value.
func (p Percentage) Int() int { return p.value }

// String formats the percentage as "N%".
func (p Percentage) String() string { return fmt.Sprintf("%d%%", p.value) }

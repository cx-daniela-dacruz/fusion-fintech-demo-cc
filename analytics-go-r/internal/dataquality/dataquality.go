// Package dataquality runs sanity checks against normalized ticks
// before they're persisted, catching feed glitches (zero prices,
// out-of-range sizes) before they pollute downstream risk numbers.
package dataquality

import "fmt"

// Issue describes a single data quality problem found in a tick.
type Issue struct {
	Field   string
	Message string
}

// Tick is the minimal shape data quality checks operate on.
type Tick struct {
	Symbol    string
	LastPrice float64
	LastSize  int64
}

const (
	minReasonablePrice = 0.0001
	maxReasonablePrice = 1_000_000.0
	maxReasonableSize  = 10_000_000
)

// CheckTick runs every registered rule against a tick and returns any
// issues found. An empty slice means the tick passed all checks.
func CheckTick(tick Tick) []Issue {
	var issues []Issue

	if tick.Symbol == "" {
		issues = append(issues, Issue{Field: "symbol", Message: "symbol must not be empty"})
	}
	if tick.LastPrice <= 0 {
		issues = append(issues, Issue{Field: "last_price", Message: "price must be positive"})
	} else if tick.LastPrice < minReasonablePrice || tick.LastPrice > maxReasonablePrice {
		issues = append(issues, Issue{
			Field:   "last_price",
			Message: fmt.Sprintf("price %.6f is outside the reasonable range", tick.LastPrice),
		})
	}
	if tick.LastSize < 0 {
		issues = append(issues, Issue{Field: "last_size", Message: "size must not be negative"})
	} else if tick.LastSize > maxReasonableSize {
		issues = append(issues, Issue{
			Field:   "last_size",
			Message: fmt.Sprintf("size %d exceeds the reasonable maximum", tick.LastSize),
		})
	}

	return issues
}

// IsClean reports whether a tick passed every check.
func IsClean(tick Tick) bool {
	return len(CheckTick(tick)) == 0
}

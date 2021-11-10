package aggregator

import (
	"math"
	"strconv"
)

// StrToFloat64 converts the provided string to its float64 representation
func StrToFloat64(v string) (float64, error) {
	vFloat, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return 0, err
	}

	return vFloat, nil
}

// TruncateFloat64ToPercentile does a round of the float to 1%
func TruncateFloat64ToPercentile(v float64) float64 {
	return math.Round(v*100) / 100
}

// PercentageChange computes the change in percents
func PercentageChange(curr, last float64) float64 {
	return math.Abs((curr-last)/last) * 100
}

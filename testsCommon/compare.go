package testsCommon

import "github.com/multiversx/mx-bridge-eth-go/core"

// EqualStringMetrics returns true if the provided metrics contain same values and keys
func EqualStringMetrics(metric1 core.StringMetrics, metric2 core.StringMetrics) bool {
	if len(metric1) != len(metric2) {
		return false
	}

	for metric, value := range metric1 {
		value2 := metric2[metric]
		if value != value2 {
			return false
		}
	}

	return true
}

// EqualIntMetrics returns true if the provided metrics contain same values and keys
func EqualIntMetrics(metric1 core.IntMetrics, metric2 core.IntMetrics) bool {
	if len(metric1) != len(metric2) {
		return false
	}

	for metric, value := range metric1 {
		value2 := metric2[metric]
		if value != value2 {
			return false
		}
	}

	return true
}

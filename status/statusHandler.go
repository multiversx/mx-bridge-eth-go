package status

import (
	"sync"

	"github.com/ElrondNetwork/elrond-eth-bridge/core"
)

type statusHandler struct {
	mutStatus     sync.RWMutex
	intMetrics    map[string]int
	stringMetrics map[string]string
	name          string
}

// NewStatusHandler creates a new instance of the status handler
func NewStatusHandler(name string) (*statusHandler, error) {
	if len(name) == 0 {
		return nil, ErrEmptyName
	}

	return &statusHandler{
		intMetrics:    make(map[string]int),
		stringMetrics: make(map[string]string),
		name:          name,
	}, nil
}

// SetIntMetric will set the metric from an int value
func (sh *statusHandler) SetIntMetric(metric string, value int) {
	sh.mutStatus.Lock()
	defer sh.mutStatus.Unlock()

	sh.intMetrics[metric] = value
}

// AddIntMetric will update the provided metric with the delta value
func (sh *statusHandler) AddIntMetric(metric string, delta int) {
	sh.mutStatus.Lock()
	defer sh.mutStatus.Unlock()

	sh.intMetrics[metric] += delta
}

// SetStringMetric will update the provided metric with string value
func (sh *statusHandler) SetStringMetric(metric string, val string) {
	sh.mutStatus.Lock()
	defer sh.mutStatus.Unlock()

	sh.stringMetrics[metric] = val
}

// GetStringMetrics returns the string metrics
func (sh *statusHandler) GetStringMetrics() core.StringMetrics {
	metrics := make(core.StringMetrics)

	sh.mutStatus.RLock()
	defer sh.mutStatus.RUnlock()

	for metric, value := range sh.stringMetrics {
		metrics[metric] = value
	}

	return metrics
}

// GetIntMetrics returns the int metrics
func (sh *statusHandler) GetIntMetrics() core.IntMetrics {
	metrics := make(core.IntMetrics)

	sh.mutStatus.RLock()
	defer sh.mutStatus.RUnlock()

	for metric, value := range sh.intMetrics {
		metrics[metric] = value
	}

	return metrics
}

// Name returns the status handler's name
func (sh *statusHandler) Name() string {
	return sh.name
}

// IsInterfaceNil returns true if there is no value under the interface
func (sh *statusHandler) IsInterfaceNil() bool {
	return sh == nil
}

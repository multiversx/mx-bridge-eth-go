package disabled

import "github.com/multiversx/mx-bridge-eth-go/core"

// StatusHandler represents the disabled status handler implementation
type StatusHandler struct {
}

// SetIntMetric does nothing
func (handler *StatusHandler) SetIntMetric(_ string, _ int) {
}

// AddIntMetric does nothing
func (handler *StatusHandler) AddIntMetric(_ string, _ int) {
}

// SetStringMetric does nothing
func (handler *StatusHandler) SetStringMetric(_ string, _ string) {
}

// Name returns an empty string
func (handler *StatusHandler) Name() string {
	return ""
}

// GetAllMetrics returns an empty map
func (handler *StatusHandler) GetAllMetrics() core.GeneralMetrics {
	return make(core.GeneralMetrics)
}

// IsInterfaceNil returns true if there is no value under the interface
func (handler *StatusHandler) IsInterfaceNil() bool {
	return handler == nil
}

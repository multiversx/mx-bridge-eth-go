package testsCommon

import "sync"

// StatusHandlerMock -
type StatusHandlerMock struct {
	mutStatus     sync.RWMutex
	intMetrics    map[string]int
	stringMetrics map[string]string
}

// NewStatusHandlerMock -
func NewStatusHandlerMock() *StatusHandlerMock {
	return &StatusHandlerMock{
		intMetrics:    make(map[string]int),
		stringMetrics: make(map[string]string),
	}
}

// SetIntMetric -
func (mock *StatusHandlerMock) SetIntMetric(metric string, value int) {
	mock.mutStatus.Lock()
	defer mock.mutStatus.Unlock()

	mock.intMetrics[metric] = value
}

// AddIntMetric -
func (mock *StatusHandlerMock) AddIntMetric(metric string, delta int) {
	mock.mutStatus.Lock()
	defer mock.mutStatus.Unlock()

	mock.intMetrics[metric] += delta
}

// SetStringMetric -
func (mock *StatusHandlerMock) SetStringMetric(metric string, val string) {
	mock.mutStatus.Lock()
	defer mock.mutStatus.Unlock()

	mock.stringMetrics[metric] = val
}

// Name -
func (mock *StatusHandlerMock) Name() string {
	return "mock"
}

// GetIntMetric -
func (mock *StatusHandlerMock) GetIntMetric(metric string) int {
	mock.mutStatus.RLock()
	defer mock.mutStatus.RUnlock()

	return mock.intMetrics[metric]
}

// GetStringMetric -
func (mock *StatusHandlerMock) GetStringMetric(metric string) string {
	mock.mutStatus.RLock()
	defer mock.mutStatus.RUnlock()

	return mock.stringMetrics[metric]
}

// IsInterfaceNil -
func (mock *StatusHandlerMock) IsInterfaceNil() bool {
	return mock == nil
}

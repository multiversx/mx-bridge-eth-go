package testsCommon

import (
	"github.com/ElrondNetwork/elrond-eth-bridge/core"
)

// StatusHandlerStub -
type StatusHandlerStub struct {
	SetIntMetricCalled    func(metric string, value int)
	AddIntMetricCalled    func(metric string, delta int)
	SetStringMetricCalled func(metric string, val string)
	NameCalled            func() string
	GetAllMetricsCalled   func() core.GeneralMetrics
}

// SetIntMetric -
func (stub *StatusHandlerStub) SetIntMetric(metric string, value int) {
	if stub.SetIntMetricCalled != nil {
		stub.SetIntMetricCalled(metric, value)
	}
}

// AddIntMetric -
func (stub *StatusHandlerStub) AddIntMetric(metric string, delta int) {
	if stub.AddIntMetricCalled != nil {
		stub.AddIntMetricCalled(metric, delta)
	}
}

// SetStringMetric -
func (stub *StatusHandlerStub) SetStringMetric(metric string, val string) {
	if stub.SetStringMetricCalled != nil {
		stub.SetStringMetricCalled(metric, val)
	}
}

// Name -
func (stub *StatusHandlerStub) Name() string {
	if stub.NameCalled != nil {
		return stub.NameCalled()
	}

	return ""
}

// GetAllMetrics -
func (stub *StatusHandlerStub) GetAllMetrics() core.GeneralMetrics {
	if stub.GetAllMetricsCalled != nil {
		return stub.GetAllMetricsCalled()
	}

	return make(core.GeneralMetrics)
}

// IsInterfaceNil -
func (stub *StatusHandlerStub) IsInterfaceNil() bool {
	return stub == nil
}

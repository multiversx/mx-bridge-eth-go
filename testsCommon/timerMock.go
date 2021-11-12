package testsCommon

import (
	"time"
)

// TimerMock -
type TimerMock struct {
	OverrideTimeAfter time.Duration
}

// NowUnix -
func (tm *TimerMock) NowUnix() int64 {
	return time.Now().Unix()
}

// Start -
func (tm *TimerMock) Start() {
}

// Close -
func (tm *TimerMock) Close() error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (tm *TimerMock) IsInterfaceNil() bool {
	return tm == nil
}

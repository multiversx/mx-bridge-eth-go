package mock

import (
	"time"
)

// TimerMock -
type TimerMock struct {
}

// After -
func (tm *TimerMock) After(d time.Duration) <-chan time.Time {
	return time.After(d)
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

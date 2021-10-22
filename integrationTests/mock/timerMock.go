package mock

import (
	"time"
)

// TimerMock -
type TimerMock struct {
	OverrideTimeAfter time.Duration
}

// After -
func (tm *TimerMock) After(d time.Duration) <-chan time.Time {
	if tm.OverrideTimeAfter == 0 {
		return time.After(d)
	}

	return time.After(tm.OverrideTimeAfter)
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

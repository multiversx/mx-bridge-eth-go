package mock

import "time"

// SyncTimerStub -
type SyncTimerStub struct {
	CloseCalled                func() error
	StartSyncingTimeCalled     func()
	ClockOffsetCalled          func() time.Duration
	FormattedCurrentTimeCalled func() string
	CurrentTimeCalled          func() time.Time
}

// Close -
func (stub *SyncTimerStub) Close() error {
	if stub.CloseCalled != nil {
		return stub.CloseCalled()
	}

	return nil
}

// StartSyncingTime -
func (stub *SyncTimerStub) StartSyncingTime() {
	if stub.StartSyncingTimeCalled != nil {
		stub.StartSyncingTimeCalled()
	}
}

// ClockOffset -
func (stub *SyncTimerStub) ClockOffset() time.Duration {
	if stub != nil {
		stub.ClockOffsetCalled()
	}

	return 0
}

// FormattedCurrentTime -
func (stub *SyncTimerStub) FormattedCurrentTime() string {
	if stub.FormattedCurrentTimeCalled != nil {
		return stub.FormattedCurrentTimeCalled()
	}

	return ""
}

// CurrentTime -
func (stub *SyncTimerStub) CurrentTime() time.Time {
	if stub.CurrentTimeCalled != nil {
		return stub.CurrentTimeCalled()
	}

	return time.Time{}
}

// IsInterfaceNil -
func (stub *SyncTimerStub) IsInterfaceNil() bool {
	return stub == nil
}

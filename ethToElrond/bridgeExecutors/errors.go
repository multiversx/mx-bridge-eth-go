package bridgeExecutors

import "errors"

// ErrNilBridge signals that a nil bridge instance has been provided
var ErrNilBridge = errors.New("nil bridge")

// ErrNilLogger signals that a nil logger instance has been provided
var ErrNilLogger = errors.New("nil logger")

// ErrNilTopologyProvider signals that a nil topology provider instance has been used
var ErrNilTopologyProvider = errors.New("nil topology provider")

// ErrNilQuorumProvider signals that a nil quorum provider instance has been used
var ErrNilQuorumProvider = errors.New("nil quorum provider")

// ErrNilTimer signals that a nil timer was provided
var ErrNilTimer = errors.New("nil timer")

// ErrNilDurationsMap signals that a nil durations map was provided
var ErrNilDurationsMap = errors.New("nil durations map")

// ErrDurationForStepNotFound signals that a duration for provided step was not found
var ErrDurationForStepNotFound = errors.New("duration for step not found")

// ErrNilStatusHandler signals that a nil status handler was provided
var ErrNilStatusHandler = errors.New("nil status handler")

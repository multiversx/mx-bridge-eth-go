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

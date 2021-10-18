package steps

import "errors"

// ErrNilBridgeExecutor signals that a nil bridge executor has been provided
var ErrNilBridgeExecutor = errors.New("nil bridge executor")

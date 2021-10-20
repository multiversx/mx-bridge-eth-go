package steps

import "errors"

// ErrNilBridgeExecutor signals that a nil bridge executor has been provided
var ErrNilBridgeExecutor = errors.New("nil bridge executor")

// ErrDuplicatedStepIdentifier signals that the same step identifier was used in 2 or more steps
var ErrDuplicatedStepIdentifier = errors.New("duplicated step identifier used in multiple steps")

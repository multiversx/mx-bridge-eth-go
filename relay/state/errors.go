package state

import "errors"

// ErrStepNotFound signals that the step was not registered (and found in the steps map)
var ErrStepNotFound = errors.New("step not found")

// ErrNilStepsMap signals that a nil steps map was provided
var ErrNilStepsMap = errors.New("nil steps map")

// ErrNilStep signals that a nil step was provided
var ErrNilStep = errors.New("nil step")

// ErrNilLogger signals that a nil logger was provided
var ErrNilLogger = errors.New("nil logger")

package status

import "errors"

// ErrEmptyName signals that an empty name was provided
var ErrEmptyName = errors.New("empty name")

// ErrStatusHandlerExists signals that a status handler with the same name was already registered
var ErrStatusHandlerExists = errors.New("status handler exists with the same name")

// ErrMissingStatusHandler signals that a missing status handler has occurred
var ErrMissingStatusHandler = errors.New("missing status handler")

// ErrNilStorer signals that a nil storer was provided
var ErrNilStorer = errors.New("nil storer")

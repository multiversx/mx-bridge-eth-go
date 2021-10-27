package relay

import "errors"

// ErrInvalidDurationConfig signals that an invalid config duration was provided
var ErrInvalidDurationConfig = errors.New("invalid config duration")

// ErrMissingDurationConfig signals that a missing config duration was detected
var ErrMissingDurationConfig = errors.New("missing config duration")

// ErrMissingConfig signals that a missing config was detected
var ErrMissingConfig = errors.New("missing config duration")

package facade

import "errors"

// ErrNilMetricsHolder signals that a nil metrics holder was provided
var ErrNilMetricsHolder = errors.New("nil metrics holder")

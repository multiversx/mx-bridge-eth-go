package eth

import "errors"

// ErrNilGasHandler signals that a nil gas handler was provided
var ErrNilGasHandler = errors.New("nil gas handler")

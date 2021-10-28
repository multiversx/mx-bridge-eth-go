package roleProvider

import "errors"

// ErrNilChainClient signals that a nil chain client was provided
var ErrNilChainClient = errors.New("nil chain client")

// ErrNilLogger signals that a nil logger was provided
var ErrNilLogger = errors.New("nil logger")

// ErrInvalidValue signals that an invalid value was provided
var ErrInvalidValue = errors.New("invalid value")

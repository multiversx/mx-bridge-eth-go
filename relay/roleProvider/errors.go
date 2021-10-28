package roleProvider

import "errors"

// ErrNilChainInteractor signals that a nil chain interactor was provided
var ErrNilChainInteractor = errors.New("nil chain interactor")

// ErrNilLogger signals that a nil logger was provided
var ErrNilLogger = errors.New("nil logger")

// ErrInvalidValue signals that an invalid value was provided
var ErrInvalidValue = errors.New("invalid value")

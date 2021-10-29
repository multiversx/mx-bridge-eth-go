package errors

import "errors"

// ErrNilHttpServer signals that a nil http server has been provided
var ErrNilHttpServer = errors.New("nil http server")

// ErrNilFacade signals that a nil facade has been provided
var ErrNilFacade = errors.New("nil facade")

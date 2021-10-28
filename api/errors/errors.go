package errors

import "errors"

// ErrNilHttpServer signals that a nil http server has been provided
var ErrNilHttpServer = errors.New("nil http server")

// ErrCannotCreateWebServer signals that the gin web server cannot be created
var ErrNilFacade = errors.New("nil facade")

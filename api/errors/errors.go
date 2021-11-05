package errors

import "errors"

// ErrNilHttpServer signals that a nil http server has been provided
var ErrNilHttpServer = errors.New("nil http server")

// ErrNilFacade signals that a nil facade has been provided
var ErrNilFacade = errors.New("nil facade")

// ErrNilAntiFloodConfig signals that a nil anti flood config has been provided
var ErrNilAntiFloodConfig = errors.New("nil antiflood config")

// ErrNilApiConfig signals that a nil api config has been provided
var ErrNilApiConfig = errors.New("nil api config")

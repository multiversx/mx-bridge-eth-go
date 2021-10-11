package contracts

import "errors"

// ErrNilTokenHandler signals that a nil token handler instance has been provided
var ErrNilTokenHandler = errors.New("nil TokenHandler instance")

// ErrInvalidNumberOfArguments signals that an invalid number of arguments has been provided
var ErrInvalidNumberOfArguments = errors.New("invalid number of arguments")

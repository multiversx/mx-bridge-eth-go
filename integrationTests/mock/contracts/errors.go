package contracts

import "errors"

// ErrNilTokenHandler signals that a nil token handler instance has been provided
var ErrNilTokenHandler = errors.New("nil TokenHandler instance")

// ErrInsufficientArguments signals that insufficient arguments were provided
var ErrInsufficientArguments = errors.New("insufficient argument")

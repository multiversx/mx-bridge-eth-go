package clients

import "errors"

// TODO: add more errors that are duplicated in different clients
var (

	// ErrNilAddressConverter signals that a nil address converter has been provided
	ErrNilAddressConverter = errors.New("nil address converter")
)

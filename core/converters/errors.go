package converters

import "errors"

var (
	// ErrNotUint64Bytes is the error returned when the provided bytes do not represent a valid uint64 number
	ErrNotUint64Bytes = errors.New("provided bytes do not represent a valid uint64 number")
)

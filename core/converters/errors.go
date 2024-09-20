package converters

import "errors"

var (
	ErrNotUint64Bytes = errors.New("provided bytes do not represent a valid uint64 number")
)

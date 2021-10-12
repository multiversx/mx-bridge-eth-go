package elrond

import "errors"

// ErrNilProxy signals that a nil ElrondProxy instance was provided
var ErrNilProxy = errors.New("nil ElrondProxy")

// ErrUnexpectedLengthOnResponse signals that an unexpected length for a response data has occurred
var ErrUnexpectedLengthOnResponse = errors.New("contract error, unexpected 0 length response data")

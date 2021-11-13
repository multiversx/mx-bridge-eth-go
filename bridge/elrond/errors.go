package elrond

import "errors"

// ErrNilProxy signals that a nil ElrondProxy instance was provided
var ErrNilProxy = errors.New("nil ElrondProxy")

// ErrNoStatusForBatchID signals that no status is available for the batch ID
var ErrNoStatusForBatchID = errors.New("no status available for batch ID")

// ErrNilBatchId signals that a nil batch ID has been provided
var ErrNilBatchId = errors.New("nil batch ID")

// ErrBatchNotFinished signals that a batch is not finalized yet
var ErrBatchNotFinished = errors.New("batch not finished yet")

// ErrMalformedBatchResponse signals that a batch response is malformed
var ErrMalformedBatchResponse = errors.New("batch response is malformed")

// ErrNilPrivateKey signals that a nil private key has been provided
var ErrNilPrivateKey = errors.New("nil private key")

// ErrNilAddressHandler signals that a nil address handler has been provided
var ErrNilAddressHandler = errors.New("nil address handler")

// ErrInvalidGasValue signals that an invalid gas value was provided
var ErrInvalidGasValue = errors.New("invalid gas value")

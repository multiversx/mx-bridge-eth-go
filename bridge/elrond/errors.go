package elrond

import "errors"

// ErrNilProxy signals that a nil ElrondProxy instance was provided
var ErrNilProxy = errors.New("nil ElrondProxy")

// ErrNoStatusForBatchID signals that no status is available for the batch ID
var ErrNoStatusForBatchID = errors.New("no status available for batch ID")

// ErrNilBatchId signals that a nil batch ID has been provided
var ErrNilBatchId = errors.New("nil batch ID")

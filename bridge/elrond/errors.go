package elrond

import "errors"

// ErrNilProxy signals that a nil ElrondProxy instance was provided
var ErrNilProxy = errors.New("nil ElrondProxy")

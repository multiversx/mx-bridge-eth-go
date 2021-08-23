package elrond

import "errors"

// ErrNilProxy signals that a nil ElrondProxy instance was provided
var ErrNilProxy = errors.New("nil ElrondProxy")

// ErrInvalidNonceUpdateInterval signals that an invalid NonceUpdateInterval value was provided
var ErrInvalidNonceUpdateInterval = errors.New("invalid config.NonceUpdateInSeconds value")

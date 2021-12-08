package topology

import (
    "errors"
)

var (
    ErrNilSortedPublicKeys = errors.New("nil sorted public keys")
    ErrInvalidStepDuration = errors.New("invalid step duration")
    ErrNilTimer            = errors.New("nil timer")
    ErrNilAddress          = errors.New("nil address")
)

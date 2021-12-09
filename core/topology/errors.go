package topology

import (
    "errors"
)

var (
    errNilSortedPublicKeys = errors.New("nil sorted public keys")
    errInvalidStepDuration = errors.New("invalid step duration")
    errNilTimer            = errors.New("nil timer")
    errNilAddress          = errors.New("nil address")
)

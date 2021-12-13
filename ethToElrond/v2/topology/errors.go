package topology

import (
	"errors"
)

var (
	errNilPublicKeysProvider = errors.New("nil public keys provider")
	errInvalidStepDuration   = errors.New("invalid step duration")
	errNilTimer              = errors.New("nil timer")
	errNilAddress            = errors.New("nil address")
)

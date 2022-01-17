package topology

import (
	"errors"
)

var (
	errNilPublicKeysProvider    = errors.New("nil public keys provider")
	errInvalidIntervalForLeader = errors.New("invalid interval for leader")
	errNilTimer                 = errors.New("nil timer")
	errEmptyAddress             = errors.New("empty address")
)

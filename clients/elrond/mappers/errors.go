package mappers

import "errors"

var (
	errNilDataGetter = errors.New("nil elrondClientDataGetter")
	errUnknownToken  = errors.New("unknown token")
)

package module

import "errors"

var (
	errNilProxy  = errors.New("nil proxy")
	errNilLogger = errors.New("nil logger")
)

package antiflood

import "errors"

var (
	errNilTopicFloodPreventer  = errors.New("nil topic flood preventer")
	errInvalidNumberOfMessages = errors.New("invalid number of messages")
	errSystemBusy              = errors.New("system busy")
)

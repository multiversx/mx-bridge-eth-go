package filters

import "errors"

var (
	errNilLogger         = errors.New("nil logger")
	errNoItemsAllowed    = errors.New("no items allowed")
	errUnsupportedMarker = errors.New("unsupported marker")
	errMissingEthPrefix  = errors.New("missing Ethereum address prefix")
)

package bridge

import (
	"errors"

	"github.com/ElrondNetwork/elrond-sdk-erdgo/data"
)

var notImplemented = errors.New("method not implemented")

// StaticAddress is an instance of Address Handler
var StaticAddress = data.NewAddressFromBytes([]byte{})

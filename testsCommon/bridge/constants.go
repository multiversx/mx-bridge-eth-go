package bridge

import (
	"errors"
)

var notImplemented = errors.New("method not implemented")

// CallDataMock -
var CallDataMock = func() []byte {
	b := []byte{
		1,
		0, 0, 0, 28,
		0, 0, 0, 3, 'a', 'b', 'c',
		0x00, 0x00, 0x00, 0x00, 0x1D, 0xCD, 0x65, 0x00, // gas limit
		0, 0, 0, 1, // numArguments
		0, 0, 0, 5, // argument 0 length
		'd', 'e', 'f', 'g', 'h', // argument 0 data
	}

	return b
}()

// EthCallDataMock -
var EthCallDataMock = func() []byte {
	b := []byte{
		0, 0, 0, 3, 'a', 'b', 'c',
		0x00, 0x00, 0x00, 0x00, 0x1D, 0xCD, 0x65, 0x00, // gas limit
		0, 0, 0, 1, // numArguments
		0, 0, 0, 5, // argument 0 length
		'd', 'e', 'f', 'g', 'h', // argument 0 data
	}

	return b
}()

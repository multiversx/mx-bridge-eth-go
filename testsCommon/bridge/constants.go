package bridge

import (
	"bytes"
	"errors"
)

var notImplemented = errors.New("method not implemented")

// CallDataMock -
var CallDataMock = func() []byte {
	b := []byte{0x01, 0x00, 0x00, 0x00, 0x03}
	b = append(b, []byte("abc")...)
	b = append(b, 0x00, 0x00, 0x00, 0x00, 0x1D, 0xCD, 0x65, 0x00) // Gas limit
	b = append(b, 0x00, 0x00, 0x00, 0x01)                         // numArguments
	b = append(b, 0x00, 0x00, 0x00, 0x05)                         // Argument 0 length
	b = append(b, bytes.Repeat([]byte{'a'}, 5)...)                // Argument 0 data
	return b
}()

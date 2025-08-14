package sui

import (
	"encoding/binary"
	"encoding/hex"
)

// AddressBytesToString converts a Sui address from bytes to hex string
func AddressBytesToString(bytes []byte) string {
	return string(bytes)
}

// AddressFromBytes returns the address from bytes as a proper Sui address string
func AddressFromBytes(bytes [32]byte) string {
	hexAddress := hex.EncodeToString(bytes[:])
	return "0x" + hexAddress
}

func EncodeManagedBufferTokenType(data []byte) string {
	lenBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(lenBytes, uint32(len(data)))
	encoded := append(lenBytes, data...)

	return hex.EncodeToString(encoded)
}

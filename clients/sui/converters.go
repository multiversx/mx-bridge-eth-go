package sui

import (
	"encoding/hex"
	"fmt"
	"strings"
)

// AddressBytesToString converts a Sui address from bytes to hex string
func AddressBytesToString(bytes []byte) string {
	return "0x" + hex.EncodeToString(bytes)
}

// AddressStringToBytes converts a Sui address from hex string to bytes
func AddressStringToBytes(address string) ([]byte, error) {
	hexStr := strings.TrimPrefix(address, "0x")

	bytes, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, fmt.Errorf("invalid hex string: %w", err)
	}

	if len(bytes) != 32 {
		return nil, fmt.Errorf("invalid Sui address length: expected 32 bytes, got %d", len(bytes))
	}

	return bytes, nil
}

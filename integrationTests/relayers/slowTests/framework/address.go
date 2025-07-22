package framework

import (
	"encoding/hex"
	"fmt"
	"strings"
	"testing"

	sdkCore "github.com/multiversx/mx-sdk-go/core"
	"github.com/multiversx/mx-sdk-go/data"
	"github.com/stretchr/testify/require"
)

// MvxAddress holds the different forms a MultiversX address might have
type MvxAddress struct {
	sdkCore.AddressHandler
	bytes  []byte
	bech32 string
	hex    string
}

// NewMvxAddressFromBytes return a new instance of MvxAddress from bytes
func NewMvxAddressFromBytes(tb testing.TB, bytes []byte) *MvxAddress {
	address := &MvxAddress{
		bytes:          make([]byte, len(bytes)),
		hex:            hex.EncodeToString(bytes),
		AddressHandler: data.NewAddressFromBytes(bytes),
	}

	var err error
	copy(address.bytes, bytes)
	address.bech32, err = addressPubkeyConverter.Encode(bytes)
	require.Nil(tb, err)

	return address
}

// NewMvxAddressFromBech32 return a new instance of MvxAddress from the bech32 string
func NewMvxAddressFromBech32(tb testing.TB, bech32 string) *MvxAddress {
	addressHandler, err := data.NewAddressFromBech32String(bech32)
	require.Nil(tb, err)

	return &MvxAddress{
		bytes:          addressHandler.AddressBytes(),
		hex:            hex.EncodeToString(addressHandler.AddressBytes()),
		bech32:         bech32,
		AddressHandler: addressHandler,
	}
}

// Bytes returns the bytes format address
func (address *MvxAddress) Bytes() []byte {
	return address.bytes
}

// Bech32 returns the bech32 string format address
func (address *MvxAddress) Bech32() string {
	return address.bech32
}

// Hex returns the hex string format address
func (address *MvxAddress) Hex() string {
	return address.hex
}

// String returns the address in bech32 format
func (address *MvxAddress) String() string {
	return address.bech32
}

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

package converters

import (
	"encoding/hex"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-sdk-go/data"
	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

func TestConvertFromByteSliceToArray(t *testing.T) {
	t.Parallel()

	buff := []byte("12345678901234567890123456789012")

	result := data.NewAddressFromBytes(buff).AddressSlice()
	assert.Equal(t, buff, result[:])
}

func TestTrimWhiteSpaceCharacters(t *testing.T) {
	t.Parallel()

	dataField := "aaII139HSAh32q782!$#*$(nc"

	input := " " + dataField
	assert.Equal(t, dataField, TrimWhiteSpaceCharacters(input))

	input = "\t " + dataField
	assert.Equal(t, dataField, TrimWhiteSpaceCharacters(input))

	input = "\t " + dataField + "\n"
	assert.Equal(t, dataField, TrimWhiteSpaceCharacters(input))

	input = "\t\n " + dataField + "\n\n\n\n\t"
	assert.Equal(t, dataField, TrimWhiteSpaceCharacters(input))
}

func TestAddressConverter_ToBech32String(t *testing.T) {
	t.Parallel()

	addrConv, err := NewAddressConverter()
	require.Nil(t, err)
	assert.False(t, check.IfNil(addrConv))

	t.Run("invalid bytes should return empty", func(t *testing.T) {
		str, errLocal := addrConv.ToBech32String([]byte("invalid"))
		assert.NotNil(t, errLocal)
		assert.Contains(t, errLocal.Error(), "wrong size when encoding address, expected length 32, received 7")
		assert.Empty(t, str)
	})
	t.Run("should work", func(t *testing.T) {
		expected := "erd1r69gk66fmedhhcg24g2c5kn2f2a5k4kvpr6jfw67dn2lyydd8cfswy6ede"
		bytes, _ := hex.DecodeString("1e8a8b6b49de5b7be10aaa158a5a6a4abb4b56cc08f524bb5e6cd5f211ad3e13")
		bech32Address, errLocal := addrConv.ToBech32String(bytes)
		assert.Equal(t, expected, bech32Address)
		assert.Nil(t, errLocal)
	})
}

func TestAddressConverter_ToHexString(t *testing.T) {
	t.Parallel()

	addrConv, err := NewAddressConverter()
	require.Nil(t, err)
	assert.False(t, check.IfNil(addrConv))

	expected := "627974657320746f20656e636f6465"
	bytes := []byte("bytes to encode")
	assert.Equal(t, expected, addrConv.ToHexString(bytes))
}

func TestAddressConverter_ToHexStringWithPrefix(t *testing.T) {
	t.Parallel()

	addrConv, err := NewAddressConverter()
	require.Nil(t, err)
	assert.False(t, check.IfNil(addrConv))

	expected := "0x627974657320746f20656e636f6465"
	bytes := []byte("bytes to encode")
	assert.Equal(t, expected, addrConv.ToHexStringWithPrefix(bytes))
}

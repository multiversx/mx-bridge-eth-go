package core

import (
	"encoding/hex"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvertFromByteSliceToArray(t *testing.T) {
	t.Parallel()

	buff := []byte("12345678901234567890123456789012")

	result := ConvertFromByteSliceToArray(buff)
	assert.Equal(t, buff, result[:])
}

func TestTrimWhiteSpaceCharacters(t *testing.T) {
	t.Parallel()

	data := "aaII139HSAh32q782!$#*$(nc"

	input := " " + data
	assert.Equal(t, data, TrimWhiteSpaceCharacters(input))

	input = "\t " + data
	assert.Equal(t, data, TrimWhiteSpaceCharacters(input))

	input = "\t " + data + "\n"
	assert.Equal(t, data, TrimWhiteSpaceCharacters(input))

	input = "\t\n " + data + "\n\n\n\n\t"
	assert.Equal(t, data, TrimWhiteSpaceCharacters(input))
}

func TestAddressConverter_ToBech32String(t *testing.T) {
	t.Parallel()

	addrConv := NewAddressConverter()
	assert.False(t, check.IfNil(addrConv))

	t.Run("invalid bytes should return empty", func(t *testing.T) {
		str := addrConv.ToBech32String([]byte("invalid"))
		assert.Equal(t, "", str)
	})
	t.Run("should work", func(t *testing.T) {
		expected := "erd1r69gk66fmedhhcg24g2c5kn2f2a5k4kvpr6jfw67dn2lyydd8cfswy6ede"
		bytes, _ := hex.DecodeString("1e8a8b6b49de5b7be10aaa158a5a6a4abb4b56cc08f524bb5e6cd5f211ad3e13")
		assert.Equal(t, expected, addrConv.ToBech32String(bytes))
	})
}

func TestAddressConverter_ToHexString(t *testing.T) {
	t.Parallel()

	addrConv := NewAddressConverter()
	assert.False(t, check.IfNil(addrConv))

	expected := "627974657320746f20656e636f6465"
	bytes := []byte("bytes to encode")
	assert.Equal(t, expected, addrConv.ToHexString(bytes))
}
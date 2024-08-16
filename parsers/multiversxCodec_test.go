package parsers

import (
	"bytes"
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/multiversx/mx-sdk-go/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestProxySCCompleteCallData() ProxySCCompleteCallData {
	ethUnhexed, _ := hex.DecodeString("880ec53af800b5cd051531672ef4fc4de233bd5d")
	completeCallData := ProxySCCompleteCallData{
		RawCallData: []byte{'A', 'B', 'C'},
		From:        common.Address{},
		Token:       "ETHUSDC-0ae8ee",
		Amount:      big.NewInt(20000),
		Nonce:       1,
	}
	completeCallData.To, _ = data.NewAddressFromBech32String("erd1qqqqqqqqqqqqqpgqsudu3a3n9yu62k5qkgcpy4j9ywl2x2gl5smsy7t4uv")
	completeCallData.From.SetBytes(ethUnhexed)

	return completeCallData
}

func TestMultiversxCodec_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var instance *MultiversxCodec
	assert.True(t, instance.IsInterfaceNil())

	instance = &MultiversxCodec{}
	assert.False(t, instance.IsInterfaceNil())
}

func TestMultiversxCodec_ExtractGasLimitFromRawCallData(t *testing.T) {
	t.Parallel()

	codec := &MultiversxCodec{}

	t.Run("empty buffer should error", func(t *testing.T) {
		t.Parallel()

		gasLimit, err := codec.ExtractGasLimitFromRawCallData(nil)
		assert.Equal(t, errBufferTooShortForMarker, err)
		assert.Zero(t, gasLimit)

		gasLimit, err = codec.ExtractGasLimitFromRawCallData(make([]byte, 0))
		assert.Equal(t, errBufferTooShortForMarker, err)
		assert.Zero(t, gasLimit)
	})
	t.Run("unexpected marker should error", func(t *testing.T) {
		t.Parallel()

		gasLimit, err := codec.ExtractGasLimitFromRawCallData([]byte{0x03})
		assert.ErrorIs(t, err, errUnexpectedMarker)
		assert.Contains(t, err.Error(), ": 3")
		assert.Zero(t, gasLimit)
	})
	t.Run("buffer contains missing data marker should error", func(t *testing.T) {
		t.Parallel()

		buff := []byte{0}

		gasLimit, err := codec.ExtractGasLimitFromRawCallData(buff)
		assert.ErrorIs(t, err, errUnexpectedMarker)
		assert.Contains(t, err.Error(), ": 0")
		assert.Zero(t, gasLimit)
	})
	t.Run("buffer to short for call data length should error", func(t *testing.T) {
		t.Parallel()

		buff := []byte{1}

		gasLimit, err := codec.ExtractGasLimitFromRawCallData(buff)
		assert.ErrorIs(t, err, errBufferTooShortForUint32)
		assert.Contains(t, err.Error(), "for len of call data")
		assert.Zero(t, gasLimit)
	})
	t.Run("buffer len for call data mismatch should error", func(t *testing.T) {
		t.Parallel()

		buff := []byte{
			1,
			0, 0, 0, 1}

		gasLimit, err := codec.ExtractGasLimitFromRawCallData(buff)
		assert.ErrorIs(t, err, errBufferLenMismatch)
		assert.Contains(t, err.Error(), "actual 0, declared 1")
		assert.Zero(t, gasLimit)
	})
	t.Run("buffer to short for function length should error", func(t *testing.T) {
		t.Parallel()

		buff := []byte{
			1,
			0, 0, 0, 0}

		gasLimit, err := codec.ExtractGasLimitFromRawCallData(buff)
		assert.ErrorIs(t, err, errBufferTooShortForLength)
		assert.Contains(t, err.Error(), "for function")
		assert.Zero(t, gasLimit)
	})
	t.Run("buffer to short for function should error", func(t *testing.T) {
		t.Parallel()

		buff := []byte{
			1,
			0, 0, 0, 4,
			0, 0, 0, 5}

		gasLimit, err := codec.ExtractGasLimitFromRawCallData(buff)
		assert.ErrorIs(t, err, errBufferTooShortForString)
		assert.Contains(t, err.Error(), "for function")
		assert.Zero(t, gasLimit)
	})
	t.Run("buffer to short for gas limit should error", func(t *testing.T) {
		t.Parallel()

		buff := []byte{
			1,
			0, 0, 0, 14,
			0, 0, 0, 3, 'a', 'b', 'c',
			0, 0, 0, 0, 0, 0, 0, // malformed gas limit (7 bytes for an uint64)
		}

		gasLimit, err := codec.ExtractGasLimitFromRawCallData(buff)
		assert.ErrorIs(t, err, errBufferTooShortForUint64)
		assert.Contains(t, err.Error(), "for gas limit")
		assert.Zero(t, gasLimit)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		buff := []byte{
			1,
			0, 0, 0, 15,
			0, 0, 0, 3, 'a', 'b', 'c',
			0, 0, 1, 2, 3, 4, 5, 6, // gas limit is 1108152157446
		}

		gasLimit, err := codec.ExtractGasLimitFromRawCallData(buff)
		assert.Nil(t, err)
		assert.Equal(t, uint64(1108152157446), gasLimit)
	})
}

func TestMultiversxCodec_DecodeProxySCCompleteCallData(t *testing.T) {
	t.Parallel()

	codec := MultiversxCodec{}
	emptyCompleteCallData := ProxySCCompleteCallData{}

	t.Run("buffer to short for Ethereum address should error", func(t *testing.T) {
		t.Parallel()

		buff := []byte{0x01}
		completeCallData, err := codec.DecodeProxySCCompleteCallData(buff)
		assert.ErrorIs(t, err, errBufferTooShortForEthAddress)
		assert.Equal(t, emptyCompleteCallData, completeCallData)
	})
	t.Run("buffer to short for MultiversX address should error", func(t *testing.T) {
		t.Parallel()

		buff := bytes.Repeat([]byte{0x01}, 20)
		buff = append(buff, 0x1)
		completeCallData, err := codec.DecodeProxySCCompleteCallData(buff)
		assert.ErrorIs(t, err, errBufferTooShortForMvxAddress)
		assert.Equal(t, emptyCompleteCallData, completeCallData)
	})
	t.Run("invalid token size bytes should error", func(t *testing.T) {
		t.Parallel()

		buff := bytes.Repeat([]byte{0x01}, 20)                 // Eth address
		buff = append(buff, bytes.Repeat([]byte{0x01}, 32)...) // Mvx address
		buff = append(buff, []byte{0x00, 0x01, 0x04}...)       // invalid token

		completeCallData, err := codec.DecodeProxySCCompleteCallData(buff)
		assert.ErrorIs(t, err, errBufferTooShortForLength)
		assert.Contains(t, err.Error(), "for token")
		assert.Equal(t, emptyCompleteCallData, completeCallData)
	})
	t.Run("invalid token size should error", func(t *testing.T) {
		t.Parallel()

		buff := bytes.Repeat([]byte{0x01}, 20)                 // Eth address
		buff = append(buff, bytes.Repeat([]byte{0x01}, 32)...) // Mvx address
		buff = append(buff, []byte{0x00, 0x00, 0x00, 0x02}...) // token length
		buff = append(buff, 0x04)                              // instead of 2 bytes for token we have only one

		completeCallData, err := codec.DecodeProxySCCompleteCallData(buff)
		assert.ErrorIs(t, err, errBufferTooShortForString)
		assert.Contains(t, err.Error(), "for token")
		assert.Equal(t, emptyCompleteCallData, completeCallData)
	})
	t.Run("invalid big int size should error", func(t *testing.T) {
		t.Parallel()

		buff := bytes.Repeat([]byte{0x01}, 20)                 // Eth address
		buff = append(buff, bytes.Repeat([]byte{0x01}, 32)...) // Mvx address
		buff = append(buff, []byte{0x00, 0x00, 0x00, 0x02}...) // token size
		buff = append(buff, []byte{0x02, 0x03}...)             // token
		buff = append(buff, []byte{0x00, 0x00, 0x00}...)       // invalid amount size

		completeCallData, err := codec.DecodeProxySCCompleteCallData(buff)
		assert.ErrorIs(t, err, errBufferTooShortForLength)
		assert.Contains(t, err.Error(), "for amount")
		assert.Equal(t, emptyCompleteCallData, completeCallData)
	})
	t.Run("invalid big int bytes should error", func(t *testing.T) {
		t.Parallel()

		buff := bytes.Repeat([]byte{0x01}, 20)                 // Eth address
		buff = append(buff, bytes.Repeat([]byte{0x01}, 32)...) // Mvx address
		buff = append(buff, []byte{0x00, 0x00, 0x00, 0x02}...) // token size
		buff = append(buff, []byte{0x02, 0x03}...)             // token
		buff = append(buff, []byte{0x00, 0x00, 0x00, 0x05}...) // amount size
		buff = append(buff, []byte{0x00}...)                   // invalid amount

		completeCallData, err := codec.DecodeProxySCCompleteCallData(buff)
		assert.ErrorIs(t, err, errBufferTooShortForBigInt)
		assert.Contains(t, err.Error(), "for amount")
		assert.Equal(t, emptyCompleteCallData, completeCallData)
	})
	t.Run("invalid nonce should error", func(t *testing.T) {
		t.Parallel()

		buff := bytes.Repeat([]byte{0x01}, 20)                 // Eth address
		buff = append(buff, bytes.Repeat([]byte{0x01}, 32)...) // Mvx address
		buff = append(buff, []byte{0x00, 0x00, 0x00, 0x02}...) // token size
		buff = append(buff, []byte{0x02, 0x03}...)             // token
		buff = append(buff, []byte{0x00, 0x00, 0x00, 0x01}...) // amount size
		buff = append(buff, []byte{0x01}...)                   // amount
		buff = append(buff, []byte{0x03, 0x04}...)             // invalid nonce

		completeCallData, err := codec.DecodeProxySCCompleteCallData(buff)
		assert.ErrorIs(t, err, errBufferTooShortForUint64)
		assert.Contains(t, err.Error(), "for nonce")
		assert.Equal(t, emptyCompleteCallData, completeCallData)
	})
	t.Run("invalid marker should work", func(t *testing.T) {
		t.Parallel()

		buff := bytes.Repeat([]byte{0x01}, 20)                                         // Eth address
		buff = append(buff, bytes.Repeat([]byte{0x01}, 32)...)                         // Mvx address
		buff = append(buff, []byte{0x00, 0x00, 0x00, 0x02}...)                         // token size
		buff = append(buff, []byte{0x02, 0x03}...)                                     // token
		buff = append(buff, []byte{0x00, 0x00, 0x00, 0x00}...)                         // amount size = 0 => amount = 0
		buff = append(buff, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}...) // nonce
		buff = append(buff, 0x03)                                                      // invalid marker

		completeCallData, err := codec.DecodeProxySCCompleteCallData(buff)
		assert.Nil(t, err)
		expectedCallData := ProxySCCompleteCallData{
			RawCallData: []byte{0x03},
			From:        common.HexToAddress("0x0101010101010101010101010101010101010101"),
			To:          data.NewAddressFromBytes(bytes.Repeat([]byte{0x01}, 32)),
			Token:       string([]byte{2, 3}),
			Amount:      big.NewInt(0),
			Nonce:       1,
		}
		assert.Equal(t, expectedCallData, completeCallData)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		//           |--------------FROM---------------------|---------------------TO----------------------------------------|-len-TK|------ETHUSDC-0ae8ee-------|-len-A-|20k|--tx-nonce=1---|-raw-call-data|
		hexedData := "880ec53af800b5cd051531672ef4fc4de233bd5d00000000000000000500871bc8f6332939a55a80b23012564523bea3291fa4370000000e455448555344432d306165386565000000024e200000000000000001414243"
		buff, err := hex.DecodeString(hexedData)
		require.Nil(t, err)

		expectedCompleteCallData := createTestProxySCCompleteCallData()
		completeCallData, err := codec.DecodeProxySCCompleteCallData(buff)
		assert.Equal(t, expectedCompleteCallData, completeCallData)
		assert.Nil(t, err)
	})
}

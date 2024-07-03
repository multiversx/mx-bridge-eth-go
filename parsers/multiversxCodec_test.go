package parsers

import (
	"bytes"
	"encoding/hex"
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/multiversx/mx-sdk-go/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testCallData = CallData{
	Type:     DataPresentProtocolMarker,
	Function: "abc",
	GasLimit: 500000000,
	Arguments: []string{
		strings.Repeat("A", 5),
		strings.Repeat("B", 50),
	},
}

func createTestProxySCCompleteCallData() ProxySCCompleteCallData {
	ethUnhexed, _ := hex.DecodeString("880ec53af800b5cd051531672ef4fc4de233bd5d")
	completeCallData := ProxySCCompleteCallData{
		CallData: CallData{
			Type:      DataPresentProtocolMarker,
			Function:  "",
			GasLimit:  50000000,
			Arguments: make([]string, 0),
		},
		From:   common.Address{},
		Token:  "ETHUSDC-0ae8ee",
		Amount: big.NewInt(20000),
		Nonce:  1,
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

func TestMultiversXCodec_EncodeDecodeCallData(t *testing.T) {
	t.Parallel()

	codec := &MultiversxCodec{}

	t.Run("with no parameters should work", func(t *testing.T) {
		t.Parallel()

		localCallData := testCallData // value copy
		localCallData.Arguments = make([]string, 0)

		buff := codec.EncodeCallData(localCallData)
		expectedBuff := []byte{0x01, 0x00, 0x00, 0x00, 0x03, 'a', 'b', 'c'}
		expectedBuff = append(expectedBuff, 0x00, 0x00, 0x00, 0x00, 0x1D, 0xCD, 0x65, 0x00) // Gas limit
		expectedBuff = append(expectedBuff, 0x00, 0x00, 0x00, 0x00)                         // numArguments
		assert.Equal(t, expectedBuff, buff)

		callData, err := codec.DecodeCallData(buff)
		require.Nil(t, err)
		assert.Equal(t, localCallData, callData)
	})
	t.Run("with parameters should work", func(t *testing.T) {
		t.Parallel()

		buff := codec.EncodeCallData(testCallData)
		expectedBuff := []byte{0x01, 0x00, 0x00, 0x00, 0x03, 'a', 'b', 'c'}
		expectedBuff = append(expectedBuff, 0x00, 0x00, 0x00, 0x00, 0x1D, 0xCD, 0x65, 0x00) // Gas limit
		expectedBuff = append(expectedBuff, 0x00, 0x00, 0x00, 0x02)                         // numArguments
		expectedBuff = append(expectedBuff, 0x00, 0x00, 0x00, 0x05)                         // Argument 0 length
		expectedBuff = append(expectedBuff, bytes.Repeat([]byte{'A'}, 5)...)                // Argument 0 data
		expectedBuff = append(expectedBuff, 0x00, 0x00, 0x00, 0x32)                         // Argument 1 length
		expectedBuff = append(expectedBuff, bytes.Repeat([]byte{'B'}, 50)...)               // Argument 1 data
		assert.Equal(t, expectedBuff, buff)

		callData, err := codec.DecodeCallData(buff)
		require.Nil(t, err)
		assert.Equal(t, testCallData, callData)
	})
}

func TestMultiversXCodec_EncodeDecodeProxySCCompleteCallData(t *testing.T) {
	t.Parallel()

	codec := &MultiversxCodec{}

	t.Run("with no parameters should work", func(t *testing.T) {
		t.Parallel()

		localCallData := createTestProxySCCompleteCallData()
		localCallData.Arguments = make([]string, 0)

		buff, err := codec.EncodeProxySCCompleteCallData(localCallData)
		require.Nil(t, err)

		callData, err := codec.DecodeProxySCCompleteCallData(buff)
		require.Nil(t, err)
		assert.Equal(t, localCallData, callData)
	})
	t.Run("with parameters should work", func(t *testing.T) {
		t.Parallel()

		localCallData := createTestProxySCCompleteCallData()
		buff, err := codec.EncodeProxySCCompleteCallData(localCallData)
		require.Nil(t, err)

		callData, err := codec.DecodeProxySCCompleteCallData(buff)
		require.Nil(t, err)
		assert.Equal(t, localCallData, callData)
	})
}

func TestMultiversxCodec_DecodeCallData(t *testing.T) {
	t.Parallel()

	codec := &MultiversxCodec{}
	emptyCallData := CallData{}

	t.Run("empty buffer should error", func(t *testing.T) {
		t.Parallel()

		callData, err := codec.DecodeCallData(nil)
		assert.Equal(t, errBufferTooShortForMarker, err)
		assert.Equal(t, emptyCallData, callData)

		callData, err = codec.DecodeCallData(make([]byte, 0))
		assert.Equal(t, errBufferTooShortForMarker, err)
		assert.Equal(t, emptyCallData, callData)
	})
	t.Run("unexpected marker should error", func(t *testing.T) {
		t.Parallel()

		callData, err := codec.DecodeCallData([]byte{0x03})
		assert.ErrorIs(t, err, errUnexpectedMarker)
		assert.Contains(t, err.Error(), ": 3")
		assert.Equal(t, emptyCallData, callData)
	})
	t.Run("buffer contains missing data marker should work", func(t *testing.T) {
		t.Parallel()

		buff := []byte{0x00}
		expectedCallData := CallData{
			Type: MissingDataProtocolMarker,
		}

		callData, err := codec.DecodeCallData(buff)
		assert.Nil(t, err)
		assert.Equal(t, expectedCallData, callData)
	})
	t.Run("buffer to short for function length should error", func(t *testing.T) {
		t.Parallel()

		buff := []byte{0x01}

		callData, err := codec.DecodeCallData(buff)
		assert.ErrorIs(t, err, errBufferTooShortForLength)
		assert.Contains(t, err.Error(), "for function")
		assert.Equal(t, emptyCallData, callData)
	})
	t.Run("buffer to short for function should error", func(t *testing.T) {
		t.Parallel()

		buff := []byte{0x01, 0x00, 0x00, 0x00, 0x05}

		callData, err := codec.DecodeCallData(buff)
		assert.ErrorIs(t, err, errBufferTooShortForString)
		assert.Contains(t, err.Error(), "for function")
		assert.Equal(t, emptyCallData, callData)
	})
	t.Run("buffer to short for gas limit should error", func(t *testing.T) {
		t.Parallel()

		buff := []byte{0x01, 0x00, 0x00, 0x00, 0x03, 'a', 'b', 'c'}
		buff = append(buff, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0) // malformed gas limit

		callData, err := codec.DecodeCallData(buff)
		assert.ErrorIs(t, err, errBufferTooShortForUint64)
		assert.Contains(t, err.Error(), "for gas limit")
		assert.Equal(t, emptyCallData, callData)
	})
	t.Run("buffer to short for num arguments should error", func(t *testing.T) {
		t.Parallel()

		buff := []byte{0x01, 0x00, 0x00, 0x00, 0x03, 'a', 'b', 'c'}
		buff = append(buff, 0x00, 0x00, 0x00, 0x00, 0x1D, 0xCD, 0x65, 0x00) // Gas limit
		buff = append(buff, 0x00, 0x00, 0x03)                               // Bad numArgument

		callData, err := codec.DecodeCallData(buff)
		assert.ErrorIs(t, err, errBufferTooShortForNumArgs)
		assert.Equal(t, emptyCallData, callData)
	})
	t.Run("buffer to short for argument length should error", func(t *testing.T) {
		t.Parallel()

		buff := []byte{0x01, 0x00, 0x00, 0x00, 0x03, 'a', 'b', 'c'}
		buff = append(buff, 0x00, 0x00, 0x00, 0x00, 0x1D, 0xCD, 0x65, 0x00) // Gas limit
		buff = append(buff, 0x00, 0x00, 0x00, 0x01)                         // numArguments
		buff = append(buff, 0x00, 0x00, 0x04)                               // Bad Argument 0 length

		callData, err := codec.DecodeCallData(buff)
		assert.ErrorIs(t, err, errBufferTooShortForLength)
		assert.Contains(t, err.Error(), "for argument 0")
		assert.Equal(t, emptyCallData, callData)
	})
	t.Run("buffer to short for argument data should error", func(t *testing.T) {
		t.Parallel()

		buff := []byte{0x01, 0x00, 0x00, 0x00, 0x03, 'a', 'b', 'c'}
		buff = append(buff, 0x00, 0x00, 0x00, 0x00, 0x1D, 0xCD, 0x65, 0x00) // Gas limit length
		buff = append(buff, 0x00, 0x00, 0x00, 0x01)                         // numArguments
		buff = append(buff, 0x00, 0x00, 0x00, 0x04)                         // Argument 0 length
		buff = append(buff, 0x00, 0x00, 0x04)                               // Bad Argument 0 data

		callData, err := codec.DecodeCallData(buff)
		assert.ErrorIs(t, err, errBufferTooShortForString)
		assert.Contains(t, err.Error(), "for argument 0")
		assert.Equal(t, emptyCallData, callData)
	})
}

func TestMultiversxCodec_EncodeProxySCCompleteCallData(t *testing.T) {
	t.Parallel()

	codec := MultiversxCodec{}

	t.Run("nil completeData.To should error", func(t *testing.T) {
		t.Parallel()

		completeCallData := ProxySCCompleteCallData{
			CallData: CallData{
				Type:      DataPresentProtocolMarker,
				Function:  "callPayable",
				GasLimit:  50000000,
				Arguments: make([]string, 0),
			},
			From:   common.Address{},
			Token:  "ETHUSDC-0ae8ee",
			Amount: big.NewInt(20000),
			Nonce:  1,
		}

		result, err := codec.EncodeProxySCCompleteCallData(completeCallData)
		assert.ErrorIs(t, err, errNilAddressHandler)
		assert.Contains(t, err.Error(), "for To field")
		assert.Nil(t, result)
	})
	t.Run("nil completeData.Amount should error", func(t *testing.T) {
		t.Parallel()

		completeCallData := ProxySCCompleteCallData{
			CallData: CallData{
				Type:      DataPresentProtocolMarker,
				Function:  "callPayable",
				GasLimit:  50000000,
				Arguments: make([]string, 0),
			},
			From:  common.Address{},
			Token: "ETHUSDC-0ae8ee",
			Nonce: 1,
		}
		completeCallData.To = data.NewAddressFromBytes(make([]byte, 0))

		result, err := codec.EncodeProxySCCompleteCallData(completeCallData)
		assert.ErrorIs(t, err, errNilAmount)
		assert.Nil(t, result)
	})
	t.Run("should work with function and no arguments", func(t *testing.T) {
		t.Parallel()

		//           |--------------FROM---------------------|---------------------TO----------------------------------------|-len-TK|------ETHUSDC-0ae8ee-------|-len-A-|20k|--tx-nonce=1---|M|-len-f-|--func-callPayable---|-gas-limit-50M-|-no-arg|
		hexedData := "880ec53af800b5cd051531672ef4fc4de233bd5d00000000000000000500871bc8f6332939a55a80b23012564523bea3291fa4370000000e455448555344432d306165386565000000024e200000000000000001010000000b63616c6c50617961626c650000000002faf08000000000"

		ethUnhexed, err := hex.DecodeString("880ec53af800b5cd051531672ef4fc4de233bd5d")
		require.Nil(t, err)
		completeCallData := ProxySCCompleteCallData{
			CallData: CallData{
				Type:      DataPresentProtocolMarker,
				Function:  "callPayable",
				GasLimit:  50000000,
				Arguments: make([]string, 0),
			},
			From:   common.Address{},
			Token:  "ETHUSDC-0ae8ee",
			Amount: big.NewInt(20000),
			Nonce:  1,
		}
		completeCallData.To, err = data.NewAddressFromBech32String("erd1qqqqqqqqqqqqqpgqsudu3a3n9yu62k5qkgcpy4j9ywl2x2gl5smsy7t4uv")
		require.Nil(t, err)
		completeCallData.From.SetBytes(ethUnhexed)

		buff, err := hex.DecodeString(hexedData)
		require.Nil(t, err)

		result, err := codec.EncodeProxySCCompleteCallData(completeCallData)
		assert.Nil(t, err)
		assert.Equal(t, buff, result)
	})
	t.Run("should work with function and with 2 arguments", func(t *testing.T) {
		t.Parallel()

		//           |--------------FROM---------------------|---------------------TO----------------------------------------|-len-TK|------ETHUSDC-0ae8ee-------|-len-A-|20k|--tx-nonce=1---|M|-len-f-|--func-callPayable---|-gas-limit-50M-|-no-arg|-arg0-l|-ABC-|-arg1-l|-DEFG--|
		hexedData := "880ec53af800b5cd051531672ef4fc4de233bd5d00000000000000000500871bc8f6332939a55a80b23012564523bea3291fa4370000000e455448555344432d306165386565000000024e200000000000000001010000000b63616c6c50617961626c650000000002faf08000000002000000034142430000000444454647"

		ethUnhexed, err := hex.DecodeString("880ec53af800b5cd051531672ef4fc4de233bd5d")
		require.Nil(t, err)
		completeCallData := ProxySCCompleteCallData{
			CallData: CallData{
				Type:     DataPresentProtocolMarker,
				Function: "callPayable",
				GasLimit: 50000000,
				Arguments: []string{
					"ABC",
					"DEFG",
				},
			},
			From:   common.Address{},
			Token:  "ETHUSDC-0ae8ee",
			Amount: big.NewInt(20000),
			Nonce:  1,
		}
		completeCallData.To, err = data.NewAddressFromBech32String("erd1qqqqqqqqqqqqqpgqsudu3a3n9yu62k5qkgcpy4j9ywl2x2gl5smsy7t4uv")
		require.Nil(t, err)
		completeCallData.From.SetBytes(ethUnhexed)

		buff, err := hex.DecodeString(hexedData)
		require.Nil(t, err)

		result, err := codec.EncodeProxySCCompleteCallData(completeCallData)
		assert.Nil(t, err)
		assert.Equal(t, buff, result)
	})
	t.Run("should work with no function and no arguments", func(t *testing.T) {
		t.Parallel()

		//           |--------------FROM---------------------|---------------------TO----------------------------------------|-len-TK|------ETHUSDC-0ae8ee-------|-len-A-|20k|--tx-nonce=1---|M|-len-f-|-gas-limit-50M-|-no-arg|
		hexedData := "880ec53af800b5cd051531672ef4fc4de233bd5d00000000000000000500871bc8f6332939a55a80b23012564523bea3291fa4370000000e455448555344432d306165386565000000024e20000000000000000101000000000000000002faf08000000000"
		completeCallData := createTestProxySCCompleteCallData()
		buff, err := hex.DecodeString(hexedData)
		require.Nil(t, err)

		result, err := codec.EncodeProxySCCompleteCallData(completeCallData)
		assert.Nil(t, err)
		assert.Equal(t, buff, result)
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
	t.Run("invalid token bytes should error", func(t *testing.T) {
		t.Parallel()

		buff := bytes.Repeat([]byte{0x01}, 20)                 // Eth address
		buff = append(buff, bytes.Repeat([]byte{0x01}, 32)...) // Mvx address
		buff = append(buff, []byte{0x00, 0x01, 0x04}...)       // invalid token

		completeCallData, err := codec.DecodeProxySCCompleteCallData(buff)
		assert.ErrorIs(t, err, errBufferTooShortForLength)
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
	t.Run("invalid nonce should error", func(t *testing.T) {
		t.Parallel()

		buff := bytes.Repeat([]byte{0x01}, 20)                                         // Eth address
		buff = append(buff, bytes.Repeat([]byte{0x01}, 32)...)                         // Mvx address
		buff = append(buff, []byte{0x00, 0x00, 0x00, 0x02}...)                         // token size
		buff = append(buff, []byte{0x02, 0x03}...)                                     // token
		buff = append(buff, []byte{0x00, 0x00, 0x00, 0x00}...)                         // amount size = 0 => amount = 0
		buff = append(buff, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}...) // nonce
		buff = append(buff, 0x03)                                                      // invalid marker

		completeCallData, err := codec.DecodeProxySCCompleteCallData(buff)
		assert.ErrorIs(t, err, errUnexpectedMarker)
		assert.Contains(t, err.Error(), ": 3")
		assert.Equal(t, emptyCompleteCallData, completeCallData)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		//           |--------------FROM---------------------|---------------------TO----------------------------------------|-len-TK|------ETHUSDC-0ae8ee-------|-len-A-|20k|--tx-nonce=1---|M|-len-f-|-gas-limit-50M-|-no-arg|
		hexedData := "880ec53af800b5cd051531672ef4fc4de233bd5d00000000000000000500871bc8f6332939a55a80b23012564523bea3291fa4370000000e455448555344432d306165386565000000024e20000000000000000101000000000000000002faf08000000000"
		buff, err := hex.DecodeString(hexedData)
		require.Nil(t, err)

		expectedCompleteCallData := createTestProxySCCompleteCallData()
		completeCallData, err := codec.DecodeProxySCCompleteCallData(buff)
		assert.Equal(t, expectedCompleteCallData, completeCallData)
		assert.Nil(t, err)
	})
}

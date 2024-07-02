package parsers

import (
	"bytes"
	"strings"
	"testing"

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

func TestMultiversXCodec_EncodeDecode(t *testing.T) {
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
		assert.ErrorIs(t, err, errBufferTooShortForGasLimit)
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

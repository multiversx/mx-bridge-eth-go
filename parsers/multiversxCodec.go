package parsers

import (
	"encoding/binary"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/multiversx/mx-bridge-eth-go/core"
	"github.com/multiversx/mx-sdk-go/data"
)

const lenEthAddress = 20
const lenMvxAddress = 32

// MultiversxCodec defines the codec operations to be used for MultiversX contracts
type MultiversxCodec struct {
}

func partiallyDecodeCallData(buff []byte, marker byte) (core.CallData, error) {
	buff, numChars, err := ExtractUint32(buff)
	if err != nil {
		return core.CallData{}, fmt.Errorf("%w for len of call data", err)
	}
	if numChars != len(buff) {
		return core.CallData{}, fmt.Errorf("%w: actual %d, declared %d", errBufferLenMismatch, len(buff), numChars)
	}

	buff, function, err := ExtractString(buff)
	if err != nil {
		return core.CallData{}, fmt.Errorf("%w for function", err)
	}

	_, gasLimit, err := ExtractUint64(buff)
	if err != nil {
		return core.CallData{}, fmt.Errorf("%w for gas limit", err)
	}

	return core.CallData{
		Type:     marker,
		Function: function,
		GasLimit: gasLimit,
	}, nil
}

// ExtractString will return the string value after extracting the length of the string from the buffer.
// The buffer returned will be trimmed out of the 4 bytes + the length of the string
func ExtractString(buff []byte) ([]byte, string, error) {
	// Ensure there's enough length for the 4 bytes for length
	if len(buff) < core.Uint32ArgBytes {
		return nil, "", errBufferTooShortForLength
	}
	argumentLength := int(binary.BigEndian.Uint32(buff[:core.Uint32ArgBytes]))
	buff = buff[core.Uint32ArgBytes:] // remove the len bytes

	// Check for the argument data
	if len(buff) < argumentLength {
		return nil, "", errBufferTooShortForString
	}
	endpointName := string(buff[:argumentLength])
	buff = buff[argumentLength:] // remove the string bytes

	return buff, endpointName, nil
}

func extractBigInt(buff []byte) ([]byte, *big.Int, error) {
	// Ensure there's enough length for the 4 bytes for length
	if len(buff) < core.Uint32ArgBytes {
		return nil, nil, errBufferTooShortForLength
	}
	argumentLength := int(binary.BigEndian.Uint32(buff[:core.Uint32ArgBytes]))
	buff = buff[core.Uint32ArgBytes:] // remove the len bytes

	// Check for the argument data
	if len(buff) < argumentLength {
		return nil, nil, errBufferTooShortForBigInt
	}

	value := big.NewInt(0).SetBytes(buff[:argumentLength])
	buff = buff[argumentLength:] // remove the value bytes

	return buff, value, nil
}

// ExtractUint64 will return the uint64 value after extracting 8 bytes from the buffer.
// The buffer returned will be trimmed out of the 8 bytes
func ExtractUint64(buff []byte) ([]byte, uint64, error) {
	// Ensure there's enough length for the 8 bytes
	if len(buff) < core.Uint64ArgBytes {
		return nil, 0, errBufferTooShortForUint64
	}

	value := binary.BigEndian.Uint64(buff[:core.Uint64ArgBytes])
	buff = buff[core.Uint64ArgBytes:]

	return buff, value, nil
}

// ExtractUint32 will return the int value after extracting 4 bytes from the buffer.
// The buffer returned will be trimmed out of the 4 bytes
func ExtractUint32(buff []byte) ([]byte, int, error) {
	// Ensure there's enough length for the 4 bytes
	if len(buff) < core.Uint32ArgBytes {
		return nil, 0, errBufferTooShortForUint32
	}
	value := int(binary.BigEndian.Uint32(buff[:core.Uint32ArgBytes]))
	buff = buff[core.Uint32ArgBytes:] // remove the len bytes

	return buff, value, nil
}

// DecodeProxySCCompleteCallData will try to decode the provided bytes into a ProxySCCompleteCallData struct
func (codec *MultiversxCodec) DecodeProxySCCompleteCallData(buff []byte) (core.ProxySCCompleteCallData, error) {
	result := core.ProxySCCompleteCallData{}

	if len(buff) < lenEthAddress {
		return core.ProxySCCompleteCallData{}, errBufferTooShortForEthAddress
	}
	result.From = common.Address{}
	result.From.SetBytes(buff[:lenEthAddress])
	buff = buff[lenEthAddress:]

	if len(buff) < lenMvxAddress {
		return core.ProxySCCompleteCallData{}, errBufferTooShortForMvxAddress
	}
	result.To = data.NewAddressFromBytes(buff[:lenMvxAddress])
	buff = buff[lenMvxAddress:]

	buff, token, err := ExtractString(buff)
	if err != nil {
		return core.ProxySCCompleteCallData{}, fmt.Errorf("%w for token", err)
	}
	result.Token = token

	buff, amount, err := extractBigInt(buff)
	if err != nil {
		return core.ProxySCCompleteCallData{}, fmt.Errorf("%w for amount", err)
	}
	result.Amount = amount

	buff, nonce, err := ExtractUint64(buff)
	if err != nil {
		return core.ProxySCCompleteCallData{}, fmt.Errorf("%w for nonce", err)
	}
	result.Nonce = nonce

	result.RawCallData = buff

	return result, nil
}

// DecodeCallData will try to decode the provided bytes into a CallData struct
func (codec *MultiversxCodec) DecodeCallData(buff []byte) (core.CallData, error) {
	if len(buff) == 0 {
		return core.CallData{}, errEmptyBuffer
	}

	marker := buff[0]
	buff = buff[1:]

	switch marker {
	case core.MissingDataProtocolMarker:
		return core.CallData{
			Type: core.MissingDataProtocolMarker,
		}, nil
	case core.DataPresentProtocolMarker:
		return decodeCallData(buff, marker)
	default:
		return core.CallData{}, fmt.Errorf("%w: %d", errUnexpectedMarker, marker)
	}
}

func decodeCallData(buff []byte, marker byte) (core.CallData, error) {
	buff, numChars, err := ExtractUint32(buff)
	if err != nil {
		return core.CallData{}, fmt.Errorf("%w when extracting complete buffer length", err)
	}
	if numChars != len(buff) {
		return core.CallData{}, fmt.Errorf("%w when checking the complete buffer length, expected %d, got %d", errBufferLenMismatch, numChars, len(buff))
	}

	buff, function, err := ExtractString(buff)
	if err != nil {
		return core.CallData{}, fmt.Errorf("%w when extracting the function", err)
	}

	buff, gasLimit, err := ExtractUint64(buff)
	if err != nil {
		return core.CallData{}, fmt.Errorf("%w when extracting the gas limit", err)
	}

	arguments, err := extractArguments(buff)
	if err != nil {
		return core.CallData{}, err
	}

	return core.CallData{
		Type:      marker,
		Function:  function,
		GasLimit:  gasLimit,
		Arguments: arguments,
	}, nil
}

func extractArguments(buff []byte) ([]string, error) {
	if len(buff) == 0 {
		return nil, fmt.Errorf("%w when parsing the arguments", errEmptyBuffer)
	}

	if len(buff) == 1 && buff[0] == core.MissingDataProtocolMarker {
		// no arguments provided
		return make([]string, 0), nil
	}

	buff = buff[1:]

	buff, numArgumentsLength, err := ExtractUint32(buff)
	if err != nil {
		return nil, fmt.Errorf("%w when extracting the number of arguments", err)
	}

	arguments := make([]string, 0)
	for i := 0; i < numArgumentsLength; i++ {
		var argument string
		buff, argument, err = ExtractString(buff)
		if err != nil {
			return nil, fmt.Errorf("%w for argument %d", err, i)
		}

		arguments = append(arguments, argument)
	}

	return arguments, nil
}

// EncodeCallDataWithLenAndMarker will provide a valid data byte slice with encoded call data parameters along with the length and marker
func (codec *MultiversxCodec) EncodeCallDataWithLenAndMarker(callData core.CallData) []byte {
	initialAlloc := 1024 * 1024 // 1MB initial buffer
	result := make([]byte, 0, initialAlloc)
	buff32Bits := make([]byte, 4)

	result = append(result, core.DataPresentProtocolMarker) // marker

	callDataBuff := codec.EncodeCallDataStrict(callData)
	binary.BigEndian.PutUint32(buff32Bits, uint32(len(callDataBuff)))

	result = append(result, buff32Bits...)
	result = append(result, callDataBuff...)

	return result
}

// EncodeCallDataStrict will encode just the provided call data. No length or marker will be added
func (codec *MultiversxCodec) EncodeCallDataStrict(callData core.CallData) []byte {
	initialAlloc := 1024 * 1024 // 1MB initial buffer
	result := make([]byte, 0, initialAlloc)

	buff32Bits := make([]byte, 4)
	buff64Bits := make([]byte, 8)

	funcLen := len(callData.Function)

	binary.BigEndian.PutUint32(buff32Bits, uint32(funcLen))
	result = append(result, buff32Bits...)        // append the function len
	result = append(result, callData.Function...) // append the function as string

	binary.BigEndian.PutUint64(buff64Bits, callData.GasLimit)
	result = append(result, buff64Bits...) // append the gas limit as 8 bytes

	if len(callData.Arguments) == 0 {
		// in case of no arguments, the contract requires that the missing data protocol marker should be provided, not
		// a 0 encoded on 4 bytes.
		result = append(result, core.MissingDataProtocolMarker)
		return result
	}

	result = append(result, core.DataPresentProtocolMarker)
	encodedArgs := codec.encodeArgs(callData.Arguments)
	result = append(result, encodedArgs...)

	return result
}

func (codec *MultiversxCodec) encodeArgs(args []string) []byte {
	buff32Bits := make([]byte, 4)

	initialAlloc := 1024 * 1024 // 1MB initial buffer
	result := make([]byte, 0, initialAlloc)

	binary.BigEndian.PutUint32(buff32Bits, uint32(len(args)))
	result = append(result, buff32Bits...) // append the number of arguments

	for _, arg := range args {
		lenArg := len(arg)
		binary.BigEndian.PutUint32(buff32Bits, uint32(lenArg))
		result = append(result, buff32Bits...) // append the length of the current argument
		result = append(result, arg...)        // append the argument as string
	}

	return result
}

// ExtractGasLimitFromRawCallData will try to extract the gas limit from the provided buffer
func (codec *MultiversxCodec) ExtractGasLimitFromRawCallData(buff []byte) (uint64, error) {
	if len(buff) == 0 {
		return 0, errBufferTooShortForMarker
	}

	marker := buff[0]
	buff = buff[1:]

	if marker != core.DataPresentProtocolMarker {
		return 0, fmt.Errorf("%w: %d", errUnexpectedMarker, marker)
	}

	callData, err := partiallyDecodeCallData(buff, marker)
	if err != nil {
		return 0, err
	}

	return callData.GasLimit, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (codec *MultiversxCodec) IsInterfaceNil() bool {
	return codec == nil
}

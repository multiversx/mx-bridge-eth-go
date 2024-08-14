package testsCommon

import (
	"encoding/binary"
	"fmt"

	bridgeCore "github.com/multiversx/mx-bridge-eth-go/core"
	"github.com/multiversx/mx-bridge-eth-go/parsers"
)

// TestMultiversXCodec is the codec helper used in testing
type TestMultiversXCodec struct {
}

// EncodeCallDataWithLenAndMarker will provide a valid data byte slice with encoded call data parameters along with the length and marker
func (codec *TestMultiversXCodec) EncodeCallDataWithLenAndMarker(callData parsers.CallData) []byte {
	initialAlloc := 1024 * 1024 // 1MB initial buffer
	result := make([]byte, 0, initialAlloc)
	buff32Bits := make([]byte, 4)

	result = append(result, bridgeCore.DataPresentProtocolMarker) // marker

	callDataBuff := codec.EncodeCallDataStrict(callData)
	binary.BigEndian.PutUint32(buff32Bits, uint32(len(callDataBuff)))

	result = append(result, buff32Bits...)
	result = append(result, callDataBuff...)

	return result
}

// EncodeCallDataStrict will encode just the provided call data. No length or marker will be added
func (codec *TestMultiversXCodec) EncodeCallDataStrict(callData parsers.CallData) []byte {
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
		result = append(result, bridgeCore.MissingDataProtocolMarker)
		return result
	}

	result = append(result, bridgeCore.DataPresentProtocolMarker)
	encodedArgs := codec.encodeArgs(callData.Arguments)
	result = append(result, encodedArgs...)

	return result
}

func (codec *TestMultiversXCodec) encodeArgs(args []string) []byte {
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

// DecodeCallData will try to decode the provided bytes into a CallData struct
func (codec *TestMultiversXCodec) DecodeCallData(buff []byte) parsers.CallData {
	if len(buff) == 0 {
		panic("empty buffer")
	}

	marker := buff[0]
	buff = buff[1:]

	switch marker {
	case bridgeCore.MissingDataProtocolMarker:
		return parsers.CallData{
			Type: bridgeCore.MissingDataProtocolMarker,
		}
	case bridgeCore.DataPresentProtocolMarker:
		return decodeCallData(buff, marker)
	default:
		panic(fmt.Sprintf("unexpected marker: %d", marker))
	}
}

func decodeCallData(buff []byte, marker byte) parsers.CallData {
	buff, numChars, err := parsers.ExtractUint32(buff)
	if err != nil {
		panic(err)
	}
	if numChars != len(buff) {
		panic("mismatch for len")
	}

	buff, function, err := parsers.ExtractString(buff)
	if err != nil {
		panic(err)
	}

	_, gasLimit, err := parsers.ExtractUint64(buff)
	if err != nil {
		panic(err)
	}

	arguments, err := extractArguments(buff)
	if err != nil {
		panic(err)
	}

	return parsers.CallData{
		Type:      marker,
		Function:  function,
		GasLimit:  gasLimit,
		Arguments: arguments,
	}
}

func extractArguments(buff []byte) ([]string, error) {
	if len(buff) == 0 {
		panic("empty buffer")
	}

	if len(buff) == 1 && buff[0] == bridgeCore.MissingDataProtocolMarker {
		// no arguments provided
		return make([]string, 0), nil
	}

	buff, numArgumentsLength, err := parsers.ExtractUint32(buff)
	if err != nil {
		panic(err)
	}

	arguments := make([]string, 0)
	for i := 0; i < numArgumentsLength; i++ {
		var argument string
		buff, argument, err = parsers.ExtractString(buff)
		if err != nil {
			return nil, fmt.Errorf("%w for argument %d", err, i)
		}

		arguments = append(arguments, argument)
	}

	return arguments, nil
}

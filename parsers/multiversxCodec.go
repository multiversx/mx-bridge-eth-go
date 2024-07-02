package parsers

import (
	"encoding/binary"
	"fmt"
)

// MultiversxCodec defines the codec operations to be used for MultiversX contracts
type MultiversxCodec struct {
}

// EncodeCallData will provide a valid data byte slice with encoded call data parameters
func (codec *MultiversxCodec) EncodeCallData(callData CallData) []byte {
	initialAlloc := 1024 * 1024 // 1MB initial buffer
	buff32Bits := make([]byte, 4)
	buff64Bits := make([]byte, 8)

	result := make([]byte, 0, initialAlloc)

	result = append(result, DataPresentProtocolMarker) // marker
	funcLen := len(callData.Function)

	binary.BigEndian.PutUint32(buff32Bits, uint32(funcLen))
	result = append(result, buff32Bits...)        // append the function len
	result = append(result, callData.Function...) // append the function as string

	binary.BigEndian.PutUint64(buff64Bits, callData.GasLimit)

	result = append(result, buff64Bits...) // append the gas limit as 8 bytes

	binary.BigEndian.PutUint32(buff32Bits, uint32(len(callData.Arguments)))
	result = append(result, buff32Bits...) // append the number of arguments

	for _, arg := range callData.Arguments {
		lenArg := len(arg)

		binary.BigEndian.PutUint32(buff32Bits, uint32(lenArg))
		result = append(result, buff32Bits...) // append the length of the current argument
		result = append(result, arg...)        // append the argument as string
	}

	return result
}

// DecodeCallData will try to decode the provided bytes into a CallData struct
func (codec *MultiversxCodec) DecodeCallData(buff []byte) (CallData, error) {
	if len(buff) == 0 {
		return CallData{}, errBufferTooShortForMarker
	}

	marker := buff[0]
	buff = buff[1:]

	switch marker {
	case MissingDataProtocolMarker:
		return CallData{
			Type: MissingDataProtocolMarker,
		}, nil
	case DataPresentProtocolMarker:
		return decodeCallData(buff, marker)
	default:
		return CallData{}, fmt.Errorf("%w: %d", errUnexpectedMarker, marker)
	}
}

func decodeCallData(buff []byte, marker byte) (CallData, error) {
	buff, function, err := extractString(buff)
	if err != nil {
		return CallData{}, fmt.Errorf("%w for function", err)
	}

	buff, gasLimit, err := extractGasLimit(buff)
	if err != nil {
		return CallData{}, err
	}

	buff, numArgumentsLength, err := extractArgumentsLen(buff)
	if err != nil {
		return CallData{}, err
	}

	arguments := make([]string, 0)
	for i := 0; i < numArgumentsLength; i++ {
		var argument string
		buff, argument, err = extractString(buff)
		if err != nil {
			return CallData{}, fmt.Errorf("%w for argument %d", err, i)
		}

		arguments = append(arguments, argument)
	}

	return CallData{
		Type:      marker,
		Function:  function,
		GasLimit:  gasLimit,
		Arguments: arguments,
	}, nil
}

func extractString(buff []byte) ([]byte, string, error) {
	// Ensure there's enough length for the 4 bytes for length
	if len(buff) < uint32ArgBytes {
		return nil, "", errBufferTooShortForLength
	}
	argumentLength := int(binary.BigEndian.Uint32(buff[:uint32ArgBytes]))
	buff = buff[uint32ArgBytes:] // remove the len bytes

	// Check for the argument data
	if len(buff) < argumentLength {
		return nil, "", errBufferTooShortForString
	}
	endpointName := string(buff[:argumentLength])
	buff = buff[argumentLength:] // remove the string bytes

	return buff, endpointName, nil
}

func extractGasLimit(buff []byte) ([]byte, uint64, error) {
	// Check for gas limit
	if len(buff) < uint64ArgBytes { // 8 bytes for gas limit
		return nil, 0, errBufferTooShortForGasLimit
	}

	gasLimit := binary.BigEndian.Uint64(buff[:uint64ArgBytes])
	buff = buff[uint64ArgBytes:]

	return buff, gasLimit, nil
}

func extractArgumentsLen(buff []byte) ([]byte, int, error) {
	// Ensure there's enough length for the 4 bytes for endpoint name length
	if len(buff) < uint32ArgBytes {
		return nil, 0, errBufferTooShortForNumArgs
	}
	length := int(binary.BigEndian.Uint32(buff[:uint32ArgBytes]))
	buff = buff[uint32ArgBytes:] // remove the len bytes

	return buff, length, nil
}

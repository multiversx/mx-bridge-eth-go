package parsers

import (
	"encoding/binary"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-sdk-go/data"
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

// EncodeProxySCCompleteCallData will provide a valid byte slice with the encoded parameters
func (codec *MultiversxCodec) EncodeProxySCCompleteCallData(completeData ProxySCCompleteCallData) ([]byte, error) {
	if check.IfNil(completeData.To) {
		return nil, fmt.Errorf("%w for To field", errNilAddressHandler)
	}
	if completeData.Amount == nil {
		return nil, errNilAmount
	}

	initialAlloc := 1024 * 1024 // 1MB initial buffer
	buff32Bits := make([]byte, 4)
	buff64Bits := make([]byte, 8)

	result := make([]byte, 0, initialAlloc)
	result = append(result, completeData.From.Bytes()...)      // append To
	result = append(result, completeData.To.AddressBytes()...) // append From

	binary.BigEndian.PutUint32(buff32Bits, uint32(len(completeData.Token)))
	result = append(result, buff32Bits...)         // append len(token)
	result = append(result, completeData.Token...) // append token

	amountBytes := big.NewInt(0).Set(completeData.Amount).Bytes()
	binary.BigEndian.PutUint32(buff32Bits, uint32(len(amountBytes)))
	result = append(result, buff32Bits...)  // append len(amount)
	result = append(result, amountBytes...) // append amount

	binary.BigEndian.PutUint64(buff64Bits, completeData.Nonce)
	result = append(result, buff64Bits...) // append nonce

	result = append(result, codec.EncodeCallData(completeData.CallData)...)

	return result, nil
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

	buff, gasLimit, err := extractUint64(buff)
	if err != nil {
		return CallData{}, fmt.Errorf("%w for gas limit", err)
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

func extractBigInt(buff []byte) ([]byte, *big.Int, error) {
	// Ensure there's enough length for the 4 bytes for length
	if len(buff) < uint32ArgBytes {
		return nil, nil, errBufferTooShortForLength
	}
	argumentLength := int(binary.BigEndian.Uint32(buff[:uint32ArgBytes]))
	buff = buff[uint32ArgBytes:] // remove the len bytes

	// Check for the argument data
	if len(buff) < argumentLength {
		return nil, nil, errBufferTooShortForBigInt
	}

	value := big.NewInt(0).SetBytes(buff[:argumentLength])
	buff = buff[argumentLength:] // remove the value bytes

	return buff, value, nil
}

func extractUint64(buff []byte) ([]byte, uint64, error) {
	if len(buff) < uint64ArgBytes { // 8 bytes for gas limit
		return nil, 0, errBufferTooShortForUint64
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

// DecodeProxySCCompleteCallData will try to decode the provided bytes into a ProxySCCompleteCallData struct
func (codec *MultiversxCodec) DecodeProxySCCompleteCallData(buff []byte) (ProxySCCompleteCallData, error) {
	result := ProxySCCompleteCallData{}

	if len(buff) < lenEthAddress {
		return ProxySCCompleteCallData{}, errBufferTooShortForEthAddress
	}
	result.From = common.Address{}
	result.From.SetBytes(buff[:lenEthAddress])
	buff = buff[lenEthAddress:]

	if len(buff) < lenMvxAddress {
		return ProxySCCompleteCallData{}, errBufferTooShortForMvxAddress
	}
	result.To = data.NewAddressFromBytes(buff[:lenMvxAddress])
	buff = buff[lenMvxAddress:]

	buff, token, err := extractString(buff)
	if err != nil {
		return ProxySCCompleteCallData{}, fmt.Errorf("%w for token", err)
	}
	result.Token = token

	buff, amount, err := extractBigInt(buff)
	if err != nil {
		return ProxySCCompleteCallData{}, fmt.Errorf("%w for amount", err)
	}
	result.Amount = amount

	buff, nonce, err := extractUint64(buff)
	if err != nil {
		return ProxySCCompleteCallData{}, fmt.Errorf("%w for nonce", err)
	}
	result.Nonce = nonce

	result.CallData, err = codec.DecodeCallData(buff)
	if err != nil {
		return ProxySCCompleteCallData{}, err
	}

	return result, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (codec *MultiversxCodec) IsInterfaceNil() bool {
	return codec == nil
}

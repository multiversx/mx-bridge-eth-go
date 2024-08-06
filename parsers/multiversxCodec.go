package parsers

import (
	"encoding/binary"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	bridgeCommon "github.com/multiversx/mx-bridge-eth-go/common"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-sdk-go/data"
)

// MultiversxCodec defines the codec operations to be used for MultiversX contracts
type MultiversxCodec struct {
}

// EncodeCallData will provide a valid data byte slice with encoded call data parameters
func (codec *MultiversxCodec) EncodeCallData(callData CallData) []byte {
	initialAlloc := 1024 * 1024 // 1MB initial buffer
	result := make([]byte, 0, initialAlloc)
	buff32Bits := make([]byte, 4)

	result = append(result, DataPresentProtocolMarker) // marker

	callDataBuff := codec.encodeInnerCallData(callData)
	binary.BigEndian.PutUint32(buff32Bits, uint32(len(callDataBuff)))

	result = append(result, buff32Bits...)
	result = append(result, callDataBuff...)

	return result
}

func (codec *MultiversxCodec) encodeInnerCallData(callData CallData) []byte {
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
		result = append(result, MissingDataProtocolMarker)
		return result
	}

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

	result := make([]byte, 0, initialAlloc)
	result = append(result, completeData.From.Bytes()...)
	result = append(result, completeData.To.AddressBytes()...)
	result = append(result, encodeToken([]byte(completeData.Token))...)
	result = append(result, encodeAmount(completeData.Amount)...)
	result = append(result, encodeNonce(completeData.Nonce)...)
	result = append(result, codec.EncodeCallData(completeData.CallData)...)

	return result, nil
}

func encodeToken(token []byte) []byte {
	result := make([]byte, 0, len(token)+4)
	buff32Bits := make([]byte, 4)

	binary.BigEndian.PutUint32(buff32Bits, uint32(len(token)))
	result = append(result, buff32Bits...) // append len(token)
	result = append(result, token...)      // append token

	return result
}

func encodeAmount(amount *big.Int) []byte {
	buff32Bits := make([]byte, 4)

	amountBytes := big.NewInt(0).Set(amount).Bytes()
	result := make([]byte, 0, len(amountBytes)+4)

	binary.BigEndian.PutUint32(buff32Bits, uint32(len(amountBytes)))
	result = append(result, buff32Bits...)  // append len(amount)
	result = append(result, amountBytes...) // append amount

	return result
}

func encodeNonce(nonce uint64) []byte {
	buff64Bits := make([]byte, 8)
	binary.BigEndian.PutUint64(buff64Bits, nonce)
	return buff64Bits
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
	buff, numChars, err := extractUint32(buff)
	if err != nil {
		return CallData{}, fmt.Errorf("%w for len of call data", err)
	}
	if numChars != len(buff) {
		return CallData{}, fmt.Errorf("%w: actual %d, declared %d", errBufferLenMismatch, len(buff), numChars)
	}

	buff, function, err := extractString(buff)
	if err != nil {
		return CallData{}, fmt.Errorf("%w for function", err)
	}

	buff, gasLimit, err := extractUint64(buff)
	if err != nil {
		return CallData{}, fmt.Errorf("%w for gas limit", err)
	}

	arguments, err := extractArguments(buff)
	if err != nil {
		return CallData{}, err
	}

	return CallData{
		Type:      marker,
		Function:  function,
		GasLimit:  gasLimit,
		Arguments: arguments,
	}, nil
}

func extractArguments(buff []byte) ([]string, error) {
	if len(buff) == 0 {
		return nil, errBufferTooShortForMarker
	}
	if len(buff) == 1 && buff[0] == MissingDataProtocolMarker {
		// no arguments provided
		return make([]string, 0), nil
	}

	buff, numArgumentsLength, err := extractUint32(buff)
	if err != nil {
		return nil, err
	}

	arguments := make([]string, 0)
	for i := 0; i < numArgumentsLength; i++ {
		var argument string
		buff, argument, err = extractString(buff)
		if err != nil {
			return nil, fmt.Errorf("%w for argument %d", err, i)
		}

		arguments = append(arguments, argument)
	}

	return arguments, nil
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
	// Ensure there's enough length for the 8 bytes
	if len(buff) < uint64ArgBytes {
		return nil, 0, errBufferTooShortForUint64
	}

	value := binary.BigEndian.Uint64(buff[:uint64ArgBytes])
	buff = buff[uint64ArgBytes:]

	return buff, value, nil
}

func extractUint32(buff []byte) ([]byte, int, error) {
	// Ensure there's enough length for the 4 bytes
	if len(buff) < uint32ArgBytes {
		return nil, 0, errBufferTooShortForUint32
	}
	value := int(binary.BigEndian.Uint32(buff[:uint32ArgBytes]))
	buff = buff[uint32ArgBytes:] // remove the len bytes

	return buff, value, nil
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

// EncodeDeposits will encode the provided deposits as a byte slice
func (codec *MultiversxCodec) EncodeDeposits(deposits []*bridgeCommon.DepositTransfer) ([]byte, error) {
	initialAlloc := 1024 * 1024 // 1MB initial buffer
	result := make([]byte, 0, initialAlloc)

	for _, dt := range deposits {
		buff, err := codec.encodeDeposit(dt)
		if err != nil {
			return nil, err
		}

		result = append(result, buff...)
	}

	return result, nil
}

// encodeDeposit will provide a valid byte slice with the encoded fields of a deposit transfer
func (codec *MultiversxCodec) encodeDeposit(deposit *bridgeCommon.DepositTransfer) ([]byte, error) {
	if deposit.Amount == nil {
		return nil, errNilAmount
	}

	initialAlloc := 1024 * 1024 // 1MB initial buffer
	result := make([]byte, 0, initialAlloc)

	result = append(result, deposit.FromBytes...)
	result = append(result, deposit.ToBytes...)
	result = append(result, encodeToken(deposit.DestinationTokenBytes)...)
	result = append(result, encodeAmount(deposit.Amount)...)
	result = append(result, encodeNonce(deposit.Nonce)...)
	result = append(result, deposit.Data...)

	return result, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (codec *MultiversxCodec) IsInterfaceNil() bool {
	return codec == nil
}

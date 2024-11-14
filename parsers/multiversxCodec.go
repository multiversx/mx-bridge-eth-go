package parsers

import (
	"encoding/binary"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	bridgeCore "github.com/multiversx/mx-bridge-eth-go/core"
	"github.com/multiversx/mx-sdk-go/data"
)

const lenEthAddress = 20
const lenMvxAddress = 32

// MultiversxCodec defines the codec operations to be used for MultiversX contracts
type MultiversxCodec struct {
}

func partiallyDecodeCallData(buff []byte, marker byte) (CallData, error) {
	buff, numChars, err := ExtractUint32(buff)
	if err != nil {
		return CallData{}, fmt.Errorf("%w for len of call data", err)
	}
	if numChars != len(buff) {
		return CallData{}, fmt.Errorf("%w: actual %d, declared %d", errBufferLenMismatch, len(buff), numChars)
	}

	buff, function, err := ExtractString(buff)
	if err != nil {
		return CallData{}, fmt.Errorf("%w for function", err)
	}

	_, gasLimit, err := ExtractUint64(buff)
	if err != nil {
		return CallData{}, fmt.Errorf("%w for gas limit", err)
	}

	return CallData{
		Type:     marker,
		Function: function,
		GasLimit: gasLimit,
	}, nil
}

// ExtractString will return the string value after extracting the length of the string from the buffer.
// The buffer returned will be trimmed out of the 4 bytes + the length of the string
func ExtractString(buff []byte) ([]byte, string, error) {
	// Ensure there's enough length for the 4 bytes for length
	if len(buff) < bridgeCore.Uint32ArgBytes {
		return nil, "", errBufferTooShortForLength
	}
	argumentLength := int(binary.BigEndian.Uint32(buff[:bridgeCore.Uint32ArgBytes]))
	buff = buff[bridgeCore.Uint32ArgBytes:] // remove the len bytes

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
	if len(buff) < bridgeCore.Uint32ArgBytes {
		return nil, nil, errBufferTooShortForLength
	}
	argumentLength := int(binary.BigEndian.Uint32(buff[:bridgeCore.Uint32ArgBytes]))
	buff = buff[bridgeCore.Uint32ArgBytes:] // remove the len bytes

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
	if len(buff) < bridgeCore.Uint64ArgBytes {
		return nil, 0, errBufferTooShortForUint64
	}

	value := binary.BigEndian.Uint64(buff[:bridgeCore.Uint64ArgBytes])
	buff = buff[bridgeCore.Uint64ArgBytes:]

	return buff, value, nil
}

// ExtractUint32 will return the int value after extracting 4 bytes from the buffer.
// The buffer returned will be trimmed out of the 4 bytes
func ExtractUint32(buff []byte) ([]byte, int, error) {
	// Ensure there's enough length for the 4 bytes
	if len(buff) < bridgeCore.Uint32ArgBytes {
		return nil, 0, errBufferTooShortForUint32
	}
	value := int(binary.BigEndian.Uint32(buff[:bridgeCore.Uint32ArgBytes]))
	buff = buff[bridgeCore.Uint32ArgBytes:] // remove the len bytes

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

	buff, token, err := ExtractString(buff)
	if err != nil {
		return ProxySCCompleteCallData{}, fmt.Errorf("%w for token", err)
	}
	result.Token = token

	buff, amount, err := extractBigInt(buff)
	if err != nil {
		return ProxySCCompleteCallData{}, fmt.Errorf("%w for amount", err)
	}
	result.Amount = amount

	buff, nonce, err := ExtractUint64(buff)
	if err != nil {
		return ProxySCCompleteCallData{}, fmt.Errorf("%w for nonce", err)
	}
	result.Nonce = nonce

	result.RawCallData = buff

	return result, nil
}

// ExtractGasLimitFromRawCallData will try to extract the gas limit from the provided buffer
func (codec *MultiversxCodec) ExtractGasLimitFromRawCallData(buff []byte) (uint64, error) {
	if len(buff) == 0 {
		return 0, errBufferTooShortForMarker
	}

	marker := buff[0]
	buff = buff[1:]

	if marker != bridgeCore.DataPresentProtocolMarker {
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

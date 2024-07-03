package testsCommon

import "github.com/multiversx/mx-bridge-eth-go/parsers"

// MultiversxCodecStub -
type MultiversxCodecStub struct {
	EncodeCallDataCalled                func(callData parsers.CallData) []byte
	EncodeProxySCCompleteCallDataCalled func(completeData parsers.ProxySCCompleteCallData) ([]byte, error)
	DecodeCallDataCalled                func(buff []byte) (parsers.CallData, error)
	DecodeProxySCCompleteCallDataCalled func(buff []byte) (parsers.ProxySCCompleteCallData, error)
}

// EncodeCallData -
func (stub *MultiversxCodecStub) EncodeCallData(callData parsers.CallData) []byte {
	if stub.EncodeCallDataCalled != nil {
		return stub.EncodeCallDataCalled(callData)
	}

	return make([]byte, 0)
}

// EncodeProxySCCompleteCallData -
func (stub *MultiversxCodecStub) EncodeProxySCCompleteCallData(completeData parsers.ProxySCCompleteCallData) ([]byte, error) {
	if stub.EncodeProxySCCompleteCallDataCalled != nil {
		return stub.EncodeProxySCCompleteCallDataCalled(completeData)
	}

	return make([]byte, 0), nil
}

// DecodeCallData -
func (stub *MultiversxCodecStub) DecodeCallData(buff []byte) (parsers.CallData, error) {
	if stub.DecodeCallDataCalled != nil {
		return stub.DecodeCallDataCalled(buff)
	}

	return parsers.CallData{}, nil
}

// DecodeProxySCCompleteCallData -
func (stub *MultiversxCodecStub) DecodeProxySCCompleteCallData(buff []byte) (parsers.ProxySCCompleteCallData, error) {
	if stub.DecodeProxySCCompleteCallDataCalled != nil {
		return stub.DecodeProxySCCompleteCallDataCalled(buff)
	}

	return parsers.ProxySCCompleteCallData{}, nil
}

// IsInterfaceNil -
func (stub *MultiversxCodecStub) IsInterfaceNil() bool {
	return stub == nil
}

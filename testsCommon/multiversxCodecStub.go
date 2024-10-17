package testsCommon

import "github.com/multiversx/mx-bridge-eth-go/parsers"

// MultiversxCodecStub -
type MultiversxCodecStub struct {
	DecodeProxySCCompleteCallDataCalled  func(buff []byte) (parsers.ProxySCCompleteCallData, error)
	ExtractGasLimitFromRawCallDataCalled func(buff []byte) (uint64, error)
}

// DecodeProxySCCompleteCallData -
func (stub *MultiversxCodecStub) DecodeProxySCCompleteCallData(buff []byte) (parsers.ProxySCCompleteCallData, error) {
	if stub.DecodeProxySCCompleteCallDataCalled != nil {
		return stub.DecodeProxySCCompleteCallDataCalled(buff)
	}

	return parsers.ProxySCCompleteCallData{}, nil
}

// ExtractGasLimitFromRawCallData -
func (stub *MultiversxCodecStub) ExtractGasLimitFromRawCallData(buff []byte) (uint64, error) {
	if stub.ExtractGasLimitFromRawCallDataCalled != nil {
		return stub.ExtractGasLimitFromRawCallDataCalled(buff)
	}

	return 0, nil
}

// IsInterfaceNil -
func (stub *MultiversxCodecStub) IsInterfaceNil() bool {
	return stub == nil
}

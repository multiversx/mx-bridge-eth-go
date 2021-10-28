package mock

// ChainClientStub -
type ChainClientStub struct {
	ExecuteVmQueryOnBridgeContractCalled func(function string, params ...[]byte) ([][]byte, error)
}

// ExecuteVmQueryOnBridgeContract -
func (cc *ChainClientStub) ExecuteVmQueryOnBridgeContract(function string, params ...[]byte) ([][]byte, error) {
	if cc.ExecuteVmQueryOnBridgeContractCalled != nil {
		return cc.ExecuteVmQueryOnBridgeContractCalled(function, params...)
	}

	return make([][]byte, 0), nil
}

// IsInterfaceNil -
func (cc *ChainClientStub) IsInterfaceNil() bool {
	return cc == nil
}

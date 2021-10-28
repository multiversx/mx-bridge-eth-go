package mock

// ChainInteractorStub -
type ChainInteractorStub struct {
	ExecuteVmQueryOnBridgeContractCalled func(function string, params ...[]byte) ([][]byte, error)
}

// ExecuteVmQueryOnBridgeContract -
func (cis *ChainInteractorStub) ExecuteVmQueryOnBridgeContract(function string, params ...[]byte) ([][]byte, error) {
	if cis.ExecuteVmQueryOnBridgeContractCalled != nil {
		return cis.ExecuteVmQueryOnBridgeContractCalled(function, params...)
	}

	return make([][]byte, 0), nil
}

// IsInterfaceNil -
func (cis *ChainInteractorStub) IsInterfaceNil() bool {
	return cis == nil
}

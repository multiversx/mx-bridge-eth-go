package interactors

// ElrondChainInteractorStub -
type ElrondChainInteractorStub struct {
	ExecuteVmQueryOnBridgeContractCalled func(function string, params ...[]byte) ([][]byte, error)
}

// ExecuteVmQueryOnBridgeContract -
func (stub *ElrondChainInteractorStub) ExecuteVmQueryOnBridgeContract(function string, params ...[]byte) ([][]byte, error) {
	if stub.ExecuteVmQueryOnBridgeContractCalled != nil {
		return stub.ExecuteVmQueryOnBridgeContractCalled(function, params...)
	}

	return nil, nil
}

// IsInterfaceNil -
func (stub *ElrondChainInteractorStub) IsInterfaceNil() bool {
	return stub == nil
}

package bridge

import "github.com/ethereum/go-ethereum/accounts/abi/bind"

// SafeContractWrapperStub -
type SafeContractWrapperStub struct {
	DepositsCountCalled func(opts *bind.CallOpts) (uint64, error)
	BatchesCountCalled  func(opts *bind.CallOpts) (uint64, error)
}

// DepositsCount -
func (stub *SafeContractWrapperStub) DepositsCount(opts *bind.CallOpts) (uint64, error) {
	if stub.DepositsCountCalled != nil {
		return stub.DepositsCountCalled(opts)
	}

	return 0, nil
}

// BatchesCount -
func (stub *SafeContractWrapperStub) BatchesCount(opts *bind.CallOpts) (uint64, error) {
	if stub.BatchesCountCalled != nil {
		return stub.BatchesCountCalled(opts)
	}

	return 0, nil
}

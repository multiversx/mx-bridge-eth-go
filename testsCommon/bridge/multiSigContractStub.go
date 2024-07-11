package bridge

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/multiversx/mx-bridge-eth-go/clients/ethereum/contract"
)

// MultiSigContractStub -
type MultiSigContractStub struct {
	GetBatchCalled         func(opts *bind.CallOpts, batchNonce *big.Int) (contract.Batch, bool, error)
	GetBatchDepositsCalled func(opts *bind.CallOpts, batchNonce *big.Int) ([]contract.Deposit, bool, error)
	GetRelayersCalled      func(opts *bind.CallOpts) ([]common.Address, error)
	WasBatchExecutedCalled func(opts *bind.CallOpts, batchNonce *big.Int) (bool, error)
	ExecuteTransferCalled  func(opts *bind.TransactOpts, tokens []common.Address, recipients []common.Address,
		amounts []*big.Int, nonces []*big.Int, batchNonce *big.Int, signatures [][]byte) (*types.Transaction, error)
	QuorumCalled                    func(opts *bind.CallOpts) (*big.Int, error)
	GetStatusesAfterExecutionCalled func(opts *bind.CallOpts, batchID *big.Int) ([]byte, bool, error)
	PausedCalled                    func(opts *bind.CallOpts) (bool, error)
}

// GetBatch -
func (stub *MultiSigContractStub) GetBatch(opts *bind.CallOpts, batchNonce *big.Int) (contract.Batch, bool, error) {
	if stub.GetBatchCalled != nil {
		return stub.GetBatchCalled(opts, batchNonce)
	}

	return contract.Batch{}, false, nil
}

// GetBatchDeposits -
func (stub *MultiSigContractStub) GetBatchDeposits(opts *bind.CallOpts, batchNonce *big.Int) ([]contract.Deposit, bool, error) {
	if stub.GetBatchCalled != nil {
		return stub.GetBatchDepositsCalled(opts, batchNonce)
	}

	return make([]contract.Deposit, 0), false, nil
}

// GetRelayers -
func (stub *MultiSigContractStub) GetRelayers(opts *bind.CallOpts) ([]common.Address, error) {
	if stub.GetRelayersCalled != nil {
		return stub.GetRelayersCalled(opts)
	}

	return make([]common.Address, 0), nil
}

// WasBatchExecuted -
func (stub *MultiSigContractStub) WasBatchExecuted(opts *bind.CallOpts, batchNonce *big.Int) (bool, error) {
	if stub.WasBatchExecutedCalled != nil {
		return stub.WasBatchExecutedCalled(opts, batchNonce)
	}

	return false, nil
}

// ExecuteTransfer -
func (stub *MultiSigContractStub) ExecuteTransfer(
	opts *bind.TransactOpts,
	tokens []common.Address,
	recipients []common.Address,
	amounts []*big.Int,
	nonces []*big.Int,
	batchNonce *big.Int,
	signatures [][]byte,
) (*types.Transaction, error) {
	if stub.ExecuteTransferCalled != nil {
		return stub.ExecuteTransferCalled(opts, tokens, recipients, amounts, nonces, batchNonce, signatures)
	}

	return nil, errNotImplemented
}

// Quorum -
func (stub *MultiSigContractStub) Quorum(opts *bind.CallOpts) (*big.Int, error) {
	if stub.QuorumCalled != nil {
		return stub.QuorumCalled(opts)
	}

	return big.NewInt(0), nil
}

// GetStatusesAfterExecution -
func (stub *MultiSigContractStub) GetStatusesAfterExecution(opts *bind.CallOpts, batchID *big.Int) ([]byte, bool, error) {
	if stub.GetStatusesAfterExecutionCalled != nil {
		return stub.GetStatusesAfterExecutionCalled(opts, batchID)
	}

	return make([]byte, 0), false, nil
}

// Paused -
func (stub *MultiSigContractStub) Paused(opts *bind.CallOpts) (bool, error) {
	if stub.PausedCalled != nil {
		return stub.PausedCalled(opts)
	}

	return false, nil
}

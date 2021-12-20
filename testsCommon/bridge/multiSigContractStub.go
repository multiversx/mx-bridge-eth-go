package bridge

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-eth-bridge/clients/ethereum/contract"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// MultiSigContractStub -
type MultiSigContractStub struct {
	GetBatchCalled         func(opts *bind.CallOpts, batchNonce *big.Int) (contract.Batch, error)
	GetRelayersCalled      func(opts *bind.CallOpts) ([]common.Address, error)
	WasBatchExecutedCalled func(opts *bind.CallOpts, batchNonce *big.Int) (bool, error)
	ExecuteTransferCalled  func(opts *bind.TransactOpts, tokens []common.Address, recipients []common.Address,
		amounts []*big.Int, batchNonce *big.Int, signatures [][]byte) (*types.Transaction, error)
	QuorumCalled                    func(opts *bind.CallOpts) (*big.Int, error)
	GetStatusesAfterExecutionCalled func(opts *bind.CallOpts, batchID *big.Int) ([]byte, error)
}

// GetBatch -
func (stub *MultiSigContractStub) GetBatch(opts *bind.CallOpts, batchNonce *big.Int) (contract.Batch, error) {
	if stub.GetBatchCalled != nil {
		return stub.GetBatchCalled(opts, batchNonce)
	}

	return contract.Batch{}, nil
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
	batchNonce *big.Int,
	signatures [][]byte,
) (*types.Transaction, error) {
	if stub.ExecuteTransferCalled != nil {
		return stub.ExecuteTransferCalled(opts, tokens, recipients, amounts, batchNonce, signatures)
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
func (stub *MultiSigContractStub) GetStatusesAfterExecution(opts *bind.CallOpts, batchID *big.Int) ([]byte, error) {
	if stub.GetStatusesAfterExecutionCalled != nil {
		return stub.GetStatusesAfterExecutionCalled(opts, batchID)
	}

	return make([]byte, 0), nil
}

package interactors

import (
	"fmt"
	"math/big"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridge/eth/contract"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

var errNotImplemented = fmt.Errorf("not implemented")

// BridgeContractStub -
type BridgeContractStub struct {
	GetNextPendingBatchCalled       func(opts *bind.CallOpts) (contract.Batch, error)
	FinishCurrentPendingBatchCalled func(opts *bind.TransactOpts, batchNonce *big.Int, newDepositStatuses []uint8, signatures [][]byte) (*types.Transaction, error)
	ExecuteTransferCalled           func(opts *bind.TransactOpts, tokens []common.Address, recipients []common.Address, amounts []*big.Int, batchNonce *big.Int, signatures [][]byte) (*types.Transaction, error)
	WasBatchExecutedCalled          func(opts *bind.CallOpts, batchNonce *big.Int) (bool, error)
	WasBatchFinishedCalled          func(opts *bind.CallOpts, batchNonce *big.Int) (bool, error)
	QuorumCalled                    func(opts *bind.CallOpts) (*big.Int, error)
	GetStatusesAfterExecutionCalled func(opts *bind.CallOpts, batchNonceElrondETH *big.Int) ([]uint8, error)
}

// GetNextPendingBatch -
func (bcs *BridgeContractStub) GetNextPendingBatch(opts *bind.CallOpts) (contract.Batch, error) {
	if bcs.GetNextPendingBatchCalled != nil {
		return bcs.GetNextPendingBatchCalled(opts)
	}

	return contract.Batch{}, nil
}

// FinishCurrentPendingBatch -
func (bcs *BridgeContractStub) FinishCurrentPendingBatch(opts *bind.TransactOpts, batchNonce *big.Int, newDepositStatuses []uint8, signatures [][]byte) (*types.Transaction, error) {
	if bcs.FinishCurrentPendingBatchCalled != nil {
		return bcs.FinishCurrentPendingBatchCalled(opts, batchNonce, newDepositStatuses, signatures)
	}

	return nil, errNotImplemented
}

// ExecuteTransfer -
func (bcs *BridgeContractStub) ExecuteTransfer(opts *bind.TransactOpts, tokens []common.Address, recipients []common.Address, amounts []*big.Int, batchNonce *big.Int, signatures [][]byte) (*types.Transaction, error) {
	if bcs.ExecuteTransferCalled != nil {
		return bcs.ExecuteTransferCalled(opts, tokens, recipients, amounts, batchNonce, signatures)
	}

	return nil, errNotImplemented
}

// WasBatchExecuted -
func (bcs *BridgeContractStub) WasBatchExecuted(opts *bind.CallOpts, batchNonce *big.Int) (bool, error) {
	if bcs.WasBatchExecutedCalled != nil {
		return bcs.WasBatchExecutedCalled(opts, batchNonce)
	}

	return true, nil
}

// WasBatchFinished -
func (bcs *BridgeContractStub) WasBatchFinished(opts *bind.CallOpts, batchNonce *big.Int) (bool, error) {
	if bcs.WasBatchFinishedCalled != nil {
		return bcs.WasBatchFinishedCalled(opts, batchNonce)
	}

	return true, nil
}

// Quorum -
func (bcs *BridgeContractStub) Quorum(opts *bind.CallOpts) (*big.Int, error) {
	if bcs.QuorumCalled != nil {
		return bcs.QuorumCalled(opts)
	}

	return big.NewInt(0), nil
}

// GetStatusesAfterExecution -
func (bcs *BridgeContractStub) GetStatusesAfterExecution(opts *bind.CallOpts, batchNonceElrondETH *big.Int) ([]uint8, error) {
	if bcs.GetStatusesAfterExecutionCalled != nil {
		return bcs.GetStatusesAfterExecutionCalled(opts, batchNonceElrondETH)
	}

	return make([]byte, 0), nil
}

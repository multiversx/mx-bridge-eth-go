package mock

import (
	"context"
	"math/big"
	"sync"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridge/eth/contract"
	"github.com/ElrondNetwork/elrond-eth-bridge/integrationTests"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// EthereumProposedStatus -
type EthereumProposedStatus struct {
	BatchNonce         *big.Int
	NewDepositStatuses []uint8
	Signatures         [][]byte
}

// EthereumProposedTransfer -
type EthereumProposedTransfer struct {
	BatchNonce *big.Int
	Tokens     []common.Address
	Recipients []common.Address
	Amounts    []*big.Int
	Signatures [][]byte
}

// EthereumChainMock -
type EthereumChainMock struct {
	mutState                         sync.RWMutex
	nonces                           map[common.Address]uint64
	pendingBatch                     contract.Batch
	proposedStatus                   *EthereumProposedStatus
	proposedTransfer                 *EthereumProposedTransfer
	GetStatusesAfterExecutionHandler func() []byte
	ProcessFinishedHandler           func()
	quorum                           int
	relayers                         []common.Address
}

// NewEthereumChainMock -
func NewEthereumChainMock() *EthereumChainMock {
	mock := &EthereumChainMock{
		nonces: make(map[common.Address]uint64),
	}
	mock.Clean()

	return mock
}

// Clean -
func (mock *EthereumChainMock) Clean() {
	mock.mutState.Lock()
	mock.pendingBatch = contract.Batch{
		Nonce: big.NewInt(0),
	}
	mock.proposedStatus = nil
	mock.proposedTransfer = nil
	mock.mutState.Unlock()
}

// BlockNumber -
func (mock *EthereumChainMock) BlockNumber(_ context.Context) (uint64, error) {
	return 0, nil
}

// NonceAt -
func (mock *EthereumChainMock) NonceAt(_ context.Context, account common.Address, _ *big.Int) (uint64, error) {
	mock.mutState.RLock()
	defer mock.mutState.RUnlock()

	return mock.nonces[account], nil
}

// ChainID -
func (mock *EthereumChainMock) ChainID(_ context.Context) (*big.Int, error) {
	return big.NewInt(1), nil
}

// GetNextPendingBatch -
func (mock *EthereumChainMock) GetNextPendingBatch(_ *bind.CallOpts) (contract.Batch, error) {
	mock.mutState.RLock()
	defer mock.mutState.RUnlock()

	return mock.pendingBatch, nil
}

// SetPendingBatch -
func (mock *EthereumChainMock) SetPendingBatch(batch contract.Batch) {
	mock.mutState.Lock()
	mock.pendingBatch = batch
	mock.mutState.Unlock()
}

// FinishCurrentPendingBatch -
func (mock *EthereumChainMock) FinishCurrentPendingBatch(_ *bind.TransactOpts, batchNonce *big.Int, newDepositStatuses []uint8, signatures [][]byte) (*types.Transaction, error) {
	status := &EthereumProposedStatus{
		BatchNonce:         batchNonce,
		NewDepositStatuses: newDepositStatuses,
		Signatures:         signatures,
	}

	mockDataField, err := integrationTests.TestMarshalizer.Marshal(status)
	if err != nil {
		panic(err)
	}

	txData := &types.LegacyTx{
		Nonce: 0,
		Data:  mockDataField,
	}
	tx := types.NewTx(txData)

	mock.mutState.Lock()
	mock.proposedStatus = status
	mock.pendingBatch = contract.Batch{
		Nonce: big.NewInt(0),
	}
	mock.mutState.Unlock()

	integrationTests.Log.Info("process finished, set status was written")
	if mock.ProcessFinishedHandler != nil {
		mock.ProcessFinishedHandler()
	}

	return tx, nil
}

// GetLastProposedStatus -
func (mock *EthereumChainMock) GetLastProposedStatus() *EthereumProposedStatus {
	mock.mutState.RLock()
	defer mock.mutState.RUnlock()

	return mock.proposedStatus
}

// ExecuteTransfer -
func (mock *EthereumChainMock) ExecuteTransfer(_ *bind.TransactOpts, tokens []common.Address, recipients []common.Address, amounts []*big.Int, batchNonce *big.Int, signatures [][]byte) (*types.Transaction, error) {
	tokensLength := len(tokens)
	recipientsLength := len(recipients)
	amountsLength := len(amounts)
	if tokensLength != recipientsLength {
		panic("tokens length & recipients length mismatch")
	}
	if recipientsLength != amountsLength {
		panic("recipients length & amounts length mismatch")
	}

	proposedTransfer := &EthereumProposedTransfer{
		BatchNonce: batchNonce,
		Tokens:     tokens,
		Recipients: recipients,
		Amounts:    amounts,
		Signatures: signatures,
	}

	mockDataField, err := integrationTests.TestMarshalizer.Marshal(proposedTransfer)
	if err != nil {
		panic(err)
	}

	txData := &types.LegacyTx{
		Nonce: 0,
		Data:  mockDataField,
	}
	tx := types.NewTx(txData)

	mock.mutState.Lock()
	mock.proposedTransfer = proposedTransfer
	mock.mutState.Unlock()

	return tx, nil
}

// GetLastProposedTransfer -
func (mock *EthereumChainMock) GetLastProposedTransfer() *EthereumProposedTransfer {
	mock.mutState.RLock()
	defer mock.mutState.RUnlock()

	return mock.proposedTransfer
}

// WasBatchExecuted -
func (mock *EthereumChainMock) WasBatchExecuted(_ *bind.CallOpts, batchNonce *big.Int) (bool, error) {
	mock.mutState.RLock()
	defer mock.mutState.RUnlock()

	if mock.proposedTransfer == nil {
		return false, nil
	}

	return batchNonce.Cmp(mock.proposedTransfer.BatchNonce) == 0, nil
}

// WasBatchFinished -
func (mock *EthereumChainMock) WasBatchFinished(_ *bind.CallOpts, batchNonce *big.Int) (bool, error) {
	mock.mutState.RLock()
	defer mock.mutState.RUnlock()

	if mock.proposedStatus == nil {
		return false, nil
	}

	return batchNonce.Cmp(mock.proposedStatus.BatchNonce) == 0, nil
}

// Quorum -
func (mock *EthereumChainMock) Quorum(_ *bind.CallOpts) (*big.Int, error) {
	mock.mutState.RLock()
	defer mock.mutState.RUnlock()

	return big.NewInt(int64(mock.quorum)), nil
}

// SetQuorum -
func (mock *EthereumChainMock) SetQuorum(quorum int) {
	mock.mutState.Lock()
	defer mock.mutState.Unlock()

	mock.quorum = quorum
}

// GetStatusesAfterExecution -
func (mock *EthereumChainMock) GetStatusesAfterExecution(_ *bind.CallOpts, _ *big.Int) ([]uint8, error) {
	return mock.GetStatusesAfterExecutionHandler(), nil
}

// GetRelayers -
func (mock *EthereumChainMock) GetRelayers(_ *bind.CallOpts) ([]common.Address, error) {
	mock.mutState.RLock()
	defer mock.mutState.RUnlock()

	return mock.relayers, nil
}

// AddRelayer -
func (mock *EthereumChainMock) AddRelayer(relayer common.Address) {
	mock.mutState.Lock()
	defer mock.mutState.Unlock()

	mock.relayers = append(mock.relayers, relayer)
}

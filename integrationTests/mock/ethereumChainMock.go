package mock

import (
	"context"
	"fmt"
	"math/big"
	"sync"

	"github.com/ElrondNetwork/elrond-eth-bridge/clients/ethereum/contract"
	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ElrondNetwork/elrond-eth-bridge/integrationTests"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// EthereumProposedTransfer -
type EthereumProposedTransfer struct {
	BatchNonce *big.Int
	Tokens     []common.Address
	Recipients []common.Address
	Amounts    []*big.Int
	Nonces     []*big.Int
	Signatures [][]byte
}

// EthereumChainMock -
type EthereumChainMock struct {
	mutState                         sync.RWMutex
	nonces                           map[common.Address]uint64
	batches                          map[uint64]*contract.Batch
	proposedTransfer                 *EthereumProposedTransfer
	GetStatusesAfterExecutionHandler func() []byte
	ProcessFinishedHandler           func()
	quorum                           int
	relayers                         []common.Address

	ProposeMultiTransferEsdtBatchCalled func()
}

// NewEthereumChainMock -
func NewEthereumChainMock() *EthereumChainMock {
	return &EthereumChainMock{
		nonces:  make(map[common.Address]uint64),
		batches: make(map[uint64]*contract.Batch),
	}
}

// SetIntMetric -
func (mock *EthereumChainMock) SetIntMetric(_ string, _ int) {}

// AddIntMetric -
func (mock *EthereumChainMock) AddIntMetric(_ string, _ int) {}

// SetStringMetric -
func (mock *EthereumChainMock) SetStringMetric(_ string, _ string) {}

// GetAllMetrics -
func (mock *EthereumChainMock) GetAllMetrics() core.GeneralMetrics {
	return make(core.GeneralMetrics)
}

// Name -
func (mock *EthereumChainMock) Name() string {
	return ""
}

// GetBatch -
func (mock *EthereumChainMock) GetBatch(_ context.Context, batchNonce *big.Int) (contract.Batch, error) {
	mock.mutState.RLock()
	defer mock.mutState.RUnlock()

	batch, found := mock.batches[batchNonce.Uint64()]
	if !found {
		return contract.Batch{}, fmt.Errorf("batch %d not found", batchNonce)
	}

	return *batch, nil
}

// GetRelayers -
func (mock *EthereumChainMock) GetRelayers(_ context.Context) ([]common.Address, error) {
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

// WasBatchExecuted -
func (mock *EthereumChainMock) WasBatchExecuted(_ context.Context, batchNonce *big.Int) (bool, error) {
	mock.mutState.RLock()
	defer mock.mutState.RUnlock()

	if mock.proposedTransfer == nil {
		return false, nil
	}

	return batchNonce.Cmp(mock.proposedTransfer.BatchNonce) == 0, nil
}

// Clean -
func (mock *EthereumChainMock) Clean() {
	mock.mutState.Lock()
	mock.batches = make(map[uint64]*contract.Batch)
	mock.proposedTransfer = nil
	mock.mutState.Unlock()
}

// ChainID -
func (mock *EthereumChainMock) ChainID(_ context.Context) (*big.Int, error) {
	return big.NewInt(1), nil
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

// AddBatch -
func (mock *EthereumChainMock) AddBatch(batch contract.Batch) {
	mock.mutState.Lock()
	mock.batches[batch.Nonce.Uint64()] = &batch
	mock.mutState.Unlock()
}

// AddDepositToBatch -
func (mock *EthereumChainMock) AddDepositToBatch(nonce uint64, deposit contract.Deposit) {
	mock.mutState.Lock()
	defer mock.mutState.Unlock()

	batch, found := mock.batches[nonce]
	if !found {
		panic(fmt.Sprintf("programming error in tests: no batch found for nonce %d", nonce))
	}

	batch.Deposits = append(batch.Deposits, deposit)
	mock.mutState.Unlock()
}

// ExecuteTransfer -
func (mock *EthereumChainMock) ExecuteTransfer(_ *bind.TransactOpts, tokens []common.Address, recipients []common.Address, amounts []*big.Int, nonces []*big.Int, batchNonce *big.Int, signatures [][]byte) (*types.Transaction, error) {
	tokensLength := len(tokens)
	recipientsLength := len(recipients)
	amountsLength := len(amounts)
	noncesLength := len(nonces)
	if tokensLength != recipientsLength {
		panic("tokens length & recipients length mismatch")
	}
	if recipientsLength != amountsLength {
		panic("recipients length & amounts length mismatch")
	}
	if tokensLength != noncesLength {
		panic("tokens length & nonces length mismatch")
	}

	proposedTransfer := &EthereumProposedTransfer{
		BatchNonce: batchNonce,
		Tokens:     tokens,
		Recipients: recipients,
		Amounts:    amounts,
		Nonces:     nonces,
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

	if mock.ProposeMultiTransferEsdtBatchCalled != nil {
		mock.ProposeMultiTransferEsdtBatchCalled()
	}

	return tx, nil
}

// Quorum -
func (mock *EthereumChainMock) Quorum(_ context.Context) (*big.Int, error) {
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
func (mock *EthereumChainMock) GetStatusesAfterExecution(_ context.Context, _ *big.Int) ([]byte, error) {
	return mock.GetStatusesAfterExecutionHandler(), nil
}

// IsInterfaceNil -
func (mock *EthereumChainMock) IsInterfaceNil() bool {
	return mock == nil
}

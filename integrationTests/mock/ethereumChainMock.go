package mock

import (
	"context"
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/multiversx/mx-bridge-eth-go/clients/ethereum/contract"
	"github.com/multiversx/mx-bridge-eth-go/core"
	"github.com/multiversx/mx-bridge-eth-go/integrationTests"
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
	mutState                            sync.RWMutex
	nonces                              map[common.Address]uint64
	batches                             map[uint64]*contract.Batch
	deposits                            map[uint64][]contract.Deposit
	proposedTransfer                    *EthereumProposedTransfer
	totalBalances                       map[common.Address]*big.Int
	mintBalances                        map[common.Address]*big.Int
	burnBalances                        map[common.Address]*big.Int
	mintBurnTokens                      map[common.Address]bool
	nativeTokens                        map[common.Address]bool
	whitelistedTokens                   map[common.Address]bool
	nativeErc20TokensBalances           map[common.Address]*big.Int
	GetStatusesAfterExecutionHandler    func() []byte
	ProcessFinishedHandler              func()
	quorum                              int
	relayers                            []common.Address
	ProposeMultiTransferEsdtBatchCalled func()
	BalanceAtCalled                     func(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error)
	FilterLogsCalled                    func(ctx context.Context, q ethereum.FilterQuery) ([]types.Log, error)
}

// NewEthereumChainMock -
func NewEthereumChainMock() *EthereumChainMock {
	return &EthereumChainMock{
		nonces:                    make(map[common.Address]uint64),
		batches:                   make(map[uint64]*contract.Batch),
		deposits:                  make(map[uint64][]contract.Deposit),
		totalBalances:             make(map[common.Address]*big.Int),
		mintBalances:              make(map[common.Address]*big.Int),
		burnBalances:              make(map[common.Address]*big.Int),
		mintBurnTokens:            make(map[common.Address]bool),
		nativeTokens:              make(map[common.Address]bool),
		whitelistedTokens:         make(map[common.Address]bool),
		nativeErc20TokensBalances: make(map[common.Address]*big.Int),
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

// GetBatchDeposits -
func (mock *EthereumChainMock) GetBatchDeposits(_ context.Context, batchNonce *big.Int) ([]contract.Deposit, error) {
	mock.mutState.RLock()
	defer mock.mutState.RUnlock()

	deposits, found := mock.deposits[batchNonce.Uint64()]
	if !found {
		return make([]contract.Deposit, 0), fmt.Errorf("deposits for batch %d not found", batchNonce)
	}

	return deposits, nil
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

	mock.deposits[nonce] = append(mock.deposits[nonce], deposit)
	batch.DepositsCount++
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

// GetLastProposedTransfer -
func (mock *EthereumChainMock) GetLastProposedTransfer() *EthereumProposedTransfer {
	mock.mutState.RLock()
	defer mock.mutState.RUnlock()

	return mock.proposedTransfer
}

// BalanceAt -
func (mock *EthereumChainMock) BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error) {
	if mock.BalanceAtCalled != nil {
		return mock.BalanceAtCalled(ctx, account, blockNumber)
	}
	return big.NewInt(0), nil
}

// FilterLogs -
func (mock *EthereumChainMock) FilterLogs(ctx context.Context, q ethereum.FilterQuery) ([]types.Log, error) {
	if mock.FilterLogsCalled != nil {
		return mock.FilterLogsCalled(ctx, q)
	}

	return []types.Log{}, nil
}

// IsPaused -
func (mock *EthereumChainMock) IsPaused(_ context.Context) (bool, error) {
	return false, nil
}

// TotalBalances -
func (mock *EthereumChainMock) TotalBalances(_ context.Context, account common.Address) (*big.Int, error) {
	mock.mutState.RLock()
	defer mock.mutState.RUnlock()

	return mock.totalBalances[account], nil
}

// UpdateTotalBalances -
func (mock *EthereumChainMock) UpdateTotalBalances(account common.Address, value *big.Int) {
	mock.mutState.Lock()
	mock.totalBalances[account] = value
	mock.mutState.Unlock()
}

// MintBalances -
func (mock *EthereumChainMock) MintBalances(_ context.Context, account common.Address) (*big.Int, error) {
	mock.mutState.RLock()
	defer mock.mutState.RUnlock()

	return mock.mintBalances[account], nil
}

// UpdateMintBalances -
func (mock *EthereumChainMock) UpdateMintBalances(account common.Address, value *big.Int) {
	mock.mutState.Lock()
	mock.mintBalances[account] = value
	mock.mutState.Unlock()
}

// BurnBalances -
func (mock *EthereumChainMock) BurnBalances(_ context.Context, account common.Address) (*big.Int, error) {
	mock.mutState.RLock()
	defer mock.mutState.RUnlock()

	return mock.burnBalances[account], nil
}

// UpdateBurnBalances -
func (mock *EthereumChainMock) UpdateBurnBalances(account common.Address, value *big.Int) {
	mock.mutState.Lock()
	mock.burnBalances[account] = value
	mock.mutState.Unlock()
}

// MintBurnTokens -
func (mock *EthereumChainMock) MintBurnTokens(_ context.Context, account common.Address) (bool, error) {
	mock.mutState.RLock()
	defer mock.mutState.RUnlock()

	return mock.mintBurnTokens[account], nil
}

// UpdateMintBurnTokens -
func (mock *EthereumChainMock) UpdateMintBurnTokens(account common.Address, value bool) {
	mock.mutState.Lock()
	mock.mintBurnTokens[account] = value
	mock.mutState.Unlock()
}

// NativeTokens -
func (mock *EthereumChainMock) NativeTokens(_ context.Context, account common.Address) (bool, error) {
	mock.mutState.RLock()
	defer mock.mutState.RUnlock()

	return mock.nativeTokens[account], nil
}

// UpdateNativeTokens -
func (mock *EthereumChainMock) UpdateNativeTokens(account common.Address, value bool) {
	mock.mutState.Lock()
	mock.nativeTokens[account] = value
	mock.mutState.Unlock()
}

// WhitelistedTokens -
func (mock *EthereumChainMock) WhitelistedTokens(_ context.Context, account common.Address) (bool, error) {
	mock.mutState.RLock()
	defer mock.mutState.RUnlock()

	return mock.whitelistedTokens[account], nil
}

// UpdateWhitelistedTokens -
func (mock *EthereumChainMock) UpdateWhitelistedTokens(account common.Address, value bool) {
	mock.mutState.Lock()
	mock.whitelistedTokens[account] = value
	mock.mutState.Unlock()
}

// IsInterfaceNil -
func (mock *EthereumChainMock) IsInterfaceNil() bool {
	return mock == nil
}

// TokenMintedBalances -
func (mock *EthereumChainMock) TokenMintedBalances(_ context.Context, erc20Token common.Address) (*big.Int, error) {
	mock.mutState.RLock()
	defer mock.mutState.RUnlock()

	value := mock.nativeErc20TokensBalances[erc20Token]

	return value, nil
}

// WhitelistedTokensMintBurn -
func (mock *EthereumChainMock) WhitelistedTokensMintBurn(_ context.Context, erc20Token common.Address) (bool, error) {
	mock.mutState.RLock()
	defer mock.mutState.RUnlock()

	_, found := mock.nativeErc20TokensBalances[erc20Token]

	return found, nil
}

// AddWhitelistedTokensMintBurn -
func (mock *EthereumChainMock) AddWhitelistedTokensMintBurn(erc20Token common.Address, nativeBalance *big.Int) {
	mock.mutState.Lock()
	mock.nativeErc20TokensBalances[erc20Token] = nativeBalance
	mock.mutState.Unlock()
}

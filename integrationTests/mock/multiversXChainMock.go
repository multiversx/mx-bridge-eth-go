package mock

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/multiversx/mx-bridge-eth-go/integrationTests"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	logger "github.com/multiversx/mx-chain-logger-go"
	sdkCore "github.com/multiversx/mx-sdk-go/core"
	"github.com/multiversx/mx-sdk-go/data"
)

var log = logger.GetOrCreate("integrationTests/mock")

// MultiversXChainMock -
type MultiversXChainMock struct {
	*multiversXContractStateMock
	mutState         sync.RWMutex
	sentTransactions map[string]*transaction.FrontendTransaction
	accounts         *multiversXAccountsMock
}

// NewMultiversXChainMock -
func NewMultiversXChainMock() *MultiversXChainMock {
	return &MultiversXChainMock{
		multiversXContractStateMock: newMultiversXContractStateMock(),
		sentTransactions:            make(map[string]*transaction.FrontendTransaction),
		accounts:                    newMultiversXAccountsMock(),
	}
}

// GetNetworkConfig -
func (mock *MultiversXChainMock) GetNetworkConfig(_ context.Context) (*data.NetworkConfig, error) {
	return &data.NetworkConfig{
		ChainID:                  "t",
		LatestTagSoftwareVersion: "",
		MinGasPrice:              1000000000,
		MinTransactionVersion:    1,
	}, nil
}

// GetNetworkStatus -
func (mock *MultiversXChainMock) GetNetworkStatus(_ context.Context, _ uint32) (*data.NetworkStatus, error) {
	return &data.NetworkStatus{}, nil
}

// GetShardOfAddress -
func (mock *MultiversXChainMock) GetShardOfAddress(_ context.Context, _ string) (uint32, error) {
	return 0, nil
}

// SendTransaction -
func (mock *MultiversXChainMock) SendTransaction(_ context.Context, transaction *transaction.FrontendTransaction) (string, error) {
	if transaction == nil {
		panic("nil transaction")
	}

	addrAsBech32 := transaction.Sender
	addressHandler, err := data.NewAddressFromBech32String(addrAsBech32)
	if err != nil {
		panic(fmt.Sprintf("%v while creating address handler for string %s", err, addrAsBech32))
	}

	hash, err := core.CalculateHash(integrationTests.TestMarshalizer, integrationTests.TestHasher, transaction)
	if err != nil {
		panic(err)
	}

	log.Info("sent MultiversX transaction", "sender", addrAsBech32, "data", string(transaction.Data))

	mock.mutState.Lock()
	defer mock.mutState.Unlock()

	mock.sentTransactions[string(hash)] = transaction
	mock.accounts.updateNonce(addressHandler, transaction.Nonce)

	mock.processTransaction(transaction)

	return hex.EncodeToString(hash), nil
}

// SendTransactions -
func (mock *MultiversXChainMock) SendTransactions(ctx context.Context, txs []*transaction.FrontendTransaction) ([]string, error) {
	hashes := make([]string, 0, len(txs))
	for _, tx := range txs {
		hash, _ := mock.SendTransaction(ctx, tx)
		hashes = append(hashes, hash)
	}

	return hashes, nil
}

// GetAllSentTransactions -
func (mock *MultiversXChainMock) GetAllSentTransactions(_ context.Context) map[string]*transaction.FrontendTransaction {
	mock.mutState.RLock()
	defer mock.mutState.RUnlock()

	txs := make(map[string]*transaction.FrontendTransaction)
	for hash, tx := range mock.sentTransactions {
		txs[hash] = tx
	}

	return txs
}

// ExecuteVMQuery -
func (mock *MultiversXChainMock) ExecuteVMQuery(_ context.Context, vmRequest *data.VmValueRequest) (*data.VmValuesResponseData, error) {
	mock.mutState.Lock()
	defer mock.mutState.Unlock()

	return mock.processVmRequests(vmRequest)
}

// GetAccount -
func (mock *MultiversXChainMock) GetAccount(_ context.Context, address sdkCore.AddressHandler) (*data.Account, error) {
	mock.mutState.Lock()
	defer mock.mutState.Unlock()

	return mock.accounts.getOrCreate(address), nil
}

// AddRelayer -
func (mock *MultiversXChainMock) AddRelayer(address sdkCore.AddressHandler) {
	mock.mutState.Lock()
	defer mock.mutState.Unlock()

	mock.relayers = append(mock.relayers, address.AddressBytes())
}

// SetLastExecutedEthBatchID -
func (mock *MultiversXChainMock) SetLastExecutedEthBatchID(lastExecutedEthBatchId uint64) {
	mock.mutState.Lock()
	defer mock.mutState.Unlock()

	mock.lastExecutedEthBatchId = lastExecutedEthBatchId
}

// SetLastExecutedEthTxId -
func (mock *MultiversXChainMock) SetLastExecutedEthTxId(lastExecutedEthTxId uint64) {
	mock.mutState.Lock()
	defer mock.mutState.Unlock()

	mock.lastExecutedEthTxId = lastExecutedEthTxId
}

// AddTokensPair -
func (mock *MultiversXChainMock) AddTokensPair(erc20 common.Address, ticker string) {
	mock.mutState.Lock()
	defer mock.mutState.Unlock()

	mock.addTokensPair(erc20, ticker)
}

// SetQuorum -
func (mock *MultiversXChainMock) SetQuorum(quorum int) {
	mock.mutState.Lock()
	defer mock.mutState.Unlock()

	mock.quorum = quorum
}

// PerformedActionID returns the performed action ID
func (mock *MultiversXChainMock) PerformedActionID() *big.Int {
	mock.mutState.RLock()
	defer mock.mutState.RUnlock()

	return mock.performedAction
}

// ProposedTransfer returns the proposed transfer that matches the performed action ID
func (mock *MultiversXChainMock) ProposedTransfer() *multiversXProposedTransfer {
	mock.mutState.RLock()
	defer mock.mutState.RUnlock()

	if mock.performedAction == nil {
		return nil
	}

	for hash, transfer := range mock.proposedTransfers {
		if HashToActionID(hash).String() == mock.performedAction.String() {
			return transfer
		}
	}

	return nil
}

// SetPendingBatch -
func (mock *MultiversXChainMock) SetPendingBatch(pendingBatch *MultiversXPendingBatch) {
	mock.mutState.Lock()
	mock.setPendingBatch(pendingBatch)
	mock.mutState.Unlock()
}

// AddDepositToCurrentBatch -
func (mock *MultiversXChainMock) AddDepositToCurrentBatch(deposit MultiversXDeposit) {
	mock.mutState.Lock()
	mock.pendingBatch.MultiversXDeposits = append(mock.pendingBatch.MultiversXDeposits, deposit)
	mock.mutState.Unlock()
}

// IsInterfaceNil -
func (mock *MultiversXChainMock) IsInterfaceNil() bool {
	return mock == nil
}

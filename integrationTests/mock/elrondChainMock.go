package mock

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"sync"

	"github.com/ElrondNetwork/elrond-eth-bridge/integrationTests"
	"github.com/ElrondNetwork/elrond-go-core/core"
	erdgoCore "github.com/ElrondNetwork/elrond-sdk-erdgo/core"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/data"
	"github.com/ethereum/go-ethereum/common"
)

// ElrondChainMock -
type ElrondChainMock struct {
	*elrondContractStateMock
	mutState         sync.RWMutex
	sentTransactions map[string]*data.Transaction
	accounts         *elrondAccountsMock
}

// NewElrondChainMock -
func NewElrondChainMock() *ElrondChainMock {
	return &ElrondChainMock{
		elrondContractStateMock: newElrondContractStateMock(),
		sentTransactions:        make(map[string]*data.Transaction),
		accounts:                newElrondAccountsMock(),
	}
}

// GetNetworkConfig -
func (mock *ElrondChainMock) GetNetworkConfig(_ context.Context) (*data.NetworkConfig, error) {
	return &data.NetworkConfig{
		ChainID:                  "t",
		LatestTagSoftwareVersion: "",
		MinGasPrice:              1000000000,
		MinTransactionVersion:    1,
	}, nil
}

// SendTransaction -
func (mock *ElrondChainMock) SendTransaction(_ context.Context, transaction *data.Transaction) (string, error) {
	if transaction == nil {
		panic("nil transaction")
	}

	addrAsBech32 := transaction.SndAddr
	addressHandler, err := data.NewAddressFromBech32String(addrAsBech32)
	if err != nil {
		panic(fmt.Sprintf("%v while creating address handler for string %s", err, addrAsBech32))
	}

	hash, err := core.CalculateHash(integrationTests.TestMarshalizer, integrationTests.TestHasher, transaction)
	if err != nil {
		panic(err)
	}

	mock.mutState.Lock()
	defer mock.mutState.Unlock()

	mock.sentTransactions[string(hash)] = transaction
	mock.accounts.updateNonce(addressHandler, transaction.Nonce)

	mock.processTransaction(transaction)

	return hex.EncodeToString(hash), nil
}

// SendTransactions -
func (mock *ElrondChainMock) SendTransactions(ctx context.Context, txs []*data.Transaction) ([]string, error) {
	hashes := make([]string, 0, len(txs))
	for _, tx := range txs {
		hash, _ := mock.SendTransaction(ctx, tx)
		hashes = append(hashes, hash)
	}

	return hashes, nil
}

// GetAllSentTransactions -
func (mock *ElrondChainMock) GetAllSentTransactions(_ context.Context) map[string]*data.Transaction {
	mock.mutState.RLock()
	defer mock.mutState.RUnlock()

	txs := make(map[string]*data.Transaction)
	for hash, tx := range mock.sentTransactions {
		txs[hash] = tx
	}

	return txs
}

// ExecuteVMQuery -
func (mock *ElrondChainMock) ExecuteVMQuery(_ context.Context, vmRequest *data.VmValueRequest) (*data.VmValuesResponseData, error) {
	mock.mutState.Lock()
	defer mock.mutState.Unlock()

	return mock.processVmRequests(vmRequest)
}

// GetAccount -
func (mock *ElrondChainMock) GetAccount(_ context.Context, address erdgoCore.AddressHandler) (*data.Account, error) {
	mock.mutState.Lock()
	defer mock.mutState.Unlock()

	return mock.accounts.getOrCreate(address), nil
}

// AddRelayer -
func (mock *ElrondChainMock) AddRelayer(address erdgoCore.AddressHandler) {
	mock.mutState.Lock()
	defer mock.mutState.Unlock()

	mock.relayers = append(mock.relayers, address.AddressBytes())
}

// SetLastExecutedEthBatchID -
func (mock *ElrondChainMock) SetLastExecutedEthBatchID(lastExecutedEthBatchId uint64) {
	mock.mutState.Lock()
	defer mock.mutState.Unlock()

	mock.lastExecutedEthBatchId = lastExecutedEthBatchId
}

// SetLastExecutedEthTxId -
func (mock *ElrondChainMock) SetLastExecutedEthTxId(lastExecutedEthTxId uint64) {
	mock.mutState.Lock()
	defer mock.mutState.Unlock()

	mock.lastExecutedEthTxId = lastExecutedEthTxId
}

// AddTokensPair -
func (mock *ElrondChainMock) AddTokensPair(erc20 common.Address, ticker string) {
	mock.mutState.Lock()
	defer mock.mutState.Unlock()

	mock.addTokensPair(erc20, ticker)
}

// SetQuorum -
func (mock *ElrondChainMock) SetQuorum(quorum int) {
	mock.mutState.Lock()
	defer mock.mutState.Unlock()

	mock.quorum = quorum
}

// PerformedActionID returns the performed action ID
func (mock *ElrondChainMock) PerformedActionID() *big.Int {
	mock.mutState.RLock()
	defer mock.mutState.RUnlock()

	return mock.performedAction
}

// ProposedTransfer returns the proposed transfer that matches the performed action ID
func (mock *ElrondChainMock) ProposedTransfer() *ElrondProposedTransfer {
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
func (mock *ElrondChainMock) SetPendingBatch(pendingBatch *ElrondPendingBatch) {
	mock.mutState.Lock()
	mock.setPendingBatch(pendingBatch)
	mock.mutState.Unlock()
}

// AddDepositToCurrentBatch -
func (mock *ElrondChainMock) AddDepositToCurrentBatch(deposit ElrondDeposit) {
	mock.mutState.Lock()
	mock.pendingBatch.ElrondDeposits = append(mock.pendingBatch.ElrondDeposits, deposit)
	mock.mutState.Unlock()
}

// IsInterfaceNil -
func (mock *ElrondChainMock) IsInterfaceNil() bool {
	return mock == nil
}

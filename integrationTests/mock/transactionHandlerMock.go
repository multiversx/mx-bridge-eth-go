package mock

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/pubkeyConverter"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/hashing/blake2b"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	apiTransaction "github.com/ElrondNetwork/elrond-go/api/transaction"
)

type transactionHandlerMock struct {
	addressConverter core.PubkeyConverter
	marshalizer      marshal.Marshalizer
	hasher           hashing.Hasher
	mutTransactions  sync.RWMutex
	transactions     map[string]*apiTransaction.SendTxRequest
}

func newTransactionHandlerMock() *transactionHandlerMock {
	thm := &transactionHandlerMock{
		marshalizer:  &marshal.GogoProtoMarshalizer{},
		hasher:       blake2b.NewBlake2b(),
		transactions: make(map[string]*apiTransaction.SendTxRequest),
	}
	thm.addressConverter, _ = pubkeyConverter.NewBech32PubkeyConverter(32, log)

	return thm
}

func (thm *transactionHandlerMock) processSendTransaction(rw http.ResponseWriter, req *http.Request) {
	bodyBytes := getBodyAsByteSlice(req)
	sendTxRequest := &apiTransaction.SendTxRequest{}

	err := json.Unmarshal(bodyBytes, sendTxRequest)
	if err != nil {
		writeResponse(rw, http.StatusInternalServerError, "", nil, fmt.Errorf("ElrondMockClient: %w", err))
		return
	}

	_, txHash, err := thm.createTransaction(sendTxRequest)
	if err != nil {
		writeResponse(rw, http.StatusInternalServerError, "", nil, fmt.Errorf("ElrondMockClient: %w", err))
		return
	}

	thm.mutTransactions.Lock()
	thm.transactions[string(txHash)] = sendTxRequest
	thm.mutTransactions.Unlock()

	writeResponse(rw, http.StatusOK, "txHash", hex.EncodeToString(txHash), nil)
}

// createTransaction will return a transaction from all the required fields
// used to create a correct hash, as any running chain will output.
func (thm *transactionHandlerMock) createTransaction(txRequest *apiTransaction.SendTxRequest) (*transaction.Transaction, []byte, error) {
	log.Debug("createTransaction", "nonce", txRequest.Nonce, "value", txRequest.Value, "receiver", txRequest.Receiver,
		"receiverUsername", txRequest.ReceiverUsername, "sender", txRequest.Sender, "senderUsername", txRequest.SenderUsername,
		"gasPrice", txRequest.GasPrice, "gasLimit", txRequest.GasLimit, "dataField", string(txRequest.Data),
		"sig", txRequest.Signature, "chainID", txRequest.ChainID, "version", txRequest.Version, "options", txRequest.Options)

	receiverAddress, err := thm.addressConverter.Decode(txRequest.Receiver)
	if err != nil {
		return nil, nil, errors.New("could not create receiver address from provided param")
	}

	senderAddress, err := thm.addressConverter.Decode(txRequest.Sender)
	if err != nil {
		return nil, nil, errors.New("could not create sender address from provided param")
	}

	signatureBytes, err := hex.DecodeString(txRequest.Signature)
	if err != nil {
		return nil, nil, errors.New("could not fetch signature bytes")
	}

	valAsBigInt, ok := big.NewInt(0).SetString(txRequest.Value, 10)
	if !ok {
		return nil, nil, errors.New("invalid value")
	}

	tx := &transaction.Transaction{
		Nonce:       txRequest.Nonce,
		Value:       valAsBigInt,
		RcvAddr:     receiverAddress,
		RcvUserName: txRequest.ReceiverUsername,
		SndAddr:     senderAddress,
		SndUserName: txRequest.SenderUsername,
		GasPrice:    txRequest.GasPrice,
		GasLimit:    txRequest.GasLimit,
		Data:        txRequest.Data,
		Signature:   signatureBytes,
		ChainID:     []byte(txRequest.ChainID),
		Version:     txRequest.Version,
		Options:     txRequest.Options,
	}

	var txHash []byte
	txHash, err = core.CalculateHash(thm.marshalizer, thm.hasher, tx)
	if err != nil {
		return nil, nil, err
	}

	return tx, txHash, nil
}

// GetTransaction will get a stored transaction based on hash
func (thm *transactionHandlerMock) GetTransaction(hash string) *apiTransaction.SendTxRequest {
	thm.mutTransactions.RLock()
	defer thm.mutTransactions.RUnlock()

	return thm.transactions[hash]
}

// GetAllTransactions will return a new map containing all transactions
func (thm *transactionHandlerMock) GetAllTransactions() map[string]*apiTransaction.SendTxRequest {
	thm.mutTransactions.RLock()
	defer thm.mutTransactions.RUnlock()

	m := make(map[string]*apiTransaction.SendTxRequest)
	for hash, tx := range thm.transactions {
		m[hash] = tx
	}

	return m
}

// CleanReceivedTransactions will clean any stored transactions
func (thm *transactionHandlerMock) CleanReceivedTransactions() {
	thm.mutTransactions.Lock()
	thm.transactions = make(map[string]*apiTransaction.SendTxRequest)
	thm.mutTransactions.Unlock()
}

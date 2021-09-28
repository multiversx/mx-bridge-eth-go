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

	_, txHash, err := thm.createTransaction(sendTxRequest.Nonce, sendTxRequest.Value, sendTxRequest.Receiver,
		sendTxRequest.ReceiverUsername, sendTxRequest.Sender, sendTxRequest.SenderUsername, sendTxRequest.GasPrice,
		sendTxRequest.GasLimit, sendTxRequest.Data, sendTxRequest.Signature, sendTxRequest.ChainID, sendTxRequest.Version,
		sendTxRequest.Options)
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
func (thm *transactionHandlerMock) createTransaction(
	nonce uint64,
	value string,
	receiver string,
	receiverUsername []byte,
	sender string,
	senderUsername []byte,
	gasPrice uint64,
	gasLimit uint64,
	dataField []byte,
	signatureHex string,
	chainID string,
	version uint32,
	options uint32,
) (*transaction.Transaction, []byte, error) {
	log.Debug("createTransaction", "nonce", nonce, "value", value, "receiver", receiver,
		"receiverUsername", receiverUsername, "sender", sender, "senderUsername", senderUsername, "gasPrice", gasPrice,
		"gasLimit", gasLimit, "dataField", string(dataField), "sig", signatureHex, "chainID", chainID, "version", version,
		"options", options)

	receiverAddress, err := thm.addressConverter.Decode(receiver)
	if err != nil {
		return nil, nil, errors.New("could not create receiver address from provided param")
	}

	senderAddress, err := thm.addressConverter.Decode(sender)
	if err != nil {
		return nil, nil, errors.New("could not create sender address from provided param")
	}

	signatureBytes, err := hex.DecodeString(signatureHex)
	if err != nil {
		return nil, nil, errors.New("could not fetch signature bytes")
	}

	valAsBigInt, ok := big.NewInt(0).SetString(value, 10)
	if !ok {
		return nil, nil, errors.New("invalid value")
	}

	tx := &transaction.Transaction{
		Nonce:       nonce,
		Value:       valAsBigInt,
		RcvAddr:     receiverAddress,
		RcvUserName: receiverUsername,
		SndAddr:     senderAddress,
		SndUserName: senderUsername,
		GasPrice:    gasPrice,
		GasLimit:    gasLimit,
		Data:        dataField,
		Signature:   signatureBytes,
		ChainID:     []byte(chainID),
		Version:     version,
		Options:     options,
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

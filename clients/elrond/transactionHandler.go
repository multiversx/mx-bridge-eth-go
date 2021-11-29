package elrond

import (
	"context"
	"encoding/hex"
	"encoding/json"

	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/builders"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/core"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/data"
)

type transactionHandler struct {
	proxy                   ElrondProxy
	relayerAddress          core.AddressHandler
	multisigAddressAsBech32 string
	nonceTxHandler          NonceTransactionsHandler
	relayerPrivateKey       crypto.PrivateKey
	singleSigner            crypto.SingleSigner
}

// sendTransactionReturningHash will try to assemble a transaction, sign it, send it and, if everything is OK, returns the transaction's hash
func (txHandler *transactionHandler) sendTransactionReturningHash(ctx context.Context, builder builders.TxDataBuilder, cost uint64) (string, error) {
	tx, err := txHandler.signTransaction(ctx, builder, cost)
	if err != nil {
		return "", err
	}

	return txHandler.nonceTxHandler.SendTransaction(context.Background(), tx)
}

func (txHandler *transactionHandler) signTransaction(ctx context.Context, builder builders.TxDataBuilder, cost uint64) (*data.Transaction, error) {
	// TODO use a smart cache here (request the configs only from time to time)
	networkConfig, err := txHandler.proxy.GetNetworkConfig(ctx)
	if err != nil {
		return nil, err
	}

	nonce, err := txHandler.nonceTxHandler.GetNonce(context.Background(), txHandler.relayerAddress)
	if err != nil {
		return nil, err
	}

	dataBytes, err := builder.ToDataBytes()
	if err != nil {
		return nil, err
	}

	tx := &data.Transaction{
		ChainID:  networkConfig.ChainID,
		Version:  networkConfig.MinTransactionVersion,
		GasLimit: cost,
		GasPrice: networkConfig.MinGasPrice,
		Nonce:    nonce,
		Data:     dataBytes,
		SndAddr:  txHandler.relayerAddress.AddressAsBech32String(),
		RcvAddr:  txHandler.multisigAddressAsBech32,
		Value:    "0",
	}

	err = txHandler.signTransactionWithPrivateKey(tx)
	if err != nil {
		return nil, err
	}

	return tx, nil
}

// signTransactionWithPrivateKey signs a transaction with the client's private key
func (txHandler *transactionHandler) signTransactionWithPrivateKey(tx *data.Transaction) error {
	tx.Signature = ""
	bytes, err := json.Marshal(&tx)
	if err != nil {
		return err
	}

	signature, err := txHandler.singleSigner.Sign(txHandler.relayerPrivateKey, bytes)
	if err != nil {
		return err
	}

	tx.Signature = hex.EncodeToString(signature)

	return nil
}

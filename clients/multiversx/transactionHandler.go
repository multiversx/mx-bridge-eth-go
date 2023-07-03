package multiversx

import (
	"context"
	"encoding/hex"
	"encoding/json"

	"github.com/multiversx/mx-chain-core-go/data/transaction"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-sdk-go/builders"
	"github.com/multiversx/mx-sdk-go/core"
)

type transactionHandler struct {
	proxy                   Proxy
	relayerAddress          core.AddressHandler
	multisigAddressAsBech32 string
	nonceTxHandler          NonceTransactionsHandler
	relayerPrivateKey       crypto.PrivateKey
	singleSigner            crypto.SingleSigner
	roleProvider            roleProvider
}

// SendTransactionReturnHash will try to assemble a transaction, sign it, send it and, if everything is OK, returns the transaction's hash
func (txHandler *transactionHandler) SendTransactionReturnHash(ctx context.Context, builder builders.TxDataBuilder, gasLimit uint64) (string, error) {
	if !txHandler.roleProvider.IsWhitelisted(txHandler.relayerAddress) {
		return "", errRelayerNotWhitelisted
	}
	tx, err := txHandler.signTransaction(ctx, builder, gasLimit)
	if err != nil {
		return "", err
	}

	return txHandler.nonceTxHandler.SendTransaction(context.Background(), tx)
}

func (txHandler *transactionHandler) signTransaction(ctx context.Context, builder builders.TxDataBuilder, gasLimit uint64) (*transaction.FrontendTransaction, error) {
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

	tx := &transaction.FrontendTransaction{
		ChainID:  networkConfig.ChainID,
		Version:  networkConfig.MinTransactionVersion,
		GasLimit: gasLimit,
		GasPrice: networkConfig.MinGasPrice,
		Nonce:    nonce,
		Data:     dataBytes,
		Sender:   txHandler.relayerAddress.AddressAsBech32String(),
		Receiver: txHandler.multisigAddressAsBech32,
		Value:    "0",
	}

	err = txHandler.signTransactionWithPrivateKey(tx)
	if err != nil {
		return nil, err
	}

	return tx, nil
}

// signTransactionWithPrivateKey signs a transaction with the client's private key
func (txHandler *transactionHandler) signTransactionWithPrivateKey(tx *transaction.FrontendTransaction) error {
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

// Close will close any sub-components it uses
func (txHandler *transactionHandler) Close() error {
	return txHandler.nonceTxHandler.Close()
}

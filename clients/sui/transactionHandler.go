package sui

import (
	"context"
	"crypto/ed25519"
	"github.com/block-vision/sui-go-sdk/models"
)

type transactionHandler struct {
	client                SuiClient
	relayerAddress        string
	relayerPrivateKey     ed25519.PrivateKey
	bridgeContractAddress string
}

func (txHandler *transactionHandler) SendTransaction(ctx context.Context, moveCallRequest models.MoveCallRequest, gasLimit uint64) (string, error) {
	moveCallRequest.Signer = txHandler.relayerAddress
	moveCallRequest.PackageObjectId = txHandler.bridgeContractAddress

	txnMetaData, err := txHandler.client.MoveCall(ctx, moveCallRequest)
	if err != nil {
		return "", err
	}

	txBlockResponse, err := txHandler.client.SignAndExecuteTransactionBlock(ctx, models.SignAndExecuteTransactionBlockRequest{
		TxnMetaData: txnMetaData,
		PriKey:      txHandler.relayerPrivateKey,
		Options: models.SuiTransactionBlockOptions{
			ShowInput:    true,
			ShowRawInput: true,
		},
		RequestType: "WaitForLocalExecution",
	})
	if err != nil {
		return "", err
	}

	return txBlockResponse.Digest, nil
}

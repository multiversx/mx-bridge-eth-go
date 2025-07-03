package sui

import (
	"context"
	"encoding/base64"
	"errors"
	"github.com/block-vision/sui-go-sdk/models"
	"github.com/block-vision/sui-go-sdk/signer"
	"github.com/block-vision/sui-go-sdk/sui"
)

type transactionHandler struct {
	client        sui.ISuiAPI
	relayerSigner *signer.Signer
}

func (txHandler *transactionHandler) SendTransactionReturnHash(ctx context.Context, moveCallRequest models.MoveCallRequest) (string, error) {
	moveCallRequest.Signer = txHandler.relayerSigner.Address

	txnMetaData, err := txHandler.client.MoveCall(ctx, moveCallRequest)
	if err != nil {
		return "", err
	}

	txBlockResponse, err := txHandler.client.SignAndExecuteTransactionBlock(ctx, models.SignAndExecuteTransactionBlockRequest{
		TxnMetaData: txnMetaData,
		PriKey:      txHandler.relayerSigner.PriKey,
		Options: models.SuiTransactionBlockOptions{
			ShowInput:         true,
			ShowEffects:       true,
			ShowObjectChanges: true,
		},
		RequestType: "WaitForLocalExecution",
	})
	if err != nil {
		return "", err
	}

	if txBlockResponse.Effects.Status.Status != "success" {
		return "", errors.New(txBlockResponse.Effects.Status.Error)
	}

	return txBlockResponse.Digest, nil
}

func (txHandler *transactionHandler) Sign(message []byte) ([]byte, error) {
	signedMessageSerializedSig, err := txHandler.relayerSigner.SignPersonalMessage(string(message))
	if err != nil {
		return nil, err
	}
	return base64.StdEncoding.DecodeString(signedMessageSerializedSig.Signature)
}

package sui

import (
	"context"
	"fmt"
	"github.com/block-vision/sui-go-sdk/models"
	"github.com/block-vision/sui-go-sdk/signer"
	"github.com/block-vision/sui-go-sdk/sui"
	"github.com/block-vision/sui-go-sdk/transaction"
	bridgeCore "github.com/multiversx/mx-bridge-eth-go/core"
	chainCore "github.com/multiversx/mx-chain-core-go/core"
)

const waitForLocalExecution = "WaitForLocalExecution"

// ArgsTxHandler holds the arguments for creating a new Sui transaction handler
type ArgsTxHandler struct {
	Proxy     Proxy
	Signer    *signer.Signer
	GasPrice  uint64
	GasBudget uint64
	Logger    chainCore.Logger
}

type transactionHandler struct {
	suiCli    *sui.Client
	signer    *signer.Signer
	gasPrice  uint64
	gasBudget uint64
	logger    chainCore.Logger
}

func NewTransactionHandler(args ArgsTxHandler) (*transactionHandler, error) {
	cli, ok := args.Proxy.(*sui.Client)
	if !ok {
		return nil, errInvalidProxyType
	}

	return &transactionHandler{
		suiCli:    cli,
		signer:    args.Signer,
		gasPrice:  args.GasPrice,
		gasBudget: args.GasBudget,
		logger:    args.Logger,
	}, nil
}

func (txHandler *transactionHandler) SendTransactionReturnHash(
	ctx context.Context,
	gasCoin *transaction.SuiObjectRef,
	calls []bridgeCore.SuiPTBOperation,
) (string, error) {
	tx := transaction.NewTransaction()
	tx.SetSuiClient(txHandler.suiCli).
		SetSigner(txHandler.signer).
		SetSender(models.SuiAddress(txHandler.signer.Address)).
		SetGasPrice(txHandler.gasPrice).
		SetGasBudget(txHandler.gasBudget).
		SetGasPayment([]transaction.SuiObjectRef{*gasCoin}).
		SetGasOwner(models.SuiAddress(txHandler.signer.Address))

	for _, call := range calls {
		args := call.ArgsFn(tx)
		tx.MoveCall(call.Package, call.Module, call.Function, call.TypeTags, args)
	}

	resp, err := tx.Execute(
		ctx,
		models.SuiTransactionBlockOptions{ShowEffects: true},
		waitForLocalExecution,
	)
	if err != nil {
		return "", err
	}
	if resp.Effects.Status.Status != successCodeAfterExecution {
		return "", fmt.Errorf("transaction failed: %s", resp.Effects.Status.Error)
	}

	return resp.Digest, nil
}

package sui

import (
	"context"
	"fmt"
	"github.com/block-vision/sui-go-sdk/models"
	"github.com/block-vision/sui-go-sdk/signer"
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
	proxy     Proxy
	signer    *signer.Signer
	gasPrice  uint64
	gasBudget uint64
	logger    chainCore.Logger
}

func NewTransactionHandler(args ArgsTxHandler) (*transactionHandler, error) {
	if args.Proxy == nil {
		return nil, errNilProxy
	}

	return &transactionHandler{
		proxy:     args.Proxy,
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
	tx.SetSigner(txHandler.signer).
		SetSender(models.SuiAddress(txHandler.signer.Address)).
		SetGasPrice(txHandler.gasPrice).
		SetGasBudget(txHandler.gasBudget).
		SetGasPayment([]transaction.SuiObjectRef{*gasCoin}).
		SetGasOwner(models.SuiAddress(txHandler.signer.Address))

	for _, call := range calls {
		args := call.ArgsFn(tx)
		tx.MoveCall(call.Package, call.Module, call.Function, call.TypeTags, args)
	}

	req, err := tx.ToSuiExecuteTransactionBlockRequest(
		ctx,
		models.SuiTransactionBlockOptions{ShowEffects: true},
		waitForLocalExecution,
	)
	if err != nil {
		return "", err
	}

	resp, err := txHandler.proxy.SuiExecuteTransactionBlock(ctx, *req)
	if err != nil {
		return "", err
	}
	if resp.Effects.Status.Status != successCodeAfterExecution {
		return "", fmt.Errorf("transaction failed: %s", resp.Effects.Status.Error)
	}

	return resp.Digest, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (txHandler *transactionHandler) IsInterfaceNil() bool {
	return txHandler == nil
}

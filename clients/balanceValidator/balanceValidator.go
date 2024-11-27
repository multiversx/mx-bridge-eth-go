package balanceValidator

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/multiversx/mx-bridge-eth-go/clients"
	bridgeCore "github.com/multiversx/mx-bridge-eth-go/core"
	"github.com/multiversx/mx-bridge-eth-go/core/batchProcessor"
	"github.com/multiversx/mx-chain-core-go/core/check"
	logger "github.com/multiversx/mx-chain-logger-go"
)

// ArgsBalanceValidator represents the DTO struct used in the NewBalanceValidator constructor function
type ArgsBalanceValidator struct {
	Log              logger.Logger
	MultiversXClient MultiversXClient
	EthereumClient   EthereumClient
}

type balanceValidator struct {
	log              logger.Logger
	multiversXClient MultiversXClient
	ethereumClient   EthereumClient
}

// NewBalanceValidator creates a new instance of type balanceValidator
func NewBalanceValidator(args ArgsBalanceValidator) (*balanceValidator, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	return &balanceValidator{
		log:              args.Log,
		multiversXClient: args.MultiversXClient,
		ethereumClient:   args.EthereumClient,
	}, nil
}

func checkArgs(args ArgsBalanceValidator) error {
	if check.IfNil(args.Log) {
		return ErrNilLogger
	}
	if check.IfNil(args.MultiversXClient) {
		return ErrNilMultiversXClient
	}
	if check.IfNil(args.EthereumClient) {
		return ErrNilEthereumClient
	}

	return nil
}

// CheckToken returns error if the bridge can not happen to the provided token due to faulty balance values in the contracts
func (validator *balanceValidator) CheckToken(ctx context.Context, ethToken common.Address, mvxToken []byte, amount *big.Int, direction batchProcessor.Direction) error {
	err := validator.checkRequiredBalance(ctx, ethToken, mvxToken, amount, direction)
	if err != nil {
		return err
	}

	isMintBurnOnEthereum, err := validator.isMintBurnOnEthereum(ctx, ethToken)
	if err != nil {
		return err
	}

	isMintBurnOnMultiversX, err := validator.isMintBurnOnMultiversX(ctx, mvxToken)
	if err != nil {
		return err
	}

	isNativeOnEthereum, err := validator.isNativeOnEthereum(ctx, ethToken)
	if err != nil {
		return err
	}

	isNativeOnMultiversX, err := validator.isNativeOnMultiversX(ctx, mvxToken)
	if err != nil {
		return err
	}

	if !isNativeOnEthereum && !isMintBurnOnEthereum {
		return fmt.Errorf("%w isNativeOnEthereum = %v, isMintBurnOnEthereum = %v", ErrInvalidSetup, isNativeOnEthereum, isMintBurnOnEthereum)
	}

	if !isNativeOnMultiversX && !isMintBurnOnMultiversX {
		return fmt.Errorf("%w isNativeOnMultiversX = %v, isMintBurnOnMultiversX = %v", ErrInvalidSetup, isNativeOnMultiversX, isMintBurnOnMultiversX)
	}

	if isNativeOnEthereum == isNativeOnMultiversX {
		return fmt.Errorf("%w isNativeOnEthereum = %v, isNativeOnMultiversX = %v", ErrInvalidSetup, isNativeOnEthereum, isNativeOnMultiversX)
	}

	ethAmount, err := validator.computeEthAmount(ctx, ethToken, isMintBurnOnEthereum, isNativeOnEthereum)
	if err != nil {
		return err
	}
	mvxAmount, err := validator.computeMvxAmount(ctx, mvxToken, isMintBurnOnMultiversX, isNativeOnMultiversX)
	if err != nil {
		return err
	}

	validator.log.Debug("balanceValidator.CheckToken",
		"ERC20 token", ethToken.String(),
		"ERC20 balance", ethAmount.String(),
		"ESDT token", mvxToken,
		"ESDT balance", mvxAmount.String(),
		"amount", amount.String(),
	)

	if ethAmount.Cmp(mvxAmount) != 0 {
		return fmt.Errorf("%w, balance for ERC20 token %s is %s and the balance for ESDT token %s is %s, direction %s",
			ErrBalanceMismatch, ethToken.String(), ethAmount.String(), mvxToken, mvxAmount.String(), direction)
	}
	return nil
}

func (validator *balanceValidator) checkRequiredBalance(ctx context.Context, ethToken common.Address, mvxToken []byte, amount *big.Int, direction batchProcessor.Direction) error {
	switch direction {
	case batchProcessor.FromMultiversX:
		return validator.ethereumClient.CheckRequiredBalance(ctx, ethToken, amount)
	case batchProcessor.ToMultiversX:
		return validator.multiversXClient.CheckRequiredBalance(ctx, mvxToken, amount)
	default:
		return fmt.Errorf("%w, direction: %s", ErrInvalidDirection, direction)
	}
}

func (validator *balanceValidator) isMintBurnOnEthereum(ctx context.Context, erc20Address common.Address) (bool, error) {
	isMintBurn, err := validator.ethereumClient.MintBurnTokens(ctx, erc20Address)
	if err != nil {
		return false, err
	}

	return isMintBurn, nil
}

func (validator *balanceValidator) isNativeOnEthereum(ctx context.Context, erc20Address common.Address) (bool, error) {
	isNative, err := validator.ethereumClient.NativeTokens(ctx, erc20Address)
	if err != nil {
		return false, err
	}
	return isNative, nil
}

func (validator *balanceValidator) isMintBurnOnMultiversX(ctx context.Context, token []byte) (bool, error) {
	isMintBurn, err := validator.multiversXClient.IsMintBurnToken(ctx, token)
	if err != nil {
		return false, err
	}
	return isMintBurn, nil
}

func (validator *balanceValidator) isNativeOnMultiversX(ctx context.Context, token []byte) (bool, error) {
	isNative, err := validator.multiversXClient.IsNativeToken(ctx, token)
	if err != nil {
		return false, err
	}
	return isNative, nil
}

func (validator *balanceValidator) computeEthAmount(
	ctx context.Context,
	token common.Address,
	isMintBurn bool,
	isNative bool,
) (*big.Int, error) {
	ethAmountInPendingBatches, err := validator.getTotalTransferAmountInPendingEthBatches(ctx, token)
	if err != nil {
		return nil, err
	}

	if !isMintBurn {
		// we need to subtract all locked balances on the Ethereum side (all pending, un-executed batches) so the balances
		// with the minted MultiversX tokens will match
		total, errTotal := validator.ethereumClient.TotalBalances(ctx, token)
		if errTotal != nil {
			return nil, errTotal
		}

		return total.Sub(total, ethAmountInPendingBatches), nil
	}

	burnBalances, err := validator.ethereumClient.BurnBalances(ctx, token)
	if err != nil {
		return nil, err
	}
	mintBalances, err := validator.ethereumClient.MintBalances(ctx, token)
	if err != nil {
		return nil, err
	}

	// we need to cancel out what was burned in advance when the deposit was registered in the contract
	burnBalances.Sub(burnBalances, ethAmountInPendingBatches)

	var ethAmount *big.Int
	if isNative {
		ethAmount = big.NewInt(0).Sub(burnBalances, mintBalances)
	} else {
		ethAmount = big.NewInt(0).Sub(mintBalances, burnBalances)
	}

	if ethAmount.Cmp(big.NewInt(0)) < 0 {
		return big.NewInt(0), fmt.Errorf("%w, ethAmount: %s", ErrNegativeAmount, ethAmount.String())
	}
	return ethAmount, nil
}

func (validator *balanceValidator) computeMvxAmount(
	ctx context.Context,
	token []byte,
	isMintBurn bool,
	isNative bool,
) (*big.Int, error) {
	mvxAmountInPendingBatches, err := validator.getTotalTransferAmountInPendingMvxBatches(ctx, token)
	if err != nil {
		return nil, err
	}

	if !isMintBurn {
		// we need to subtract all locked balances on the MultiversX side (all pending, un-executed batches) so the balances
		// with the minted Ethereum tokens will match
		total, errTotal := validator.multiversXClient.TotalBalances(ctx, token)
		if errTotal != nil {
			return nil, errTotal
		}

		return total.Sub(total, mvxAmountInPendingBatches), nil
	}

	burnBalances, err := validator.multiversXClient.BurnBalances(ctx, token)
	if err != nil {
		return nil, err
	}
	mintBalances, err := validator.multiversXClient.MintBalances(ctx, token)
	if err != nil {
		return nil, err
	}
	var mvxAmount *big.Int

	// we need to cancel out what was burned in advance when the deposit was registered in the contract
	burnBalances.Sub(burnBalances, mvxAmountInPendingBatches)

	if isNative {
		mvxAmount = big.NewInt(0).Sub(burnBalances, mintBalances)
	} else {
		mvxAmount = big.NewInt(0).Sub(mintBalances, burnBalances)
	}

	if mvxAmount.Cmp(big.NewInt(0)) < 0 {
		return big.NewInt(0), fmt.Errorf("%w, mvxAmount: %s", ErrNegativeAmount, mvxAmount.String())
	}
	return mvxAmount, nil
}

func getTotalAmountFromBatch(batch *bridgeCore.TransferBatch, token []byte) *big.Int {
	amount := big.NewInt(0)
	for _, deposit := range batch.Deposits {
		if bytes.Equal(deposit.SourceTokenBytes, token) {
			amount.Add(amount, deposit.Amount)
		}
	}

	return amount
}

func (validator *balanceValidator) getTotalTransferAmountInPendingMvxBatches(ctx context.Context, mvxToken []byte) (*big.Int, error) {
	batchID, err := validator.multiversXClient.GetLastMvxBatchID(ctx)
	if err != nil {
		return nil, err
	}

	var batch *bridgeCore.TransferBatch
	amount := big.NewInt(0)
	for {
		batch, err = validator.multiversXClient.GetBatch(ctx, batchID)
		if errors.Is(err, clients.ErrNoBatchAvailable) {
			return amount, nil
		}
		if err != nil {
			return nil, err
		}

		wasExecuted, errWasExecuted := validator.ethereumClient.WasExecuted(ctx, batch.ID)
		if errWasExecuted != nil {
			return nil, errWasExecuted
		}
		if wasExecuted {
			return amount, nil
		}

		amountFromBatch := getTotalAmountFromBatch(batch, mvxToken)
		amount.Add(amount, amountFromBatch)
		batchID-- // go to the previous batch
	}
}

func (validator *balanceValidator) getTotalTransferAmountInPendingEthBatches(ctx context.Context, ethToken common.Address) (*big.Int, error) {
	batchID, err := validator.multiversXClient.GetLastExecutedEthBatchID(ctx)
	if err != nil {
		return nil, err
	}

	var batch *bridgeCore.TransferBatch
	amount := big.NewInt(0)
	for {
		batch, _, err = validator.ethereumClient.GetBatch(ctx, batchID+1) // we take all batches, regardless if they are final or not
		if err != nil {
			return nil, err
		}

		isBatchInvalid := batch.ID != batchID+1 || len(batch.Deposits) == 0
		if isBatchInvalid {
			return amount, nil
		}

		amountFromBatch := getTotalAmountFromBatch(batch, ethToken.Bytes())
		amount.Add(amount, amountFromBatch)
		batchID++
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (validator *balanceValidator) IsInterfaceNil() bool {
	return validator == nil
}

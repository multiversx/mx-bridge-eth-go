package ethereum

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/multiversx/mx-bridge-eth-go/clients/ethereum"
	"github.com/multiversx/mx-bridge-eth-go/core/batchProcessor"
	"github.com/multiversx/mx-chain-core-go/core/check"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var zero = big.NewInt(0)

const timeBetweenBatchIDChecks = time.Millisecond * 100

// ArgsMigrationBatchCreator is the argument for the NewMigrationBatchCreator constructor
type ArgsMigrationBatchCreator struct {
	MvxDataGetter        MvxDataGetter
	Erc20ContractsHolder Erc20ContractsHolder
	SafeContractAddress  common.Address
	EthereumChainWrapper EthereumChainWrapper
	Logger               logger.Logger
}

type migrationBatchCreator struct {
	mvxDataGetter        MvxDataGetter
	erc20ContractsHolder Erc20ContractsHolder
	safeContractAddress  common.Address
	ethereumChainWrapper EthereumChainWrapper
	logger               logger.Logger
}

// NewMigrationBatchCreator creates a new instance of type migrationBatchCreator that is able to generate the migration batch output file
func NewMigrationBatchCreator(args ArgsMigrationBatchCreator) (*migrationBatchCreator, error) {
	if check.IfNil(args.MvxDataGetter) {
		return nil, errNilMvxDataGetter
	}
	if check.IfNil(args.Erc20ContractsHolder) {
		return nil, errNilErc20ContractsHolder
	}
	if check.IfNilReflect(args.EthereumChainWrapper) {
		return nil, errNilEthereumChainWrapper
	}
	if check.IfNil(args.Logger) {
		return nil, errNilLogger
	}

	return &migrationBatchCreator{
		mvxDataGetter:        args.MvxDataGetter,
		erc20ContractsHolder: args.Erc20ContractsHolder,
		safeContractAddress:  args.SafeContractAddress,
		logger:               args.Logger,
		ethereumChainWrapper: args.EthereumChainWrapper,
	}, nil
}

// CreateBatchInfo creates an instance of type BatchInfo
func (creator *migrationBatchCreator) CreateBatchInfo(ctx context.Context, newSafeAddress common.Address, partialMigration map[string]*FloatWrapper) (*BatchInfo, error) {
	creator.logger.Info("started the batch creation process...")

	depositStart := uint64(0) // deposits inside a batch are not tracked, we can start from 0

	creator.logger.Info("will try to find a usable batch ID, please wait, this might take a while...")

	startTime := time.Now()
	freeBatchID, err := creator.findAnUsableBatchID(ctx, timeBetweenBatchIDChecks)
	endTime := time.Now()
	if err != nil {
		return nil, err
	}

	creator.logger.Info("fetched Ethereum contracts state",
		"free batch ID", freeBatchID, "time took", endTime.Sub(startTime))

	if partialMigration == nil {
		partialMigration = make(map[string]*FloatWrapper)
	}

	tokensList, err := creator.getTokensList(ctx, partialMigration)
	if err != nil {
		return nil, err
	}

	creator.logger.Info("fetched known tokens", "tokens", strings.Join(tokensList, ", "))

	deposits, err := creator.fetchERC20ContractsAddresses(ctx, tokensList, depositStart)
	if err != nil {
		return nil, err
	}

	creator.logger.Info("fetched ERC20 contract addresses")

	err = creator.fetchBalances(ctx, deposits, partialMigration)
	if err != nil {
		return nil, err
	}

	creator.logger.Info("fetched balances contract addresses")

	return creator.assembleBatchInfo(freeBatchID, deposits, newSafeAddress)
}

func (creator *migrationBatchCreator) findAnUsableBatchID(ctx context.Context, timeBetweenChecks time.Duration) (uint64, error) {
	highBatchID := uint64(100000)
	lowBatchID := uint64(1)
	increaseHigh := uint64(100000)

	batchesUsedMap := make(map[uint64]bool)
	for {
		err := creator.checkAvailableBatch(ctx, highBatchID, batchesUsedMap, timeBetweenChecks)
		if err != nil {
			return 0, err
		}

		err = creator.checkAvailableBatch(ctx, lowBatchID, batchesUsedMap, timeBetweenChecks)
		if err != nil {
			return 0, err
		}

		if !batchesUsedMap[lowBatchID] {
			// beginning of the interval optimization
			return lowBatchID, nil
		}

		if batchesUsedMap[highBatchID] && batchesUsedMap[lowBatchID] {
			// high was too low
			highBatchID += increaseHigh
			continue
		}

		mid := (highBatchID + lowBatchID) / 2
		if mid == lowBatchID {
			// high and low are so close that their middle value between them is actually the low value, return high
			return highBatchID, nil
		}

		err = creator.checkAvailableBatch(ctx, mid, batchesUsedMap, timeBetweenChecks)
		if err != nil {
			return 0, err
		}

		if batchesUsedMap[mid] {
			// the middle value was set, bring the low value to it and restart the checking process
			lowBatchID = mid
		} else {
			// the middle value was not set, bring the high value to it and restart the checking process
			highBatchID = mid
		}
	}
}

func (creator *migrationBatchCreator) checkAvailableBatch(
	ctx context.Context,
	batchID uint64,
	batchesUsedMap map[uint64]bool,
	timeBetweenChecks time.Duration,
) error {
	_, checked := batchesUsedMap[batchID]
	if checked {
		return nil
	}

	time.Sleep(timeBetweenChecks)
	wasExecuted, err := creator.ethereumChainWrapper.WasBatchExecuted(ctx, big.NewInt(0).SetUint64(batchID))
	if err != nil {
		return err
	}

	batchesUsedMap[batchID] = wasExecuted
	return nil
}

func (creator *migrationBatchCreator) getTokensList(ctx context.Context, partialMigration map[string]*FloatWrapper) ([]string, error) {
	tokens, err := creator.mvxDataGetter.GetAllKnownTokens(ctx)
	if err != nil {
		return nil, err
	}
	if len(tokens) == 0 {
		return nil, fmt.Errorf("%w when calling the getAllKnownTokens function on the safe contract", errEmptyTokensList)
	}

	stringTokens := make([]string, 0, len(tokens))
	for _, token := range tokens {
		if len(partialMigration) > 1 && partialMigration[string(token)] == nil {
			// partial migration was set, but for the current token in this deposit a value was not given
			// skip this deposit
			continue
		}

		stringTokens = append(stringTokens, string(token))
	}

	return stringTokens, nil
}

func (creator *migrationBatchCreator) fetchERC20ContractsAddresses(ctx context.Context, tokensList []string, lastDepositNonce uint64) ([]*DepositInfo, error) {
	deposits := make([]*DepositInfo, 0, len(tokensList))
	for idx, token := range tokensList {
		response, err := creator.mvxDataGetter.GetERC20AddressForTokenId(ctx, []byte(token))
		if err != nil {
			return nil, err
		}
		if len(response) != 1 {
			return nil, fmt.Errorf("%w when querying the safe contract for token %s",
				errWrongERC20AddressResponse, token)
		}

		deposit := &DepositInfo{
			DepositNonce:          lastDepositNonce + uint64(1+idx),
			Token:                 token,
			ContractAddressString: common.BytesToAddress(response[0]).String(),
			ContractAddress:       common.BytesToAddress(response[0]),
			AmountString:          "",
		}

		deposits = append(deposits, deposit)
	}

	return deposits, nil
}

func (creator *migrationBatchCreator) fetchBalances(ctx context.Context, deposits []*DepositInfo, partialMigration map[string]*FloatWrapper) error {
	for _, deposit := range deposits {
		balance, err := creator.erc20ContractsHolder.BalanceOf(ctx, deposit.ContractAddress, creator.safeContractAddress)
		if err != nil {
			return fmt.Errorf("%w for address %s in ERC20 contract %s", err, creator.safeContractAddress.String(), deposit.ContractAddress.String())
		}

		decimals, err := creator.erc20ContractsHolder.Decimals(ctx, deposit.ContractAddress)
		if err != nil {
			return fmt.Errorf("%w for in ERC20 contract %s", err, deposit.ContractAddress.String())
		}
		deposit.Decimals = decimals

		trimValue := partialMigration[deposit.Token]
		trimIsNeeded := trimValue != nil && !trimValue.IsMax
		if trimIsNeeded {
			denominatedTrimAmount := big.NewFloat(0).Set(trimValue.Float)
			multiplier := big.NewInt(10)
			multiplier.Exp(multiplier, big.NewInt(int64(deposit.Decimals)), nil)
			denominatedTrimAmount.Mul(denominatedTrimAmount, big.NewFloat(0).SetInt(multiplier))

			newBalance := big.NewInt(0)
			denominatedTrimAmount.Int(newBalance)
			if balance.Cmp(newBalance) > 0 {
				creator.logger.Warn("applied denominated value", "balance", balance.String(), "new value to consider", newBalance.String())
				balance = newBalance
			} else {
				creator.logger.Warn("can not apply denominated value as the balance is under the provided value, will use the whole balance", "balance", balance.String())
			}
		}

		deposit.Amount = balance
		deposit.AmountString = balance.String()

		divider := big.NewInt(10)
		divider.Exp(divider, big.NewInt(int64(decimals)), nil)

		deposit.DenominatedAmount = big.NewFloat(0).SetInt(balance)
		deposit.DenominatedAmount.Quo(deposit.DenominatedAmount, big.NewFloat(0).SetInt(divider))
		deposit.DenominatedAmountString = deposit.DenominatedAmount.Text('f', -1)
	}

	return nil
}

func (creator *migrationBatchCreator) assembleBatchInfo(usableBatchID uint64, deposits []*DepositInfo, newSafeAddress common.Address) (*BatchInfo, error) {
	batchInfo := &BatchInfo{
		OldSafeContractAddress: creator.safeContractAddress.String(),
		NewSafeContractAddress: newSafeAddress.String(),
		BatchID:                usableBatchID,
		DepositsInfo:           make([]*DepositInfo, 0, len(deposits)),
	}

	for _, deposit := range deposits {
		if deposit.Amount.Cmp(zero) <= 0 {
			continue
		}

		batchInfo.DepositsInfo = append(batchInfo.DepositsInfo, deposit)
	}

	var err error
	batchInfo.MessageHash, err = creator.computeMessageHash(batchInfo)
	if err != nil {
		return nil, err
	}

	return batchInfo, nil
}

func (creator *migrationBatchCreator) computeMessageHash(batch *BatchInfo) (common.Hash, error) {
	tokens := make([]common.Address, 0, len(batch.DepositsInfo))
	recipients := make([]common.Address, 0, len(batch.DepositsInfo))
	amounts := make([]*big.Int, 0, len(batch.DepositsInfo))
	nonces := make([]*big.Int, 0, len(batch.DepositsInfo))
	for _, deposit := range batch.DepositsInfo {
		tokens = append(tokens, deposit.ContractAddress)
		recipients = append(recipients, common.HexToAddress(batch.NewSafeContractAddress))
		amounts = append(amounts, deposit.Amount)
		nonces = append(nonces, big.NewInt(0).SetUint64(deposit.DepositNonce))
	}

	args := &batchProcessor.ArgListsBatch{
		EthTokens:  tokens,
		Recipients: recipients,
		Amounts:    amounts,
		Nonces:     nonces,
	}

	return ethereum.GenerateMessageHash(args, batch.BatchID)
}

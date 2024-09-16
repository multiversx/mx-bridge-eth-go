package ethereum

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/multiversx/mx-bridge-eth-go/clients/ethereum"
	"github.com/multiversx/mx-bridge-eth-go/core/batchProcessor"
	"github.com/multiversx/mx-chain-core-go/core/check"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var zero = big.NewInt(0)

// ArgsMigrationBatchCreator is the argument for the NewMigrationBatchCreator constructor
type ArgsMigrationBatchCreator struct {
	MvxDataGetter        MvxDataGetter
	Erc20ContractsHolder Erc20ContractsHolder
	SafeContractAddress  common.Address
	SafeContractWrapper  SafeContractWrapper
	Logger               logger.Logger
}

type migrationBatchCreator struct {
	mvxDataGetter        MvxDataGetter
	erc20ContractsHolder Erc20ContractsHolder
	safeContractAddress  common.Address
	safeContractWrapper  SafeContractWrapper
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
	if check.IfNilReflect(args.SafeContractWrapper) {
		return nil, errNilSafeContractWrapper
	}
	if check.IfNil(args.Logger) {
		return nil, errNilLogger
	}

	return &migrationBatchCreator{
		mvxDataGetter:        args.MvxDataGetter,
		erc20ContractsHolder: args.Erc20ContractsHolder,
		safeContractAddress:  args.SafeContractAddress,
		safeContractWrapper:  args.SafeContractWrapper,
		logger:               args.Logger,
	}, nil
}

// CreateBatchInfo creates an instance of type BatchInfo
func (creator *migrationBatchCreator) CreateBatchInfo(ctx context.Context, newSafeAddress common.Address) (*BatchInfo, error) {
	creator.logger.Info("started the batch creation process...")

	batchesCount, err := creator.safeContractWrapper.BatchesCount(&bind.CallOpts{Context: ctx})
	if err != nil {
		return nil, err
	}

	depositsCount, err := creator.safeContractWrapper.DepositsCount(&bind.CallOpts{Context: ctx})
	if err != nil {
		return nil, err
	}

	creator.logger.Info("fetched Ethereum contracts state", "batches count", batchesCount, "deposits count", depositsCount)

	tokensList, err := creator.getTokensList(ctx)
	if err != nil {
		return nil, err
	}

	creator.logger.Info("fetched known tokens", "tokens", strings.Join(tokensList, ", "))

	deposits, err := creator.fetchERC20ContractsAddresses(ctx, tokensList, depositsCount)
	if err != nil {
		return nil, err
	}

	creator.logger.Info("fetched ERC20 contract addresses")

	err = creator.fetchBalances(ctx, deposits)
	if err != nil {
		return nil, err
	}

	creator.logger.Info("fetched balances contract addresses")

	return creator.assembleBatchInfo(batchesCount, deposits, newSafeAddress)
}

func (creator *migrationBatchCreator) getTokensList(ctx context.Context) ([]string, error) {
	tokens, err := creator.mvxDataGetter.GetAllKnownTokens(ctx)
	if err != nil {
		return nil, err
	}
	if len(tokens) == 0 {
		return nil, fmt.Errorf("%w when calling the getAllKnownTokens function on the safe contract", errEmptyTokensList)
	}

	stringTokens := make([]string, 0, len(tokens))
	for _, token := range tokens {
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

func (creator *migrationBatchCreator) fetchBalances(ctx context.Context, deposits []*DepositInfo) error {
	for _, deposit := range deposits {
		balance, err := creator.erc20ContractsHolder.BalanceOf(ctx, deposit.ContractAddress, creator.safeContractAddress)
		if err != nil {
			return fmt.Errorf("%w for address %s in ERC20 contract %s", err, creator.safeContractAddress.String(), deposit.ContractAddress.String())
		}

		deposit.Amount = balance
		deposit.AmountString = balance.String()
	}

	return nil
}

func (creator *migrationBatchCreator) assembleBatchInfo(batchesCount uint64, deposits []*DepositInfo, newSafeAddress common.Address) (*BatchInfo, error) {
	batchInfo := &BatchInfo{
		OldSafeContractAddress: creator.safeContractAddress.String(),
		NewSafeContractAddress: newSafeAddress.String(),
		BatchID:                batchesCount + 1,
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

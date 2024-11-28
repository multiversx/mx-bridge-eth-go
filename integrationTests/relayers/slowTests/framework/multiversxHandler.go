package framework

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/multiversx/mx-sdk-go/data"
	"github.com/stretchr/testify/require"
)

const (
	minRelayerStake          = "10000000000000000000" // 10 EGLD
	esdtIssueCost            = "50000000000000000"    // 0.05 EGLD
	emptyAddress             = "erd1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq6gq4hu"
	esdtSystemSCAddress      = "erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqzllls8a5w6u"
	slashAmount              = "00"
	zeroStringValue          = "0"
	canAddSpecialRoles       = "canAddSpecialRoles"
	trueStr                  = "true"
	esdtRoleLocalMint        = "ESDTRoleLocalMint"
	esdtRoleLocalBurn        = "ESDTRoleLocalBurn"
	hexTrue                  = "01"
	hexFalse                 = "00"
	gwei                     = "GWEI"
	maxBridgedAmountForToken = "500000"
	deployGasLimit           = 150000000 // 150 million
	setCallsGasLimit         = 80000000  // 80 million
	issueTokenGasLimit       = 70000000  // 70 million
	createDepositGasLimit    = 20000000  // 20 million
	generalSCCallGasLimit    = 50000000  // 50 million
	gasLimitPerDataByte      = 1500

	aggregatorContractPath    = "testdata/contracts/mvx/multiversx-price-aggregator-sc.wasm"
	wrapperContractPath       = "testdata/contracts/mvx/bridged-tokens-wrapper.wasm"
	multiTransferContractPath = "testdata/contracts/mvx/multi-transfer-esdt.wasm"
	safeContractPath          = "testdata/contracts/mvx/esdt-safe.wasm"
	multisigContractPath      = "testdata/contracts/mvx/multisig.wasm"
	bridgeProxyContractPath   = "testdata/contracts/mvx/bridge-proxy.wasm"
	testCallerContractPath    = "testdata/contracts/mvx/test-caller.wasm"

	setBridgeProxyContractAddressFunction                = "setBridgeProxyContractAddress"
	setWrappingContractAddressFunction                   = "setWrappingContractAddress"
	changeOwnerAddressFunction                           = "ChangeOwnerAddress"
	setEsdtSafeOnMultiTransferFunction                   = "setEsdtSafeOnMultiTransfer"
	setEsdtSafeAddressFunction                           = "setEsdtSafeAddress"
	stakeFunction                                        = "stake"
	unpauseFunction                                      = "unpause"
	unpauseEsdtSafeFunction                              = "unpauseEsdtSafe"
	unpauseProxyFunction                                 = "unpauseProxy"
	pauseEsdtSafeFunction                                = "pauseEsdtSafe"
	pauseFunction                                        = "pause"
	issueFunction                                        = "issue"
	setSpecialRoleFunction                               = "setSpecialRole"
	esdtTransferFunction                                 = "ESDTTransfer"
	setPairDecimalsFunction                              = "setPairDecimals"
	addWrappedTokenFunction                              = "addWrappedToken"
	depositLiquidityFunction                             = "depositLiquidity"
	whitelistTokenFunction                               = "whitelistToken"
	addMappingFunction                                   = "addMapping"
	esdtSafeAddTokenToWhitelistFunction                  = "esdtSafeAddTokenToWhitelist"
	esdtSafeSetMaxBridgedAmountForTokenFunction          = "esdtSafeSetMaxBridgedAmountForToken"
	multiTransferEsdtSetMaxBridgedAmountForTokenFunction = "multiTransferEsdtSetMaxBridgedAmountForToken"
	submitBatchFunction                                  = "submitBatch"
	unwrapTokenCreateTransactionFunction                 = "unwrapTokenCreateTransaction"
	createTransactionFunction                            = "createTransaction"
	setBridgedTokensWrapperAddressFunction               = "setBridgedTokensWrapperAddress"
	setMultiTransferAddressFunction                      = "setMultiTransferAddress"
	withdrawRefundFeesForEthereumFunction                = "withdrawRefundFeesForEthereum"
	getRefundFeesForEthereumFunction                     = "getRefundFeesForEthereum"
	withdrawTransactionFeesFunction                      = "withdrawTransactionFees"
	getTransactionFeesFunction                           = "getTransactionFees"
	initSupplyMintBurnEsdtSafe                           = "initSupplyMintBurnEsdtSafe"
	initSupplyEsdtSafe                                   = "initSupplyEsdtSafe"
	getMintBalances                                      = "getMintBalances"
	getBurnBalances                                      = "getBurnBalances"
	getTotalBalances                                     = "getTotalBalances"
)

var (
	feeInt = big.NewInt(50)
)

// MultiversxHandler will handle all the operations on the MultiversX side
type MultiversxHandler struct {
	testing.TB
	*KeysStore
	Quorum         string
	TokensRegistry TokensRegistry
	ChainSimulator ChainSimulatorWrapper

	AggregatorAddress         *MvxAddress
	WrapperAddress            *MvxAddress
	SafeAddress               *MvxAddress
	MultisigAddress           *MvxAddress
	MultiTransferAddress      *MvxAddress
	ScProxyAddress            *MvxAddress
	CalleeScAddress           *MvxAddress
	ESDTSystemContractAddress *MvxAddress
}

// NewMultiversxHandler will create the handler that will adapt all test operations on MultiversX
func NewMultiversxHandler(
	tb testing.TB,
	ctx context.Context,
	keysStore *KeysStore,
	tokensRegistry TokensRegistry,
	chainSimulator ChainSimulatorWrapper,
	quorum string,
) *MultiversxHandler {
	handler := &MultiversxHandler{
		TB:             tb,
		KeysStore:      keysStore,
		TokensRegistry: tokensRegistry,
		ChainSimulator: chainSimulator,
		Quorum:         quorum,
	}

	handler.ESDTSystemContractAddress = NewMvxAddressFromBech32(handler, esdtSystemSCAddress)

	handler.ChainSimulator.GenerateBlocksUntilEpochReached(ctx, 1)

	handler.ChainSimulator.FundWallets(ctx, handler.WalletsToFundOnMultiversX())
	handler.ChainSimulator.GenerateBlocks(ctx, 1)

	return handler
}

// DeployAndSetContracts will deploy all required contracts on MultiversX side and do the proper wiring
func (handler *MultiversxHandler) DeployAndSetContracts(ctx context.Context) {
	handler.deployContracts(ctx)

	handler.wireMultiTransfer(ctx)
	handler.wireSCProxy(ctx)
	handler.wireSafe(ctx)

	handler.changeOwners(ctx)
	handler.finishSettings(ctx)
}

func (handler *MultiversxHandler) deployContracts(ctx context.Context) {
	// deploy aggregator
	stakeValue, _ := big.NewInt(0).SetString(minRelayerStake, 10)
	aggregatorDeployParams := []string{
		hex.EncodeToString([]byte("EGLD")),
		hex.EncodeToString(stakeValue.Bytes()),
		"01",
		"02",
		"03",
	}

	for _, oracleKey := range handler.OraclesKeys {
		aggregatorDeployParams = append(aggregatorDeployParams, oracleKey.MvxAddress.Hex())
	}

	hash := ""
	handler.AggregatorAddress, hash, _ = handler.ChainSimulator.DeploySC(
		ctx,
		aggregatorContractPath,
		handler.OwnerKeys.MvxSk,
		deployGasLimit,
		aggregatorDeployParams,
	)
	require.NotEqual(handler, emptyAddress, handler.AggregatorAddress)
	log.Info("Deploy: aggregator contract", "address", handler.AggregatorAddress, "transaction hash", hash, "num oracles", len(handler.OraclesKeys))

	// deploy wrapper
	handler.WrapperAddress, hash, _ = handler.ChainSimulator.DeploySC(
		ctx,
		wrapperContractPath,
		handler.OwnerKeys.MvxSk,
		deployGasLimit,
		[]string{},
	)
	require.NotEqual(handler, emptyAddress, handler.WrapperAddress)
	log.Info("Deploy: wrapper contract", "address", handler.WrapperAddress, "transaction hash", hash)

	// deploy multi-transfer
	handler.MultiTransferAddress, hash, _ = handler.ChainSimulator.DeploySC(
		ctx,
		multiTransferContractPath,
		handler.OwnerKeys.MvxSk,
		deployGasLimit,
		[]string{},
	)
	require.NotEqual(handler, emptyAddress, handler.MultiTransferAddress)
	log.Info("Deploy: multi-transfer contract", "address", handler.MultiTransferAddress, "transaction hash", hash)

	// deploy safe
	handler.SafeAddress, hash, _ = handler.ChainSimulator.DeploySC(
		ctx,
		safeContractPath,
		handler.OwnerKeys.MvxSk,
		deployGasLimit,
		[]string{
			handler.AggregatorAddress.Hex(),
			handler.MultiTransferAddress.Hex(),
			"01",
		},
	)
	require.NotEqual(handler, emptyAddress, handler.SafeAddress)
	log.Info("Deploy: safe contract", "address", handler.SafeAddress, "transaction hash", hash)

	// deploy bridge proxy
	handler.ScProxyAddress, hash, _ = handler.ChainSimulator.DeploySC(
		ctx,
		bridgeProxyContractPath,
		handler.OwnerKeys.MvxSk,
		deployGasLimit,
		[]string{
			handler.MultiTransferAddress.Hex(),
		},
	)
	require.NotEqual(handler, emptyAddress, handler.ScProxyAddress)
	log.Info("Deploy: SC proxy contract", "address", handler.ScProxyAddress, "transaction hash", hash)

	// deploy multisig
	minRelayerStakeInt, _ := big.NewInt(0).SetString(minRelayerStake, 10)
	minRelayerStakeHex := hex.EncodeToString(minRelayerStakeInt.Bytes())
	params := []string{
		handler.SafeAddress.Hex(),
		handler.MultiTransferAddress.Hex(),
		handler.ScProxyAddress.Hex(),
		minRelayerStakeHex,
		slashAmount,
		handler.Quorum}
	for _, relayerKeys := range handler.RelayersKeys {
		params = append(params, relayerKeys.MvxAddress.Hex())
	}
	handler.MultisigAddress, hash, _ = handler.ChainSimulator.DeploySC(
		ctx,
		multisigContractPath,
		handler.OwnerKeys.MvxSk,
		deployGasLimit,
		params,
	)
	require.NotEqual(handler, emptyAddress, handler.MultisigAddress)
	log.Info("Deploy: multisig contract", "address", handler.MultisigAddress, "transaction hash", hash)

	// deploy test-caller
	handler.CalleeScAddress, hash, _ = handler.ChainSimulator.DeploySC(
		ctx,
		testCallerContractPath,
		handler.OwnerKeys.MvxSk,
		deployGasLimit,
		[]string{},
	)
	require.NotEqual(handler, emptyAddress, handler.CalleeScAddress)
	log.Info("Deploy: test-caller contract", "address", handler.CalleeScAddress, "transaction hash", hash)
}

func (handler *MultiversxHandler) wireMultiTransfer(ctx context.Context) {
	// setBridgeProxyContractAddress
	hash, txResult := handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.MvxSk,
		handler.MultiTransferAddress,
		zeroStringValue,
		setCallsGasLimit,
		setBridgeProxyContractAddressFunction,
		[]string{
			handler.ScProxyAddress.Hex(),
		},
	)
	log.Info("Set in multi-transfer contract the SC proxy contract", "transaction hash", hash, "status", txResult.Status)

	// setWrappingContractAddress
	hash, txResult = handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.MvxSk,
		handler.MultiTransferAddress,
		zeroStringValue,
		setCallsGasLimit,
		setWrappingContractAddressFunction,
		[]string{
			handler.WrapperAddress.Hex(),
		},
	)
	log.Info("Set in multi-transfer contract the wrapper contract", "transaction hash", hash, "status", txResult.Status)
}

func (handler *MultiversxHandler) wireSCProxy(ctx context.Context) {
	// setBridgedTokensWrapper in SC bridge proxy
	hash, txResult := handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.MvxSk,
		handler.ScProxyAddress,
		zeroStringValue,
		setCallsGasLimit,
		setBridgedTokensWrapperAddressFunction,
		[]string{
			handler.WrapperAddress.Hex(),
		},
	)
	log.Info("Set in SC proxy contract the wrapper contract", "transaction hash", hash, "status", txResult.Status)

	// setMultiTransferAddress in SC bridge proxy
	hash, txResult = handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.MvxSk,
		handler.ScProxyAddress,
		zeroStringValue,
		setCallsGasLimit,
		setMultiTransferAddressFunction,
		[]string{
			handler.MultiTransferAddress.Hex(),
		},
	)
	log.Info("Set in SC proxy contract the multi-transfer contract", "transaction hash", hash, "status", txResult.Status)

	// setEsdtSafeAddress on bridge proxy
	hash, txResult = handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.MvxSk,
		handler.ScProxyAddress,
		zeroStringValue,
		setCallsGasLimit,
		setEsdtSafeAddressFunction,
		[]string{
			handler.SafeAddress.Hex(),
		},
	)
	log.Info("Set in SC proxy contract the safe contract", "transaction hash", hash, "status", txResult.Status)
}

func (handler *MultiversxHandler) wireSafe(ctx context.Context) {
	// setBridgedTokensWrapperAddress
	hash, txResult := handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.MvxSk,
		handler.SafeAddress,
		zeroStringValue,
		setCallsGasLimit,
		setBridgedTokensWrapperAddressFunction,
		[]string{
			handler.WrapperAddress.Hex(),
		},
	)
	log.Info("Set in safe contract the wrapper contract", "transaction hash", hash, "status", txResult.Status)

	//setBridgeProxyContractAddress
	hash, txResult = handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.MvxSk,
		handler.SafeAddress,
		zeroStringValue,
		setCallsGasLimit,
		setBridgeProxyContractAddressFunction,
		[]string{
			handler.ScProxyAddress.Hex(),
		},
	)
	log.Info("Set in safe contract the SC proxy contract", "transaction hash", hash, "status", txResult.Status)
}

func (handler *MultiversxHandler) changeOwners(ctx context.Context) {
	// ChangeOwnerAddress for safe
	hash, txResult := handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.MvxSk,
		handler.SafeAddress,
		zeroStringValue,
		setCallsGasLimit,
		changeOwnerAddressFunction,
		[]string{
			handler.MultisigAddress.Hex(),
		},
	)
	log.Info("ChangeOwnerAddress for safe contract", "transaction hash", hash, "status", txResult.Status)

	// ChangeOwnerAddress for multi-transfer
	hash, txResult = handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.MvxSk,
		handler.MultiTransferAddress,
		zeroStringValue,
		setCallsGasLimit,
		changeOwnerAddressFunction,
		[]string{
			handler.MultisigAddress.Hex(),
		},
	)
	log.Info("ChangeOwnerAddress for multi-transfer contract", "transaction hash", hash, "status", txResult.Status)

	// ChangeOwnerAddress for bridge proxy
	hash, txResult = handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.MvxSk,
		handler.ScProxyAddress,
		zeroStringValue,
		setCallsGasLimit,
		changeOwnerAddressFunction,
		[]string{
			handler.MultisigAddress.Hex(),
		},
	)
	log.Info("ChangeOwnerAddress for SC proxy contract", "transaction hash", hash, "status", txResult.Status)
}

func (handler *MultiversxHandler) finishSettings(ctx context.Context) {
	// unpause sc proxy
	hash, txResult := handler.callContractNoParams(ctx, handler.MultisigAddress, unpauseProxyFunction)
	log.Info("Un-paused SC proxy contract", "transaction hash", hash, "status", txResult.Status)

	// setEsdtSafeOnMultiTransfer
	hash, txResult = handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.MvxSk,
		handler.MultisigAddress,
		zeroStringValue,
		setCallsGasLimit,
		setEsdtSafeOnMultiTransferFunction,
		[]string{},
	)
	log.Info("Set in multisig contract the safe contract (automatically)", "transaction hash", hash, "status", txResult.Status)

	// stake relayers on multisig
	handler.stakeAddressesOnContract(ctx, handler.MultisigAddress, handler.RelayersKeys)

	// stake relayers on price aggregator
	handler.stakeAddressesOnContract(ctx, handler.AggregatorAddress, handler.OraclesKeys)

	// unpause multisig
	hash, txResult = handler.callContractNoParams(ctx, handler.MultisigAddress, unpauseFunction)
	log.Info("Un-paused multisig contract", "transaction hash", hash, "status", txResult.Status)

	handler.UnPauseContractsAfterTokenChanges(ctx)
}

// CheckForZeroBalanceOnReceivers will check that the balances for all provided tokens are 0 for the test address and the test SC call address
func (handler *MultiversxHandler) CheckForZeroBalanceOnReceivers(ctx context.Context, tokens ...TestTokenParams) {
	for _, params := range tokens {
		handler.CheckForZeroBalanceOnReceiversForToken(ctx, params)
	}
}

// CheckForZeroBalanceOnReceiversForToken will check that the balance for the test address and the test SC call address is 0
func (handler *MultiversxHandler) CheckForZeroBalanceOnReceiversForToken(ctx context.Context, token TestTokenParams) {
	balance := handler.GetESDTUniversalTokenBalance(ctx, handler.BobKeys.MvxAddress, token.AbstractTokenIdentifier)
	require.Equal(handler, big.NewInt(0).String(), balance.String())

	balance = handler.GetESDTUniversalTokenBalance(ctx, handler.CalleeScAddress, token.AbstractTokenIdentifier)
	require.Equal(handler, big.NewInt(0).String(), balance.String())
}

// GetESDTUniversalTokenBalance will return the universal ESDT token's balance
func (handler *MultiversxHandler) GetESDTUniversalTokenBalance(
	ctx context.Context,
	address *MvxAddress,
	abstractTokenIdentifier string,
) *big.Int {
	token := handler.TokensRegistry.GetTokenData(abstractTokenIdentifier)
	require.NotNil(handler, token)

	balanceString := handler.ChainSimulator.GetESDTBalance(ctx, address, token.MvxUniversalToken)

	balance, ok := big.NewInt(0).SetString(balanceString, 10)
	require.True(handler, ok)

	return balance
}

// GetESDTChainSpecificTokenBalance will return the chain specific ESDT token's balance
func (handler *MultiversxHandler) GetESDTChainSpecificTokenBalance(
	ctx context.Context,
	address *MvxAddress,
	abstractTokenIdentifier string,
) *big.Int {
	token := handler.TokensRegistry.GetTokenData(abstractTokenIdentifier)
	require.NotNil(handler, token)

	balanceString := handler.ChainSimulator.GetESDTBalance(ctx, address, token.MvxChainSpecificToken)

	balance, ok := big.NewInt(0).SetString(balanceString, 10)
	require.True(handler, ok)

	return balance
}

func (handler *MultiversxHandler) callContractNoParams(ctx context.Context, contract *MvxAddress, endpoint string) (string, *data.TransactionOnNetwork) {
	return handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.MvxSk,
		contract,
		zeroStringValue,
		setCallsGasLimit,
		endpoint,
		[]string{},
	)
}

// UnPauseContractsAfterTokenChanges can unpause contracts after token changes
func (handler *MultiversxHandler) UnPauseContractsAfterTokenChanges(ctx context.Context) {
	// unpause safe
	hash, txResult := handler.callContractNoParams(ctx, handler.MultisigAddress, unpauseEsdtSafeFunction)
	log.Info("unpaused safe executed", "hash", hash, "status", txResult.Status)

	// unpause wrapper
	hash, txResult = handler.callContractNoParams(ctx, handler.WrapperAddress, unpauseFunction)
	log.Info("unpaused wrapper executed", "hash", hash, "status", txResult.Status)

	// unpause aggregator
	hash, txResult = handler.callContractNoParams(ctx, handler.AggregatorAddress, unpauseFunction)
	log.Info("unpaused aggregator executed", "hash", hash, "status", txResult.Status)
}

// PauseContractsForTokenChanges can pause contracts for token changes
func (handler *MultiversxHandler) PauseContractsForTokenChanges(ctx context.Context) {
	// pause safe
	hash, txResult := handler.callContractNoParams(ctx, handler.MultisigAddress, pauseEsdtSafeFunction)
	log.Info("paused safe executed", "hash", hash, "status", txResult.Status)

	// pause aggregator
	hash, txResult = handler.callContractNoParams(ctx, handler.AggregatorAddress, pauseFunction)
	log.Info("paused aggregator executed", "hash", hash, "status", txResult.Status)

	// pause wrapper
	hash, txResult = handler.callContractNoParams(ctx, handler.WrapperAddress, pauseFunction)
	log.Info("paused wrapper executed", "hash", hash, "status", txResult.Status)
}

func (handler *MultiversxHandler) stakeAddressesOnContract(ctx context.Context, contract *MvxAddress, allKeys []KeysHolder) {
	for _, keys := range allKeys {
		hash, txResult := handler.ChainSimulator.SendTx(
			ctx,
			keys.MvxSk,
			contract,
			minRelayerStake,
			setCallsGasLimit,
			[]byte(stakeFunction),
		)
		log.Info(fmt.Sprintf("Address %s staked on contract %s with transaction hash %s, status %s", keys.MvxAddress, contract, hash, txResult.Status))
	}
}

// IssueAndWhitelistToken will issue and whitelist the token on MultiversX
func (handler *MultiversxHandler) IssueAndWhitelistToken(ctx context.Context, params IssueTokenParams) {
	if params.HasChainSpecificToken {
		handler.issueAndWhitelistTokensWithChainSpecific(ctx, params)
	} else {
		handler.issueAndWhitelistTokens(ctx, params)
	}
}

func (handler *MultiversxHandler) issueAndWhitelistTokensWithChainSpecific(ctx context.Context, params IssueTokenParams) {
	handler.issueUniversalToken(ctx, params)
	handler.issueChainSpecificToken(ctx, params)
	handler.setLocalRolesForUniversalTokenOnWrapper(ctx, params)
	handler.transferChainSpecificTokenToSCs(ctx, params)
	handler.addUniversalTokenToWrapper(ctx, params)
	handler.whitelistTokenOnWrapper(ctx, params)
	handler.setRolesForSpecificTokenOnSafe(ctx, params)
	handler.addMappingInMultisig(ctx, params)
	handler.whitelistTokenOnMultisig(ctx, params)
	handler.setInitialSupply(ctx, params)
	handler.setPairDecimalsOnAggregator(ctx, params)
	handler.setMaxBridgeAmountOnSafe(ctx, params)
	handler.setMaxBridgeAmountOnMultitransfer(ctx, params)
}

func (handler *MultiversxHandler) issueAndWhitelistTokens(ctx context.Context, params IssueTokenParams) {
	handler.issueUniversalToken(ctx, params)

	tkData := handler.TokensRegistry.GetTokenData(params.AbstractTokenIdentifier)
	handler.TokensRegistry.RegisterChainSpecificToken(params.AbstractTokenIdentifier, tkData.MvxUniversalToken)

	handler.setRolesForSpecificTokenOnSafe(ctx, params)
	handler.addMappingInMultisig(ctx, params)
	handler.whitelistTokenOnMultisig(ctx, params)
	handler.setInitialSupply(ctx, params)
	handler.setPairDecimalsOnAggregator(ctx, params)
	handler.setMaxBridgeAmountOnSafe(ctx, params)
	handler.setMaxBridgeAmountOnMultitransfer(ctx, params)
}

func (handler *MultiversxHandler) issueUniversalToken(ctx context.Context, params IssueTokenParams) {
	token := handler.TokensRegistry.GetTokenData(params.AbstractTokenIdentifier)
	require.NotNil(handler, token)

	valueToMintInt, ok := big.NewInt(0).SetString(params.ValueToMintOnMvx, 10)
	require.True(handler, ok)

	// issue universal token
	hash, txResult := handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.MvxSk,
		handler.ESDTSystemContractAddress,
		esdtIssueCost,
		issueTokenGasLimit,
		issueFunction,
		[]string{
			hex.EncodeToString([]byte(params.MvxUniversalTokenDisplayName)),
			hex.EncodeToString([]byte(params.MvxUniversalTokenTicker)),
			hex.EncodeToString(valueToMintInt.Bytes()),
			fmt.Sprintf("%02x", params.NumOfDecimalsUniversal),
			hex.EncodeToString([]byte(canAddSpecialRoles)),
			hex.EncodeToString([]byte(trueStr))})
	mvxUniversalToken := handler.getTokenNameFromResult(*txResult)
	require.Greater(handler, len(mvxUniversalToken), 0)
	handler.TokensRegistry.RegisterUniversalToken(params.AbstractTokenIdentifier, mvxUniversalToken)
	log.Info("issue universal token tx executed", "hash", hash, "status", txResult.Status, "token", mvxUniversalToken, "owner", handler.OwnerKeys.MvxAddress)
}

func (handler *MultiversxHandler) issueChainSpecificToken(ctx context.Context, params IssueTokenParams) {
	valueToMintInt, ok := big.NewInt(0).SetString(params.ValueToMintOnMvx, 10)
	require.True(handler, ok)

	hash, txResult := handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.MvxSk,
		handler.ESDTSystemContractAddress,
		esdtIssueCost,
		issueTokenGasLimit,
		issueFunction,
		[]string{
			hex.EncodeToString([]byte(params.MvxChainSpecificTokenDisplayName)),
			hex.EncodeToString([]byte(params.MvxChainSpecificTokenTicker)),
			hex.EncodeToString(valueToMintInt.Bytes()),
			fmt.Sprintf("%02x", params.NumOfDecimalsChainSpecific),
			hex.EncodeToString([]byte(canAddSpecialRoles)),
			hex.EncodeToString([]byte(trueStr))})
	mvxChainSpecificToken := handler.getTokenNameFromResult(*txResult)
	require.Greater(handler, len(mvxChainSpecificToken), 0)
	handler.TokensRegistry.RegisterChainSpecificToken(params.AbstractTokenIdentifier, mvxChainSpecificToken)
	log.Info("issue chain specific token tx executed", "hash", hash, "status", txResult.Status, "token", mvxChainSpecificToken, "owner", handler.OwnerKeys.MvxAddress)
}

func (handler *MultiversxHandler) setLocalRolesForUniversalTokenOnWrapper(ctx context.Context, params IssueTokenParams) {
	tkData := handler.TokensRegistry.GetTokenData(params.AbstractTokenIdentifier)

	// set local roles bridged tokens wrapper
	hash, txResult := handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.MvxSk,
		handler.ESDTSystemContractAddress,
		zeroStringValue,
		setCallsGasLimit,
		setSpecialRoleFunction,
		[]string{
			hex.EncodeToString([]byte(tkData.MvxUniversalToken)),
			handler.WrapperAddress.Hex(),
			hex.EncodeToString([]byte(esdtRoleLocalMint)),
			hex.EncodeToString([]byte(esdtRoleLocalBurn))})
	log.Info("set local roles bridged tokens wrapper tx executed", "hash", hash, "status", txResult.Status)
}

func (handler *MultiversxHandler) transferChainSpecificTokenToSCs(ctx context.Context, params IssueTokenParams) {
	valueToMintInt, ok := big.NewInt(0).SetString(params.ValueToMintOnMvx, 10)
	require.True(handler, ok)

	tkData := handler.TokensRegistry.GetTokenData(params.AbstractTokenIdentifier)

	// transfer to wrapper sc
	initialMintValue := valueToMintInt.Div(valueToMintInt, big.NewInt(3))
	hash, txResult := handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.MvxSk,
		handler.WrapperAddress,
		zeroStringValue,
		setCallsGasLimit,
		esdtTransferFunction,
		[]string{
			hex.EncodeToString([]byte(tkData.MvxChainSpecificToken)),
			hex.EncodeToString(initialMintValue.Bytes()),
			hex.EncodeToString([]byte(depositLiquidityFunction))})
	log.Info("transfer to wrapper sc tx executed", "hash", hash, "status", txResult.Status)

	// transfer to safe sc
	hash, txResult = handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.MvxSk,
		handler.SafeAddress,
		zeroStringValue,
		setCallsGasLimit,
		esdtTransferFunction,
		[]string{
			hex.EncodeToString([]byte(tkData.MvxChainSpecificToken)),
			hex.EncodeToString(initialMintValue.Bytes())})
	log.Info("transfer to safe sc tx executed", "hash", hash, "status", txResult.Status)
}

func (handler *MultiversxHandler) addUniversalTokenToWrapper(ctx context.Context, params IssueTokenParams) {
	tkData := handler.TokensRegistry.GetTokenData(params.AbstractTokenIdentifier)

	// add wrapped token
	hash, txResult := handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.MvxSk,
		handler.WrapperAddress,
		zeroStringValue,
		setCallsGasLimit,
		addWrappedTokenFunction,
		[]string{
			hex.EncodeToString([]byte(tkData.MvxUniversalToken)),
			fmt.Sprintf("%02x", params.NumOfDecimalsUniversal),
		})
	log.Info("add wrapped token tx executed", "hash", hash, "status", txResult.Status)
}

func (handler *MultiversxHandler) whitelistTokenOnWrapper(ctx context.Context, params IssueTokenParams) {
	tkData := handler.TokensRegistry.GetTokenData(params.AbstractTokenIdentifier)

	// wrapper whitelist token
	hash, txResult := handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.MvxSk,
		handler.WrapperAddress,
		zeroStringValue,
		setCallsGasLimit,
		whitelistTokenFunction,
		[]string{
			hex.EncodeToString([]byte(tkData.MvxChainSpecificToken)),
			fmt.Sprintf("%02x", params.NumOfDecimalsChainSpecific),
			hex.EncodeToString([]byte(tkData.MvxUniversalToken))})
	log.Info("wrapper whitelist token tx executed", "hash", hash, "status", txResult.Status)
}

func (handler *MultiversxHandler) setRolesForSpecificTokenOnSafe(ctx context.Context, params IssueTokenParams) {
	tkData := handler.TokensRegistry.GetTokenData(params.AbstractTokenIdentifier)

	// set local roles esdt safe
	hash, txResult := handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.MvxSk,
		handler.ESDTSystemContractAddress,
		zeroStringValue,
		setCallsGasLimit,
		setSpecialRoleFunction,
		[]string{
			hex.EncodeToString([]byte(tkData.MvxChainSpecificToken)),
			handler.SafeAddress.Hex(),
			hex.EncodeToString([]byte(esdtRoleLocalMint)),
			hex.EncodeToString([]byte(esdtRoleLocalBurn))})
	log.Info("set local roles esdt safe tx executed", "hash", hash, "status", txResult.Status)
}

func (handler *MultiversxHandler) addMappingInMultisig(ctx context.Context, params IssueTokenParams) {
	tkData := handler.TokensRegistry.GetTokenData(params.AbstractTokenIdentifier)

	// add mapping
	hash, txResult := handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.MvxSk,
		handler.MultisigAddress,
		zeroStringValue,
		setCallsGasLimit,
		addMappingFunction,
		[]string{
			hex.EncodeToString(tkData.EthErc20Address.Bytes()),
			hex.EncodeToString([]byte(tkData.MvxChainSpecificToken))})
	log.Info("add mapping tx executed", "hash", hash, "status", txResult.Status)
}

func (handler *MultiversxHandler) whitelistTokenOnMultisig(ctx context.Context, params IssueTokenParams) {
	tkData := handler.TokensRegistry.GetTokenData(params.AbstractTokenIdentifier)

	// whitelist token
	hash, txResult := handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.MvxSk,
		handler.MultisigAddress,
		zeroStringValue,
		setCallsGasLimit,
		esdtSafeAddTokenToWhitelistFunction,
		[]string{
			hex.EncodeToString([]byte(tkData.MvxChainSpecificToken)),
			hex.EncodeToString([]byte(params.MvxChainSpecificTokenTicker)),
			getHexBool(params.IsMintBurnOnMvX),
			getHexBool(params.IsNativeOnMvX),
			hex.EncodeToString(zeroValueBigInt.Bytes()), // total_balance
			hex.EncodeToString(zeroValueBigInt.Bytes()), // mint_balance
			hex.EncodeToString(zeroValueBigInt.Bytes()), // burn_balance
		})
	log.Info("whitelist token tx executed", "hash", hash, "status", txResult.Status)
}

func (handler *MultiversxHandler) setInitialSupply(ctx context.Context, params IssueTokenParams) {
	tkData := handler.TokensRegistry.GetTokenData(params.AbstractTokenIdentifier)

	// set initial supply
	if len(params.InitialSupplyValue) > 0 {
		initialSupply, okConvert := big.NewInt(0).SetString(params.InitialSupplyValue, 10)
		require.True(handler, okConvert)

		if params.IsMintBurnOnMvX {
			mintAmount := big.NewInt(0)
			burnAmount := big.NewInt(0)

			if params.IsNativeOnMvX {
				burnAmount = initialSupply
			} else {
				mintAmount = initialSupply
			}

			hash, txResult := handler.ChainSimulator.ScCall(
				ctx,
				handler.OwnerKeys.MvxSk,
				handler.MultisigAddress,
				zeroStringValue,
				setCallsGasLimit,
				initSupplyMintBurnEsdtSafe,
				[]string{
					hex.EncodeToString([]byte(tkData.MvxChainSpecificToken)),
					hex.EncodeToString(mintAmount.Bytes()),
					hex.EncodeToString(burnAmount.Bytes()),
				},
			)

			log.Info("initial supply tx executed", "hash", hash, "status", txResult.Status,
				"initial mint", mintAmount.String(), "initial burned", burnAmount.String())
		} else {
			hash, txResult := handler.ChainSimulator.ScCall(
				ctx,
				handler.OwnerKeys.MvxSk,
				handler.MultisigAddress,
				zeroStringValue,
				setCallsGasLimit,
				esdtTransferFunction,
				[]string{
					hex.EncodeToString([]byte(tkData.MvxChainSpecificToken)),
					hex.EncodeToString(initialSupply.Bytes()),
					hex.EncodeToString([]byte(initSupplyEsdtSafe)),
					hex.EncodeToString([]byte(tkData.MvxChainSpecificToken)),
					hex.EncodeToString(initialSupply.Bytes()),
				})

			log.Info("initial supply tx executed", "hash", hash, "status", txResult.Status,
				"initial value", params.InitialSupplyValue)
		}
	}
}

func (handler *MultiversxHandler) setPairDecimalsOnAggregator(ctx context.Context, params IssueTokenParams) {
	// setPairDecimals on aggregator
	hash, txResult := handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.MvxSk,
		handler.AggregatorAddress,
		zeroStringValue,
		setCallsGasLimit,
		setPairDecimalsFunction,
		[]string{
			hex.EncodeToString([]byte(gwei)),
			hex.EncodeToString([]byte(params.MvxChainSpecificTokenTicker)),
			fmt.Sprintf("%02x", params.NumOfDecimalsChainSpecific)})
	log.Info("setPairDecimals tx executed", "hash", hash, "status", txResult.Status)
}

func (handler *MultiversxHandler) setMaxBridgeAmountOnSafe(ctx context.Context, params IssueTokenParams) {
	tkData := handler.TokensRegistry.GetTokenData(params.AbstractTokenIdentifier)

	// safe set max bridge amount for token
	maxBridgedAmountForTokenInt, _ := big.NewInt(0).SetString(maxBridgedAmountForToken, 10)
	hash, txResult := handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.MvxSk,
		handler.MultisigAddress,
		zeroStringValue,
		setCallsGasLimit,
		esdtSafeSetMaxBridgedAmountForTokenFunction,
		[]string{
			hex.EncodeToString([]byte(tkData.MvxChainSpecificToken)),
			hex.EncodeToString(maxBridgedAmountForTokenInt.Bytes())})
	log.Info("safe set max bridge amount for token tx executed", "hash", hash, "status", txResult.Status)
}

func (handler *MultiversxHandler) setMaxBridgeAmountOnMultitransfer(ctx context.Context, params IssueTokenParams) {
	tkData := handler.TokensRegistry.GetTokenData(params.AbstractTokenIdentifier)

	// multi-transfer set max bridge amount for token
	maxBridgedAmountForTokenInt, _ := big.NewInt(0).SetString(maxBridgedAmountForToken, 10)
	hash, txResult := handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.MvxSk,
		handler.MultisigAddress,
		zeroStringValue,
		setCallsGasLimit,
		multiTransferEsdtSetMaxBridgedAmountForTokenFunction,
		[]string{
			hex.EncodeToString([]byte(tkData.MvxChainSpecificToken)),
			hex.EncodeToString(maxBridgedAmountForTokenInt.Bytes())})
	log.Info("multi-transfer set max bridge amount for token tx executed", "hash", hash, "status", txResult.Status)
}

func (handler *MultiversxHandler) getTokenNameFromResult(txResult data.TransactionOnNetwork) string {
	for _, event := range txResult.Logs.Events {
		if event.Identifier == issueFunction {
			require.Greater(handler, len(event.Topics), 1)

			return string(event.Topics[0])
		}
	}

	require.Fail(handler, "did not find the event with the issue identifier")
	return ""
}

// SubmitAggregatorBatch will submit the aggregator batch
func (handler *MultiversxHandler) SubmitAggregatorBatch(ctx context.Context, params IssueTokenParams) {
	txHashes := make([]string, 0, len(handler.OraclesKeys))
	for _, key := range handler.OraclesKeys {
		hash := handler.submitAggregatorBatchForKey(ctx, key, params)
		txHashes = append(txHashes, hash)
	}

	for _, hash := range txHashes {
		txResult := handler.ChainSimulator.GetTransactionResult(ctx, hash)
		log.Info("submit aggregator batch tx", "hash", hash, "status", txResult.Status)
	}
}

func (handler *MultiversxHandler) submitAggregatorBatchForKey(ctx context.Context, key KeysHolder, params IssueTokenParams) string {
	timestamp := handler.ChainSimulator.GetBlockchainTimeStamp(ctx)
	require.Greater(handler, timestamp, uint64(0), "something went wrong and the chain simulator returned 0 for the current timestamp")

	timestampAsBigInt := big.NewInt(0).SetUint64(timestamp)

	hash := handler.ChainSimulator.ScCallWithoutGenerateBlocks(
		ctx,
		key.MvxSk,
		handler.AggregatorAddress,
		zeroStringValue,
		setCallsGasLimit,
		submitBatchFunction,
		[]string{
			hex.EncodeToString([]byte(gwei)),
			hex.EncodeToString([]byte(params.MvxChainSpecificTokenTicker)),
			hex.EncodeToString(timestampAsBigInt.Bytes()),
			hex.EncodeToString(feeInt.Bytes()),
			fmt.Sprintf("%02x", params.NumOfDecimalsChainSpecific)})

	log.Info("submit aggregator batch tx sent", "transaction hash", hash, "submitter", key.MvxAddress.Bech32())

	return hash
}

// SendDepositTransactionFromMultiversx will send the deposit transaction from MultiversX
func (handler *MultiversxHandler) SendDepositTransactionFromMultiversx(ctx context.Context, from KeysHolder, to KeysHolder, token *TokenData, params TestTokenParams, value *big.Int) {
	if params.HasChainSpecificToken {
		handler.unwrapCreateTransaction(ctx, token, from, to, value)
		return
	}

	handler.createTransactionWithoutUnwrap(ctx, token, from, to, value)
}

func (handler *MultiversxHandler) createTransactionWithoutUnwrap(
	ctx context.Context,
	token *TokenData,
	from KeysHolder,
	to KeysHolder,
	value *big.Int,
) {
	// create transaction params
	params := []string{
		hex.EncodeToString([]byte(token.MvxUniversalToken)),
		hex.EncodeToString(value.Bytes()),
		hex.EncodeToString([]byte(createTransactionFunction)),
		hex.EncodeToString(to.EthAddress.Bytes()),
	}
	dataField := strings.Join(params, "@")

	hash, txResult := handler.ChainSimulator.ScCall(
		ctx,
		from.MvxSk,
		handler.SafeAddress,
		zeroStringValue,
		createDepositGasLimit+gasLimitPerDataByte*uint64(len(dataField)),
		esdtTransferFunction,
		params,
	)
	log.Info("MultiversX->Ethereum createTransaction sent", "hash", hash, "token", token.MvxUniversalToken, "status", txResult.Status)
}

func (handler *MultiversxHandler) unwrapCreateTransaction(ctx context.Context, token *TokenData, from KeysHolder, to KeysHolder, value *big.Int) {
	// create transaction params
	params := []string{
		hex.EncodeToString([]byte(token.MvxUniversalToken)),
		hex.EncodeToString(value.Bytes()),
		hex.EncodeToString([]byte(unwrapTokenCreateTransactionFunction)),
		hex.EncodeToString([]byte(token.MvxChainSpecificToken)),
		hex.EncodeToString(handler.SafeAddress.Bytes()),
		hex.EncodeToString(to.EthAddress.Bytes()),
	}
	dataField := strings.Join(params, "@")

	hash, txResult := handler.ChainSimulator.ScCall(
		ctx,
		from.MvxSk,
		handler.WrapperAddress,
		zeroStringValue,
		createDepositGasLimit+gasLimitPerDataByte*uint64(len(dataField)),
		esdtTransferFunction,
		params,
	)
	log.Info("MultiversX->Ethereum unwrapCreateTransaction sent", "hash", hash, "token", token.MvxUniversalToken, "status", txResult.Status)
}

// TestWithdrawFees will try to withdraw the fees for the provided token from the safe contract to the owner
func (handler *MultiversxHandler) TestWithdrawFees(
	ctx context.Context,
	token string,
	expectedDeltaForRefund *big.Int,
	expectedDeltaForAccumulated *big.Int,
) {
	handler.withdrawFees(ctx, token, expectedDeltaForRefund, getRefundFeesForEthereumFunction, withdrawRefundFeesForEthereumFunction)
	handler.withdrawFees(ctx, token, expectedDeltaForAccumulated, getTransactionFeesFunction, withdrawTransactionFeesFunction)
}

func (handler *MultiversxHandler) withdrawFees(ctx context.Context,
	token string,
	expectedDelta *big.Int,
	getFunction string,
	withdrawFunction string,
) {
	queryParams := []string{
		hex.EncodeToString([]byte(token)),
	}
	responseData := handler.ChainSimulator.ExecuteVMQuery(ctx, handler.SafeAddress, getFunction, queryParams)
	require.Greater(handler, len(responseData), 0)
	value := big.NewInt(0).SetBytes(responseData[0])
	require.Equal(handler, expectedDelta.String(), value.String())
	if expectedDelta.Cmp(zeroValueBigInt) == 0 {
		return
	}

	handler.ChainSimulator.GenerateBlocks(ctx, 5) // ensure block finality
	initialBalanceStr := handler.ChainSimulator.GetESDTBalance(ctx, handler.OwnerKeys.MvxAddress, token)
	initialBalance, ok := big.NewInt(0).SetString(initialBalanceStr, 10)
	require.True(handler, ok)

	handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.MvxSk,
		handler.MultisigAddress,
		zeroStringValue,
		generalSCCallGasLimit,
		withdrawFunction,
		[]string{
			hex.EncodeToString([]byte(token)),
		},
	)

	handler.ChainSimulator.GenerateBlocks(ctx, 5) // ensure block finality
	finalBalanceStr := handler.ChainSimulator.GetESDTBalance(ctx, handler.OwnerKeys.MvxAddress, token)
	finalBalance, ok := big.NewInt(0).SetString(finalBalanceStr, 10)
	require.True(handler, ok)

	require.Equal(handler, expectedDelta, finalBalance.Sub(finalBalance, initialBalance),
		fmt.Sprintf("mismatch on balance check after the call to %s: initial balance: %s, final balance %s, expected delta: %s",
			withdrawFunction, initialBalanceStr, finalBalanceStr, expectedDelta.String()))
}

// TransferToken is able to create an ESDT transfer
func (handler *MultiversxHandler) TransferToken(ctx context.Context, source KeysHolder, receiver KeysHolder, amount *big.Int, params IssueTokenParams) {
	tkData := handler.TokensRegistry.GetTokenData(params.AbstractTokenIdentifier)

	// transfer to receiver, so it will have funds to carry on with the deposits
	hash, txResult := handler.ChainSimulator.ScCall(
		ctx,
		source.MvxSk,
		receiver.MvxAddress,
		zeroStringValue,
		createDepositGasLimit,
		esdtTransferFunction,
		[]string{
			hex.EncodeToString([]byte(tkData.MvxUniversalToken)),
			hex.EncodeToString(amount.Bytes())})

	log.Info("transfer to tx executed",
		"source address", source.MvxAddress.Bech32(),
		"receiver", receiver.MvxAddress.Bech32(),
		"token", tkData.MvxUniversalToken,
		"amount", amount.String(),
		"hash", hash, "status", txResult.Status)
}

// GetTotalBalancesForToken will return the total locked balance for the provided token
func (handler *MultiversxHandler) GetTotalBalancesForToken(ctx context.Context, token string) *big.Int {
	queryParams := []string{
		hex.EncodeToString([]byte(token)),
	}
	responseData := handler.ChainSimulator.ExecuteVMQuery(ctx, handler.SafeAddress, getTotalBalances, queryParams)
	require.Greater(handler, len(responseData), 0)
	value := big.NewInt(0).SetBytes(responseData[0])
	return value
}

// GetMintedAmountForToken will return mint balance for token
func (handler *MultiversxHandler) GetMintedAmountForToken(ctx context.Context, token string) *big.Int {
	queryParams := []string{
		hex.EncodeToString([]byte(token)),
	}
	responseData := handler.ChainSimulator.ExecuteVMQuery(ctx, handler.SafeAddress, getMintBalances, queryParams)
	require.Greater(handler, len(responseData), 0)
	value := big.NewInt(0).SetBytes(responseData[0])
	return value
}

// GetBurnedAmountForToken will return burn balance of token
func (handler *MultiversxHandler) GetBurnedAmountForToken(ctx context.Context, token string) *big.Int {
	queryParams := []string{
		hex.EncodeToString([]byte(token)),
	}
	responseData := handler.ChainSimulator.ExecuteVMQuery(ctx, handler.SafeAddress, getBurnBalances, queryParams)
	require.Greater(handler, len(responseData), 0)
	value := big.NewInt(0).SetBytes(responseData[0])
	return value
}

func getHexBool(input bool) string {
	if input {
		return hexTrue
	}

	return hexFalse
}

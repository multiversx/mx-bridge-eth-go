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
	setEsdtSafeOnWrapperFunction                         = "setEsdtSafeContractAddress"
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
	createTransactionFunction                            = "createTransaction"
	unwrapTokenFunction                                  = "unwrapToken"
	setBridgedTokensWrapperAddressFunction               = "setBridgedTokensWrapperAddress"
	setMultiTransferAddressFunction                      = "setMultiTransferAddress"
	withdrawRefundFeesForEthereumFunction                = "withdrawRefundFeesForEthereum"
	getRefundFeesForEthereumFunction                     = "getRefundFeesForEthereum"
	withdrawTransactionFeesFunction                      = "withdrawTransactionFees"
	getTransactionFeesFunction                           = "getTransactionFees"
	initSupplyMintBurnEsdtSafe                           = "initSupplyMintBurnEsdtSafe"
	initSupplyEsdtSafe                                   = "initSupplyEsdtSafe"
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

	AggregatorAddress    *MvxAddress
	WrapperAddress       *MvxAddress
	SafeAddress          *MvxAddress
	MultisigAddress      *MvxAddress
	MultiTransferAddress *MvxAddress
	ScProxyAddress       *MvxAddress
	TestCallerAddress    *MvxAddress
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
	handler.wireWrapper(ctx)
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
	handler.TestCallerAddress, hash, _ = handler.ChainSimulator.DeploySC(
		ctx,
		testCallerContractPath,
		handler.OwnerKeys.MvxSk,
		deployGasLimit,
		[]string{},
	)
	require.NotEqual(handler, emptyAddress, handler.TestCallerAddress)
	log.Info("Deploy: test-caller contract", "address", handler.TestCallerAddress, "transaction hash", hash)
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

func (handler *MultiversxHandler) wireWrapper(ctx context.Context) {
	// setEsdtSafeOnWrapper
	hash, txResult := handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.MvxSk,
		handler.WrapperAddress,
		zeroStringValue,
		setCallsGasLimit,
		setEsdtSafeOnWrapperFunction,
		[]string{
			handler.SafeAddress.Hex(),
		},
	)
	log.Info("Set in wrapper contract the safe contract", "transaction hash", hash, "status", txResult.Status)

	// setBridgeProxyContractAddress
	hash, txResult = handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.MvxSk,
		handler.WrapperAddress,
		zeroStringValue,
		setCallsGasLimit,
		setBridgeProxyContractAddressFunction,
		[]string{
			handler.ScProxyAddress.Hex(),
		},
	)
	log.Info("Set in wrapper contract the SC proxy contract", "transaction hash", hash, "status", txResult.Status)
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
	balance := handler.GetESDTUniversalTokenBalance(ctx, handler.TestKeys.MvxAddress, token.AbstractTokenIdentifier)
	require.Equal(handler, big.NewInt(0).String(), balance.String())

	balance = handler.GetESDTUniversalTokenBalance(ctx, handler.TestCallerAddress, token.AbstractTokenIdentifier)
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
	token := handler.TokensRegistry.GetTokenData(params.AbstractTokenIdentifier)
	require.NotNil(handler, token)

	esdtAddress := NewMvxAddressFromBech32(handler, esdtSystemSCAddress)

	// issue universal token
	hash, txResult := handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.MvxSk,
		esdtAddress,
		esdtIssueCost,
		issueTokenGasLimit,
		issueFunction,
		[]string{
			hex.EncodeToString([]byte(params.MvxUniversalTokenDisplayName)),
			hex.EncodeToString([]byte(params.MvxUniversalTokenTicker)),
			"00",
			fmt.Sprintf("%02x", params.NumOfDecimalsUniversal),
			hex.EncodeToString([]byte(canAddSpecialRoles)),
			hex.EncodeToString([]byte(trueStr))})
	mvxUniversalToken := handler.getTokenNameFromResult(*txResult)
	require.Greater(handler, len(mvxUniversalToken), 0)
	handler.TokensRegistry.RegisterUniversalToken(params.AbstractTokenIdentifier, mvxUniversalToken)
	log.Info("issue universal token tx executed", "hash", hash, "status", txResult.Status, "token", mvxUniversalToken, "owner", handler.OwnerKeys.MvxAddress)

	// issue chain specific token
	valueToMintInt, ok := big.NewInt(0).SetString(params.ValueToMintOnMvx, 10)
	require.True(handler, ok)

	hash, txResult = handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.MvxSk,
		esdtAddress,
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

	// set local roles bridged tokens wrapper
	hash, txResult = handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.MvxSk,
		esdtAddress,
		zeroStringValue,
		setCallsGasLimit,
		setSpecialRoleFunction,
		[]string{
			hex.EncodeToString([]byte(mvxUniversalToken)),
			handler.WrapperAddress.Hex(),
			hex.EncodeToString([]byte(esdtRoleLocalMint)),
			hex.EncodeToString([]byte(esdtRoleLocalBurn))})
	log.Info("set local roles bridged tokens wrapper tx executed", "hash", hash, "status", txResult.Status)

	// transfer to wrapper sc
	initialMintValue := valueToMintInt.Div(valueToMintInt, big.NewInt(3))
	hash, txResult = handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.MvxSk,
		handler.WrapperAddress,
		zeroStringValue,
		setCallsGasLimit,
		esdtTransferFunction,
		[]string{
			hex.EncodeToString([]byte(mvxChainSpecificToken)),
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
			hex.EncodeToString([]byte(mvxChainSpecificToken)),
			hex.EncodeToString(initialMintValue.Bytes())})
	log.Info("transfer to safe sc tx executed", "hash", hash, "status", txResult.Status)

	// add wrapped token
	hash, txResult = handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.MvxSk,
		handler.WrapperAddress,
		zeroStringValue,
		setCallsGasLimit,
		addWrappedTokenFunction,
		[]string{
			hex.EncodeToString([]byte(mvxUniversalToken)),
			fmt.Sprintf("%02x", params.NumOfDecimalsUniversal),
		})
	log.Info("add wrapped token tx executed", "hash", hash, "status", txResult.Status)

	// wrapper whitelist token
	hash, txResult = handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.MvxSk,
		handler.WrapperAddress,
		zeroStringValue,
		setCallsGasLimit,
		whitelistTokenFunction,
		[]string{
			hex.EncodeToString([]byte(mvxChainSpecificToken)),
			fmt.Sprintf("%02x", params.NumOfDecimalsChainSpecific),
			hex.EncodeToString([]byte(mvxUniversalToken))})
	log.Info("wrapper whitelist token tx executed", "hash", hash, "status", txResult.Status)

	// set local roles esdt safe
	hash, txResult = handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.MvxSk,
		esdtAddress,
		zeroStringValue,
		setCallsGasLimit,
		setSpecialRoleFunction,
		[]string{
			hex.EncodeToString([]byte(mvxChainSpecificToken)),
			handler.SafeAddress.Hex(),
			hex.EncodeToString([]byte(esdtRoleLocalMint)),
			hex.EncodeToString([]byte(esdtRoleLocalBurn))})
	log.Info("set local roles esdt safe tx executed", "hash", hash, "status", txResult.Status)

	// add mapping
	hash, txResult = handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.MvxSk,
		handler.MultisigAddress,
		zeroStringValue,
		setCallsGasLimit,
		addMappingFunction,
		[]string{
			hex.EncodeToString(token.EthErc20Address.Bytes()),
			hex.EncodeToString([]byte(mvxChainSpecificToken))})
	log.Info("add mapping tx executed", "hash", hash, "status", txResult.Status)

	// whitelist token
	hash, txResult = handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.MvxSk,
		handler.MultisigAddress,
		zeroStringValue,
		setCallsGasLimit,
		esdtSafeAddTokenToWhitelistFunction,
		[]string{
			hex.EncodeToString([]byte(mvxChainSpecificToken)),
			hex.EncodeToString([]byte(params.MvxChainSpecificTokenTicker)),
			getHexBool(params.IsMintBurnOnMvX),
			getHexBool(params.IsNativeOnMvX),
			hex.EncodeToString(zeroValueBigInt.Bytes()), // total_balance
			hex.EncodeToString(zeroValueBigInt.Bytes()), // mint_balance
			hex.EncodeToString(zeroValueBigInt.Bytes()), // burn_balance
		})
	log.Info("whitelist token tx executed", "hash", hash, "status", txResult.Status)

	// set initial supply
	if len(params.InitialSupplyValue) > 0 {
		initialSupply, okConvert := big.NewInt(0).SetString(params.InitialSupplyValue, 10)
		require.True(handler, okConvert)

		if params.IsMintBurnOnMvX {
			hash, txResult = handler.ChainSimulator.ScCall(
				ctx,
				handler.OwnerKeys.MvxSk,
				handler.MultisigAddress,
				zeroStringValue,
				setCallsGasLimit,
				initSupplyMintBurnEsdtSafe,
				[]string{
					hex.EncodeToString([]byte(mvxChainSpecificToken)),
					hex.EncodeToString(initialSupply.Bytes()),
					hex.EncodeToString([]byte{0}),
				},
			)
			log.Info("initial supply tx executed", "hash", hash, "status", txResult.Status,
				"initial mint", params.InitialSupplyValue, "initial burned", "0")
		} else {
			hash, txResult = handler.ChainSimulator.ScCall(
				ctx,
				handler.OwnerKeys.MvxSk,
				handler.MultisigAddress,
				zeroStringValue,
				setCallsGasLimit,
				esdtTransferFunction,
				[]string{
					hex.EncodeToString([]byte(mvxChainSpecificToken)),
					hex.EncodeToString(initialSupply.Bytes()),
					hex.EncodeToString([]byte(initSupplyEsdtSafe)),
					hex.EncodeToString([]byte(mvxChainSpecificToken)),
					hex.EncodeToString(initialSupply.Bytes()),
				})

			log.Info("initial supply tx executed", "hash", hash, "status", txResult.Status,
				"initial value", params.InitialSupplyValue)
		}
	}

	// setPairDecimals on aggregator
	hash, txResult = handler.ChainSimulator.ScCall(
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

	// safe set max bridge amount for token
	maxBridgedAmountForTokenInt, _ := big.NewInt(0).SetString(maxBridgedAmountForToken, 10)
	hash, txResult = handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.MvxSk,
		handler.MultisigAddress,
		zeroStringValue,
		setCallsGasLimit,
		esdtSafeSetMaxBridgedAmountForTokenFunction,
		[]string{
			hex.EncodeToString([]byte(mvxChainSpecificToken)),
			hex.EncodeToString(maxBridgedAmountForTokenInt.Bytes())})
	log.Info("safe set max bridge amount for token tx executed", "hash", hash, "status", txResult.Status)

	// multi-transfer set max bridge amount for token
	hash, txResult = handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.MvxSk,
		handler.MultisigAddress,
		zeroStringValue,
		setCallsGasLimit,
		multiTransferEsdtSetMaxBridgedAmountForTokenFunction,
		[]string{
			hex.EncodeToString([]byte(mvxChainSpecificToken)),
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

// CreateDepositsOnMultiversxForToken will send the deposit transactions on MultiversX returning how many tokens should be minted on Ethereum
func (handler *MultiversxHandler) CreateDepositsOnMultiversxForToken(
	ctx context.Context,
	params TestTokenParams,
) *big.Int {
	token := handler.TokensRegistry.GetTokenData(params.AbstractTokenIdentifier)
	require.NotNil(handler, token)

	valueToMintOnEthereum := big.NewInt(0)
	for _, operation := range params.TestOperations {
		if operation.ValueToSendFromMvX == nil {
			continue
		}

		valueToMintOnEthereum.Add(valueToMintOnEthereum, operation.ValueToSendFromMvX)

		// transfer to sender tx
		hash, txResult := handler.ChainSimulator.ScCall(
			ctx,
			handler.OwnerKeys.MvxSk,
			handler.TestKeys.MvxAddress,
			zeroStringValue,
			createDepositGasLimit,
			esdtTransferFunction,
			[]string{
				hex.EncodeToString([]byte(token.MvxChainSpecificToken)),
				hex.EncodeToString(operation.ValueToSendFromMvX.Bytes())})
		log.Info("transfer to sender tx executed", "hash", hash, "status", txResult.Status)

		// send tx to safe contract
		scCallParams := []string{
			hex.EncodeToString([]byte(token.MvxChainSpecificToken)),
			hex.EncodeToString(operation.ValueToSendFromMvX.Bytes()),
			hex.EncodeToString([]byte(createTransactionFunction)),
			hex.EncodeToString(handler.TestKeys.EthAddress.Bytes()),
		}
		dataField := strings.Join(scCallParams, "@")

		hash, txResult = handler.ChainSimulator.ScCall(
			ctx,
			handler.TestKeys.MvxSk,
			handler.SafeAddress,
			zeroStringValue,
			createDepositGasLimit+gasLimitPerDataByte*uint64(len(dataField)),
			esdtTransferFunction,
			scCallParams)
		log.Info("MultiversX->Ethereum transaction sent", "hash", hash, "status", txResult.Status)
	}

	return valueToMintOnEthereum
}

// SendDepositTransactionFromMultiversx will send the deposit transaction from MultiversX
func (handler *MultiversxHandler) SendDepositTransactionFromMultiversx(ctx context.Context, token *TokenData, value *big.Int) {
	// unwrap token
	paramsUnwrap := []string{
		hex.EncodeToString([]byte(token.MvxUniversalToken)),
		hex.EncodeToString(value.Bytes()),
		hex.EncodeToString([]byte(unwrapTokenFunction)),
		hex.EncodeToString([]byte(token.MvxChainSpecificToken)),
	}

	hash, txResult := handler.ChainSimulator.ScCall(
		ctx,
		handler.TestKeys.MvxSk,
		handler.WrapperAddress,
		zeroStringValue,
		createDepositGasLimit,
		esdtTransferFunction,
		paramsUnwrap,
	)
	log.Info("unwrap transaction sent", "hash", hash, "token", token.MvxUniversalToken, "status", txResult.Status)

	// send tx to safe contract
	params := []string{
		hex.EncodeToString([]byte(token.MvxChainSpecificToken)),
		hex.EncodeToString(value.Bytes()),
		hex.EncodeToString([]byte(createTransactionFunction)),
		hex.EncodeToString(handler.TestKeys.EthAddress.Bytes()),
	}
	dataField := strings.Join(params, "@")

	hash, txResult = handler.ChainSimulator.ScCall(
		ctx,
		handler.TestKeys.MvxSk,
		handler.SafeAddress,
		zeroStringValue,
		createDepositGasLimit+gasLimitPerDataByte*uint64(len(dataField)),
		esdtTransferFunction,
		params)
	log.Info("MultiversX->Ethereum transaction sent", "hash", hash, "status", txResult.Status)
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

func getHexBool(input bool) string {
	if input {
		return hexTrue
	}

	return hexFalse
}

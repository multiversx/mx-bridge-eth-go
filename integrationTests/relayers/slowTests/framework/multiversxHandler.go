package framework

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/transaction"
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
	testHelperContractPath    = "testdata/contracts/mvx/helper-contract.wasm"

	changeOwnerAddressFunction                           = "ChangeOwnerAddress"
	moveRefundBatchToSafeFromChildContractFunction       = "moveRefundBatchToSafeFromChildContract"
	getCurrentRefundBatchFunction                        = "getCurrentRefundBatch"
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
	withdrawRefundFeesForEthereumFunction                = "withdrawRefundFeesForEthereum"
	getRefundFeesForEthereumFunction                     = "getRefundFeesForEthereum"
	withdrawTransactionFeesFunction                      = "withdrawTransactionFees"
	getTransactionFeesFunction                           = "getTransactionFees"
	initSupplyMintBurnEsdtSafeFunction                   = "initSupplyMintBurnEsdtSafe"
	initSupplyEsdtSafeFunction                           = "initSupplyEsdtSafe"
	getMintBalancesFunction                              = "getMintBalances"
	getBurnBalancesFunction                              = "getBurnBalances"
	getTotalBalancesFunction                             = "getTotalBalances"
	getTokenLiquidityFunction                            = "getTokenLiquidity"
	setCalleeFunction                                    = "setCallee"
	callDepositFunction                                  = "callDeposit"
	executeFunction                                      = "execute"
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
	TestHelperAddress         *MvxAddress
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
			"01",
		},
	)
	require.NotEqual(handler, emptyAddress, handler.SafeAddress)
	log.Info("Deploy: safe contract", "address", handler.SafeAddress, "transaction hash", hash)

	handler.DeployBridgeProxy(ctx)

	handler.DeployMultisig(ctx)

	handler.DeployTestCaller(ctx)
}

// DeployBridgeProxy will deploy the bridge proxy contract
func (handler *MultiversxHandler) DeployBridgeProxy(ctx context.Context) {
	hash := ""
	handler.ScProxyAddress, hash, _ = handler.ChainSimulator.DeploySC(
		ctx,
		bridgeProxyContractPath,
		handler.OwnerKeys.MvxSk,
		deployGasLimit,
		make([]string, 0),
	)
	require.NotEqual(handler, emptyAddress, handler.ScProxyAddress)
	log.Info("Deploy: Bridge Proxy contract", "address", handler.ScProxyAddress, "transaction hash", hash)
}

// DeployMultisig will deploy the multisig contract
func (handler *MultiversxHandler) DeployMultisig(ctx context.Context) {
	minRelayerStakeInt, _ := big.NewInt(0).SetString(minRelayerStake, 10)
	minRelayerStakeHex := hex.EncodeToString(minRelayerStakeInt.Bytes())
	params := []string{
		handler.SafeAddress.Hex(),
		handler.MultiTransferAddress.Hex(),
		handler.ScProxyAddress.Hex(),
		handler.WrapperAddress.Hex(),
		handler.AggregatorAddress.Hex(),
		minRelayerStakeHex,
		slashAmount,
		handler.Quorum}
	for _, relayerKeys := range handler.RelayersKeys {
		params = append(params, relayerKeys.MvxAddress.Hex())
	}

	hash := ""
	handler.MultisigAddress, hash, _ = handler.ChainSimulator.DeploySC(
		ctx,
		multisigContractPath,
		handler.OwnerKeys.MvxSk,
		deployGasLimit,
		params,
	)
	require.NotEqual(handler, emptyAddress, handler.MultisigAddress)
	log.Info("Deploy: multisig contract", "address", handler.MultisigAddress, "transaction hash", hash)
}

// DeployTestCaller will deploy the test caller contract
func (handler *MultiversxHandler) DeployTestCaller(ctx context.Context) {
	hash := ""
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

func (handler *MultiversxHandler) changeOwners(ctx context.Context) {
	// ChangeOwnerAddress for safe
	params := []string{
		handler.MultisigAddress.Hex(),
	}
	hash, txResult := handler.scCallAndCheckTx(
		ctx,
		handler.OwnerKeys,
		handler.SafeAddress,
		zeroStringValue,
		setCallsGasLimit,
		changeOwnerAddressFunction,
		params)

	log.Info("ChangeOwnerAddress for safe contract", "transaction hash", hash, "status", txResult.Status)

	// ChangeOwnerAddress for multi-transfer
	params = []string{
		handler.MultisigAddress.Hex(),
	}
	hash, txResult = handler.scCallAndCheckTx(
		ctx,
		handler.OwnerKeys,
		handler.MultiTransferAddress,
		zeroStringValue,
		setCallsGasLimit,
		changeOwnerAddressFunction,
		params)

	log.Info("ChangeOwnerAddress for multi-transfer contract", "transaction hash", hash, "status", txResult.Status)

	// ChangeOwnerAddress for bridge proxy
	handler.ChangeOwnerForBridgeProxy(ctx)
}

// ChangeOwnerForBridgeProxy will change the owner for the bridge proxy contract
func (handler *MultiversxHandler) ChangeOwnerForBridgeProxy(ctx context.Context) {
	params := []string{
		handler.MultisigAddress.Hex(),
	}
	hash, txResult := handler.scCallAndCheckTx(
		ctx,
		handler.OwnerKeys,
		handler.ScProxyAddress,
		zeroStringValue,
		setCallsGasLimit,
		changeOwnerAddressFunction,
		params)

	log.Info("ChangeOwnerAddress for SC proxy contract", "transaction hash", hash, "status", txResult.Status)
}

func (handler *MultiversxHandler) finishSettings(ctx context.Context) {
	// unpause sc proxy
	handler.UnpauseBridgeProxy(ctx)

	// stake relayers on multisig
	handler.stakeAddressesOnContract(ctx, handler.MultisigAddress, handler.RelayersKeys)

	// stake relayers on price aggregator
	handler.stakeAddressesOnContract(ctx, handler.AggregatorAddress, handler.OraclesKeys)

	// unpause multisig
	hash, txResult := handler.callContractNoParams(ctx, handler.MultisigAddress, unpauseFunction)
	log.Info("Un-paused multisig contract", "transaction hash", hash, "status", txResult.Status)

	handler.UnPauseContractsAfterTokenChanges(ctx)
}

// UnpauseBridgeProxy will unpause the bridge proxy contract
func (handler *MultiversxHandler) UnpauseBridgeProxy(ctx context.Context) {
	hash, txResult := handler.callContractNoParams(ctx, handler.MultisigAddress, unpauseProxyFunction)
	log.Info("Un-paused SC proxy contract", "transaction hash", hash, "status", txResult.Status)
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
	return handler.scCallAndCheckTx(
		ctx,
		handler.OwnerKeys,
		contract,
		zeroStringValue,
		setCallsGasLimit,
		endpoint,
		[]string{})
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
		hash, txResult, _ := handler.ChainSimulator.SendTx(
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
	if params.PreventWhitelist {
		return
	}
	handler.setLocalRolesForUniversalTokenOnWrapper(ctx, params)
	handler.addUniversalTokenToWrapper(ctx, params)
	handler.whitelistTokenOnWrapper(ctx, params)
	handler.transferChainSpecificTokenToSCs(ctx, params)
	handler.setRolesForSpecificTokenOnSafe(ctx, params)
	handler.addMappingInMultisig(ctx, params)
	handler.whitelistTokenOnMultisig(ctx, params)
	handler.setInitialSupply(ctx, params)
	handler.setPairDecimalsOnAggregator(ctx, params)
	handler.SetMaxBridgeAmountOnSafe(ctx, params, maxBridgedAmountForToken)
	handler.setMaxBridgeAmountOnMultitransfer(ctx, params)
}

func (handler *MultiversxHandler) issueAndWhitelistTokens(ctx context.Context, params IssueTokenParams) {
	handler.issueUniversalToken(ctx, params)

	tkData := handler.TokensRegistry.GetTokenData(params.AbstractTokenIdentifier)
	handler.TokensRegistry.RegisterChainSpecificToken(params.AbstractTokenIdentifier, tkData.MvxUniversalToken)

	if params.PreventWhitelist {
		return
	}

	handler.setRolesForSpecificTokenOnSafe(ctx, params)
	handler.addMappingInMultisig(ctx, params)
	handler.whitelistTokenOnMultisig(ctx, params)
	handler.setInitialSupply(ctx, params)
	handler.setPairDecimalsOnAggregator(ctx, params)
	handler.SetMaxBridgeAmountOnSafe(ctx, params, maxBridgedAmountForToken)
	handler.setMaxBridgeAmountOnMultitransfer(ctx, params)
}

func (handler *MultiversxHandler) issueUniversalToken(ctx context.Context, params IssueTokenParams) {
	token := handler.TokensRegistry.GetTokenData(params.AbstractTokenIdentifier)
	require.NotNil(handler, token)

	valueToMintInt, ok := big.NewInt(0).SetString(params.ValueToMintOnMvx, 10)
	require.True(handler, ok)

	// issue universal token
	scCallParams := []string{
		hex.EncodeToString([]byte(params.MvxUniversalTokenDisplayName)),
		hex.EncodeToString([]byte(params.MvxUniversalTokenTicker)),
		hex.EncodeToString(valueToMintInt.Bytes()),
		fmt.Sprintf("%02x", params.NumOfDecimalsUniversal),
		hex.EncodeToString([]byte(canAddSpecialRoles)),
		hex.EncodeToString([]byte(trueStr))}

	hash, txResult := handler.scCallAndCheckTx(
		ctx,
		handler.OwnerKeys,
		handler.ESDTSystemContractAddress,
		esdtIssueCost,
		issueTokenGasLimit,
		issueFunction,
		scCallParams)
	mvxUniversalToken := handler.getTokenNameFromResult(*txResult)
	require.Greater(handler, len(mvxUniversalToken), 0)
	handler.TokensRegistry.RegisterUniversalToken(params.AbstractTokenIdentifier, mvxUniversalToken)
	log.Info("issue universal token tx executed", "hash", hash, "status", txResult.Status, "token", mvxUniversalToken, "owner", handler.OwnerKeys.MvxAddress)
}

func (handler *MultiversxHandler) issueChainSpecificToken(ctx context.Context, params IssueTokenParams) {
	valueToMintInt, ok := big.NewInt(0).SetString(params.ValueToMintOnMvx, 10)
	require.True(handler, ok)

	scCallParams := []string{
		hex.EncodeToString([]byte(params.MvxChainSpecificTokenDisplayName)),
		hex.EncodeToString([]byte(params.MvxChainSpecificTokenTicker)),
		hex.EncodeToString(valueToMintInt.Bytes()),
		fmt.Sprintf("%02x", params.NumOfDecimalsChainSpecific),
		hex.EncodeToString([]byte(canAddSpecialRoles)),
		hex.EncodeToString([]byte(trueStr))}

	hash, txResult := handler.scCallAndCheckTx(
		ctx,
		handler.OwnerKeys,
		handler.ESDTSystemContractAddress,
		esdtIssueCost,
		issueTokenGasLimit,
		issueFunction,
		scCallParams)
	mvxChainSpecificToken := handler.getTokenNameFromResult(*txResult)
	require.Greater(handler, len(mvxChainSpecificToken), 0)
	handler.TokensRegistry.RegisterChainSpecificToken(params.AbstractTokenIdentifier, mvxChainSpecificToken)
	log.Info("issue chain specific token tx executed", "hash", hash, "status", txResult.Status, "token", mvxChainSpecificToken, "owner", handler.OwnerKeys.MvxAddress)
}

func (handler *MultiversxHandler) setLocalRolesForUniversalTokenOnWrapper(ctx context.Context, params IssueTokenParams) {
	tkData := handler.TokensRegistry.GetTokenData(params.AbstractTokenIdentifier)

	// set local roles bridged tokens wrapper
	scCallParams := []string{
		hex.EncodeToString([]byte(tkData.MvxUniversalToken)),
		handler.WrapperAddress.Hex(),
		hex.EncodeToString([]byte(esdtRoleLocalMint)),
		hex.EncodeToString([]byte(esdtRoleLocalBurn))}

	hash, txResult := handler.scCallAndCheckTx(
		ctx,
		handler.OwnerKeys,
		handler.ESDTSystemContractAddress,
		zeroStringValue,
		setCallsGasLimit,
		setSpecialRoleFunction,
		scCallParams)

	log.Info("set local roles bridged tokens wrapper tx executed", "hash", hash, "status", txResult.Status)
}

func (handler *MultiversxHandler) transferChainSpecificTokenToSCs(ctx context.Context, params IssueTokenParams) {
	valueToMintInt, ok := big.NewInt(0).SetString(params.ValueToMintOnMvx, 10)
	require.True(handler, ok)

	tkData := handler.TokensRegistry.GetTokenData(params.AbstractTokenIdentifier)

	// transfer to wrapper sc
	initialMintValue := valueToMintInt.Div(valueToMintInt, big.NewInt(3))
	scCallParams := []string{
		hex.EncodeToString([]byte(tkData.MvxChainSpecificToken)),
		hex.EncodeToString(initialMintValue.Bytes()),
		hex.EncodeToString([]byte(depositLiquidityFunction))}

	hash, txResult := handler.scCallAndCheckTx(
		ctx,
		handler.OwnerKeys,
		handler.WrapperAddress,
		zeroStringValue,
		setCallsGasLimit,
		esdtTransferFunction,
		scCallParams)

	log.Info("transfer to wrapper sc tx executed", "hash", hash, "status", txResult.Status)

	// transfer to safe sc
	scCallParams = []string{
		hex.EncodeToString([]byte(tkData.MvxChainSpecificToken)),
		hex.EncodeToString(initialMintValue.Bytes())}

	hash, txResult = handler.scCallAndCheckTx(
		ctx,
		handler.OwnerKeys,
		handler.SafeAddress,
		zeroStringValue,
		setCallsGasLimit,
		esdtTransferFunction,
		scCallParams)

	log.Info("transfer to safe sc tx executed", "hash", hash, "status", txResult.Status)
}

func (handler *MultiversxHandler) addUniversalTokenToWrapper(ctx context.Context, params IssueTokenParams) {
	tkData := handler.TokensRegistry.GetTokenData(params.AbstractTokenIdentifier)

	// add wrapped token
	scCallParams := []string{
		hex.EncodeToString([]byte(tkData.MvxUniversalToken)),
		fmt.Sprintf("%02x", params.NumOfDecimalsUniversal),
	}

	hash, txResult := handler.scCallAndCheckTx(
		ctx,
		handler.OwnerKeys,
		handler.WrapperAddress,
		zeroStringValue,
		setCallsGasLimit,
		addWrappedTokenFunction,
		scCallParams)

	log.Info("add wrapped token tx executed", "hash", hash, "status", txResult.Status)
}

func (handler *MultiversxHandler) whitelistTokenOnWrapper(ctx context.Context, params IssueTokenParams) {
	tkData := handler.TokensRegistry.GetTokenData(params.AbstractTokenIdentifier)

	// wrapper whitelist token
	scCallParams := []string{
		hex.EncodeToString([]byte(tkData.MvxChainSpecificToken)),
		fmt.Sprintf("%02x", params.NumOfDecimalsChainSpecific),
		hex.EncodeToString([]byte(tkData.MvxUniversalToken))}

	hash, txResult := handler.scCallAndCheckTx(
		ctx,
		handler.OwnerKeys,
		handler.WrapperAddress,
		zeroStringValue,
		setCallsGasLimit,
		whitelistTokenFunction,
		scCallParams)

	log.Info("wrapper whitelist token tx executed", "hash", hash, "status", txResult.Status)
}

func (handler *MultiversxHandler) setRolesForSpecificTokenOnSafe(ctx context.Context, params IssueTokenParams) {
	tkData := handler.TokensRegistry.GetTokenData(params.AbstractTokenIdentifier)

	// set local roles esdt safe
	scCallParams := []string{
		hex.EncodeToString([]byte(tkData.MvxChainSpecificToken)),
		handler.SafeAddress.Hex(),
		hex.EncodeToString([]byte(esdtRoleLocalMint)),
		hex.EncodeToString([]byte(esdtRoleLocalBurn))}

	hash, txResult := handler.scCallAndCheckTx(
		ctx,
		handler.OwnerKeys,
		handler.ESDTSystemContractAddress,
		zeroStringValue,
		setCallsGasLimit,
		setSpecialRoleFunction,
		scCallParams)

	log.Info("set local roles esdt safe tx executed", "hash", hash, "status", txResult.Status)
}

func (handler *MultiversxHandler) addMappingInMultisig(ctx context.Context, params IssueTokenParams) {
	tkData := handler.TokensRegistry.GetTokenData(params.AbstractTokenIdentifier)

	// add mapping
	scCallParams := []string{
		hex.EncodeToString(tkData.EthErc20Address.Bytes()),
		hex.EncodeToString([]byte(tkData.MvxChainSpecificToken))}

	hash, txResult := handler.scCallAndCheckTx(
		ctx,
		handler.OwnerKeys,
		handler.MultisigAddress,
		zeroStringValue,
		setCallsGasLimit,
		addMappingFunction,
		scCallParams)

	log.Info("add mapping tx executed", "hash", hash, "status", txResult.Status)
}

func (handler *MultiversxHandler) whitelistTokenOnMultisig(ctx context.Context, params IssueTokenParams) {
	tkData := handler.TokensRegistry.GetTokenData(params.AbstractTokenIdentifier)

	// whitelist token
	scCallParams := []string{
		hex.EncodeToString([]byte(tkData.MvxChainSpecificToken)),
		hex.EncodeToString([]byte(params.MvxChainSpecificTokenTicker)),
		getHexBool(params.IsMintBurnOnMvX),
		getHexBool(params.IsNativeOnMvX),
		hex.EncodeToString(zeroValueBigInt.Bytes()), // total_balance
		hex.EncodeToString(zeroValueBigInt.Bytes()), // mint_balance
		hex.EncodeToString(zeroValueBigInt.Bytes()), // burn_balance
	}

	hash, txResult := handler.scCallAndCheckTx(
		ctx,
		handler.OwnerKeys,
		handler.MultisigAddress,
		zeroStringValue,
		setCallsGasLimit,
		esdtSafeAddTokenToWhitelistFunction,
		scCallParams)

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

			scCallParams := []string{
				hex.EncodeToString([]byte(tkData.MvxChainSpecificToken)),
				hex.EncodeToString(mintAmount.Bytes()),
				hex.EncodeToString(burnAmount.Bytes()),
			}
			hash, txResult := handler.scCallAndCheckTx(
				ctx,
				handler.OwnerKeys,
				handler.MultisigAddress,
				zeroStringValue,
				setCallsGasLimit,
				initSupplyMintBurnEsdtSafeFunction,
				scCallParams)

			log.Info("initial supply tx executed", "hash", hash, "status", txResult.Status,
				"initial mint", mintAmount.String(), "initial burned", burnAmount.String())
		} else {
			scCallParams := []string{
				hex.EncodeToString([]byte(tkData.MvxChainSpecificToken)),
				hex.EncodeToString(initialSupply.Bytes()),
				hex.EncodeToString([]byte(initSupplyEsdtSafeFunction)),
				hex.EncodeToString([]byte(tkData.MvxChainSpecificToken)),
				hex.EncodeToString(initialSupply.Bytes()),
			}
			hash, txResult := handler.scCallAndCheckTx(
				ctx,
				handler.OwnerKeys,
				handler.MultisigAddress,
				zeroStringValue,
				setCallsGasLimit,
				esdtTransferFunction,
				scCallParams)

			log.Info("initial supply tx executed", "hash", hash, "status", txResult.Status,
				"initial value", params.InitialSupplyValue)
		}
	}
}

func (handler *MultiversxHandler) setPairDecimalsOnAggregator(ctx context.Context, params IssueTokenParams) {
	// setPairDecimals on aggregator
	scCallParams := []string{
		hex.EncodeToString([]byte(gwei)),
		hex.EncodeToString([]byte(params.MvxChainSpecificTokenTicker)),
		fmt.Sprintf("%02x", params.NumOfDecimalsChainSpecific)}
	hash, txResult := handler.scCallAndCheckTx(
		ctx,
		handler.OwnerKeys,
		handler.AggregatorAddress,
		zeroStringValue,
		setCallsGasLimit,
		setPairDecimalsFunction,
		scCallParams)

	log.Info("setPairDecimals tx executed", "hash", hash, "status", txResult.Status)
}

// SetMaxBridgeAmountOnSafe will set the max bridge amount for the token on the safe contract
func (handler *MultiversxHandler) SetMaxBridgeAmountOnSafe(ctx context.Context, params IssueTokenParams, maxBridgedAmountForToken string) {
	tkData := handler.TokensRegistry.GetTokenData(params.AbstractTokenIdentifier)

	// safe set max bridge amount for token
	maxBridgedAmountForTokenInt, _ := big.NewInt(0).SetString(maxBridgedAmountForToken, 10)
	scCallParams := []string{
		hex.EncodeToString([]byte(tkData.MvxChainSpecificToken)),
		hex.EncodeToString(maxBridgedAmountForTokenInt.Bytes())}
	hash, txResult := handler.scCallAndCheckTx(
		ctx,
		handler.OwnerKeys,
		handler.MultisigAddress,
		zeroStringValue,
		setCallsGasLimit,
		esdtSafeSetMaxBridgedAmountForTokenFunction,
		scCallParams)

	log.Info("safe set max bridge amount for token tx executed", "hash", hash, "status", txResult.Status)
}

func (handler *MultiversxHandler) setMaxBridgeAmountOnMultitransfer(ctx context.Context, params IssueTokenParams) {
	tkData := handler.TokensRegistry.GetTokenData(params.AbstractTokenIdentifier)

	// multi-transfer set max bridge amount for token
	maxBridgedAmountForTokenInt, _ := big.NewInt(0).SetString(maxBridgedAmountForToken, 10)
	scCallParams := []string{
		hex.EncodeToString([]byte(tkData.MvxChainSpecificToken)),
		hex.EncodeToString(maxBridgedAmountForTokenInt.Bytes())}
	hash, txResult := handler.scCallAndCheckTx(
		ctx,
		handler.OwnerKeys,
		handler.MultisigAddress,
		zeroStringValue,
		setCallsGasLimit,
		multiTransferEsdtSetMaxBridgedAmountForTokenFunction,
		scCallParams)

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
func (handler *MultiversxHandler) SubmitAggregatorBatch(ctx context.Context, params IssueTokenParams, price *big.Int) {
	txHashes := make([]string, 0, len(handler.OraclesKeys))
	for _, key := range handler.OraclesKeys {
		hash := handler.submitAggregatorBatchForKey(ctx, key, params, price)
		txHashes = append(txHashes, hash)
	}

	for _, hash := range txHashes {
		txResult, _ := handler.ChainSimulator.GetTransactionResult(ctx, hash)
		log.Info("submit aggregator batch tx", "hash", hash, "status", txResult.Status)
	}
}

func (handler *MultiversxHandler) submitAggregatorBatchForKey(ctx context.Context, key KeysHolder, params IssueTokenParams, price *big.Int) string {
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
			hex.EncodeToString(price.Bytes()),
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

	hash, txResult := handler.scCallAndCheckTx(
		ctx,
		from,
		handler.SafeAddress,
		zeroStringValue,
		createDepositGasLimit+gasLimitPerDataByte*uint64(len(dataField)),
		esdtTransferFunction,
		params)

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

	hash, txResult := handler.scCallAndCheckTx(
		ctx,
		from,
		handler.WrapperAddress,
		zeroStringValue,
		createDepositGasLimit+gasLimitPerDataByte*uint64(len(dataField)),
		esdtTransferFunction,
		params)

	log.Info("MultiversX->Ethereum unwrapCreateTransaction sent", "hash", hash, "token", token.MvxUniversalToken, "status", txResult.Status)
}

// SendWrongDepositTransactionFromMultiversx will send a wrong deposit transaction from MultiversX
func (handler *MultiversxHandler) SendWrongDepositTransactionFromMultiversx(ctx context.Context, from KeysHolder, to KeysHolder, token *TokenData, value *big.Int) {
	params := []string{
		hex.EncodeToString([]byte(token.MvxUniversalToken)),
		hex.EncodeToString(value.Bytes()),
		hex.EncodeToString([]byte(unwrapTokenCreateTransactionFunction)),
		hex.EncodeToString([]byte(token.MvxChainSpecificToken)),
		hex.EncodeToString(to.EthAddress.Bytes()),
	}
	dataField := strings.Join(params, "@")

	_, txResult, txStatus := handler.ChainSimulator.ScCall(
		ctx,
		from.MvxSk,
		handler.WrapperAddress,
		zeroStringValue,
		createDepositGasLimit+gasLimitPerDataByte*uint64(len(dataField)),
		esdtTransferFunction,
		params,
	)

	_, err := json.MarshalIndent(txResult, "", "  ")
	require.Nil(handler, err)
	require.Equal(handler, transaction.TxStatusFail, txStatus)
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

	params := []string{
		hex.EncodeToString([]byte(token)),
	}
	handler.scCallAndCheckTx(
		ctx,
		handler.OwnerKeys,
		handler.MultisigAddress,
		zeroStringValue,
		generalSCCallGasLimit,
		withdrawFunction,
		params)

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
	scCallParams := []string{
		hex.EncodeToString([]byte(tkData.MvxUniversalToken)),
		hex.EncodeToString(amount.Bytes())}
	hash, txResult := handler.scCallAndCheckTx(
		ctx,
		source,
		receiver.MvxAddress,
		zeroStringValue,
		createDepositGasLimit,
		esdtTransferFunction,
		scCallParams)

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
	responseData := handler.ChainSimulator.ExecuteVMQuery(ctx, handler.SafeAddress, getTotalBalancesFunction, queryParams)
	require.Greater(handler, len(responseData), 0)
	value := big.NewInt(0).SetBytes(responseData[0])
	return value
}

// GetMintedAmountForToken will return mint balance for token
func (handler *MultiversxHandler) GetMintedAmountForToken(ctx context.Context, token string) *big.Int {
	queryParams := []string{
		hex.EncodeToString([]byte(token)),
	}
	responseData := handler.ChainSimulator.ExecuteVMQuery(ctx, handler.SafeAddress, getMintBalancesFunction, queryParams)
	require.Greater(handler, len(responseData), 0)
	value := big.NewInt(0).SetBytes(responseData[0])
	return value
}

// GetBurnedAmountForToken will return burn balance of token
func (handler *MultiversxHandler) GetBurnedAmountForToken(ctx context.Context, token string) *big.Int {
	queryParams := []string{
		hex.EncodeToString([]byte(token)),
	}
	responseData := handler.ChainSimulator.ExecuteVMQuery(ctx, handler.SafeAddress, getBurnBalancesFunction, queryParams)
	require.Greater(handler, len(responseData), 0)
	value := big.NewInt(0).SetBytes(responseData[0])
	return value
}

// MoveRefundBatchToSafe will move the refund batch from the multisig to the safe
func (handler *MultiversxHandler) MoveRefundBatchToSafe(ctx context.Context) {
	hash, txResult := handler.scCallAndCheckTx(
		ctx,
		handler.OwnerKeys,
		handler.MultisigAddress,
		zeroStringValue,
		setCallsGasLimit,
		moveRefundBatchToSafeFromChildContractFunction,
		[]string{})
	log.Info("Moved refund batch from Multisig to EsdtSafe", "transaction hash", hash, "status", txResult.Status)
}

// HasRefundBatch will check if there is a refund batch in the multisig
func (handler *MultiversxHandler) HasRefundBatch(ctx context.Context) bool {
	responseData := handler.ChainSimulator.ExecuteVMQuery(ctx, handler.MultisigAddress, getCurrentRefundBatchFunction, []string{})
	return len(responseData) != 0
}

// GetWrapperLiquidity invokes the getTokenLiquidity function, returning the stored liquidity on the wrapper contract
func (handler *MultiversxHandler) GetWrapperLiquidity(ctx context.Context, token string) *big.Int {
	queryParams := []string{
		hex.EncodeToString([]byte(token)),
	}
	responseData := handler.ChainSimulator.ExecuteVMQuery(ctx, handler.WrapperAddress, getTokenLiquidityFunction, queryParams)
	require.Greater(handler, len(responseData), 0)
	value := big.NewInt(0).SetBytes(responseData[0])
	return value
}

func (handler *MultiversxHandler) scCallAndCheckTx(
	ctx context.Context,
	sender KeysHolder,
	receiver *MvxAddress,
	value string,
	gasLimit uint64,
	function string,
	params []string) (string, *data.TransactionOnNetwork) {
	hash, txResult, txStatus := handler.ChainSimulator.ScCall(
		ctx,
		sender.MvxSk,
		receiver,
		value,
		gasLimit,
		function,
		params,
	)

	jsonData, err := json.MarshalIndent(txResult, "", "  ")
	require.Nil(handler, err)
	require.Equal(handler, transaction.TxStatusSuccess, txStatus, fmt.Sprintf("tx hash: %s,\n tx: %s", hash, string(jsonData)))

	log.Info(fmt.Sprintf("Transaction hash %s, status %s", hash, txResult.Status))

	return hash, txResult
}

// DeployTestHelperContract will deploy the test helper contract
func (handler *MultiversxHandler) DeployTestHelperContract(ctx context.Context) {
	hash := ""
	handler.TestHelperAddress, hash, _ = handler.ChainSimulator.DeploySC(
		ctx,
		testHelperContractPath,
		handler.OwnerKeys.MvxSk,
		deployGasLimit,
		make([]string, 0),
	)
	require.NotEqual(handler, emptyAddress, handler.TestHelperAddress)
	log.Info("Deploy: Test helper contract", "address", handler.TestHelperAddress, "transaction hash", hash)
}

// SetBridgeProxyAddressOnHelper will set the bridge proxy address on the helper contract
func (handler *MultiversxHandler) SetBridgeProxyAddressOnHelper(ctx context.Context) {
	params := []string{
		hex.EncodeToString(handler.ScProxyAddress.Bytes()),
	}
	hash, txResult := handler.scCallAndCheckTx(
		ctx,
		handler.OwnerKeys,
		handler.TestHelperAddress,
		zeroStringValue,
		setCallsGasLimit,
		setCalleeFunction,
		params)

	log.Info("set bridge proxy address to helper", "hash", hash, "status", txResult.Status)
}

// CallDepositOnBridgeProxy will call the callDeposit function from test helper contract which will call the deposit function on the bridge proxy
func (handler *MultiversxHandler) CallDepositOnBridgeProxy(ctx context.Context, ethTx EthTransaction, batchId uint64, token *TokenData) {
	buffNonce := make([]byte, 8)
	binary.BigEndian.PutUint64(buffNonce, ethTx.Nonce)

	buffBatchId := make([]byte, 8)
	binary.BigEndian.PutUint64(buffBatchId, batchId)

	params := []string{
		hex.EncodeToString([]byte(token.MvxUniversalToken)),
		hex.EncodeToString(ethTx.Amount.Bytes()),
		hex.EncodeToString([]byte(callDepositFunction)),
		hex.EncodeToString(buffBatchId),
		hex.EncodeToString(ethTx.From),
		hex.EncodeToString(ethTx.To),
		hex.EncodeToString([]byte(ethTx.TokenID)),
		hex.EncodeToString(ethTx.Amount.Bytes()),
		hex.EncodeToString(buffNonce),
		hex.EncodeToString(ethTx.CallData),
	}

	hash, txResult := handler.scCallAndCheckTx(
		ctx,
		handler.OwnerKeys,
		handler.TestHelperAddress,
		zeroStringValue,
		generalSCCallGasLimit,
		esdtTransferFunction,
		params)

	log.Info("deposit on bridge proxy tx executed", "hash", hash, "status", txResult.Status)
}

// ExecuteDepositWithoutGenerateBlocks will execute the deposit on the bridge proxy without generating blocks
func (handler *MultiversxHandler) ExecuteDepositWithoutGenerateBlocks(ctx context.Context, depositId uint64, txNonce uint64) string {
	buffDepositId := make([]byte, 8)
	binary.BigEndian.PutUint64(buffDepositId, depositId)

	params := []string{
		executeFunction,
		hex.EncodeToString(buffDepositId),
	}
	txData := strings.Join(params, "@")

	hash := handler.ChainSimulator.SendTxWithNonceWithoutGenerateBlocks(
		ctx,
		handler.SCExecutorKeys.MvxSk,
		handler.ScProxyAddress,
		txNonce,
		zeroStringValue,
		generalSCCallGasLimit,
		[]byte(txData),
	)
	log.Info("execute deposit on bridge proxy tx sent", "transaction hash", hash)

	return hash
}

func getHexBool(input bool) string {
	if input {
		return hexTrue
	}

	return hexFalse
}

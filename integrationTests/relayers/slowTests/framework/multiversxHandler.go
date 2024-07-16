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
	esdtIssueCost            = "5000000000000000000"  // 5 EGLD
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

	aggregatorContractPath    = "testdata/contracts/mvx/aggregator.wasm"
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
	stakeFunction                                        = "stake"
	unpauseFunction                                      = "unpause"
	unpauseEsdtSafeFunction                              = "unpauseEsdtSafe"
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

	AggregatorAddress *MvxAddress
	WrapperAddress    *MvxAddress
	SafeAddress       *MvxAddress
	MultisigAddress   *MvxAddress
	ScProxyAddress    *MvxAddress
	TestCallerAddress *MvxAddress
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

// DeployContracts will deploy all required contracts on MultiversX side
func (handler *MultiversxHandler) DeployContracts(ctx context.Context) {
	// deploy aggregator
	stakeValue, _ := big.NewInt(0).SetString(minRelayerStake, 10)
	aggregatorDeployParams := []string{
		hex.EncodeToString([]byte("EGLD")),
		hex.EncodeToString(stakeValue.Bytes()),
		"01",
		"01",
		"01",
		handler.OwnerKeys.MvxAddress.Hex(),
	}

	handler.AggregatorAddress = handler.ChainSimulator.DeploySC(
		ctx,
		aggregatorContractPath,
		handler.OwnerKeys.MvxSk,
		aggregatorDeployParams,
	)
	require.NotEqual(handler, emptyAddress, handler.AggregatorAddress)
	log.Info("aggregator contract deployed", "address", handler.AggregatorAddress.Bech32())

	// deploy wrapper
	handler.WrapperAddress = handler.ChainSimulator.DeploySC(
		ctx,
		wrapperContractPath,
		handler.OwnerKeys.MvxSk,
		[]string{},
	)
	require.NotEqual(handler, emptyAddress, handler.WrapperAddress)
	log.Info("wrapper contract deployed", "address", handler.WrapperAddress.Bech32())

	// deploy multi-transfer
	multiTransferAddress := handler.ChainSimulator.DeploySC(
		ctx,
		multiTransferContractPath,
		handler.OwnerKeys.MvxSk,
		[]string{},
	)
	require.NotEqual(handler, emptyAddress, multiTransferAddress)
	log.Info("multi-transfer contract deployed", "address", multiTransferAddress.Bech32())

	// deploy safe
	handler.SafeAddress = handler.ChainSimulator.DeploySC(
		ctx,
		safeContractPath,
		handler.OwnerKeys.MvxSk,
		[]string{
			handler.AggregatorAddress.Hex(),
			multiTransferAddress.Hex(),
			"01",
		},
	)
	require.NotEqual(handler, emptyAddress, handler.SafeAddress)
	log.Info("safe contract deployed", "address", handler.SafeAddress.Bech32())

	// deploy multisig
	minRelayerStakeInt, _ := big.NewInt(0).SetString(minRelayerStake, 10)
	minRelayerStakeHex := hex.EncodeToString(minRelayerStakeInt.Bytes())
	params := []string{
		handler.SafeAddress.Hex(),
		multiTransferAddress.Hex(),
		minRelayerStakeHex,
		slashAmount,
		handler.Quorum}
	for _, relayerKeys := range handler.RelayersKeys {
		params = append(params, relayerKeys.MvxAddress.Hex())
	}
	handler.MultisigAddress = handler.ChainSimulator.DeploySC(
		ctx,
		multisigContractPath,
		handler.OwnerKeys.MvxSk,
		params,
	)
	require.NotEqual(handler, emptyAddress, handler.MultisigAddress)
	log.Info("multisig contract deployed", "address", handler.MultisigAddress)

	// deploy bridge proxy
	handler.ScProxyAddress = handler.ChainSimulator.DeploySC(
		ctx,
		bridgeProxyContractPath,
		handler.OwnerKeys.MvxSk,
		[]string{
			multiTransferAddress.Hex(),
		},
	)
	require.NotEqual(handler, emptyAddress, handler.ScProxyAddress)
	log.Info("bridge proxy contract deployed", "address", handler.ScProxyAddress)

	// deploy test-caller
	handler.TestCallerAddress = handler.ChainSimulator.DeploySC(
		ctx,
		testCallerContractPath,
		handler.OwnerKeys.MvxSk,
		[]string{},
	)
	require.NotEqual(handler, emptyAddress, handler.TestCallerAddress)
	log.Info("test-caller contract deployed", "address", handler.TestCallerAddress)

	// setBridgeProxyContractAddress
	hash := handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.MvxSk,
		multiTransferAddress,
		zeroStringValue,
		setBridgeProxyContractAddressFunction,
		[]string{
			handler.ScProxyAddress.Hex(),
		},
	)
	txResult := handler.ChainSimulator.GetTransactionResult(ctx, hash)
	log.Info("setBridgeProxyContractAddress tx executed", "hash", hash, "status", txResult.Status)

	// setWrappingContractAddress
	hash = handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.MvxSk,
		multiTransferAddress,
		zeroStringValue,
		setWrappingContractAddressFunction,
		[]string{
			handler.WrapperAddress.Hex(),
		},
	)
	txResult = handler.ChainSimulator.GetTransactionResult(ctx, hash)
	log.Info("setWrappingContractAddress tx executed", "hash", hash, "status", txResult.Status)

	// ChangeOwnerAddress for safe
	hash = handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.MvxSk,
		handler.SafeAddress,
		zeroStringValue,
		changeOwnerAddressFunction,
		[]string{
			handler.MultisigAddress.Hex(),
		},
	)
	txResult = handler.ChainSimulator.GetTransactionResult(ctx, hash)
	log.Info("ChangeOwnerAddress for safe tx executed", "hash", hash, "status", txResult.Status)

	// ChangeOwnerAddress for multi-transfer
	hash = handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.MvxSk,
		multiTransferAddress,
		zeroStringValue,
		changeOwnerAddressFunction,
		[]string{
			handler.MultisigAddress.Hex(),
		},
	)
	txResult = handler.ChainSimulator.GetTransactionResult(ctx, hash)
	log.Info("ChangeOwnerAddress for multi-transfer tx executed", "hash", hash, "status", txResult.Status)

	// unpause sc proxy
	hash = handler.callContractNoParams(ctx, handler.ScProxyAddress, unpauseFunction)
	txResult = handler.ChainSimulator.GetTransactionResult(ctx, hash)
	log.Info("unpaused sc proxy executed", "hash", hash, "status", txResult.Status)

	// ChangeOwnerAddress for bridge proxy
	hash = handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.MvxSk,
		handler.ScProxyAddress,
		zeroStringValue,
		changeOwnerAddressFunction,
		[]string{
			handler.MultisigAddress.Hex(),
		},
	)
	txResult = handler.ChainSimulator.GetTransactionResult(ctx, hash)
	log.Info("ChangeOwnerAddress for bridge proxy tx executed", "hash", hash, "status", txResult.Status)

	// setEsdtSafeOnMultiTransfer
	hash = handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.MvxSk,
		handler.MultisigAddress,
		zeroStringValue,
		setEsdtSafeOnMultiTransferFunction,
		[]string{},
	)
	txResult = handler.ChainSimulator.GetTransactionResult(ctx, hash)
	log.Info("setEsdtSafeOnMultiTransfer tx executed", "hash", hash, "status", txResult.Status)

	// stake relayers on multisig
	handler.stakeAddressesOnContract(ctx, handler.MultisigAddress, handler.RelayersKeys)

	// stake relayers on price aggregator
	handler.stakeAddressesOnContract(ctx, handler.AggregatorAddress, []KeysHolder{handler.OwnerKeys})

	// unpause multisig
	hash = handler.callContractNoParams(ctx, handler.MultisigAddress, unpauseFunction)
	txResult = handler.ChainSimulator.GetTransactionResult(ctx, hash)
	log.Info("unpaused multisig executed", "hash", hash, "status", txResult.Status)

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

func (handler *MultiversxHandler) callContractNoParams(ctx context.Context, contract *MvxAddress, endpoint string) string {
	return handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.MvxSk,
		contract,
		zeroStringValue,
		endpoint,
		[]string{},
	)
}

// UnPauseContractsAfterTokenChanges can unpause contracts after token changes
func (handler *MultiversxHandler) UnPauseContractsAfterTokenChanges(ctx context.Context) {
	// unpause safe
	hash := handler.callContractNoParams(ctx, handler.MultisigAddress, unpauseEsdtSafeFunction)
	txResult := handler.ChainSimulator.GetTransactionResult(ctx, hash)
	log.Info("unpaused safe executed", "hash", hash, "status", txResult.Status)

	// unpause wrapper
	hash = handler.callContractNoParams(ctx, handler.WrapperAddress, unpauseFunction)
	txResult = handler.ChainSimulator.GetTransactionResult(ctx, hash)
	log.Info("unpaused wrapper executed", "hash", hash, "status", txResult.Status)

	// unpause aggregator
	hash = handler.callContractNoParams(ctx, handler.AggregatorAddress, unpauseFunction)
	txResult = handler.ChainSimulator.GetTransactionResult(ctx, hash)
	log.Info("unpaused aggregator executed", "hash", hash, "status", txResult.Status)
}

// PauseContractsForTokenChanges can pause contracts for token changes
func (handler *MultiversxHandler) PauseContractsForTokenChanges(ctx context.Context) {
	// pause safe
	hash := handler.callContractNoParams(ctx, handler.MultisigAddress, pauseEsdtSafeFunction)
	txResult := handler.ChainSimulator.GetTransactionResult(ctx, hash)
	log.Info("paused safe executed", "hash", hash, "status", txResult.Status)

	// pause aggregator
	hash = handler.callContractNoParams(ctx, handler.AggregatorAddress, pauseFunction)
	txResult = handler.ChainSimulator.GetTransactionResult(ctx, hash)
	log.Info("paused aggregator executed", "hash", hash, "status", txResult.Status)

	// pause wrapper
	hash = handler.callContractNoParams(ctx, handler.WrapperAddress, pauseFunction)
	txResult = handler.ChainSimulator.GetTransactionResult(ctx, hash)
	log.Info("paused wrapper executed", "hash", hash, "status", txResult.Status)
}

func (handler *MultiversxHandler) stakeAddressesOnContract(ctx context.Context, contract *MvxAddress, allKeys []KeysHolder) {
	for _, keys := range allKeys {
		hash := handler.ChainSimulator.SendTx(ctx, keys.MvxSk, contract, minRelayerStake, []byte(stakeFunction))
		txResult := handler.ChainSimulator.GetTransactionResult(ctx, hash)

		log.Info(fmt.Sprintf("address %s staked on contract %s with hash %s, status %s", keys.MvxAddress.Bech32(), contract, hash, txResult.Status))
	}
}

// IssueAndWhitelistToken will issue and whitelist the token on MultiversX
func (handler *MultiversxHandler) IssueAndWhitelistToken(ctx context.Context, params IssueTokenParams) {
	token := handler.TokensRegistry.GetTokenData(params.AbstractTokenIdentifier)
	require.NotNil(handler, token)

	esdtAddress := NewMvxAddressFromBech32(handler, esdtSystemSCAddress)

	// issue universal token
	hash := handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.MvxSk,
		esdtAddress,
		esdtIssueCost,
		issueFunction,
		[]string{
			hex.EncodeToString([]byte(params.MvxUniversalTokenDisplayName)),
			hex.EncodeToString([]byte(params.MvxUniversalTokenTicker)),
			"00",
			fmt.Sprintf("%02x", params.NumOfDecimalsUniversal),
			hex.EncodeToString([]byte(canAddSpecialRoles)),
			hex.EncodeToString([]byte(trueStr))})
	txResult := handler.ChainSimulator.GetTransactionResult(ctx, hash)
	mvxUniversalToken := handler.getTokenNameFromResult(*txResult)
	handler.TokensRegistry.RegisterUniversalToken(params.AbstractTokenIdentifier, mvxUniversalToken)
	log.Info("issue universal token tx executed", "hash", hash, "status", txResult.Status, "token", mvxUniversalToken, "owner", handler.OwnerKeys.MvxAddress.Bech32())

	// issue chain specific token
	valueToMintInt, ok := big.NewInt(0).SetString(params.ValueToMintOnMvx, 10)
	require.True(handler, ok)

	hash = handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.MvxSk,
		esdtAddress,
		esdtIssueCost,
		issueFunction,
		[]string{
			hex.EncodeToString([]byte(params.MvxChainSpecificTokenDisplayName)),
			hex.EncodeToString([]byte(params.MvxChainSpecificTokenTicker)),
			hex.EncodeToString(valueToMintInt.Bytes()),
			fmt.Sprintf("%02x", params.NumOfDecimalsChainSpecific),
			hex.EncodeToString([]byte(canAddSpecialRoles)),
			hex.EncodeToString([]byte(trueStr))})
	txResult = handler.ChainSimulator.GetTransactionResult(ctx, hash)
	mvxChainSpecificToken := handler.getTokenNameFromResult(*txResult)
	handler.TokensRegistry.RegisterChainSpecificToken(params.AbstractTokenIdentifier, mvxChainSpecificToken)
	log.Info("issue chain specific token tx executed", "hash", hash, "status", txResult.Status, "token", mvxChainSpecificToken, "owner", handler.OwnerKeys.MvxAddress.Bech32())

	// set local roles bridged tokens wrapper
	hash = handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.MvxSk,
		esdtAddress,
		zeroStringValue,
		setSpecialRoleFunction,
		[]string{
			hex.EncodeToString([]byte(mvxUniversalToken)),
			handler.WrapperAddress.Hex(),
			hex.EncodeToString([]byte(esdtRoleLocalMint)),
			hex.EncodeToString([]byte(esdtRoleLocalBurn))})
	txResult = handler.ChainSimulator.GetTransactionResult(ctx, hash)
	log.Info("set local roles bridged tokens wrapper tx executed", "hash", hash, "status", txResult.Status)

	// transfer to wrapper sc
	initialMintValue := valueToMintInt.Div(valueToMintInt, big.NewInt(3))
	hash = handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.MvxSk,
		handler.WrapperAddress,
		zeroStringValue,
		esdtTransferFunction,
		[]string{
			hex.EncodeToString([]byte(mvxChainSpecificToken)),
			hex.EncodeToString(initialMintValue.Bytes()),
			hex.EncodeToString([]byte(depositLiquidityFunction))})
	txResult = handler.ChainSimulator.GetTransactionResult(ctx, hash)

	log.Info("transfer to wrapper sc tx executed", "hash", hash, "status", txResult.Status)

	// transfer to safe sc
	hash = handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.MvxSk,
		handler.SafeAddress,
		zeroStringValue,
		esdtTransferFunction,
		[]string{
			hex.EncodeToString([]byte(mvxChainSpecificToken)),
			hex.EncodeToString(initialMintValue.Bytes())})
	txResult = handler.ChainSimulator.GetTransactionResult(ctx, hash)
	log.Info("transfer to safe sc tx executed", "hash", hash, "status", txResult.Status)

	// add wrapped token
	hash = handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.MvxSk,
		handler.WrapperAddress,
		zeroStringValue,
		addWrappedTokenFunction,
		[]string{
			hex.EncodeToString([]byte(mvxUniversalToken)),
			fmt.Sprintf("%02x", params.NumOfDecimalsUniversal),
		})
	txResult = handler.ChainSimulator.GetTransactionResult(ctx, hash)
	log.Info("add wrapped token tx executed", "hash", hash, "status", txResult.Status)

	// wrapper whitelist token
	hash = handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.MvxSk,
		handler.WrapperAddress,
		zeroStringValue,
		whitelistTokenFunction,
		[]string{
			hex.EncodeToString([]byte(mvxChainSpecificToken)),
			fmt.Sprintf("%02x", params.NumOfDecimalsChainSpecific),
			hex.EncodeToString([]byte(mvxUniversalToken))})
	txResult = handler.ChainSimulator.GetTransactionResult(ctx, hash)

	log.Info("wrapper whitelist token tx executed", "hash", hash, "status", txResult.Status)

	// set local roles esdt safe
	hash = handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.MvxSk,
		esdtAddress,
		zeroStringValue,
		setSpecialRoleFunction,
		[]string{
			hex.EncodeToString([]byte(mvxChainSpecificToken)),
			handler.SafeAddress.Hex(),
			hex.EncodeToString([]byte(esdtRoleLocalMint)),
			hex.EncodeToString([]byte(esdtRoleLocalBurn))})
	txResult = handler.ChainSimulator.GetTransactionResult(ctx, hash)
	log.Info("set local roles esdt safe tx executed", "hash", hash, "status", txResult.Status)

	// add mapping
	hash = handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.MvxSk,
		handler.MultisigAddress,
		zeroStringValue,
		addMappingFunction,
		[]string{
			hex.EncodeToString(token.EthErc20Address.Bytes()),
			hex.EncodeToString([]byte(mvxChainSpecificToken))})
	txResult = handler.ChainSimulator.GetTransactionResult(ctx, hash)
	log.Info("add mapping tx executed", "hash", hash, "status", txResult.Status)

	// whitelist token
	hash = handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.MvxSk,
		handler.MultisigAddress,
		zeroStringValue,
		esdtSafeAddTokenToWhitelistFunction,
		[]string{
			hex.EncodeToString([]byte(mvxChainSpecificToken)),
			hex.EncodeToString([]byte(params.MvxChainSpecificTokenTicker)),
			getHexBool(params.IsMintBurnOnMvX),
			getHexBool(params.IsNativeOnMvX)})
	txResult = handler.ChainSimulator.GetTransactionResult(ctx, hash)
	log.Info("whitelist token tx executed", "hash", hash, "status", txResult.Status)

	// setPairDecimals on aggregator
	hash = handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.MvxSk,
		handler.AggregatorAddress,
		zeroStringValue,
		setPairDecimalsFunction,
		[]string{
			hex.EncodeToString([]byte(gwei)),
			hex.EncodeToString([]byte(params.MvxChainSpecificTokenTicker)),
			fmt.Sprintf("%02x", params.NumOfDecimalsChainSpecific)})
	txResult = handler.ChainSimulator.GetTransactionResult(ctx, hash)
	log.Info("setPairDecimals tx executed", "hash", hash, "status", txResult.Status)

	// safe set max bridge amount for token
	maxBridgedAmountForTokenInt, _ := big.NewInt(0).SetString(maxBridgedAmountForToken, 10)
	hash = handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.MvxSk,
		handler.MultisigAddress,
		zeroStringValue,
		esdtSafeSetMaxBridgedAmountForTokenFunction,
		[]string{
			hex.EncodeToString([]byte(mvxChainSpecificToken)),
			hex.EncodeToString(maxBridgedAmountForTokenInt.Bytes())})
	txResult = handler.ChainSimulator.GetTransactionResult(ctx, hash)
	log.Info("safe set max bridge amount for token tx executed", "hash", hash, "status", txResult.Status)

	// multi-transfer set max bridge amount for token
	hash = handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.MvxSk,
		handler.MultisigAddress,
		zeroStringValue,
		multiTransferEsdtSetMaxBridgedAmountForTokenFunction,
		[]string{
			hex.EncodeToString([]byte(mvxChainSpecificToken)),
			hex.EncodeToString(maxBridgedAmountForTokenInt.Bytes())})
	txResult = handler.ChainSimulator.GetTransactionResult(ctx, hash)
	log.Info("multi-transfer set max bridge amount for token tx executed", "hash", hash, "status", txResult.Status)
}

func (handler *MultiversxHandler) getTokenNameFromResult(txResult data.TransactionOnNetwork) string {
	resultData := txResult.ScResults[0].Data
	splittedData := strings.Split(resultData, "@")
	if len(splittedData) < 2 {
		require.Fail(handler, fmt.Sprintf("received invalid data received while issuing: %s", resultData))
	}

	newUniversalTokenBytes, err := hex.DecodeString(splittedData[1])
	require.NoError(handler, err)

	return string(newUniversalTokenBytes)
}

// SubmitAggregatorBatch will submit the aggregator batch
func (handler *MultiversxHandler) SubmitAggregatorBatch(ctx context.Context, params IssueTokenParams) {
	timestamp := handler.ChainSimulator.GetBlockchainTimeStamp(ctx)
	require.Greater(handler, timestamp, uint64(0), "something went wrong and the chain simulator returned 0 for the current timestamp")

	timestampAsBigInt := big.NewInt(0).SetUint64(timestamp)

	hash := handler.ChainSimulator.ScCall(
		ctx,
		handler.OwnerKeys.MvxSk,
		handler.AggregatorAddress,
		zeroStringValue,
		submitBatchFunction,
		[]string{
			hex.EncodeToString([]byte(gwei)),
			hex.EncodeToString([]byte(params.MvxChainSpecificTokenTicker)),
			hex.EncodeToString(timestampAsBigInt.Bytes()),
			hex.EncodeToString(feeInt.Bytes()),
			fmt.Sprintf("%02x", params.NumOfDecimalsChainSpecific)})
	txResult := handler.ChainSimulator.GetTransactionResult(ctx, hash)
	log.Info("submit aggregator batch tx executed", "hash", hash, "submitter", handler.OwnerKeys.MvxAddress.Bech32(), "status", txResult.Status)
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
		hash := handler.ChainSimulator.ScCall(
			ctx,
			handler.OwnerKeys.MvxSk,
			handler.TestKeys.MvxAddress,
			zeroStringValue,
			esdtTransferFunction,
			[]string{
				hex.EncodeToString([]byte(token.MvxChainSpecificToken)),
				hex.EncodeToString(operation.ValueToSendFromMvX.Bytes())})
		txResult := handler.ChainSimulator.GetTransactionResult(ctx, hash)
		log.Info("transfer to sender tx executed", "hash", hash, "status", txResult.Status)

		// send tx to safe contract
		scCallParams := []string{
			hex.EncodeToString([]byte(token.MvxChainSpecificToken)),
			hex.EncodeToString(operation.ValueToSendFromMvX.Bytes()),
			hex.EncodeToString([]byte(createTransactionFunction)),
			hex.EncodeToString(handler.TestKeys.EthAddress.Bytes()),
		}
		hash = handler.ChainSimulator.ScCall(
			ctx,
			handler.TestKeys.MvxSk,
			handler.SafeAddress,
			zeroStringValue,
			esdtTransferFunction,
			scCallParams)
		txResult = handler.ChainSimulator.GetTransactionResult(ctx, hash)
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

	hash := handler.ChainSimulator.ScCall(
		ctx,
		handler.TestKeys.MvxSk,
		handler.WrapperAddress,
		zeroStringValue,
		esdtTransferFunction,
		paramsUnwrap,
	)
	txResult := handler.ChainSimulator.GetTransactionResult(ctx, hash)
	log.Info("unwrap transaction sent", "hash", hash, "token", token.MvxUniversalToken, "status", txResult.Status)

	// send tx to safe contract
	params := []string{
		hex.EncodeToString([]byte(token.MvxChainSpecificToken)),
		hex.EncodeToString(value.Bytes()),
		hex.EncodeToString([]byte(createTransactionFunction)),
		hex.EncodeToString(handler.TestKeys.EthAddress.Bytes()),
	}

	hash = handler.ChainSimulator.ScCall(
		ctx,
		handler.TestKeys.MvxSk,
		handler.SafeAddress,
		zeroStringValue,
		esdtTransferFunction,
		params)
	txResult = handler.ChainSimulator.GetTransactionResult(ctx, hash)
	log.Info("MultiversX->Ethereum transaction sent", "hash", hash, "status", txResult.Status)
}

func getHexBool(input bool) string {
	if input {
		return hexTrue
	}

	return hexFalse
}

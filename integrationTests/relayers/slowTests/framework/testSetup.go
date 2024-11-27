package framework

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"path"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/multiversx/mx-bridge-eth-go/config"
	"github.com/multiversx/mx-bridge-eth-go/executors/multiversx/module"
	sdkCore "github.com/multiversx/mx-sdk-go/core"
	"github.com/stretchr/testify/require"
)

// framework constants
const (
	LogStepMarker                              = "#################################### %s ####################################"
	proxyCacherExpirationSeconds               = 600
	proxyMaxNoncesDelta                        = 7
	NumRelayers                                = 3
	NumOracles                                 = 3
	quorum                                     = "03"
	mvxHrp                                     = "erd"
	firstBridge                  currentBridge = "first bridge"
	secondBridge                 currentBridge = "second bridge"
)

type currentBridge string

type CurrentActorState struct {
	isSender           bool
	HasToReceiveRefund bool
}

var (
	senderState = CurrentActorState{
		isSender:           true,
		HasToReceiveRefund: false,
	}
	receiverState = CurrentActorState{
		isSender:           false,
		HasToReceiveRefund: false,
	}
)

// TestSetup is the struct that holds all subcomponents for the testing infrastructure
type TestSetup struct {
	testing.TB
	TokensRegistry
	*KeysStore
	Bridge                 *BridgeComponents
	EthereumHandler        *EthereumHandler
	MultiversxHandler      *MultiversxHandler
	WorkingDir             string
	ChainSimulator         ChainSimulatorWrapper
	ScCallerKeys           KeysHolder
	ScCallerModuleInstance SCCallerModule

	ctxCancel          func()
	Ctx                context.Context
	mutBalances        sync.RWMutex
	esdtBalanceForSafe map[string]*big.Int
	mvxBalances        map[string]map[string]*big.Int
	ethBalances        map[string]map[string]*big.Int

	numScCallsInTest uint32
}

// NewTestSetup creates a new e2e test setup
func NewTestSetup(tb testing.TB) *TestSetup {
	log.Info(fmt.Sprintf(LogStepMarker, "starting setup"))

	setup := &TestSetup{
		TB:                 tb,
		TokensRegistry:     NewTokenRegistry(tb),
		WorkingDir:         tb.TempDir(),
		esdtBalanceForSafe: make(map[string]*big.Int),
		mvxBalances:        make(map[string]map[string]*big.Int),
		ethBalances:        make(map[string]map[string]*big.Int),
	}
	setup.KeysStore = NewKeysStore(tb, setup.WorkingDir, NumRelayers, NumOracles)

	setup.mvxBalances[setup.AliceKeys.MvxAddress.String()] = make(map[string]*big.Int)
	setup.mvxBalances[setup.BobKeys.MvxAddress.String()] = make(map[string]*big.Int)
	setup.mvxBalances[setup.CharlieKeys.MvxAddress.String()] = make(map[string]*big.Int)

	setup.ethBalances[setup.AliceKeys.EthAddress.String()] = make(map[string]*big.Int)
	setup.ethBalances[setup.BobKeys.EthAddress.String()] = make(map[string]*big.Int)
	setup.ethBalances[setup.CharlieKeys.EthAddress.String()] = make(map[string]*big.Int)

	// create a test context
	setup.Ctx, setup.ctxCancel = context.WithCancel(context.Background())

	setup.EthereumHandler = NewEthereumHandler(tb, setup.Ctx, setup.KeysStore, setup.TokensRegistry, quorum)
	setup.EthereumHandler.DeployContracts(setup.Ctx)

	setup.createChainSimulatorWrapper()
	setup.MultiversxHandler = NewMultiversxHandler(tb, setup.Ctx, setup.KeysStore, setup.TokensRegistry, setup.ChainSimulator, quorum)
	setup.MultiversxHandler.DeployAndSetContracts(setup.Ctx)

	return setup
}

func (setup *TestSetup) createChainSimulatorWrapper() {
	// create a new working directory
	tmpDir := path.Join(setup.TempDir(), "test")
	err := os.MkdirAll(tmpDir, os.ModePerm)
	require.NoError(setup, err)

	// start the chain simulator
	args := ArgChainSimulatorWrapper{
		TB:                           setup.TB,
		ProxyCacherExpirationSeconds: proxyCacherExpirationSeconds,
		ProxyMaxNoncesDelta:          proxyMaxNoncesDelta,
	}
	setup.ChainSimulator = CreateChainSimulatorWrapper(args)
	require.NoError(setup, err)
}

// StartRelayersAndScModule will start the bridge and the SC execution module
func (setup *TestSetup) StartRelayersAndScModule() {
	log.Info(fmt.Sprintf(LogStepMarker, "starting relayers & sc execution module"))

	// start relayers
	setup.Bridge = NewBridgeComponents(
		setup.TB,
		setup.WorkingDir,
		setup.ChainSimulator,
		setup.EthereumHandler.EthChainWrapper,
		setup.EthereumHandler.Erc20ContractsHolder,
		setup.EthereumHandler.SimulatedChain,
		NumRelayers,
		setup.EthereumHandler.SafeAddress.Hex(),
		setup.MultiversxHandler.SafeAddress,
		setup.MultiversxHandler.MultisigAddress,
	)

	setup.startScCallerModule()
}

func (setup *TestSetup) startScCallerModule() {
	cfg := config.ScCallsModuleConfig{
		ScProxyBech32Address:            setup.MultiversxHandler.ScProxyAddress.Bech32(),
		ExtraGasToExecute:               60_000_000,  // 60 million: this ensures that a SC call with 0 gas limit is refunded
		MaxGasLimitToUse:                249_999_999, // max cross shard limit
		GasLimitForOutOfGasTransactions: 30_000_000,  // gas to use when a higher than max allowed is encountered
		NetworkAddress:                  setup.ChainSimulator.GetNetworkAddress(),
		ProxyMaxNoncesDelta:             5,
		ProxyFinalityCheck:              false,
		ProxyCacherExpirationSeconds:    60, // 1 minute
		ProxyRestAPIEntityType:          string(sdkCore.Proxy),
		IntervalToResendTxsInSeconds:    1,
		PrivateKeyFile:                  path.Join(setup.WorkingDir, SCCallerFilename),
		PollingIntervalInMillis:         1000, // 1 second
		Filter: config.PendingOperationsFilterConfig{
			AllowedEthAddresses: []string{"*"},
			AllowedMvxAddresses: []string{"*"},
			AllowedTokens:       []string{"*"},
		},
		TransactionChecks: config.TransactionChecksConfig{
			CheckTransactionResults:    true,
			CloseAppOnError:            false,
			ExecutionTimeoutInSeconds:  2,
			TimeInSecondsBetweenChecks: 1,
		},
	}

	var err error
	setup.ScCallerModuleInstance, err = module.NewScCallsModule(cfg, log, nil)
	require.Nil(setup, err)
	log.Info("started SC calls module", "monitoring SC proxy address", setup.MultiversxHandler.ScProxyAddress)
}

// IssueAndConfigureTokens will issue and configure the provided tokens on both chains
func (setup *TestSetup) IssueAndConfigureTokens(tokens ...TestTokenParams) {
	log.Info(fmt.Sprintf(LogStepMarker, fmt.Sprintf("issuing %d tokens", len(tokens))))

	require.Greater(setup, len(tokens), 0)

	setup.EthereumHandler.PauseContractsForTokenChanges(setup.Ctx)
	setup.MultiversxHandler.PauseContractsForTokenChanges(setup.Ctx)

	for _, token := range tokens {
		setup.processNumScCallsOperations(token)
		setup.AddToken(token.IssueTokenParams)
		setup.EthereumHandler.IssueAndWhitelistToken(setup.Ctx, token.IssueTokenParams)
		setup.MultiversxHandler.IssueAndWhitelistToken(setup.Ctx, token.IssueTokenParams)
		if token.IsNativeOnMvX {
			setup.transferTokensToMvxTestKey(token) // TODO: (Next PRs) this will be moved an batch creation time
		}
		setup.ChainSimulator.GenerateBlocks(setup.Ctx, 10)

		esdtBalanceForSafe := setup.MultiversxHandler.GetESDTChainSpecificTokenBalance(setup.Ctx, setup.MultiversxHandler.SafeAddress, token.AbstractTokenIdentifier)

		setup.mutBalances.Lock()
		setup.esdtBalanceForSafe[token.AbstractTokenIdentifier] = esdtBalanceForSafe

		setup.mvxBalances[setup.AliceKeys.MvxAddress.String()][token.AbstractTokenIdentifier] = setup.getTokenBalanceForAddress(true, setup.AliceKeys, token.AbstractTokenIdentifier)
		setup.mvxBalances[setup.BobKeys.MvxAddress.String()][token.AbstractTokenIdentifier] = setup.getTokenBalanceForAddress(true, setup.BobKeys, token.AbstractTokenIdentifier)
		setup.mvxBalances[setup.CharlieKeys.MvxAddress.String()][token.AbstractTokenIdentifier] = setup.getTokenBalanceForAddress(true, setup.CharlieKeys, token.AbstractTokenIdentifier)

		setup.ethBalances[setup.AliceKeys.EthAddress.String()][token.AbstractTokenIdentifier] = setup.getTokenBalanceForAddress(false, setup.AliceKeys, token.AbstractTokenIdentifier)
		setup.ethBalances[setup.BobKeys.EthAddress.String()][token.AbstractTokenIdentifier] = setup.getTokenBalanceForAddress(false, setup.BobKeys, token.AbstractTokenIdentifier)
		setup.ethBalances[setup.CharlieKeys.EthAddress.String()][token.AbstractTokenIdentifier] = setup.getTokenBalanceForAddress(false, setup.CharlieKeys, token.AbstractTokenIdentifier)

		setup.mutBalances.Unlock()

		log.Info("recorded the ESDT balance for safe contract", "token", token.AbstractTokenIdentifier, "balance", esdtBalanceForSafe.String())
	}

	setup.EthereumHandler.UnPauseContractsAfterTokenChanges(setup.Ctx)
	setup.MultiversxHandler.UnPauseContractsAfterTokenChanges(setup.Ctx)

	for _, token := range tokens {
		setup.MultiversxHandler.SubmitAggregatorBatch(setup.Ctx, token.IssueTokenParams)
	}
}

func (setup *TestSetup) getTokenBalanceForAddress(isOnMvx bool, holder KeysHolder, token string) *big.Int {
	if isOnMvx {
		return setup.MultiversxHandler.GetESDTUniversalTokenBalance(setup.Ctx, holder.MvxAddress, token)
	}

	return setup.EthereumHandler.GetBalance(holder.EthAddress, token)
}

func (setup *TestSetup) processNumScCallsOperations(token TestTokenParams) {
	for _, op := range token.TestOperations {
		if len(op.MvxSCCallData) > 0 || op.MvxForceSCCall {
			atomic.AddUint32(&setup.numScCallsInTest, 1)
		}
	}
}

// GetNumScCallsOperations returns the number of SC calls in this test setup
func (setup *TestSetup) GetNumScCallsOperations() uint32 {
	return atomic.LoadUint32(&setup.numScCallsInTest)
}

// IsTransferDoneFromEthereum returns true if all provided tokens are bridged from Ethereum towards MultiversX
func (setup *TestSetup) IsTransferDoneFromEthereum(sender KeysHolder, receiver KeysHolder, tokens ...TestTokenParams) bool {
	isDone := true
	for _, params := range tokens {
		isDone = isDone && setup.isTransferDoneFromEthereumForToken(sender, receiver, params)
	}

	return isDone
}

func (setup *TestSetup) isTransferDoneFromEthereumForToken(sender, receiver KeysHolder, params TestTokenParams) bool {
	// if token is prevented from whitelist, we can't check the balances
	if params.PreventWhitelist {
		return true
	}

	if !setup.checkHolderEthBalanceForToken(sender, senderState, params) {
		return false
	}

	if !setup.checkHolderMvxBalanceForToken(receiver, receiverState, params) {
		return false
	}

	if !setup.checkContractMvxBalanceForToken(params) {
		return false
	}

	if !setup.checkTokenOnEthFirstBridge(params) {
		return false
	}

	return setup.checkTokenOnMvxSecondBridge(params)
}

func (setup *TestSetup) checkHolderEthBalanceForToken(holder KeysHolder, actorState CurrentActorState, params TestTokenParams) bool {
	balanceMapping, exists := setup.getBalanceMappingForAddress(holder.EthAddress.String())
	if !exists {
		return false
	}

	actualBalance := setup.EthereumHandler.GetBalance(holder.EthAddress, params.AbstractTokenIdentifier)
	return setup.checkHolderBalanceForTokenHelper(balanceMapping, params, actualBalance, holder.EthAddress.String(), actorState)
}

func (setup *TestSetup) checkHolderMvxBalanceForToken(holder KeysHolder, actorState CurrentActorState, params TestTokenParams) bool {
	balanceMapping, exists := setup.getBalanceMappingForAddress(holder.MvxAddress.String())
	if !exists {
		return false
	}

	actualBalance := setup.MultiversxHandler.GetESDTUniversalTokenBalance(setup.Ctx, holder.MvxAddress, params.AbstractTokenIdentifier)
	return setup.checkHolderBalanceForTokenHelper(balanceMapping, params, actualBalance, holder.MvxAddress.String(), actorState)
}

func (setup *TestSetup) getBalanceMappingForAddress(address string) (map[string]*big.Int, bool) {
	setup.mutBalances.Lock()
	defer setup.mutBalances.Unlock()

	if strings.HasPrefix(address, mvxHrp) {
		balanceMapping, exists := setup.mvxBalances[address]
		return balanceMapping, exists
	}

	balanceMapping, exists := setup.ethBalances[address]
	return balanceMapping, exists
}

func (setup *TestSetup) checkHolderBalanceForTokenHelper(balanceMapping map[string]*big.Int, params TestTokenParams, actualBalance *big.Int, address string, actorState CurrentActorState) bool {
	extraBalances := setup.getExtraBalanceForHolder(address, params)
	expectedBalance := big.NewInt(0).Set(balanceMapping[params.AbstractTokenIdentifier])

	expectedBalance.Add(expectedBalance, extraBalances.ReceivedAmount)
	if actorState.isSender {
		expectedBalance.Add(expectedBalance, extraBalances.SentAmount)
		if actorState.HasToReceiveRefund {
			expectedBalance.Add(expectedBalance, extraBalances.RefundAmount)
		}
	}

	return actualBalance.String() == expectedBalance.String()
}

func (setup *TestSetup) getExtraBalanceForHolder(address string, params TestTokenParams) ExtraBalanceHolder {
	setup.mutBalances.Lock()
	defer setup.mutBalances.Unlock()

	holderName := AddressZero
	if address != "" {
		holderName = setup.AddressToName[address]
	}

	return params.ExtraBalances[holderName]
}

func (setup *TestSetup) checkContractMvxBalanceForToken(params TestTokenParams) bool {
	mvxBalance := setup.MultiversxHandler.GetESDTUniversalTokenBalance(setup.Ctx, setup.MultiversxHandler.CalleeScAddress, params.AbstractTokenIdentifier)
	expectedValueOnContract := big.NewInt(0)
	for _, operation := range params.TestOperations {
		if operation.ValueToTransferToMvx == nil {
			continue
		}
		if len(operation.MvxSCCallData) == 0 && !operation.MvxForceSCCall {
			continue
		}
		if operation.MvxFaultySCCall {
			continue
		}

		expectedValueOnContract.Add(expectedValueOnContract, operation.ValueToTransferToMvx)
	}

	return mvxBalance.String() == expectedValueOnContract.String()
}

func (setup *TestSetup) checkTokenOnEthFirstBridge(params TestTokenParams) bool {
	if params.IsMintBurnOnEth {
		return setup.checkEthBurnedTokenBalance(params)
	}

	return setup.checkEthLockedBalanceForToken(params, firstBridge)
}

func (setup *TestSetup) checkTokenOnMvxSecondBridge(params TestTokenParams) bool {
	if params.IsMintBurnOnMvX {
		return setup.checkMvxMintedBalanceForToken(params)
	}

	return setup.checkMvxLockedBalanceForToken(params, secondBridge)
}

func (setup *TestSetup) checkEthLockedBalanceForToken(params TestTokenParams, bridgeNumber currentBridge) bool {
	token := setup.GetTokenData(params.AbstractTokenIdentifier)
	lockedTokens := setup.EthereumHandler.GetTotalBalancesForToken(setup.Ctx, token.EthErc20Address)

	expectedValue := setup.computeExpectedValueToMvx(params)

	if params.InitialSupplyValue != "" {
		if initialSupply, ok := new(big.Int).SetString(params.InitialSupplyValue, 10); ok {
			expectedValue.Add(expectedValue, initialSupply)
		}
	}

	if bridgeNumber == secondBridge {
		expectedValue.Sub(expectedValue, setup.computeExpectedValueFromMvx(params))     // unlock amount of tokens sent back from Eth to Mvx
		expectedValue.Sub(expectedValue, setup.getMvxTotalRefundAmountForToken(params)) // unlock possible refund amount to be bridged back to Eth
	}

	return lockedTokens.String() == expectedValue.String()
}

func (setup *TestSetup) checkEthBurnedTokenBalance(params TestTokenParams) bool {
	token := setup.GetTokenData(params.AbstractTokenIdentifier)
	burnedTokens := setup.EthereumHandler.GetBurnBalanceForToken(setup.Ctx, token.EthErc20Address)

	expectedValue := setup.computeExpectedValueToMvx(params)

	if params.IsNativeOnEth {
		if params.InitialSupplyValue != "" {
			if initialSupply, ok := new(big.Int).SetString(params.InitialSupplyValue, 10); ok {
				expectedValue.Add(expectedValue, initialSupply)
			}
		}
	}

	return burnedTokens.String() == expectedValue.String()
}

func (setup *TestSetup) checkMvxMintedBalanceForToken(params TestTokenParams) bool {
	token := setup.GetTokenData(params.AbstractTokenIdentifier)
	mintedTokens := setup.MultiversxHandler.GetMintedAmountForToken(setup.Ctx, token.MvxChainSpecificToken)

	expectedValue := setup.computeExpectedValueToMvx(params)

	if params.IsNativeOnEth {
		if params.InitialSupplyValue != "" {
			if initialSupply, ok := new(big.Int).SetString(params.InitialSupplyValue, 10); ok {
				expectedValue.Add(expectedValue, initialSupply)
			}
		}
	}

	return mintedTokens.String() == expectedValue.String()
}

func (setup *TestSetup) computeExpectedValueToMvx(params TestTokenParams) *big.Int {
	expectedValue := big.NewInt(0)
	for _, operation := range params.TestOperations {
		if operation.ValueToTransferToMvx == nil {
			continue
		}
		if operation.IsFaultyDeposit {
			continue
		}

		expectedValue.Add(expectedValue, operation.ValueToTransferToMvx)
	}

	return expectedValue
}

func (setup *TestSetup) computeExpectedValueFromMvx(params TestTokenParams) *big.Int {
	expectedValue := big.NewInt(0)
	for _, operation := range params.TestOperations {
		if operation.ValueToSendFromMvX == nil {
			continue
		}
		if operation.IsFaultyDeposit {
			continue
		}

		expectedValue.Add(expectedValue, operation.ValueToSendFromMvX)
		expectedValue.Sub(expectedValue, feeInt)
	}

	return expectedValue
}

// IsTransferDoneFromEthereumWithRefund returns true if all provided tokens are bridged from Ethereum towards MultiversX including refunds
func (setup *TestSetup) IsTransferDoneFromEthereumWithRefund(holder KeysHolder, tokens ...TestTokenParams) bool {
	isDone := true
	for _, params := range tokens {
		isDone = isDone && setup.isTransferDoneFromEthWithRefundForToken(holder, params)
	}

	return isDone
}

func (setup *TestSetup) isTransferDoneFromEthWithRefundForToken(holder KeysHolder, params TestTokenParams) bool {
	// if token is prevented from whitelist, we can't check the balances
	if params.PreventWhitelist {
		return true
	}

	actorState := CurrentActorState{
		isSender:           true,
		HasToReceiveRefund: true,
	}

	return setup.checkHolderEthBalanceForToken(holder, actorState, params)
}

// IsTransferDoneFromMultiversX returns true if all provided tokens are bridged from MultiversX towards Ethereum
func (setup *TestSetup) IsTransferDoneFromMultiversX(sender KeysHolder, receiver KeysHolder, tokens ...TestTokenParams) bool {
	isDone := true
	for _, params := range tokens {
		isDone = isDone && setup.isTransferDoneFromMultiversXForToken(sender, receiver, params)
	}

	return isDone
}

func (setup *TestSetup) isTransferDoneFromMultiversXForToken(sender, receiver KeysHolder, params TestTokenParams) bool {
	if params.PreventWhitelist {
		return true
	}

	if !setup.checkHolderMvxBalanceForToken(sender, senderState, params) {
		return false
	}
	if !setup.checkHolderEthBalanceForToken(receiver, receiverState, params) {
		return false
	}

	if !setup.checkMvxBalanceForSafe(params) {
		return false
	}

	if !setup.checkTokenOnMvxFirstBridge(params) {
		return false
	}

	return setup.checkTokenOnEthSecondBridge(params)
}

func (setup *TestSetup) checkMvxBalanceForSafe(params TestTokenParams) bool {
	setup.mutBalances.Lock()
	initialBalance := setup.esdtBalanceForSafe[params.AbstractTokenIdentifier]
	setup.mutBalances.Unlock()

	expectedBalance := new(big.Int).Add(initialBalance, params.ESDTSafeExtraBalance)
	actualBalance := setup.MultiversxHandler.GetESDTChainSpecificTokenBalance(
		setup.Ctx, setup.MultiversxHandler.SafeAddress, params.AbstractTokenIdentifier,
	)

	return expectedBalance.Cmp(actualBalance) == 0
}

func (setup *TestSetup) checkTokenOnMvxFirstBridge(params TestTokenParams) bool {
	if params.IsMintBurnOnMvX {
		return setup.checkMvxBurnedTokenBalance(params)
	}

	return setup.checkMvxLockedBalanceForToken(params, firstBridge)
}

func (setup *TestSetup) checkTokenOnEthSecondBridge(params TestTokenParams) bool {
	if params.IsMintBurnOnEth {
		return setup.checkEthMintedBalanceForToken(params)
	}

	return setup.checkEthLockedBalanceForToken(params, secondBridge)
}

func (setup *TestSetup) checkMvxLockedBalanceForToken(params TestTokenParams, bridgeNumber currentBridge) bool {
	token := setup.GetTokenData(params.AbstractTokenIdentifier)
	lockedTokens := setup.MultiversxHandler.GetTotalBalancesForToken(setup.Ctx, token.MvxChainSpecificToken)

	expectedValue := setup.computeExpectedValueFromMvx(params)

	if params.InitialSupplyValue != "" {
		if initialSupply, ok := new(big.Int).SetString(params.InitialSupplyValue, 10); ok {
			expectedValue.Add(expectedValue, initialSupply)
		}
	}

	if bridgeNumber == secondBridge {
		expectedValue.Sub(expectedValue, setup.computeExpectedValueToMvx(params))
		expectedValue.Add(expectedValue, setup.getMvxTotalRefundAmountForToken(params)) // lock possible refund amount from failed SC call on Mvx
	}

	return lockedTokens.String() == expectedValue.String()
}

func (setup *TestSetup) checkMvxBurnedTokenBalance(params TestTokenParams) bool {
	token := setup.GetTokenData(params.AbstractTokenIdentifier)
	burnedTokens := setup.MultiversxHandler.GetBurnedAmountForToken(setup.Ctx, token.MvxChainSpecificToken)

	expectedValue := setup.computeExpectedValueFromMvx(params)

	if params.IsNativeOnMvX {
		if params.InitialSupplyValue != "" {
			if initialSupply, ok := new(big.Int).SetString(params.InitialSupplyValue, 10); ok {
				expectedValue.Add(expectedValue, initialSupply)
			}
		}
	} else {
		expectedValue.Add(expectedValue, setup.getMvxTotalRefundAmountForToken(params)) // burn possible refund amount to be bridged back to Eth
	}

	return burnedTokens.String() == expectedValue.String()
}

func (setup *TestSetup) checkEthMintedBalanceForToken(params TestTokenParams) bool {
	token := setup.GetTokenData(params.AbstractTokenIdentifier)
	mintedTokens := setup.EthereumHandler.GetMintBalanceForToken(setup.Ctx, token.EthErc20Address)

	expectedValue := setup.computeExpectedValueFromMvx(params)

	if params.IsNativeOnMvX {
		if params.InitialSupplyValue != "" {
			if initialSupply, ok := new(big.Int).SetString(params.InitialSupplyValue, 10); ok {
				expectedValue.Add(expectedValue, initialSupply)
			}
		}
	} else {
		expectedValue.Add(expectedValue, setup.getMvxTotalRefundAmountForToken(params)) // mint possible refund amount from failed SC call on Mvx
	}

	return mintedTokens.String() == expectedValue.String()
}

func (setup *TestSetup) getMvxTotalRefundAmountForToken(params TestTokenParams) *big.Int {
	totalRefund := big.NewInt(0)
	for _, operation := range params.TestOperations {
		if len(operation.MvxSCCallData) == 0 && !operation.MvxForceSCCall {
			continue
		}
		if !operation.MvxFaultySCCall {
			continue
		}

		// the balance should be bridged back to the receiver on Ethereum - fee
		totalRefund.Add(totalRefund, operation.ValueToTransferToMvx)
		totalRefund.Sub(totalRefund, feeInt)
	}
	return totalRefund
}

// CreateBatchOnMultiversX will create deposits that will be gathered in a batch on MultiversX
func (setup *TestSetup) CreateBatchOnMultiversX(tokensParams ...TestTokenParams) {
	for _, params := range tokensParams {
		setup.createBatchOnMultiversXForToken(params)
	}
}

func (setup *TestSetup) createBatchOnMultiversXForToken(params TestTokenParams) {
	token := setup.GetTokenData(params.AbstractTokenIdentifier)
	require.NotNil(setup, token)

	// TODO: transfer only required amount for deposit to the test key
	valueToMintOnEthereum := setup.createDepositOnMultiversxForToken(setup.AliceKeys, setup.BobKeys, params)
	setup.EthereumHandler.Mint(setup.Ctx, params, valueToMintOnEthereum)
}

func (setup *TestSetup) transferTokensToMvxTestKey(params TestTokenParams) {
	depositValue := big.NewInt(0)
	for _, operation := range params.TestOperations {
		if operation.ValueToSendFromMvX == nil {
			continue
		}

		depositValue.Add(depositValue, operation.ValueToSendFromMvX)
	}

	setup.MultiversxHandler.TransferToken(
		setup.Ctx,
		setup.OwnerKeys,
		setup.AliceKeys,
		depositValue,
		params.IssueTokenParams,
	)
}

// SendFromMultiversxToEthereum will create the deposits that will be gathered in a batch on MultiversX (without mint on Ethereum)
func (setup *TestSetup) SendFromMultiversxToEthereum(from KeysHolder, to KeysHolder, tokensParams ...TestTokenParams) {
	for _, params := range tokensParams {
		_ = setup.createDepositOnMultiversxForToken(from, to, params)
	}
}

func (setup *TestSetup) createDepositOnMultiversxForToken(from KeysHolder, to KeysHolder, params TestTokenParams) *big.Int {
	token := setup.GetTokenData(params.AbstractTokenIdentifier)
	require.NotNil(setup, token)

	depositValue := big.NewInt(0)
	for _, operation := range params.TestOperations {
		if operation.ValueToSendFromMvX == nil {
			continue
		}

		if operation.InvalidReceiver != nil {
			invalidReceiver := common.Address(operation.InvalidReceiver)
			to = KeysHolder{EthAddress: invalidReceiver}
		}

		depositValue.Add(depositValue, operation.ValueToSendFromMvX)

		if operation.IsFaultyDeposit || params.PreventWhitelist {
			setup.MultiversxHandler.SendWrongDepositTransactionFromMultiversx(setup.Ctx, from, to, token, operation.ValueToSendFromMvX)
		} else {
			setup.MultiversxHandler.SendDepositTransactionFromMultiversx(setup.Ctx, from, to, token, params, operation.ValueToSendFromMvX)
		}
	}

	return depositValue
}

// CreateBatchOnEthereum will create deposits that will be gathered in a batch on Ethereum
func (setup *TestSetup) CreateBatchOnEthereum(mvxCalleeScAddress sdkCore.AddressHandler, tokensParams ...TestTokenParams) {
	for _, params := range tokensParams {
		setup.createBatchOnEthereumForToken(mvxCalleeScAddress, params)
	}

	// wait until batch is settled
	setup.EthereumHandler.SettleBatchOnEthereum()
}

func (setup *TestSetup) createBatchOnEthereumForToken(mvxCalleeScAddress sdkCore.AddressHandler, params TestTokenParams) {
	token := setup.GetTokenData(params.AbstractTokenIdentifier)
	require.NotNil(setup, token)

	// TODO: transfer only required amount for deposit to the test key
	setup.createDepositOnEthereumForToken(setup.AliceKeys, setup.BobKeys, mvxCalleeScAddress, params)
}

// SendFromEthereumToMultiversX will create the deposits that will be gathered in a batch on Ethereum
func (setup *TestSetup) SendFromEthereumToMultiversX(from KeysHolder, to KeysHolder, mvxTestCallerAddress sdkCore.AddressHandler, tokensParams ...TestTokenParams) {
	for _, params := range tokensParams {
		setup.createDepositOnEthereumForToken(from, to, mvxTestCallerAddress, params)
	}
}

func (setup *TestSetup) createDepositOnEthereumForToken(from KeysHolder, to KeysHolder, targetSCAddress sdkCore.AddressHandler, params TestTokenParams) {
	token := setup.GetTokenData(params.AbstractTokenIdentifier)
	require.NotNil(setup, token)
	require.NotNil(setup, token.EthErc20Contract)

	allowanceValue := big.NewInt(0)
	for _, operation := range params.TestOperations {
		if operation.ValueToTransferToMvx == nil {
			continue
		}

		allowanceValue.Add(allowanceValue, operation.ValueToTransferToMvx)
	}

	if allowanceValue.Cmp(zeroValueBigInt) > 0 {
		setup.EthereumHandler.ApproveForToken(setup.Ctx, token, from, setup.EthereumHandler.SafeAddress, allowanceValue)
	}

	for _, operation := range params.TestOperations {
		if operation.ValueToTransferToMvx == nil {
			continue
		}

		if operation.InvalidReceiver != nil {
			invalidReceiver := NewMvxAddressFromBytes(setup, operation.InvalidReceiver)
			to = KeysHolder{MvxAddress: invalidReceiver}
		}

		setup.EthereumHandler.SendDepositTransactionFromEthereum(setup.Ctx, from, to, targetSCAddress, token, operation)
	}
}

// TestWithdrawTotalFeesOnEthereumForTokens will test the withdrawal functionality for the provided test tokens
func (setup *TestSetup) TestWithdrawTotalFeesOnEthereumForTokens(tokensParams ...TestTokenParams) {
	for _, param := range tokensParams {
		token := setup.TokensRegistry.GetTokenData(param.AbstractTokenIdentifier)

		expectedRefund := big.NewInt(0)
		expectedAccumulated := big.NewInt(0)

		if param.PreventWhitelist {
			continue
		}

		for _, operation := range param.TestOperations {
			if operation.InvalidReceiver != nil {
				expectedRefund.Add(expectedRefund, feeInt)
			}

			if operation.ValueToSendFromMvX == nil {
				continue
			}
			if operation.ValueToSendFromMvX.Cmp(zeroValueBigInt) == 0 {
				continue
			}
			if operation.IsFaultyDeposit {
				continue
			}

			expectedAccumulated.Add(expectedAccumulated, feeInt)
		}

		setup.MultiversxHandler.TestWithdrawFees(setup.Ctx, token.MvxChainSpecificToken, expectedRefund, expectedAccumulated)
	}
}

// Close will close the test subcomponents
func (setup *TestSetup) Close() {
	log.Info(fmt.Sprintf(LogStepMarker, "closing relayers & sc execution module"))

	setup.Bridge.CloseRelayers()
	require.NoError(setup, setup.EthereumHandler.Close())

	setup.ctxCancel()
	_ = setup.ScCallerModuleInstance.Close()
}

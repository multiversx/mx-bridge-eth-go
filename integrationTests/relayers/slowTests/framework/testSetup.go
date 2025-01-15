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
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/multiversx/mx-bridge-eth-go/config"
	"github.com/multiversx/mx-bridge-eth-go/executors/multiversx/module"
	"github.com/multiversx/mx-sdk-go/blockchain"
	sdkCore "github.com/multiversx/mx-sdk-go/core"
	"github.com/stretchr/testify/require"
)

// framework constants
const (
	LogStepMarker                = "#################################### %s ####################################"
	proxyCacherExpirationSeconds = 600
	proxyMaxNoncesDelta          = 7
	NumRelayers                  = 3
	NumOracles                   = 3
	quorum                       = "03"
	mvxHrp                       = "erd"
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
	ProxyWrapperInstance   *proxyWrapper

	ctxCancel   func()
	Ctx         context.Context
	mutBalances sync.RWMutex
	mvxBalances map[string]map[string]*big.Int
	ethBalances map[string]map[string]*big.Int

	numScCallsInTest uint32
}

// NewTestSetup creates a new e2e test setup
func NewTestSetup(tb testing.TB) *TestSetup {
	log.Info(fmt.Sprintf(LogStepMarker, "starting setup"))

	setup := &TestSetup{
		TB:             tb,
		TokensRegistry: NewTokenRegistry(tb),
		WorkingDir:     tb.TempDir(),
		mvxBalances:    make(map[string]map[string]*big.Int),
		ethBalances:    make(map[string]map[string]*big.Int),
	}
	setup.KeysStore = NewKeysStore(tb, setup.WorkingDir, NumRelayers, NumOracles)

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
		General: config.GeneralScCallsModuleConfig{
			ScProxyBech32Addresses: []string{
				setup.MultiversxHandler.ScProxyAddress.Bech32(),
			},
			NetworkAddress:               setup.ChainSimulator.GetNetworkAddress(),
			ProxyMaxNoncesDelta:          7,
			ProxyFinalityCheck:           true,
			ProxyCacherExpirationSeconds: 60,
			ProxyRestAPIEntityType:       string(sdkCore.Proxy),
			IntervalToResendTxsInSeconds: 1,
			PrivateKeyFile:               path.Join(setup.WorkingDir, SCCallerFilename),
		},
		ScCallsExecutor: config.ScCallsExecutorConfig{
			ExtraGasToExecute:               60_000_000,  // 60 million: this ensures that a SC call with 0 gas limit is refunded
			MaxGasLimitToUse:                249_999_999, // max cross shard limit
			GasLimitForOutOfGasTransactions: 30_000_000,  // gas to use when a higher than max allowed is encountered
			PollingIntervalInMillis:         1000,        // 1 second
			TTLForFailedRefundIdInSeconds:   1,           // 1 second
		},
		RefundExecutor: config.RefundExecutorConfig{
			GasToExecute:                  30_000_000,
			PollingIntervalInMillis:       1000,
			TTLForFailedRefundIdInSeconds: 1,
		},
		Filter: config.PendingOperationsFilterConfig{
			AllowedEthAddresses: []string{"*"},
			AllowedMvxAddresses: []string{"*"},
			AllowedTokens:       []string{"*"},
		},
		Logs: config.LogsConfig{},
		TransactionChecks: config.TransactionChecksConfig{
			ExecutionTimeoutInSeconds:  2,
			TimeInSecondsBetweenChecks: 1,
		},
	}

	argsProxy := blockchain.ArgsProxy{
		ProxyURL:            cfg.General.NetworkAddress,
		SameScState:         false,
		ShouldBeSynced:      false,
		FinalityCheck:       cfg.General.ProxyFinalityCheck,
		AllowedDeltaToFinal: cfg.General.ProxyMaxNoncesDelta,
		CacheExpirationTime: time.Second * time.Duration(cfg.General.ProxyCacherExpirationSeconds),
		EntityType:          sdkCore.RestAPIEntityType(cfg.General.ProxyRestAPIEntityType),
	}

	proxy, err := blockchain.NewProxy(argsProxy)
	require.Nil(setup, err)

	setup.ProxyWrapperInstance = NewProxyWrapper(proxy)

	argsScCallsModule := module.ArgsScCallsModule{
		Config: cfg,
		Proxy:  setup.ProxyWrapperInstance,
		Log:    log,
	}

	setup.ScCallerModuleInstance, err = module.NewScCallsModule(argsScCallsModule)
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

		setup.mutBalances.Lock()
		setup.initMvxInitialBalancesForUniversalUnsafe(token,
			setup.AliceKeys.MvxAddress,
			setup.BobKeys.MvxAddress,
			setup.CharlieKeys.MvxAddress,
			setup.MultiversxHandler.WrapperAddress,
			setup.MultiversxHandler.CalleeScAddress,
		)
		setup.initMvxInitialBalancesForChainSpecificUnsafe(token,
			setup.MultiversxHandler.SafeAddress,
			setup.MultiversxHandler.WrapperAddress,
		)

		setup.initEthInitialBalancesUnsafe(token,
			setup.AliceKeys.EthAddress,
			setup.BobKeys.EthAddress,
			setup.CharlieKeys.EthAddress,
			setup.EthereumHandler.SafeAddress,
		)
		setup.mutBalances.Unlock()

		esdtBalanceForSafe := setup.MultiversxHandler.GetESDTChainSpecificTokenBalance(setup.Ctx, setup.MultiversxHandler.SafeAddress, token.AbstractTokenIdentifier)
		log.Info("recorded the ESDT balance for safe contract", "token", token.AbstractTokenIdentifier, "balance", esdtBalanceForSafe.String())
	}

	setup.EthereumHandler.UnPauseContractsAfterTokenChanges(setup.Ctx)
	setup.MultiversxHandler.UnPauseContractsAfterTokenChanges(setup.Ctx)

	for _, token := range tokens {
		setup.MultiversxHandler.SubmitAggregatorBatch(setup.Ctx, token.IssueTokenParams, token.MvxToEthFee)
	}
}

func (setup *TestSetup) initMvxInitialBalancesForUniversalUnsafe(token TestTokenParams, addresses ...*MvxAddress) {
	for _, addr := range addresses {
		if setup.mvxBalances[addr.String()] == nil {
			setup.mvxBalances[addr.String()] = make(map[string]*big.Int)
		}

		setup.mvxBalances[addr.String()][token.AbstractTokenIdentifier] = setup.MultiversxHandler.GetESDTUniversalTokenBalance(setup.Ctx, addr, token.AbstractTokenIdentifier)
	}
}

func (setup *TestSetup) initMvxInitialBalancesForChainSpecificUnsafe(token TestTokenParams, addresses ...*MvxAddress) {
	for _, addr := range addresses {
		if setup.mvxBalances[addr.String()] == nil {
			setup.mvxBalances[addr.String()] = make(map[string]*big.Int)
		}

		setup.mvxBalances[addr.String()][token.AbstractTokenIdentifier] = setup.MultiversxHandler.GetESDTChainSpecificTokenBalance(setup.Ctx, addr, token.AbstractTokenIdentifier)
	}
}

func (setup *TestSetup) initEthInitialBalancesUnsafe(token TestTokenParams, addresses ...common.Address) {
	for _, addr := range addresses {
		if setup.ethBalances[addr.String()] == nil {
			setup.ethBalances[addr.String()] = make(map[string]*big.Int)
		}

		setup.ethBalances[addr.String()][token.AbstractTokenIdentifier] = setup.EthereumHandler.GetBalance(addr, token.AbstractTokenIdentifier)
	}
}

func (setup *TestSetup) processNumScCallsOperations(token TestTokenParams) {
	for _, op := range token.TestOperations {
		if len(op.MvxSCCallData) > 0 || op.MvxForceSCCall {
			atomic.AddUint32(&setup.numScCallsInTest, 1)
			if op.MvxFaultySCCall {
				// one more call for the refund operation
				atomic.AddUint32(&setup.numScCallsInTest, 1)
			}
		}
	}
}

// GetNumScCallsOperations returns the number of SC calls in this test setup
func (setup *TestSetup) GetNumScCallsOperations() uint32 {
	return atomic.LoadUint32(&setup.numScCallsInTest)
}

// AreAllTransfersCompleted returns true if the delta balances match the current test users' balances
func (setup *TestSetup) AreAllTransfersCompleted(halfBridgeIdentifier HalfBridgeIdentifier, tokens ...TestTokenParams) bool {
	isDone := true
	for _, params := range tokens {
		isDone = isDone && setup.isTransferDone(halfBridgeIdentifier, params)
	}

	return isDone
}

func (setup *TestSetup) isTransferDone(halfBridgeIdentifier HalfBridgeIdentifier, token TestTokenParams) bool {
	// if token is prevented from whitelist, we can't check the balances
	if token.PreventWhitelist {
		return true
	}

	deltaBalancesMap := token.DeltaBalances[halfBridgeIdentifier]
	require.NotNil(setup, deltaBalancesMap)

	for entityName, deltaBalances := range deltaBalancesMap {
		if !setup.isBalanceOkOnMvx(entityName, deltaBalances, token) {
			return false
		}

		if !setup.isBalanceOkOnEth(entityName, deltaBalances.OnEth, token) {
			return false
		}
	}

	return true
}

func (setup *TestSetup) isBalanceOkOnMvx(entityName string, deltaBalance *DeltaBalanceHolder, token TestTokenParams) bool {
	address := setup.getMvxAddressFromEntityName(entityName)

	intialBalance := setup.getBalanceMappingForAddressAndToken(address.Bech32(), token)
	expectedBalance := big.NewInt(0).Set(deltaBalance.OnMvx)
	expectedBalance.Add(expectedBalance, intialBalance)

	var actualBalance *big.Int
	switch deltaBalance.MvxToken {
	case UniversalToken:
		actualBalance = setup.MultiversxHandler.GetESDTUniversalTokenBalance(setup.Ctx, address, token.AbstractTokenIdentifier)
	case ChainSpecificToken:
		actualBalance = setup.MultiversxHandler.GetESDTChainSpecificTokenBalance(setup.Ctx, address, token.AbstractTokenIdentifier)
	default:
		require.Fail(setup, fmt.Sprintf("Unknown balance type %s for entity name %s", deltaBalance.MvxToken, entityName))
	}

	return expectedBalance.String() == actualBalance.String()
}

func (setup *TestSetup) getMvxAddressFromEntityName(entityName string) *MvxAddress {
	switch entityName {
	case Alice:
		return setup.AliceKeys.MvxAddress
	case Bob:
		return setup.BobKeys.MvxAddress
	case Charlie:
		return setup.CharlieKeys.MvxAddress
	case SafeSC:
		return setup.MultiversxHandler.SafeAddress
	case WrapperSC:
		return setup.MultiversxHandler.WrapperAddress
	case CalledTestSC:
		return setup.MultiversxHandler.CalleeScAddress
	}

	require.Fail(setup, fmt.Sprintf("getMvxAddressFromEntityName: unknown entity name %s", entityName))
	return nil
}

func (setup *TestSetup) getBalanceMappingForAddressAndToken(address string, token TestTokenParams) *big.Int {
	setup.mutBalances.Lock()
	defer setup.mutBalances.Unlock()

	if strings.HasPrefix(address, mvxHrp) {
		balanceMapping, exists := setup.mvxBalances[address]
		if !exists {
			return big.NewInt(0)
		}

		return balanceMapping[token.AbstractTokenIdentifier]
	}

	balanceMapping, exists := setup.ethBalances[address]
	if !exists {
		return big.NewInt(0)
	}

	return balanceMapping[token.AbstractTokenIdentifier]
}

func (setup *TestSetup) isBalanceOkOnEth(entityName string, expectedDeltaBalance *big.Int, token TestTokenParams) bool {
	address, shouldCheck := setup.getEthAddressFromEntityName(entityName)
	if !shouldCheck {
		return true
	}

	intialBalance := setup.getBalanceMappingForAddressAndToken(address.String(), token)
	expectedBalance := big.NewInt(0).Set(expectedDeltaBalance)
	expectedBalance.Add(expectedBalance, intialBalance)

	actualBalance := setup.EthereumHandler.GetBalance(address, token.AbstractTokenIdentifier)

	return expectedBalance.String() == actualBalance.String()
}

func (setup *TestSetup) getEthAddressFromEntityName(entityName string) (common.Address, bool) {
	switch entityName {
	case Alice:
		return setup.AliceKeys.EthAddress, true
	case Bob:
		return setup.BobKeys.EthAddress, true
	case Charlie:
		return setup.CharlieKeys.EthAddress, true
	case SafeSC:
		return setup.EthereumHandler.SafeAddress, true
	case WrapperSC, CalledTestSC:
		return common.Address{}, false
	}

	require.Fail(setup, fmt.Sprintf("getEthAddressFromEntityName: unknown entity name %s", entityName))
	return common.Address{}, false
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

	setup.transferTokensToMvxTestKey(params, setup.AliceKeys)
	setup.ChainSimulator.GenerateBlocks(setup.Ctx, 10)

	setup.mutBalances.Lock()
	setup.initMvxInitialBalancesForUniversalUnsafe(params, setup.AliceKeys.MvxAddress)
	setup.mutBalances.Unlock()

	_ = setup.createDepositOnMultiversxForToken(setup.AliceKeys, setup.BobKeys, params)
}

func (setup *TestSetup) transferTokensToMvxTestKey(params TestTokenParams, holder KeysHolder) {
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
		holder,
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

		if operation.InvalidReceiver != nil && !setup.hasCallData(operation) {
			to = KeysHolder{EthAddress: operation.InvalidReceiver.(common.Address)}
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

	setup.transferTokensToEthTestKey(params, setup.AliceKeys)

	setup.mutBalances.Lock()
	setup.initEthInitialBalancesUnsafe(params, setup.AliceKeys.EthAddress)
	setup.mutBalances.Unlock()

	setup.createDepositOnEthereumForToken(setup.AliceKeys, setup.BobKeys, mvxCalleeScAddress, params)
}

func (setup *TestSetup) transferTokensToEthTestKey(params TestTokenParams, holder KeysHolder) {
	depositValue := big.NewInt(0)
	for _, operation := range params.TestOperations {
		if operation.ValueToTransferToMvx == nil {
			continue
		}

		depositValue.Add(depositValue, operation.ValueToTransferToMvx)
	}

	if params.MultipleSpendings != nil {
		depositValue.Mul(depositValue, params.MultipleSpendings)
	}

	setup.EthereumHandler.TransferToken(
		setup.Ctx,
		params,
		setup.DepositorKeys,
		holder,
		depositValue)
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
			invalidReceiver := NewMvxAddressFromBytes(setup, operation.InvalidReceiver.([]byte))

			if setup.hasCallData(operation) {
				targetSCAddress = invalidReceiver
			} else {
				to = KeysHolder{MvxAddress: invalidReceiver}
			}
		}

		setup.EthereumHandler.SendDepositTransactionFromEthereum(setup.Ctx, from, to, targetSCAddress, token, operation)
	}
}

func (setup *TestSetup) hasCallData(operation TokenOperations) bool {
	return len(operation.MvxSCCallData) != 0 || operation.MvxForceSCCall
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
			if operation.IsFaultyDeposit {
				continue
			}

			if operation.InvalidReceiver != nil {
				expectedRefund.Add(expectedRefund, param.MvxToEthFee)
			}

			if operation.ValueToSendFromMvX == nil {
				continue
			}
			if operation.ValueToSendFromMvX.Cmp(zeroValueBigInt) == 0 {
				continue
			}

			expectedAccumulated.Add(expectedAccumulated, param.MvxToEthFee)
		}

		setup.MultiversxHandler.TestWithdrawFees(setup.Ctx, token.MvxChainSpecificToken, expectedRefund, expectedAccumulated)
	}
}

// CheckCorrectnessOnMintBurnTokens will check the correctness on the mint/burn tokens
func (setup *TestSetup) CheckCorrectnessOnMintBurnTokens(tokens ...TestTokenParams) {
	for _, params := range tokens {
		setup.checkTotalMintBurnOnMvx(params)
		setup.checkMintBurnOnEth(params)
		setup.checkSafeContractMintBurnOnMvx(params)
	}
}

// ExecuteSpecialChecks will trigger the special checks
func (setup *TestSetup) ExecuteSpecialChecks(tokens ...TestTokenParams) {
	for _, params := range tokens {
		setup.executeSpecialChecks(params)
	}
}

func (setup *TestSetup) checkTotalMintBurnOnMvx(token TestTokenParams) {
	tokenData := setup.TokensRegistry.GetTokenData(token.AbstractTokenIdentifier)

	esdtSupplyForUniversal := setup.MultiversxHandler.ChainSimulator.GetESDTSupplyValues(setup.Ctx, tokenData.MvxUniversalToken)
	require.Equal(setup, token.MintBurnChecks.MvxTotalUniversalMint.String(), esdtSupplyForUniversal.Minted, fmt.Sprintf("token: %s", tokenData.MvxUniversalToken))
	require.Equal(setup, token.MintBurnChecks.MvxTotalUniversalBurn.String(), esdtSupplyForUniversal.Burned, fmt.Sprintf("token: %s", tokenData.MvxUniversalToken))

	if tokenData.MvxUniversalToken == tokenData.MvxChainSpecificToken {
		// we do not have a chain specific token, we can return true here
		return
	}

	esdtSupplyForChainSpecific := setup.MultiversxHandler.ChainSimulator.GetESDTSupplyValues(setup.Ctx, tokenData.MvxChainSpecificToken)
	require.Equal(setup, token.MintBurnChecks.MvxTotalChainSpecificMint.String(), esdtSupplyForChainSpecific.Minted, fmt.Sprintf("token: %s", tokenData.MvxChainSpecificToken))
	require.Equal(setup, token.MintBurnChecks.MvxTotalChainSpecificBurn.String(), esdtSupplyForChainSpecific.Burned, fmt.Sprintf("token: %s", tokenData.MvxChainSpecificToken))
}

func (setup *TestSetup) checkMintBurnOnEth(token TestTokenParams) {
	tokenData := setup.GetTokenData(token.AbstractTokenIdentifier)

	minted := setup.EthereumHandler.GetMintBalanceForToken(setup.Ctx, tokenData.EthErc20Address)
	require.Equal(setup, token.MintBurnChecks.EthSafeMintValue.String(), minted.String(), fmt.Sprintf("eth safe contract, token: %s", tokenData.EthErc20Address.String()))

	burned := setup.EthereumHandler.GetBurnBalanceForToken(setup.Ctx, tokenData.EthErc20Address)
	require.Equal(setup, token.MintBurnChecks.EthSafeBurnValue.String(), burned.String(), fmt.Sprintf("eth safe contract, token: %s", tokenData.EthErc20Address.String()))
}

func (setup *TestSetup) checkSafeContractMintBurnOnMvx(token TestTokenParams) {
	tokenData := setup.TokensRegistry.GetTokenData(token.AbstractTokenIdentifier)

	minted := setup.MultiversxHandler.GetMintedAmountForToken(setup.Ctx, tokenData.MvxChainSpecificToken)
	require.Equal(setup, token.MintBurnChecks.MvxSafeMintValue.String(), minted.String(), fmt.Sprintf("Mvx safe contract, token: %s", tokenData.MvxChainSpecificToken))

	burn := setup.MultiversxHandler.GetBurnedAmountForToken(setup.Ctx, tokenData.MvxChainSpecificToken)
	require.Equal(setup, token.MintBurnChecks.MvxSafeBurnValue.String(), burn.String(), fmt.Sprintf("Mvx safe contract, token: %s", tokenData.MvxChainSpecificToken))
}

func (setup *TestSetup) executeSpecialChecks(token TestTokenParams) {
	tokenData := setup.TokensRegistry.GetTokenData(token.AbstractTokenIdentifier)

	actualValue := setup.MultiversxHandler.GetWrapperLiquidity(setup.Ctx, tokenData.MvxChainSpecificToken)
	initialBalance := setup.getBalanceMappingForAddressAndToken(setup.MultiversxHandler.WrapperAddress.Bech32(), token)
	expectedValue := big.NewInt(0).Add(initialBalance, token.SpecialChecks.WrapperDeltaLiquidityCheck)

	require.Equal(setup, expectedValue.String(), actualValue.String(), fmt.Sprintf("wrapper contract, token: %s", tokenData.MvxChainSpecificToken))
}

// Close will close the test subcomponents
func (setup *TestSetup) Close() {
	log.Info(fmt.Sprintf(LogStepMarker, "closing relayers & sc execution module"))

	setup.Bridge.CloseRelayers()
	require.NoError(setup, setup.EthereumHandler.Close())

	setup.ctxCancel()
	_ = setup.ScCallerModuleInstance.Close()
}

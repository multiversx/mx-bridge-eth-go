package framework

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"path"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/multiversx/mx-bridge-eth-go/config"
	"github.com/multiversx/mx-bridge-eth-go/executors/multiversx/module"
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
)

var addressToName = make(map[string]string)

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

	addressToName[setup.AliceKeys.EthAddress.String()] = "Alice"
	addressToName[setup.BobKeys.EthAddress.String()] = "Bob"
	addressToName[setup.CharlieKeys.EthAddress.String()] = "Charlie"
	addressToName[setup.AliceKeys.MvxAddress.String()] = "Alice"
	addressToName[setup.BobKeys.MvxAddress.String()] = "Bob"
	addressToName[setup.CharlieKeys.MvxAddress.String()] = "Charlie"

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

		setup.ChainSimulator.GenerateBlocks(setup.Ctx, 10)

		esdtBalanceForSafe := setup.MultiversxHandler.GetESDTChainSpecificTokenBalance(setup.Ctx, setup.MultiversxHandler.SafeAddress, token.AbstractTokenIdentifier)
		AliceMvxBalance := setup.MultiversxHandler.GetESDTUniversalTokenBalance(setup.Ctx, setup.AliceKeys.MvxAddress, token.AbstractTokenIdentifier)
		AliceEthBalance := setup.EthereumHandler.GetBalance(setup.AliceKeys.EthAddress, token.AbstractTokenIdentifier)

		//fmt.Println("@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@", AliceMvxBalance)

		setup.mutBalances.Lock()
		setup.esdtBalanceForSafe[token.AbstractTokenIdentifier] = esdtBalanceForSafe

		setup.mvxBalances[setup.AliceKeys.MvxAddress.String()][token.AbstractTokenIdentifier] = AliceMvxBalance
		setup.mvxBalances[setup.BobKeys.MvxAddress.String()][token.AbstractTokenIdentifier] = big.NewInt(0)
		setup.mvxBalances[setup.CharlieKeys.MvxAddress.String()][token.AbstractTokenIdentifier] = big.NewInt(0)

		setup.ethBalances[setup.AliceKeys.EthAddress.String()][token.AbstractTokenIdentifier] = AliceEthBalance
		setup.ethBalances[setup.BobKeys.EthAddress.String()][token.AbstractTokenIdentifier] = big.NewInt(0)
		setup.ethBalances[setup.CharlieKeys.EthAddress.String()][token.AbstractTokenIdentifier] = big.NewInt(0)

		setup.mutBalances.Unlock()

		log.Info("recorded the ESDT balance for safe contract", "token", token.AbstractTokenIdentifier, "balance", esdtBalanceForSafe.String())
		log.Info("recorded the ETH balance for Alice", "token", token.AbstractTokenIdentifier, "balance", AliceEthBalance.String())
	}

	setup.EthereumHandler.UnPauseContractsAfterTokenChanges(setup.Ctx)
	setup.MultiversxHandler.UnPauseContractsAfterTokenChanges(setup.Ctx)

	for _, token := range tokens {
		setup.MultiversxHandler.SubmitAggregatorBatch(setup.Ctx, token.IssueTokenParams)
	}
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

func (setup *TestSetup) isTransferDoneFromEthereumForToken(sender KeysHolder, receiver KeysHolder, params TestTokenParams) bool {
	okSender := setup.checkHolderEthBalanceForToken(sender, true, params)
	okReceiver := setup.checkHolderMvxBalanceForToken(receiver, false, params)
	okContract := setup.checkContractMvxBalanceForToken(params)

	return okSender && okReceiver && okContract
}

func (setup *TestSetup) checkHolderMvxBalanceForToken(holder KeysHolder, isSender bool, params TestTokenParams) bool {
	setup.mutBalances.Lock()

	addr := holder.MvxAddress.String()
	balanceMapping, exists := setup.mvxBalances[addr]
	if !exists {
		return false
	}
	expectedBalance := big.NewInt(0).Set(balanceMapping[params.AbstractTokenIdentifier])
	holderName := addressToName[addr]
	transferAmounts := params.EthTestAddrsExtraBalances[holderName]

	if isSender {
		expectedBalance.Add(expectedBalance, transferAmounts[0])
	} else {
		expectedBalance.Add(expectedBalance, transferAmounts[1])
	}

	balanceMapping[params.AbstractTokenIdentifier] = expectedBalance

	setup.mutBalances.Unlock()

	actualBalance := setup.MultiversxHandler.GetESDTUniversalTokenBalance(setup.Ctx, holder.MvxAddress, params.AbstractTokenIdentifier)

	fmt.Println("------------MULTIVERSX-------------------")
	fmt.Println("actualBalance: ", actualBalance)
	fmt.Println("expectedBalance: ", expectedBalance)
	fmt.Println("-------------------------------")
	return actualBalance.String() == expectedBalance.String()
}

func (setup *TestSetup) checkContractMvxBalanceForToken(params TestTokenParams) bool {
	mvxBalance := setup.MultiversxHandler.GetESDTUniversalTokenBalance(setup.Ctx, setup.MultiversxHandler.CalleeScAddress, params.AbstractTokenIdentifier)
	expectedValueOnContract := big.NewInt(0)
	for _, operation := range params.TestOperations {
		if operation.ValueToTransferToMvx == nil {
			continue
		}

		if len(operation.MvxSCCallData) > 0 || operation.MvxForceSCCall {
			if !operation.MvxFaultySCCall {
				expectedValueOnContract.Add(expectedValueOnContract, operation.ValueToTransferToMvx)
			}
		}
	}

	return mvxBalance.String() == expectedValueOnContract.String()
}

// IsTransferDoneFromEthereumWithRefund returns true if all provided tokens are bridged from Ethereum towards MultiversX including refunds
func (setup *TestSetup) IsTransferDoneFromEthereumWithRefund(tokens ...TestTokenParams) bool {
	isDone := true
	for _, params := range tokens {
		isDone = isDone && setup.isTransferDoneFromEthereumWithRefundForToken(params)
	}

	return isDone
}

func (setup *TestSetup) isTransferDoneFromEthereumWithRefundForToken(params TestTokenParams) bool {
	expectedValueOnReceiver := big.NewInt(0)
	for _, operation := range params.TestOperations {
		valueToTransferToMvx := big.NewInt(0)
		if operation.ValueToTransferToMvx != nil {
			valueToTransferToMvx.Set(operation.ValueToTransferToMvx)
		}

		valueToSendFromMvX := big.NewInt(0)
		if operation.ValueToSendFromMvX != nil {
			valueToSendFromMvX.Set(operation.ValueToSendFromMvX)
			// we subtract the fee also
			expectedValueOnReceiver.Sub(expectedValueOnReceiver, feeInt)
		}

		expectedValueOnReceiver.Add(expectedValueOnReceiver, big.NewInt(0).Sub(valueToSendFromMvX, valueToTransferToMvx))
		if len(operation.MvxSCCallData) > 0 || operation.MvxForceSCCall {
			if operation.MvxFaultySCCall {
				// the balance should be bridged back to the receiver on Ethereum - fee
				expectedValueOnReceiver.Add(expectedValueOnReceiver, valueToTransferToMvx)
				expectedValueOnReceiver.Sub(expectedValueOnReceiver, feeInt)
			}
		}
	}

	receiverBalance := setup.EthereumHandler.GetBalance(setup.BobKeys.EthAddress, params.AbstractTokenIdentifier)
	return receiverBalance.String() == expectedValueOnReceiver.String()
}

// IsTransferDoneFromMultiversX returns true if all provided tokens are bridged from MultiversX towards Ethereum
func (setup *TestSetup) IsTransferDoneFromMultiversX(sender KeysHolder, receiver KeysHolder, tokens ...TestTokenParams) bool {
	isDone := true
	for _, params := range tokens {
		isDone = isDone && setup.isTransferDoneFromMultiversXForToken(sender, receiver, params)
	}

	return isDone
}

func (setup *TestSetup) isTransferDoneFromMultiversXForToken(sender KeysHolder, receiver KeysHolder, params TestTokenParams) bool {
	setup.mutBalances.Lock()
	initialBalanceForSafe := setup.esdtBalanceForSafe[params.AbstractTokenIdentifier]
	setup.mutBalances.Unlock()

	okSender := setup.checkHolderMvxBalanceForToken(sender, true, params)
	okReceiver := setup.checkHolderEthBalanceForToken(receiver, false, params)

	expectedEsdtSafe := big.NewInt(0).Add(initialBalanceForSafe, params.ESDTSafeExtraBalance)
	balanceForSafe := setup.MultiversxHandler.GetESDTChainSpecificTokenBalance(setup.Ctx, setup.MultiversxHandler.SafeAddress, params.AbstractTokenIdentifier)
	isSafeContractOnCorrectBalance := expectedEsdtSafe.String() == balanceForSafe.String()

	return okSender && okReceiver && isSafeContractOnCorrectBalance
}

func (setup *TestSetup) checkHolderEthBalanceForToken(holder KeysHolder, isSender bool, params TestTokenParams) bool {
	setup.mutBalances.Lock()

	addr := holder.EthAddress.String()
	balanceMapping, exists := setup.ethBalances[addr]
	if !exists {
		return false
	}
	expectedBalance := big.NewInt(0).Set(balanceMapping[params.AbstractTokenIdentifier])
	holderName := addressToName[addr]
	transferAmounts := params.EthTestAddrsExtraBalances[holderName]

	setup.mutBalances.Unlock()

	if isSender {
		expectedBalance.Add(expectedBalance, transferAmounts[0])
	} else {
		expectedBalance.Add(expectedBalance, transferAmounts[1])
	}

	actualBalance := setup.EthereumHandler.GetBalance(holder.EthAddress, params.AbstractTokenIdentifier)

	fmt.Println("-------------ETHEREUM------------------")
	fmt.Println("expectedBalance: ", expectedBalance)
	fmt.Println("actualBalance: ", actualBalance)
	fmt.Println("-------------------------------")
	return actualBalance.String() == expectedBalance.String()
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

	//setup.transferTokensToTestKey(params)
	valueToMintOnEthereum := setup.createdDepositOnMultiversxForToken(setup.AliceKeys, setup.BobKeys, params)
	setup.EthereumHandler.Mint(setup.Ctx, params, valueToMintOnEthereum)
}

func (setup *TestSetup) transferTokensToTestKey(params TestTokenParams) {
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
		_ = setup.createdDepositOnMultiversxForToken(from, to, params)
	}
}

func (setup *TestSetup) createdDepositOnMultiversxForToken(from KeysHolder, to KeysHolder, params TestTokenParams) *big.Int {
	token := setup.GetTokenData(params.AbstractTokenIdentifier)
	require.NotNil(setup, token)

	depositValue := big.NewInt(0)
	for _, operation := range params.TestOperations {
		if operation.ValueToSendFromMvX == nil {
			continue
		}

		depositValue.Add(depositValue, operation.ValueToSendFromMvX)
		setup.MultiversxHandler.SendDepositTransactionFromMultiversx(setup.Ctx, from, to, token, operation.ValueToSendFromMvX)
	}

	return depositValue
}

// TestWithdrawTotalFeesOnEthereumForTokens will test the withdrawal functionality for the provided test tokens
func (setup *TestSetup) TestWithdrawTotalFeesOnEthereumForTokens(tokensParams ...TestTokenParams) {
	for _, param := range tokensParams {
		token := setup.TokensRegistry.GetTokenData(param.AbstractTokenIdentifier)

		expectedAccumulated := big.NewInt(0)
		for _, operation := range param.TestOperations {
			if operation.ValueToSendFromMvX == nil {
				continue
			}
			if operation.ValueToSendFromMvX.Cmp(zeroValueBigInt) == 0 {
				continue
			}

			expectedAccumulated.Add(expectedAccumulated, feeInt)
		}

		setup.MultiversxHandler.TestWithdrawFees(setup.Ctx, token.MvxChainSpecificToken, zeroValueBigInt, expectedAccumulated)
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

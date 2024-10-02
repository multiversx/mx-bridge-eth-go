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

	ctxCancel             func()
	Ctx                   context.Context
	mutBalances           sync.RWMutex
	esdtBalanceForSafe    map[string]*big.Int
	ethBalanceTestAddress map[string]*big.Int
	numScCallsInTest      uint32
}

// NewTestSetup creates a new e2e test setup
func NewTestSetup(tb testing.TB) *TestSetup {
	log.Info(fmt.Sprintf(LogStepMarker, "starting setup"))

	setup := &TestSetup{
		TB:                    tb,
		TokensRegistry:        NewTokenRegistry(tb),
		WorkingDir:            tb.TempDir(),
		esdtBalanceForSafe:    make(map[string]*big.Int),
		ethBalanceTestAddress: make(map[string]*big.Int),
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
		ScProxyBech32Address:         setup.MultiversxHandler.ScProxyAddress.Bech32(),
		ExtraGasToExecute:            60_000_000, // 60 million: this ensures that a SC call with 0 gas limit is refunded
		NetworkAddress:               setup.ChainSimulator.GetNetworkAddress(),
		ProxyMaxNoncesDelta:          5,
		ProxyFinalityCheck:           false,
		ProxyCacherExpirationSeconds: 60, // 1 minute
		ProxyRestAPIEntityType:       string(sdkCore.Proxy),
		IntervalToResendTxsInSeconds: 1,
		PrivateKeyFile:               path.Join(setup.WorkingDir, SCCallerFilename),
		PollingIntervalInMillis:      1000, // 1 second
		FilterConfig: config.PendingOperationsFilterConfig{
			AllowedEthAddresses: []string{"*"},
			AllowedMvxAddresses: []string{"*"},
			AllowedTokens:       []string{"*"},
		},
	}

	var err error
	setup.ScCallerModuleInstance, err = module.NewScCallsModule(cfg, log)
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

		esdtBalanceForSafe := setup.MultiversxHandler.GetESDTChainSpecificTokenBalance(setup.Ctx, setup.MultiversxHandler.SafeAddress, token.AbstractTokenIdentifier)
		ethBalanceForTestAddr := setup.EthereumHandler.GetBalance(setup.TestKeys.EthAddress, token.AbstractTokenIdentifier)

		setup.mutBalances.Lock()
		setup.esdtBalanceForSafe[token.AbstractTokenIdentifier] = esdtBalanceForSafe
		setup.ethBalanceTestAddress[token.AbstractTokenIdentifier] = ethBalanceForTestAddr
		setup.mutBalances.Unlock()

		log.Info("recorded the ESDT balance for safe contract", "token", token.AbstractTokenIdentifier, "balance", esdtBalanceForSafe.String())
		log.Info("recorded the ETH balance for test address", "token", token.AbstractTokenIdentifier, "balance", ethBalanceForTestAddr.String())
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
func (setup *TestSetup) IsTransferDoneFromEthereum(tokens ...TestTokenParams) bool {
	isDone := true
	for _, params := range tokens {
		isDone = isDone && setup.isTransferDoneFromEthereumForToken(params)
	}

	return isDone
}

func (setup *TestSetup) isTransferDoneFromEthereumForToken(params TestTokenParams) bool {
	expectedValueOnReceiver := big.NewInt(0)
	expectedValueOnContract := big.NewInt(0)
	for _, operation := range params.TestOperations {
		if operation.ValueToTransferToMvx == nil {
			continue
		}

		if len(operation.MvxSCCallData) > 0 || operation.MvxForceSCCall {
			if !operation.MvxFaultySCCall {
				expectedValueOnContract.Add(expectedValueOnContract, operation.ValueToTransferToMvx)
			}
		} else {
			expectedValueOnReceiver.Add(expectedValueOnReceiver, operation.ValueToTransferToMvx)
		}
	}

	receiverBalance := setup.MultiversxHandler.GetESDTUniversalTokenBalance(setup.Ctx, setup.TestKeys.MvxAddress, params.AbstractTokenIdentifier)
	if receiverBalance.String() != expectedValueOnReceiver.String() {
		return false
	}

	contractBalance := setup.MultiversxHandler.GetESDTUniversalTokenBalance(setup.Ctx, setup.MultiversxHandler.TestCallerAddress, params.AbstractTokenIdentifier)
	return contractBalance.String() == expectedValueOnContract.String()
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

	receiverBalance := setup.EthereumHandler.GetBalance(setup.TestKeys.EthAddress, params.AbstractTokenIdentifier)
	return receiverBalance.String() == expectedValueOnReceiver.String()
}

// IsTransferDoneFromMultiversX returns true if all provided tokens are bridged from MultiversX towards Ethereum
func (setup *TestSetup) IsTransferDoneFromMultiversX(tokens ...TestTokenParams) bool {
	isDone := true
	for _, params := range tokens {
		isDone = isDone && setup.isTransferDoneFromMultiversXForToken(params)
	}

	return isDone
}

func (setup *TestSetup) isTransferDoneFromMultiversXForToken(params TestTokenParams) bool {
	setup.mutBalances.Lock()
	initialBalanceForSafe := setup.esdtBalanceForSafe[params.AbstractTokenIdentifier]
	expectedReceiver := big.NewInt(0).Set(setup.ethBalanceTestAddress[params.AbstractTokenIdentifier])
	expectedReceiver.Add(expectedReceiver, params.EthTestAddrExtraBalance)
	setup.mutBalances.Unlock()

	ethTestBalance := setup.EthereumHandler.GetBalance(setup.TestKeys.EthAddress, params.AbstractTokenIdentifier)
	isTransferDoneFromMultiversX := ethTestBalance.String() == expectedReceiver.String()

	expectedEsdtSafe := big.NewInt(0).Add(initialBalanceForSafe, params.ESDTSafeExtraBalance)
	balanceForSafe := setup.MultiversxHandler.GetESDTChainSpecificTokenBalance(setup.Ctx, setup.MultiversxHandler.SafeAddress, params.AbstractTokenIdentifier)
	isSafeContractOnCorrectBalance := expectedEsdtSafe.String() == balanceForSafe.String()

	return isTransferDoneFromMultiversX && isSafeContractOnCorrectBalance
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

	valueToMintOnEthereum := setup.MultiversxHandler.CreateDepositsOnMultiversxForToken(setup.Ctx, params)

	setup.EthereumHandler.Mint(setup.Ctx, params, valueToMintOnEthereum)
}

// SendFromMultiversxToEthereum will create the deposits that will be gathered in a batch on MultiversX (without mint on Ethereum)
func (setup *TestSetup) SendFromMultiversxToEthereum(tokensParams ...TestTokenParams) {
	for _, params := range tokensParams {
		setup.sendFromMultiversxToEthereumForToken(params)
	}
}

func (setup *TestSetup) sendFromMultiversxToEthereumForToken(params TestTokenParams) {
	token := setup.GetTokenData(params.AbstractTokenIdentifier)
	require.NotNil(setup, token)

	for _, operation := range params.TestOperations {
		if operation.ValueToSendFromMvX == nil {
			continue
		}

		setup.MultiversxHandler.SendDepositTransactionFromMultiversx(setup.Ctx, token, operation.ValueToSendFromMvX)
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

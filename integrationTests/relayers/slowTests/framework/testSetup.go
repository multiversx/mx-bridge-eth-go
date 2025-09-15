package framework

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
	"path"
	"strings"
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
	PeerChainHandler       PeerChainHandler
	MultiversxHandler      *MultiversxHandler
	WorkingDir             string
	ChainSimulator         ChainSimulatorWrapper
	ScCallerKeys           KeysHolder
	ScCallerModuleInstance SCCallerModule
	peerChainType          ChainType

	ctxCancel                   func()
	Ctx                         context.Context
	mutBalances                 sync.RWMutex
	esdtBalanceForSafe          map[string]*big.Int
	peerChainBalanceTestAddress map[string]*big.Int
	numScCallsInTest            uint32
}

// NewTestSetup creates a new e2e test setup
func NewTestSetup(tb testing.TB, chainType ChainType) *TestSetup {
	log.Info(fmt.Sprintf(LogStepMarker, "starting setup"))

	setup := &TestSetup{
		TB:                          tb,
		TokensRegistry:              NewTokenRegistry(tb),
		WorkingDir:                  tb.TempDir(),
		peerChainType:               chainType,
		esdtBalanceForSafe:          make(map[string]*big.Int),
		peerChainBalanceTestAddress: make(map[string]*big.Int),
	}
	setup.KeysStore = NewKeysStore(tb, setup.WorkingDir, NumRelayers, NumOracles)

	// create a test context
	setup.Ctx, setup.ctxCancel = context.WithCancel(context.Background())

	switch chainType {
	case ChainTypeEthereum:
		setup.PeerChainHandler = NewEthereumHandler(tb, setup.Ctx, setup.KeysStore, setup.TokensRegistry, quorum)
	case ChainTypeSui:
		argsSuiChainSimulatorWrapper := ArgsSuiChainSimulatorWrapper{
			TB:    tb,
			Owner: setup.OwnerKeys,
		}
		suiChainSimulator := CreateSuiChainSimulatorWrapper(argsSuiChainSimulatorWrapper)
		setup.PeerChainHandler = NewSuiHandler(tb, setup.KeysStore, setup.TokensRegistry, suiChainSimulator, quorum)
	}

	setup.PeerChainHandler.DeployContracts(setup.Ctx)

	setup.createChainSimulatorWrapper()
	setup.MultiversxHandler = NewMultiversxHandler(tb, setup.Ctx, setup.KeysStore, setup.TokensRegistry, setup.ChainSimulator, quorum)
	setup.MultiversxHandler.DeployAndSetContracts(setup.Ctx, chainType)

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
	switch handler := setup.PeerChainHandler.(type) {
	case *EthereumHandler:
		setup.Bridge = NewEthereumBridgeComponents(
			setup.TB,
			setup.WorkingDir,
			setup.ChainSimulator,
			handler.EthChainWrapper,
			handler.Erc20ContractsHolder,
			handler.SimulatedChain,
			NumRelayers,
			handler.SafeAddress.Hex(),
			setup.MultiversxHandler.SafeAddress,
			setup.MultiversxHandler.MultisigAddress,
		)
	case *SuiHandler:
		setup.Bridge = NewSuiBridgeComponents(
			setup.TB,
			setup.WorkingDir,
			setup.ChainSimulator,
			handler.SuiChainSimulator,
			NumRelayers,
			handler.PackageID,
			setup.MultiversxHandler.SafeAddress,
			setup.MultiversxHandler.MultisigAddress,
			handler.BridgeObjectID,
			handler.SafeObjectID,
			handler.BridgeInitialSharedVersion,
			handler.SafeInitialSharedVersion,
		)
	default:
		panic(fmt.Sprintf("unsupported peer chain handler type: %T", handler))
	}

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

	setup.PeerChainHandler.PauseContractsForTokenChanges(setup.Ctx)
	setup.MultiversxHandler.PauseContractsForTokenChanges(setup.Ctx)

	for _, token := range tokens {
		setup.processNumScCallsOperations(token)
		setup.AddToken(token.IssueTokenParams)
		setup.PeerChainHandler.IssueAndWhitelistToken(setup.Ctx, token.IssueTokenParams)
		setup.MultiversxHandler.IssueAndWhitelistToken(setup.Ctx, token.IssueTokenParams)

		esdtBalanceForSafe := setup.MultiversxHandler.GetESDTChainSpecificTokenBalance(setup.Ctx, setup.MultiversxHandler.SafeAddress, token.AbstractTokenIdentifier)
		var peerChainBalanceForTestAddr *big.Int
		switch setup.peerChainType {
		case ChainTypeEthereum:
			peerChainBalanceForTestAddr = setup.PeerChainHandler.GetBalance(setup.Ctx, setup.TestKeys.EthAddress.Bytes(), token.AbstractTokenIdentifier)
		case ChainTypeSui:
			peerChainBalanceForTestAddr = setup.PeerChainHandler.GetBalance(setup.Ctx, setup.TestKeys.SuiAddress, token.AbstractTokenIdentifier)
		}

		setup.mutBalances.Lock()
		setup.esdtBalanceForSafe[token.AbstractTokenIdentifier] = esdtBalanceForSafe
		setup.peerChainBalanceTestAddress[token.AbstractTokenIdentifier] = peerChainBalanceForTestAddr
		setup.mutBalances.Unlock()

		log.Info("recorded the ESDT balance for safe contract", "token", token.AbstractTokenIdentifier, "balance", esdtBalanceForSafe.String())
		log.Info("recorded the peer chain balance for test address", "token", token.AbstractTokenIdentifier, "balance", peerChainBalanceForTestAddr.String())
	}

	setup.PeerChainHandler.UnPauseContractsAfterTokenChanges(setup.Ctx)
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

// IsTransferDoneFromPeerChain returns true if all provided tokens are bridged from peer chain towards MultiversX
func (setup *TestSetup) IsTransferDoneFromPeerChain(tokens ...TestTokenParams) bool {
	isDone := true
	for _, params := range tokens {
		isDone = isDone && setup.isTransferDoneFromPeerChainForToken(params)
	}

	return isDone
}

func (setup *TestSetup) isTransferDoneFromPeerChainForToken(params TestTokenParams) bool {
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

// IsTransferDoneFromPeerChainWithRefund returns true if all provided tokens are bridged from peer chain towards MultiversX including refunds
func (setup *TestSetup) IsTransferDoneFromPeerChainWithRefund(tokens ...TestTokenParams) bool {
	isDone := true
	for _, params := range tokens {
		isDone = isDone && setup.isTransferDoneFromPeerChainWithRefundForToken(params)
	}

	return isDone
}

func (setup *TestSetup) isTransferDoneFromPeerChainWithRefundForToken(params TestTokenParams) bool {
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
				// the balance should be bridged back to the receiver on peer chain - fee
				expectedValueOnReceiver.Add(expectedValueOnReceiver, valueToTransferToMvx)
				expectedValueOnReceiver.Sub(expectedValueOnReceiver, feeInt)
			}
		}
	}

	var receiverBalance *big.Int
	switch setup.peerChainType {
	case ChainTypeEthereum:
		receiverBalance = setup.PeerChainHandler.GetBalance(setup.Ctx, setup.TestKeys.EthAddress.Bytes(), params.AbstractTokenIdentifier)
	case ChainTypeSui:
		receiverBalance = setup.PeerChainHandler.GetBalance(setup.Ctx, setup.TestKeys.SuiAddress, params.AbstractTokenIdentifier)
	}

	return receiverBalance.String() == expectedValueOnReceiver.String()
}

// IsTransferDoneFromMultiversX returns true if all provided tokens are bridged from MultiversX towards peer chain
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
	expectedReceiver := big.NewInt(0).Set(setup.peerChainBalanceTestAddress[params.AbstractTokenIdentifier])
	expectedReceiver.Add(expectedReceiver, params.PeerChainTestAddrExtraBalance)
	setup.mutBalances.Unlock()

	var peerChainTestBalance *big.Int
	switch setup.peerChainType {
	case ChainTypeEthereum:
		peerChainTestBalance = setup.PeerChainHandler.GetBalance(setup.Ctx, setup.TestKeys.EthAddress.Bytes(), params.AbstractTokenIdentifier)
	case ChainTypeSui:
		peerChainTestBalance = setup.PeerChainHandler.GetBalance(setup.Ctx, setup.TestKeys.SuiAddress, params.AbstractTokenIdentifier)
	}
	isTransferDoneFromMultiversX := peerChainTestBalance.String() == expectedReceiver.String()

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

	setup.transferTokensToTestKey(params)
	valueToMintOnPeerChain := setup.sendFromMultiversxToPeerChainForToken(params)
	setup.PeerChainHandler.Mint(setup.Ctx, params, valueToMintOnPeerChain)
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
		setup.TestKeys,
		depositValue,
		params,
	)
}

// SendFromMultiversxToPeerChain will create the deposits that will be gathered in a batch on MultiversX (without mint on peer chain)
func (setup *TestSetup) SendFromMultiversxToPeerChain(tokensParams ...TestTokenParams) {
	for _, params := range tokensParams {
		if params.IsLocked {
			setup.transferTokensToTestKey(params)
		}
		_ = setup.sendFromMultiversxToPeerChainForToken(params)
	}
}

func (setup *TestSetup) sendFromMultiversxToPeerChainForToken(params TestTokenParams) *big.Int {
	token := setup.GetTokenData(params.AbstractTokenIdentifier)
	require.NotNil(setup, token)

	depositValue := big.NewInt(0)
	for _, operation := range params.TestOperations {
		if operation.ValueToSendFromMvX == nil {
			continue
		}

		depositValue.Add(depositValue, operation.ValueToSendFromMvX)

		var receiverAddress []byte
		switch setup.peerChainType {
		case ChainTypeEthereum:
			receiverAddress = setup.TestKeys.EthAddress.Bytes()
		case ChainTypeSui:
			receiverAddress, _ = hex.DecodeString(strings.TrimPrefix(string(setup.TestKeys.SuiAddress), "0x"))
		}
		setup.MultiversxHandler.SendDepositTransactionFromMultiversx(setup.Ctx, token, params, operation.ValueToSendFromMvX, receiverAddress)
	}

	return depositValue
}

// TestWithdrawTotalFeesOnPeerChainForTokens will test the withdrawal functionality for the provided test tokens
func (setup *TestSetup) TestWithdrawTotalFeesOnPeerChainForTokens(tokensParams ...TestTokenParams) {
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
	require.NoError(setup, setup.PeerChainHandler.Close())

	setup.ctxCancel()
	_ = setup.ScCallerModuleInstance.Close()
}

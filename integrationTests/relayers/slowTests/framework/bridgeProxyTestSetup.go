package framework

import (
	"context"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	scAddress = "erd1qqqqqqqqqqqqqpgqvc7gdl0p4s97guh498wgz75k8sav6sjfjlwqh679jy"
)

// BridgeProxyTestSetup is the struct that holds all subcomponents for the testing infrastructure
type BridgeProxyTestSetup struct {
	testing.TB
	TokensRegistry
	*KeysStore
	MultiversxHandler *MultiversxHandler
	ChainSimulator    ChainSimulatorWrapper
	workingDir        string

	ctxCancel func()
	Ctx       context.Context
}

// NewSetup creates a new test setup for bridge proxy
func NewSetup(tb testing.TB) *BridgeProxyTestSetup {
	setup := &BridgeProxyTestSetup{
		TB:             tb,
		TokensRegistry: NewTokenRegistry(tb),
		workingDir:     tb.TempDir(),
	}
	setup.KeysStore = NewKeysStore(tb, setup.workingDir, NumRelayers, NumOracles)

	setup.Ctx, setup.ctxCancel = context.WithCancel(context.Background())

	setup.createChainSimulatorWrapper()

	setup.MultiversxHandler = NewMultiversxHandler(setup.TB, setup.Ctx, setup.KeysStore, setup.TokensRegistry, setup.ChainSimulator, quorum)

	setup.deployAndSetContracts()

	return setup
}

func (setup *BridgeProxyTestSetup) createChainSimulatorWrapper() {
	// create a new working directory
	tmpDir := path.Join(setup.TempDir(), "test")
	err := os.MkdirAll(tmpDir, os.ModePerm)
	require.NoError(setup, err)

	// start the chain simulator
	args := ArgChainSimulatorWrapper{
		TB:                           setup.TB,
		ProxyCacherExpirationSeconds: 600,
		ProxyMaxNoncesDelta:          7,
	}
	setup.ChainSimulator = CreateChainSimulatorWrapper(args)
	require.NoError(setup, err)
}

func (setup *BridgeProxyTestSetup) deployAndSetContracts() {
	setup.deployContracts()
	setup.MultiversxHandler.ChangeOwnerForBridgeProxy(setup.Ctx)
	setup.MultiversxHandler.UnpauseBridgeProxy(setup.Ctx)
	setup.MultiversxHandler.SetBridgeProxyAddressOnHelper(setup.Ctx)
}

func (setup *BridgeProxyTestSetup) deployContracts() {
	setup.MultiversxHandler.DeployBridgeProxy(setup.Ctx)

	setup.MultiversxHandler.DeployTestHelperContract(setup.Ctx)

	setup.MultiversxHandler.MultiTransferAddress = setup.MultiversxHandler.TestHelperAddress
	setup.MultiversxHandler.SafeAddress = NewMvxAddressFromBech32(setup.TB, scAddress)
	setup.MultiversxHandler.WrapperAddress = NewMvxAddressFromBech32(setup.TB, scAddress)
	setup.MultiversxHandler.AggregatorAddress = NewMvxAddressFromBech32(setup.TB, scAddress)
	setup.MultiversxHandler.DeployMultisig(setup.Ctx)

	setup.MultiversxHandler.DeployTestCaller(setup.Ctx)
}

// IssueToken adds a token to the registry and issues it
func (setup *BridgeProxyTestSetup) IssueToken(token TestTokenParams) {
	setup.AddToken(token.IssueTokenParams)
	setup.MultiversxHandler.IssueAndWhitelistToken(setup.Ctx, token.IssueTokenParams)
}

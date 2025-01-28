//go:build slow

package slowTests

import (
	"bytes"
	"context"
	"encoding/binary"
	"math/big"
	"os"
	"path"
	"testing"

	bridgeCore "github.com/multiversx/mx-bridge-eth-go/core"
	"github.com/multiversx/mx-bridge-eth-go/integrationTests/relayers/slowTests/framework"
	"github.com/stretchr/testify/require"
)

const (
	numRelayers = 3
	numOracles  = 3
	quorum      = "03"
	scAddress   = "erd1qqqqqqqqqqqqqpgqvc7gdl0p4s97guh498wgz75k8sav6sjfjlwqh679jy"
)

type BridgeProxyTestSetup struct {
	testing.TB
	framework.TokensRegistry
	*framework.KeysStore
	multiversxHandler *framework.MultiversxHandler
	workingDir        string
	chainSimulator    framework.ChainSimulatorWrapper

	ctxCancel func()
	Ctx       context.Context
}

func NewSetup(tb testing.TB) *BridgeProxyTestSetup {
	setup := &BridgeProxyTestSetup{
		TB:             tb,
		TokensRegistry: framework.NewTokenRegistry(tb),
		workingDir:     tb.TempDir(),
	}
	setup.KeysStore = framework.NewKeysStore(tb, setup.workingDir, numRelayers, numOracles)

	setup.Ctx, setup.ctxCancel = context.WithCancel(context.Background())

	setup.createChainSimulatorWrapper()

	setup.multiversxHandler = framework.NewMultiversxHandler(setup.TB, setup.Ctx, setup.KeysStore, setup.TokensRegistry, setup.chainSimulator, quorum)

	setup.deployAndSetContracts()

	return setup
}

func (setup *BridgeProxyTestSetup) createChainSimulatorWrapper() {
	// create a new working directory
	tmpDir := path.Join(setup.TempDir(), "test")
	err := os.MkdirAll(tmpDir, os.ModePerm)
	require.NoError(setup, err)

	// start the chain simulator
	args := framework.ArgChainSimulatorWrapper{
		TB:                           setup.TB,
		ProxyCacherExpirationSeconds: 600,
		ProxyMaxNoncesDelta:          7,
	}
	setup.chainSimulator = framework.CreateChainSimulatorWrapper(args)
	require.NoError(setup, err)
}

func (setup *BridgeProxyTestSetup) deployAndSetContracts() {
	setup.deployContracts()
	setup.multiversxHandler.ChangeOwnerForBridgeProxy(setup.Ctx)
	setup.multiversxHandler.UnpauseBridgeProxy(setup.Ctx)
	setup.multiversxHandler.SetBridgeProxyAddressOnHelper(setup.Ctx)
}

func (setup *BridgeProxyTestSetup) deployContracts() {
	setup.multiversxHandler.DeployBridgeProxy(setup.Ctx)

	setup.multiversxHandler.DeployTestHelperContract(setup.Ctx)

	setup.multiversxHandler.MultiTransferAddress = setup.multiversxHandler.TestHelperAddress
	setup.multiversxHandler.SafeAddress = framework.NewMvxAddressFromBech32(setup.TB, scAddress)
	setup.multiversxHandler.WrapperAddress = framework.NewMvxAddressFromBech32(setup.TB, scAddress)
	setup.multiversxHandler.AggregatorAddress = framework.NewMvxAddressFromBech32(setup.TB, scAddress)
	setup.multiversxHandler.DeployMultisig(setup.Ctx)

	setup.multiversxHandler.DeployTestCaller(setup.Ctx)
}

func (setup *BridgeProxyTestSetup) generateAndIssueToken() framework.TestTokenParams {
	token := GenerateUnlistedTokenFromMvx()
	setup.AddToken(token.IssueTokenParams)
	setup.multiversxHandler.IssueAndWhitelistToken(setup.Ctx, token.IssueTokenParams)

	return token
}

func TestBridgeProxy(t *testing.T) {
	setup := NewSetup(t)

	token := setup.generateAndIssueToken()

	ethTx1 := framework.EthTransaction{
		From:     bytes.Repeat([]byte{0x01}, 20),
		To:       setup.multiversxHandler.CalleeScAddress.AddressBytes(),
		TokenID:  token.MvxUniversalTokenTicker,
		Amount:   big.NewInt(100),
		Nonce:    1,
		CallData: prependLenAndDataMarker(createScCallData("callPayable", 500000000)),
	}

	ethTx2 := framework.EthTransaction{
		From:     bytes.Repeat([]byte{0x02}, 20),
		To:       setup.multiversxHandler.CalleeScAddress.AddressBytes(),
		TokenID:  token.MvxUniversalTokenTicker,
		Amount:   big.NewInt(2000),
		Nonce:    2,
		CallData: prependLenAndDataMarker(createScCallData("callPayable", 500000000)),
	}

	// deposit txs in bridge proxy
	tokenData := setup.TokensRegistry.GetTokenData(token.AbstractTokenIdentifier)
	setup.multiversxHandler.CallDepositOnBridgeProxy(setup.Ctx, ethTx1, 0, tokenData)
	setup.multiversxHandler.CallDepositOnBridgeProxy(setup.Ctx, ethTx2, 0, tokenData)

	// make 2 execute calls for same deposit
	setup.multiversxHandler.ExecuteDepositWithoutGenerateBlocks(setup.Ctx, 1, 0)
	setup.multiversxHandler.ExecuteDepositWithoutGenerateBlocks(setup.Ctx, 1, 1)

	setup.chainSimulator.GenerateBlocks(setup.Ctx, 10)
}

func prependLenAndDataMarker(input []byte) []byte {
	buff32 := make([]byte, bridgeCore.Uint32ArgBytes)
	binary.BigEndian.PutUint32(buff32, uint32(len(input)))

	prefix := append([]byte{bridgeCore.DataPresentProtocolMarker}, buff32...)

	return append(prefix, input...)
}

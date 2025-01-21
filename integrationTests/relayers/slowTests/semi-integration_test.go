//go:build slow

package slowTests

import (
	"bytes"
	"context"
	"github.com/multiversx/mx-bridge-eth-go/integrationTests/relayers/slowTests/framework"
	"github.com/stretchr/testify/require"
	"math/big"
	"os"
	"path"
	"testing"
)

type BridgeProxyTestSetup struct {
	testing.TB
	*framework.KeysStore
	MultiversxHandler *framework.MultiversxHandler
	WorkingDir        string
	ChainSimulator    framework.ChainSimulatorWrapper

	ctxCancel func()
	Ctx       context.Context
}

func NewSetup(tb testing.TB) *BridgeProxyTestSetup {
	setup := &BridgeProxyTestSetup{
		TB:         tb,
		WorkingDir: tb.TempDir(),
	}
	setup.KeysStore = framework.NewKeysStore(tb, setup.WorkingDir, 3, 3)

	setup.Ctx, setup.ctxCancel = context.WithCancel(context.Background())

	setup.createChainSimulatorWrapper()

	setup.MultiversxHandler = framework.NewMultiversxHandler(setup.TB, setup.Ctx, setup.KeysStore, framework.NewTokenRegistry(tb), setup.ChainSimulator, "03")

	setup.deployContracts()

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
	setup.ChainSimulator = framework.CreateChainSimulatorWrapper(args)
	require.NoError(setup, err)
}

func (setup *BridgeProxyTestSetup) deployContracts() {
	setup.MultiversxHandler.DeployBridgeProxy(setup.Ctx)

	setup.MultiversxHandler.DeployTestHelperContract(setup.Ctx)

	// change MultiTransfer address in MultiSig to be the helper contract
	setup.MultiversxHandler.MultiTransferAddress = setup.MultiversxHandler.TestHelperAddress
	setup.MultiversxHandler.SafeAddress = framework.NewMvxAddressFromBech32(setup.TB, "erd1qqqqqqqqqqqqqpgqvc7gdl0p4s97guh498wgz75k8sav6sjfjlwqh679jy")
	setup.MultiversxHandler.ScProxyAddress = framework.NewMvxAddressFromBech32(setup.TB, "erd1qqqqqqqqqqqqqpgqvc7gdl0p4s97guh498wgz75k8sav6sjfjlwqh679jy")
	setup.MultiversxHandler.WrapperAddress = framework.NewMvxAddressFromBech32(setup.TB, "erd1qqqqqqqqqqqqqpgqvc7gdl0p4s97guh498wgz75k8sav6sjfjlwqh679jy")
	setup.MultiversxHandler.AggregatorAddress = framework.NewMvxAddressFromBech32(setup.TB, "erd1qqqqqqqqqqqqqpgqvc7gdl0p4s97guh498wgz75k8sav6sjfjlwqh679jy")

	setup.MultiversxHandler.DeployMultisig(setup.Ctx)
}

func TestBridgeProxy(t *testing.T) {
	setup := NewSetup(t)

	// send deposit transactions to bridge proxy
	ethTx1 := framework.EthTransaction{
		From:     bytes.Repeat([]byte{0x01}, 20),
		To:       bytes.Repeat([]byte{0x02}, 32),
		TokenID:  "EUROC-123456",
		Amount:   big.NewInt(100),
		Nonce:    1,
		CallData: nil,
	}

	// Define the second EthTransaction
	ethTx2 := framework.EthTransaction{
		From:     bytes.Repeat([]byte{0x03}, 20),
		To:       bytes.Repeat([]byte{0x04}, 32),
		TokenID:  "USDC-654321",
		Amount:   big.NewInt(2000),
		Nonce:    2,
		CallData: nil,
	}

	// set bridge-proxy address in helper
	setup.MultiversxHandler.SetBridgeProxyAddressOnHelper(setup.Ctx)

	// create Deposit Tx in handler
	setup.MultiversxHandler.CallDepositOnBridgeProxy(setup.Ctx, ethTx1, 0)
	setup.MultiversxHandler.CallDepositOnBridgeProxy(setup.Ctx, ethTx2, 0)
}

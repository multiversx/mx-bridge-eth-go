package slowTests

import (
	"context"
	"github.com/multiversx/mx-bridge-eth-go/integrationTests/relayers/slowTests/framework"
	"github.com/stretchr/testify/require"
	"os"
	"path"
	"testing"
)

type testSetup struct {
	testing.TB
	*framework.KeysStore
	MultiversxHandler *framework.MultiversxHandler
	WorkingDir        string
	ChainSimulator    framework.ChainSimulatorWrapper

	ctxCancel func()
	Ctx       context.Context
}

func NewSetup(tb testing.TB) *testSetup {
	setup := &testSetup{
		TB:         tb,
		WorkingDir: tb.TempDir(),
	}
	setup.KeysStore = framework.NewKeysStore(tb, setup.WorkingDir, 1, 1)

	setup.Ctx, setup.ctxCancel = context.WithCancel(context.Background())

	setup.createChainSimulatorWrapper()

	//setup.MultiversxHandler = framework.NewMultiversxHandler(setup.TB, setup.Ctx, setup.KeysStore,,setup.ChainSimulator, "03")

	//framework.CreateChainSimulatorWrapper()
}

func (setup *testSetup) createChainSimulatorWrapper() {
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

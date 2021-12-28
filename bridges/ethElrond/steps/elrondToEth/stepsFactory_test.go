package elrondToEth

import (
	"testing"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridges/ethElrond"
	bridgeTests "github.com/ElrondNetwork/elrond-eth-bridge/testsCommon/bridge"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateSteps_Errors(t *testing.T) {
	t.Parallel()

	steps, err := CreateSteps(nil)

	assert.Nil(t, steps)
	assert.Equal(t, ethElrond.ErrNilExecutor, err)
}

func TestCreateSteps_ShouldWork(t *testing.T) {
	t.Parallel()

	steps, err := CreateSteps(bridgeTests.NewBridgeExecutorStub())

	require.NotNil(t, steps)
	require.Nil(t, err)
	require.Equal(t, NumSteps, len(steps))
}

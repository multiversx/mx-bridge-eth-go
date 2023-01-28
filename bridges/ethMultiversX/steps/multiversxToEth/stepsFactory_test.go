package multiversxtoeth

import (
	"testing"

	"github.com/multiversx/mx-bridge-eth-go/bridges/ethMultiversX"
	bridgeTests "github.com/multiversx/mx-bridge-eth-go/testsCommon/bridge"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateSteps_Errors(t *testing.T) {
	t.Parallel()

	steps, err := CreateSteps(nil)

	assert.Nil(t, steps)
	assert.Equal(t, ethmultiversx.ErrNilExecutor, err)
}

func TestCreateSteps_ShouldWork(t *testing.T) {
	t.Parallel()

	steps, err := CreateSteps(bridgeTests.NewBridgeExecutorStub())

	require.NotNil(t, steps)
	require.Nil(t, err)
	require.Equal(t, NumSteps, len(steps))
}

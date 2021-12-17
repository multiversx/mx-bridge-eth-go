package steps

import (
	"testing"

	v2 "github.com/ElrondNetwork/elrond-eth-bridge/ethToElrond/v2"
	"github.com/ElrondNetwork/elrond-eth-bridge/ethToElrond/v2/ethToElrond"
	"github.com/ElrondNetwork/elrond-eth-bridge/testsCommon/bridgeV2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateSteps_Errors(t *testing.T) {
	t.Parallel()

	steps, err := CreateSteps(nil)

	assert.Nil(t, steps)
	assert.Equal(t, v2.ErrNilExecutor, err)
}

func TestCreateSteps_ShouldWork(t *testing.T) {
	t.Parallel()

	steps, err := CreateSteps(bridgeV2.NewBridgeExecutorStub())

	require.NotNil(t, steps)
	require.Nil(t, err)
	require.Equal(t, ethToElrond.NumSteps, len(steps))
}

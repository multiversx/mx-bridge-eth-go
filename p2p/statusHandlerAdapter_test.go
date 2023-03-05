package p2p

import (
	"context"
	"strings"
	"testing"

	"github.com/multiversx/mx-bridge-eth-go/core"
	"github.com/multiversx/mx-bridge-eth-go/testsCommon"
	p2pMocks "github.com/multiversx/mx-bridge-eth-go/testsCommon/p2p"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/stretchr/testify/assert"
)

func createMockArgsStatusHandlerAdapter() ArgsStatusHandlerAdapter {
	return ArgsStatusHandlerAdapter{
		StatusHandler: testsCommon.NewStatusHandlerMock("test"),
		Messenger:     &p2pMocks.MessengerStub{},
	}
}

func TestNewStatusHandlerAdapter(t *testing.T) {
	t.Parallel()

	t.Run("nil status handler", func(t *testing.T) {
		args := createMockArgsStatusHandlerAdapter()
		args.StatusHandler = nil

		adapter, err := NewStatusHandlerAdapter(args)
		assert.Equal(t, ErrNilStatusHandler, err)
		assert.True(t, check.IfNil(adapter))
	})
	t.Run("nil messenger", func(t *testing.T) {
		args := createMockArgsStatusHandlerAdapter()
		args.Messenger = nil

		adapter, err := NewStatusHandlerAdapter(args)
		assert.Equal(t, ErrNilMessenger, err)
		assert.True(t, check.IfNil(adapter))
	})
	t.Run("should work", func(t *testing.T) {
		args := createMockArgsStatusHandlerAdapter()

		adapter, err := NewStatusHandlerAdapter(args)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(adapter))
	})
}

func TestStatusHandlerAdapter_Execute(t *testing.T) {
	t.Parallel()

	hostAddresses := []string{"address 1", "address 2"}
	connectedAddresses := []string{"connected address 1", "connected address 2", "connected address 3"}

	args := createMockArgsStatusHandlerAdapter()
	args.Messenger = &p2pMocks.MessengerStub{
		AddressesCalled: func() []string {
			return hostAddresses
		},
		ConnectedAddressesCalled: func() []string {
			return connectedAddresses
		},
	}

	adapter, _ := NewStatusHandlerAdapter(args)
	err := adapter.Execute(context.TODO())
	assert.Nil(t, err)

	expectedMetric := make(core.GeneralMetrics)
	expectedMetric[core.MetricConnectedP2PAddresses] = strings.Join(connectedAddresses, " ")
	expectedMetric[core.MetricRelayerP2PAddresses] = strings.Join(hostAddresses, " ")

	metrics := adapter.GetAllMetrics()
	assert.Equal(t, 2, len(metrics))
	assert.Equal(t, expectedMetric, metrics)
}

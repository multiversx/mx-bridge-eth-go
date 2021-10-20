package bridgeExecutors

import (
	"testing"

	"github.com/ElrondNetwork/elrond-eth-bridge/relay/ethToElrond/bridgeExecutors/mock"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/stretchr/testify/assert"
)

func createMockArgs() ArgsEthElrondBridgeExecutor {
	return ArgsEthElrondBridgeExecutor{
		ExecutorName:      "executorMock",
		Logger:            logger.GetOrCreate("test"),
		SourceBridge:      &mock.BridgeStub{},
		DestinationBridge: &mock.BridgeStub{},
		TopologyProvider:  &mock.TopologyProviderStub{},
		QuorumProvider:    &mock.QuorumProviderStub{},
	}
}

func TestNewbridgeExecutors(t *testing.T) {
	t.Parallel()
	t.Run("nil source bridge", func(t *testing.T) {
		args := createMockArgs()
		args.SourceBridge = nil
		executor, err := NewEthElrondBridgeExecutor(args)

		assert.Nil(t, executor)
		assert.Equal(t, ErrNilBridge, err)
	})
	t.Run("nil destination bridge", func(t *testing.T) {
		args := createMockArgs()
		args.DestinationBridge = nil
		executor, err := NewEthElrondBridgeExecutor(args)

		assert.Nil(t, executor)
		assert.Equal(t, ErrNilBridge, err)
	})
	t.Run("nil logger", func(t *testing.T) {
		args := createMockArgs()
		args.Logger = nil
		executor, err := NewEthElrondBridgeExecutor(args)

		assert.Nil(t, executor)
		assert.Equal(t, ErrNilLogger, err)
	})
	t.Run("nil topology provider", func(t *testing.T) {
		args := createMockArgs()
		args.TopologyProvider = nil
		executor, err := NewEthElrondBridgeExecutor(args)

		assert.Nil(t, executor)
		assert.Equal(t, ErrNilTopologyProvider, err)
	})
	t.Run("nil logger", func(t *testing.T) {
		args := createMockArgs()
		args.QuorumProvider = nil
		executor, err := NewEthElrondBridgeExecutor(args)

		assert.Nil(t, executor)
		assert.Equal(t, ErrNilQuorumProvider, err)
	})
}

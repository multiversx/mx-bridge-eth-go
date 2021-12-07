package v2

import (
	"testing"

	"github.com/ElrondNetwork/elrond-eth-bridge/testsCommon/bridgeV2"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/stretchr/testify/assert"
)

func createMockEthToElrondExecutorArgs() ArgsEthToElrondBridgeExecutor {
	return ArgsEthToElrondBridgeExecutor{
		Log:              logger.GetOrCreate("test"),
		TopologyProvider: &bridgeV2.TopologyProviderStub{},
		ElrondClient:     &bridgeV2.ElrondClientStub{},
		EthereumClient:   nil,
	}
}

func TestNewEthToElrondBridgeExecutor(t *testing.T) {
	t.Parallel()

	t.Run("nil logger should error", func(t *testing.T) {
		args := createMockEthToElrondExecutorArgs()
		args.Log = nil
		executor, err := NewEthToElrondBridgeExecutor(args)

		assert.True(t, check.IfNil(executor))
		assert.Equal(t, errNilLogger, err)
	})
	t.Run("nil elrond client should error", func(t *testing.T) {
		args := createMockEthToElrondExecutorArgs()
		args.ElrondClient = nil
		executor, err := NewEthToElrondBridgeExecutor(args)

		assert.True(t, check.IfNil(executor))
		assert.Equal(t, errNilElrondClient, err)
	})
	t.Run("nil ethereum client should error", func(t *testing.T) {
		args := createMockEthToElrondExecutorArgs()
		args.EthereumClient = nil
		executor, err := NewEthToElrondBridgeExecutor(args)

		assert.True(t, check.IfNil(executor))
		assert.Equal(t, errNilEthereumClient, err)
	})
}

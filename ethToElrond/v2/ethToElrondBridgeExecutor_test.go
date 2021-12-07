package v2

import "testing"

func createMockEthToElrondExecutorArgs() ArgsEthToElrondBridgeExecutor {

}

func TestNewEthToElrondBridgeExecutor(t *testing.T) {
	t.Parallel()

	t.Run("nil logger should error", func(t *testing.T) {
		args := createMockEthToElrondExecutorArgs()
	})
}

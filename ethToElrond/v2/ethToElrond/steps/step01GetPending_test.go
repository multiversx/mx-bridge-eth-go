package steps

import (
	"context"
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ElrondNetwork/elrond-eth-bridge/ethToElrond/v2/ethToElrond"
	"github.com/ElrondNetwork/elrond-eth-bridge/testsCommon/bridgeV2"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/stretchr/testify/require"
)

var expectedErr = errors.New("expected error")
var testLogger = logger.GetOrCreate("test")

func TestExecute(t *testing.T) {
	t.Parallel()
	t.Run("GetLastExecutedEthBatchIDFromElrond gives error", func(t *testing.T) {

		stub := createStubExecutor()
		stub.GetLastExecutedEthBatchIDFromElrondCalled = func(ctx context.Context) (uint64, error) {
			return 0, expectedErr
		}
		step := getPendingStep{
			bridge: stub,
		}
		si, _ := step.Execute(context.Background())
		require.Equal(t, core.StepIdentifier(ethToElrond.GettingPendingBatchFromEthereum), si)
	})

}

func createStubExecutor() *bridgeV2.EthToElrondBridgeStub {
	stub := bridgeV2.NewEthToElrondBridgeStub()
	stub.GetLoggerCalled = func() logger.Logger {
		return testLogger
	}
	return stub
}

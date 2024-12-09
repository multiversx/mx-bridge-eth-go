//go:build slow

package slowTests

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-bridge-eth-go/integrationTests/relayers/slowTests/framework"
)

type flowType string

const (
	startFromEthereumFlow   flowType = "start from Ethereum"
	startFromMultiversXFlow flowType = "start from MultiversX"
)

type testFlow struct {
	testing.TB
	flowType
	setup                        *framework.TestSetup
	firstHalfBridgeDone          bool
	secondHalfBridgeDone         bool
	tokens                       []framework.TestTokenParams
	messageAfterFirstHalfBridge  string
	messageAfterSecondHalfBridge string
	handlerAfterFirstHalfBridge  func(flow *testFlow)
	handlerToStartFirstBridge    func(flow *testFlow)
}

func (flow *testFlow) process() (finished bool) {
	if len(flow.tokens) == 0 {
		return true
	}
	if flow.firstHalfBridgeDone && flow.secondHalfBridgeDone {
		return true
	}

	if !flow.firstHalfBridgeDone {
		transferDoneForFirstHalf := flow.setup.AreAllTransfersCompleted(framework.FirstHalfBridge, flow.tokens...)
		if transferDoneForFirstHalf {
			flow.firstHalfBridgeDone = true
			log.Info(fmt.Sprintf(framework.LogStepMarker, flow.messageAfterFirstHalfBridge))

			flow.handlerAfterFirstHalfBridge(flow)
		}

		return false
	}

	if flow.setup.MultiversxHandler.HasRefundBatch(flow.setup.Ctx) {
		flow.setup.MultiversxHandler.MoveRefundBatchToSafe(flow.setup.Ctx)
	}

	//TODO: move this logic into the SC calls executor
	flow.setup.MultiversxHandler.RefundAllFromScBridgeProxy(flow.setup.Ctx)

	transferDoneForSecondHalf := flow.setup.AreAllTransfersCompleted(framework.SecondHalfBridge, flow.tokens...)
	if !flow.secondHalfBridgeDone && transferDoneForSecondHalf {
		flow.secondHalfBridgeDone = true
		log.Info(fmt.Sprintf(framework.LogStepMarker, flow.messageAfterSecondHalfBridge))
		return true
	}

	return false
}

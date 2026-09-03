package slowTests

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-bridge-eth-go/integrationTests/relayers/slowTests/framework"
)

// FlowType defines the flow type
type FlowType string

// definition of the defined test flows
const (
	StartFromEthereumFlow   FlowType = "start from Ethereum"
	StartFromMultiversXFlow FlowType = "start from MultiversX"
)

// TestFlow defines the struct used in a test flow
type TestFlow struct {
	testing.TB
	FlowType
	Setup                        *framework.TestSetup
	FirstHalfBridgeDone          bool
	SecondHalfBridgeDone         bool
	Tokens                       []framework.TestTokenParams
	MessageAfterFirstHalfBridge  string
	MessageAfterSecondHalfBridge string
	HandlerAfterFirstHalfBridge  func(flow *TestFlow)
	HandlerToStartFirstBridge    func(flow *TestFlow)
}

// Process is the flow's main process function
func (flow *TestFlow) Process() (finished bool) {
	if len(flow.Tokens) == 0 {
		return true
	}
	if flow.FirstHalfBridgeDone && flow.SecondHalfBridgeDone {
		return true
	}

	if !flow.FirstHalfBridgeDone {
		transferDoneForFirstHalf := flow.Setup.AreAllTransfersCompleted(framework.FirstHalfBridge, flow.Tokens...)
		if transferDoneForFirstHalf {
			flow.FirstHalfBridgeDone = true
			log.Info(fmt.Sprintf(framework.LogStepMarker, flow.MessageAfterFirstHalfBridge))

			flow.HandlerAfterFirstHalfBridge(flow)
		}

		return false
	}

	if flow.Setup.MultiversxHandler.HasRefundBatch(flow.Setup.Ctx) {
		flow.Setup.MultiversxHandler.MoveRefundBatchToSafe(flow.Setup.Ctx)
	}

	transferDoneForSecondHalf := flow.Setup.AreAllTransfersCompleted(framework.SecondHalfBridge, flow.Tokens...)
	if !flow.SecondHalfBridgeDone && transferDoneForSecondHalf {
		flow.Setup.CheckCorrectnessOnMintBurnTokens(flow.Tokens...)
		flow.Setup.ExecuteSpecialChecks(flow.Tokens...)

		flow.SecondHalfBridgeDone = true
		log.Info(fmt.Sprintf(framework.LogStepMarker, flow.MessageAfterSecondHalfBridge))

		return true
	}

	return false
}

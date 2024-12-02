//go:build slow

package slowTests

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-bridge-eth-go/integrationTests/relayers/slowTests/framework"
)

type startsFromEthereumFlow struct {
	testing.TB
	setup        *framework.TestSetup
	ethToMvxDone bool
	mvxToEthDone bool
	tokens       []framework.TestTokenParams
}

func (flow *startsFromEthereumFlow) process() (finished bool) {
	if len(flow.tokens) == 0 {
		return true
	}
	if flow.mvxToEthDone && flow.ethToMvxDone {
		return true
	}

	if !flow.ethToMvxDone {
		transferDoneForFirstHalf := flow.setup.AreAllTransfersCompleted(framework.FirstHalfBridge, flow.tokens...)
		if transferDoneForFirstHalf {
			flow.ethToMvxDone = true
			log.Info(fmt.Sprintf(framework.LogStepMarker, "Ethereum->MultiversX transfer finished, now sending back to Ethereum..."))

			flow.setup.SendFromMultiversxToEthereum(flow.setup.BobKeys, flow.setup.CharlieKeys, flow.tokens...)
		}

		return false
	}

	transferDoneForSecondHalf := flow.setup.AreAllTransfersCompleted(framework.SecondHalfBridge, flow.tokens...)
	if !flow.mvxToEthDone && transferDoneForSecondHalf {
		flow.mvxToEthDone = true
		log.Info(fmt.Sprintf(framework.LogStepMarker, "MultiversX<->Ethereum from Ethereum transfers done"))
		return true
	}

	return false
}

func (flow *startsFromEthereumFlow) areTokensFullyRefunded() bool {
	if len(flow.tokens) == 0 {
		return true
	}
	if !flow.ethToMvxDone {
		return false // regular flow is not completed
	}

	if flow.setup.MultiversxHandler.HasRefundBatch(flow.setup.Ctx) {
		flow.setup.MultiversxHandler.MoveRefundBatchToSafe(flow.setup.Ctx)
	}

	return flow.setup.IsTransferDoneFromEthereumWithRefund(flow.setup.AliceKeys, flow.tokens...)
}

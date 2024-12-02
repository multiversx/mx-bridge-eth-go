package slowTests

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-bridge-eth-go/integrationTests/relayers/slowTests/framework"
)

// TODO next PRs: remove duplicated code for startsFromMultiversXFlow, startsFromEthereumFlow and startsFromEthereumEdgecaseFlow
type startsFromMultiversXFlow struct {
	testing.TB
	setup        *framework.TestSetup
	ethToMvxDone bool
	mvxToEthDone bool
	tokens       []framework.TestTokenParams
}

func (flow *startsFromMultiversXFlow) process() (finished bool) {
	if len(flow.tokens) == 0 {
		return true
	}
	if flow.mvxToEthDone && flow.ethToMvxDone {
		return true
	}

	if !flow.mvxToEthDone {
		transferDoneForFirstHalf := flow.setup.AreAllTransfersCompleted(framework.FirstHalfBridge, flow.tokens...)
		if transferDoneForFirstHalf {
			flow.mvxToEthDone = true
			log.Info(fmt.Sprintf(framework.LogStepMarker, "MultiversX->Ethereum transfer finished, now sending back to MultiversX..."))

			flow.setup.SendFromEthereumToMultiversX(flow.setup.BobKeys, flow.setup.CharlieKeys, flow.setup.MultiversxHandler.CalleeScAddress, flow.tokens...)
		}
	}
	if !flow.mvxToEthDone {
		// return here, no reason to check downwards
		return false
	}

	transferDoneForSecondHalf := flow.setup.AreAllTransfersCompleted(framework.SecondHalfBridge, flow.tokens...)
	if !flow.ethToMvxDone && transferDoneForSecondHalf {
		flow.ethToMvxDone = true
		log.Info(fmt.Sprintf(framework.LogStepMarker, "MultiversX<->Ethereum from MultiversX transfers done"))
		return true
	}

	return false
}

func (flow *startsFromMultiversXFlow) areTokensFullyRefunded() bool {
	if len(flow.tokens) == 0 {
		return true
	}
	if !flow.ethToMvxDone {
		return false // regular flow is not completed
	}

	return flow.setup.IsTransferDoneFromEthereumWithRefund(flow.tokens...)
}

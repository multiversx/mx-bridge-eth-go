//go:build slow

package slowTests

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-bridge-eth-go/integrationTests/relayers/slowTests/framework"
)

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

	isTransferDoneFromMultiversX := flow.setup.IsTransferDoneFromMultiversX(flow.tokens...)
	if !flow.mvxToEthDone && isTransferDoneFromMultiversX {
		flow.mvxToEthDone = true
		log.Info(fmt.Sprintf(framework.LogStepMarker, "MultiversX->PeerChain transfer finished, now sending back to MultiversX..."))

		flow.setup.PeerChainHandler.SendFromPeerChainToMultiversX(flow.setup.Ctx, flow.setup.MultiversxHandler.TestCallerAddress, flow.tokens...)
	}
	if !flow.mvxToEthDone {
		// return here, no reason to check downwards
		return false
	}

	isTransferDoneFromPeerChain := flow.setup.IsTransferDoneFromPeerChain(flow.tokens...)
	if !flow.ethToMvxDone && isTransferDoneFromPeerChain {
		flow.ethToMvxDone = true
		log.Info(fmt.Sprintf(framework.LogStepMarker, "MultiversX<->PeerChain from MultiversX transfers done"))
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

	return flow.setup.IsTransferDoneFromPeerChainWithRefund(flow.tokens...)
}

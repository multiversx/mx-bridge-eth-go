//go:build slow

package slowTests

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-bridge-eth-go/integrationTests/relayers/slowTests/framework"
)

type startsFromEthereumEdgecaseFlow struct {
	testing.TB
	setup        *framework.TestSetup
	ethToMvxDone bool
	mvxToEthDone bool
	tokens       []framework.TestTokenParams
}

func (flow *startsFromEthereumEdgecaseFlow) process() (finished bool) {
	if len(flow.tokens) == 0 {
		return true
	}
	if flow.mvxToEthDone && flow.ethToMvxDone {
		return true
	}

	isTransferDoneFromEthereum := flow.setup.IsTransferDoneFromEthereum(flow.setup.AliceKeys, flow.setup.BobKeys, flow.tokens...)
	if !flow.ethToMvxDone && isTransferDoneFromEthereum {
		flow.ethToMvxDone = true
		log.Info(fmt.Sprintf(framework.LogStepMarker, "Ethereum->MultiversX transfer finished, now sending back to Ethereum & another round from Ethereum..."))

		flow.setup.SendFromMultiversxToEthereum(flow.setup.BobKeys, flow.setup.AliceKeys, flow.tokens...)
		flow.setup.SendFromEthereumToMultiversX(flow.setup.AliceKeys, flow.setup.BobKeys, flow.setup.MultiversxHandler.CalleeScAddress, flow.tokens...)
	}
	if !flow.ethToMvxDone {
		// return here, no reason to check downwards
		return false
	}

	isTransferDoneFromMultiversX := flow.setup.IsTransferDoneFromMultiversX(flow.setup.BobKeys, flow.setup.AliceKeys, flow.tokens...)
	if !flow.mvxToEthDone && isTransferDoneFromMultiversX {
		flow.mvxToEthDone = true
		log.Info(fmt.Sprintf(framework.LogStepMarker, "MultiversX<->Ethereum from Ethereum transfers done"))
		return true
	}

	return false
}
//go:build slow

package slowTests

import (
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
		log.Info("MultiversX->Ethereum transfer finished, now sending back to MultiversX...")

		flow.setup.EthereumHandler.SendFromEthereumToMultiversX(flow.setup.Ctx, flow.setup.MultiversxHandler.TestCallerAddress, flow.tokens...)
	}
	if !flow.mvxToEthDone {
		// return here, no reason to check downwards
		return false
	}

	isTransferDoneFromEthereum := flow.setup.IsTransferDoneFromEthereum(flow.tokens...)
	if !flow.ethToMvxDone && isTransferDoneFromEthereum {
		flow.ethToMvxDone = true
		log.Info("MultiversX<->Ethereum transfers done")
		return true
	}

	return false
}

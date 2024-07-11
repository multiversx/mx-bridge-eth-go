//go:build slow

package slowTests

import "testing"

type startsFromMultiversXFlow struct {
	testing.TB
	testSetup    *simulatedSetup
	ethToMvxDone bool
	mvxToEthDone bool
	tokens       []testTokenParams
}

func (flow *startsFromMultiversXFlow) process() (finished bool) {
	if len(flow.tokens) == 0 {
		return true
	}
	if flow.mvxToEthDone && flow.ethToMvxDone {
		return true
	}

	isTransferDoneFromMultiversX := flow.testSetup.isTransferDoneFromMultiversX(flow.tokens...)
	if !flow.mvxToEthDone && isTransferDoneFromMultiversX {
		flow.mvxToEthDone = true
		log.Info("MultiversX->Ethereum transfer finished, now sending back to MultiversX...")

		flow.testSetup.sendFromEthereumToMultiversX(flow.tokens...)
	}
	if !flow.mvxToEthDone {
		// return here, no reason to check downwards
		return false
	}

	isTransferDoneFromEthereum := flow.testSetup.isTransferDoneFromEthereum(flow.tokens...)
	if !flow.ethToMvxDone && isTransferDoneFromEthereum {
		flow.ethToMvxDone = true
		log.Info("MultiversX<->Ethereum transfers done")
		return true
	}

	return false
}

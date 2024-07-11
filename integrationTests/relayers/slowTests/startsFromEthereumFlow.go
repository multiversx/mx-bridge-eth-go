//go:build slow

package slowTests

import "testing"

type startsFromEthereumFlow struct {
	testing.TB
	testSetup    *simulatedSetup
	ethToMvxDone bool
	mvxToEthDone bool
	tokens       []testTokenParams
}

func (flow *startsFromEthereumFlow) process() (finished bool) {
	if len(flow.tokens) == 0 {
		return true
	}
	if flow.mvxToEthDone && flow.ethToMvxDone {
		return true
	}

	isTransferDoneFromEthereum := flow.testSetup.isTransferDoneFromEthereum(flow.tokens...)
	if !flow.ethToMvxDone && isTransferDoneFromEthereum {
		flow.ethToMvxDone = true
		log.Info("Ethereum->MultiversX transfer finished, now sending back to Ethereum...")

		flow.testSetup.sendFromMultiversxToEthereum(flow.tokens...)
	}
	if !flow.ethToMvxDone {
		// return here, no reason to check downwards
		return false
	}

	isTransferDoneFromMultiversX := flow.testSetup.isTransferDoneFromMultiversX(flow.tokens...)
	if !flow.mvxToEthDone && isTransferDoneFromMultiversX {
		flow.mvxToEthDone = true
		log.Info("MultiversX<->Ethereum transfers done")
		return true
	}

	return false
}

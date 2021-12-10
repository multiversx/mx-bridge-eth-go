package ethToElrond

const (
	// GettingPendingBatchFromEthereum is the step identifier for fetching the pending batch from the Ethereum chain
	GettingPendingBatchFromEthereum = "get pending batch from ethereum"

	// GettingActionIdForProposeTransfer is the step identifier for fetching the action ID for propose transfer on Elrond
	GettingActionIdForProposeTransfer = "get action ID for propose transfer"

	// ProposingTransferOnElrond is the step idetifier for proposing transfer on Elrond
	ProposingTransferOnElrond = "propose transfer"

	// SigningProposedTransferOnElrond is the step identifier for signing proposed transfer
	SigningProposedTransferOnElrond = "sign proposed transfer"

	// WaitingForQuorum is the step identifier for waiting until the quorum is reached
	WaitingForQuorum = "wait for quorum"

	// PerformingActionID is the step identifier for performing the ActionID on Elrond
	PerformingActionID = "perform action"

	// NoFailing indicates that the states machine performed also the last step without any error
	NoFailing = "noFailing"

	// numSteps indicates how many steps the
	numSteps = 6
)

var StepList = [numSteps + 1]string{
	GettingPendingBatchFromEthereum,
	GettingActionIdForProposeTransfer,
	ProposingTransferOnElrond,
	SigningProposedTransferOnElrond,
	WaitingForQuorum,
	PerformingActionID,
	NoFailing,
}

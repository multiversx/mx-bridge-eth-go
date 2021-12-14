package ethToElrond

const (
	// GettingPendingBatchFromEthereum is the step identifier for fetching the pending batch from the Ethereum chain
	GettingPendingBatchFromEthereum = "get pending batch from ethereum"

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
	numSteps = 5
)

// FailingStepList is the list of all steps where from Ethereum to elrond flow indicating
// at which step one relayer may fail or NoFailing in case all steps were executed successfully
var FailingStepList = [numSteps + 1]string{
	GettingPendingBatchFromEthereum,
	ProposingTransferOnElrond,
	SigningProposedTransferOnElrond,
	WaitingForQuorum,
	PerformingActionID,
	NoFailing,
}

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

	// numSteps indicates how many steps the state machine for Ethereum -> Elrond flow has
	NumSteps = 5
)

package ethtomultiversx

const (
	// GettingPendingBatchFromEthereum is the step identifier for fetching the pending batch from the Ethereum chain
	GettingPendingBatchFromEthereum = "get pending batch from Ethereum"

	// ProposingTransferOnMultiversX is the step idetifier for proposing transfer on MultiversX
	ProposingTransferOnMultiversX = "propose transfer"

	// SigningProposedTransferOnMultiversX is the step identifier for signing proposed transfer
	SigningProposedTransferOnMultiversX = "sign proposed transfer"

	// WaitingForQuorum is the step identifier for waiting until the quorum is reached
	WaitingForQuorum = "wait for quorum"

	// PerformingActionID is the step identifier for performing the ActionID on MultiversX
	PerformingActionID = "perform action"

	// NumSteps indicates how many steps the state machine for Ethereum -> MultiversX flow has
	NumSteps = 5
)

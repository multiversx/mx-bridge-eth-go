package ethtomultiversx

const (
	// GettingPendingBatchFromEthereum is the step identifier for fetching the pending batch from the Ethereum chain
	GettingPendingBatchFromEthereum = "get pending batch from Ethereum"

	// ProposingTransferOnMultiversX is the step identifier for proposing transfer on MultiversX
	ProposingTransferOnMultiversX = "propose transfer"

	// ProposingSCTransfersOnMultiversX is the step identifier for proposing smart contract executions on MultiversX
	ProposingSCTransfersOnMultiversX = "propose sc transfer"

	// SigningProposedTransferOnMultiversX is the step identifier for signing proposed transfer
	SigningProposedTransferOnMultiversX = "sign proposed transfer"

	// SigningProposedSCTransferOnMultiversX is the step identifier for signing proposed smart contract transfers
	SigningProposedSCTransferOnMultiversX = "sign proposed transfer"

	// WaitingForQuorum is the step identifier for waiting until the quorum is reached
	WaitingForQuorum = "wait for quorum"

	// PerformingActionID is the step identifier for performing the ActionID on MultiversX
	PerformingActionID = "perform action"

	// NumSteps indicates how many steps the state machine for Ethereum -> MultiversX flow has
	NumSteps = 5
)

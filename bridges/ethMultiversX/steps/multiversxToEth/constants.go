package multiversxtoeth

const (
	// GettingPendingBatchFromMultiversX is the step identifier for fetching the pending batch from the MultiversX chain
	GettingPendingBatchFromMultiversX = "get pending batch from MultiversX"

	// SigningProposedTransferOnEthereum is the step identifier for signing proposed transfer
	SigningProposedTransferOnEthereum = "sign proposed transfer"

	// WaitingForQuorumOnTransfer is the step identifier for waiting until the quorum is reached
	WaitingForQuorumOnTransfer = "wait for quorum on transfer"

	// PerformingTransfer is the step identifier for performing the transfer on Ethereum
	PerformingTransfer = "perform transfer"

	// WaitingTransferConfirmation is the step identifier for waiting the transfer confirmation on Ethereum
	WaitingTransferConfirmation = "wait transfer confirmating"

	// ResolvingSetStatusOnMultiversX is the step identifier for resolving set status on MultiversX
	ResolvingSetStatusOnMultiversX = "resolve set status"

	// ProposingSetStatusOnMultiversX is the step idetifier for proposing set status action on MultiversX
	ProposingSetStatusOnMultiversX = "propose set status"

	// SigningProposedSetStatusOnMultiversX is the step identifier for signing proposed set status action
	SigningProposedSetStatusOnMultiversX = "sign proposed set status"

	// WaitingForQuorumOnSetStatus is the step identifier for waiting until the quorum is reached
	WaitingForQuorumOnSetStatus = "wait for quorum on set status"

	// PerformingSetStatus is the step identifier for performing the set status action on MultiversX
	PerformingSetStatus = "perform set status"

	// NumSteps indicates how many steps the state machine for MultiversX -> Ethereum flow has
	NumSteps = 10
)

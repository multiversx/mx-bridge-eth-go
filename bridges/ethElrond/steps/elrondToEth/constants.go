package elrondToEth

const (
	// GettingPendingBatchFromElrond is the step identifier for fetching the pending batch from the Elrond chain
	GettingPendingBatchFromElrond = "get pending batch from Elrond"

	// SigningProposedTransferOnEthereum is the step identifier for signing proposed transfer
	SigningProposedTransferOnEthereum = "sign proposed transfer"

	// WaitingForQuorumOnTransfer is the step identifier for waiting until the quorum is reached
	WaitingForQuorumOnTransfer = "wait for quorum on transfer"

	// PerformingTransfer is the step identifier for performing the transfer on Ethereum
	PerformingTransfer = "perform transfer"

	// WaitingTransferConfirmation is the step identifier for waiting the transfer confirmation on Ethereum
	WaitingTransferConfirmation = "wait transfer confirmating"

	// ResolvingSetStatusOnElrond is the step idetifier for resolving set status on Elrond
	ResolvingSetStatusOnElrond = "resolve set status"

	// ProposingSetStatusOnElrond is the step idetifier for proposing set status action on Elrond
	ProposingSetStatusOnElrond = "propose set status"

	// SigningProposedSetStatusOnElrond is the step identifier for signing proposed set status action
	SigningProposedSetStatusOnElrond = "sign proposed set status"

	// WaitingForQuorumOnSetStatus is the step identifier for waiting until the quorum is reached
	WaitingForQuorumOnSetStatus = "wait for quorum on set status"

	// PerformingSetStatus is the step identifier for performing the set status action on Elrond
	PerformingSetStatus = "perform set status"

	// NumSteps indicates how many steps the state machine for Elrond -> Ethereum flow has
	NumSteps = 10
)

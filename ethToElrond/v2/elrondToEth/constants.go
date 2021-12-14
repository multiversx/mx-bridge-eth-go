package elrondToEth

const (
	// GettingPendingBatchFromElrond is the step identifier for fetching the pending batch from the Elrond chain
	GettingPendingBatchFromElrond = "get pending batch from ethereum"

	// SigningProposedTransferOnEthereum is the step identifier for signing proposed transfer
	SigningProposedTransferOnEthereum = "sign proposed transfer"

	// WaitingForQuorumOnTransfer is the step identifier for waiting until the quorum is reached
	WaitingForQuorumOnTransfer = "wait for quorum on transfer"

	// PerformingTransfer is the step identifier for performing the transfer on Ethereum
	PerformingTransfer = "perform transfer"

	// WaitingTransferConfirmation is the step identifier for waiting the transfer confirmation on Ethereum
	WaitingTransferConfirmation = "wait transfer confirmationg"

	// ProposingSetStatusOnElrond is the step idetifier for proposing set status action on Elrond
	ProposingSetStatusOnElrond = "propose set status"

	// SigningProposedSetStatusOnElrond is the step identifier for signing proposed set status action
	SigningProposedSetStatusOnElrond = "sign proposed set status"

	// WaitingForQuorumOnSetStatus is the step identifier for waiting until the quorum is reached
	WaitingForQuorumOnSetStatus = "wait for quorum on set status"

	// PerformingSetStatus is the step identifier for performing the set status action on Elrond
	PerformingSetStatus = "perform set status"

	// NoFailing indicates that the states machine performed also the last step without any error
	NoFailing = "noFailing"

	// numSteps indicates how many steps the
	numSteps = 9
)

// FailingStepList is the list of all steps where from Ethereum to elrond flow indicating
// at which step one relayer may fail or NoFailing in case all steps were executed successfully
var FailingStepList = [numSteps + 1]string{
	GettingPendingBatchFromElrond,
	SigningProposedTransferOnEthereum,
	WaitingForQuorumOnTransfer,
	PerformingTransfer,
	ProposingSetStatusOnElrond,
	SigningProposedSetStatusOnElrond,
	WaitingForQuorumOnSetStatus,
	PerformingSetStatus,
	NoFailing,
}

package ethToElrond

const (
	// GettingPending is the step definition for the pending transactions check
	GettingPending = "getting the pending transactions"

	// ProposingTransfer is the step definition for the propose transfer operation
	ProposingTransfer = "proposing transfer"

	// WaitingSignaturesForProposeTransfer is the step definition for the signature gathering process for propose transfer
	WaitingSignaturesForProposeTransfer = "waiting signatures for propose transfer"

	// ExecutingTransfer is the step definition for the execution of the transfer operation on the destination chain
	ExecutingTransfer = "executing transfer"

	// ProposingSetStatus is the step definition for the propose set status operation
	ProposingSetStatus = "proposing set status"

	// WaitingSignaturesForProposeSetStatus is the step definition for the signature gathering process for propose set status
	WaitingSignaturesForProposeSetStatus = "waiting signatures for propose set status"

	// ExecutingSetStatus is the step definition for the execution of the set status operation on the source chain
	ExecutingSetStatus = "executing set status"
)

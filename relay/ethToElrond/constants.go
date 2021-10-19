package ethToElrond

const (
	// GetPending is the step definition for the pending transactions check
	GetPending = "get pending transactions"

	// ProposeTransfer is the step definition for the propose transfer operation
	ProposeTransfer = "propose transfer"

	// WaitForSignaturesForProposeTransfer is the step definition for the signature gathering process for propose transfer
	WaitForSignaturesForProposeTransfer = "wait for signatures for transfer propose"

	// ExecuteTransfer is the step definition for the execution of the transfer operation on the destination chain
	ExecuteTransfer = "execute transfer"

	// ProposeSetStatus is the step definition for the propose set status operation
	ProposeSetStatus = "propose set status"

	// WaitForSignaturesForProposeSetStatus is the step definition for the signature gathering process for propose set status
	WaitForSignaturesForProposeSetStatus = "wait for signatures for set status propose"

	// ExecuteSetStatus is the step definition for the execution of the set status operation on the source chain
	ExecuteSetStatus = "execute set status"
)

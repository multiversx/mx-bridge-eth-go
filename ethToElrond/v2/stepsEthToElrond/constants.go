package stepsEthToElrond

const (
	// GetPendingBatchFromEthereum is the step identifier for fetching the pending batch from the Ethereum chain
	GetPendingBatchFromEthereum = "get pending batch from ethereum"

	// GetActionIdForProposeStep is the step identifier for fetching the action ID for propose transfer on Elrond
	GetActionIdForProposeStep = "get action ID for propose transfer"

	// ProposeTransferOnElrond is the step idetifier for proposing transfer on Elrond
	ProposeTransferOnElrond = "propose transfer"

	// SignProposedTransferOnElrond is the step identifier for signing proposed transfer
	SignProposedTransferOnElrond = "sign proposed transfer"

	// WaitForQuorum is the step identifier for waiting until the quorum is reached
	WaitForQuorum = "wait for quorum"

	// PerformActionID is the step identifier for performing the ActionID on Elrond
	PerformActionID = "perform action"
)

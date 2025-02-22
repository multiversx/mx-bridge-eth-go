package ESDTtransferIssues

import "github.com/multiversx/mx-chain-core-go/data/transaction"

// FailedTransactionNotifier defines the operations of a failed transaction notifier
type FailedTransactionNotifier interface {
	BeforeSendingTransaction(tx *transaction.FrontendTransaction)
	ClearInternalStateData()
	RegisterHandler(handler func(txData string, numCalls int))
	ClearNotifierHandlers()
}

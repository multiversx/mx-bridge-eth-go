package framework

import (
	"strings"
	"sync"

	"github.com/multiversx/mx-chain-core-go/data/transaction"
)

const txDataTokensSeparator = "@"

type failedTransactionsNotifier struct {
	targetFunction      string
	numFailuresAccepted int
	mutData             sync.Mutex
	dataMap             map[string]int
	mutNotifiers        sync.RWMutex
	notifiers           []func(txData string, numCalls int)
}

// NewFailedTransactionsNotifier creates a new failed transaction notifier
func NewFailedTransactionsNotifier(
	targetFunction string,
	numFailuresAccepted int,
) *failedTransactionsNotifier {
	return &failedTransactionsNotifier{
		targetFunction:      strings.ToLower(targetFunction),
		numFailuresAccepted: numFailuresAccepted,
		dataMap:             make(map[string]int),
	}
}

// BeforeSendingTransaction should be called before sending the transaction to the chain simulator / test network
func (notifier *failedTransactionsNotifier) BeforeSendingTransaction(tx *transaction.FrontendTransaction) {
	if !notifier.shouldProcess(tx) {
		return
	}

	txData := string(tx.Data)
	shouldProcess, numCalls := notifier.processTxData(txData)
	if !shouldProcess {
		return
	}

	notifier.notifyAll(txData, numCalls)
}

func (notifier *failedTransactionsNotifier) shouldProcess(tx *transaction.FrontendTransaction) bool {
	if tx == nil {
		return false
	}

	txData := string(tx.Data)
	txDataTokens := strings.Split(txData, txDataTokensSeparator)
	return strings.ToLower(txDataTokens[0]) == notifier.targetFunction
}

func (notifier *failedTransactionsNotifier) processTxData(txData string) (bool, int) {
	notifier.mutData.Lock()
	defer notifier.mutData.Unlock()

	notifier.dataMap[txData]++
	encountered := notifier.dataMap[txData]

	return encountered > notifier.numFailuresAccepted, encountered
}

func (notifier *failedTransactionsNotifier) notifyAll(txData string, numCalls int) {
	notifier.mutNotifiers.RLock()
	defer notifier.mutNotifiers.RUnlock()

	for _, handler := range notifier.notifiers {
		handler(txData, numCalls)
	}
}

// ClearInternalStateData clears the internal state data (not the notifiers list)
func (notifier *failedTransactionsNotifier) ClearInternalStateData() {
	notifier.mutData.Lock()
	defer notifier.mutData.Unlock()

	notifier.dataMap = make(map[string]int)
}

// RegisterHandler registers a new handler
func (notifier *failedTransactionsNotifier) RegisterHandler(handler func(txData string, numCalls int)) {
	if handler == nil {
		return
	}

	notifier.mutNotifiers.Lock()
	notifier.notifiers = append(notifier.notifiers, handler)
	notifier.mutNotifiers.Unlock()
}

// ClearNotifierHandlers clears the internal list of notifiers
func (notifier *failedTransactionsNotifier) ClearNotifierHandlers() {
	notifier.mutNotifiers.Lock()
	notifier.notifiers = make([]func(txData string, numCalls int), 0)
	notifier.mutNotifiers.Unlock()
}

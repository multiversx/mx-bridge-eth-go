package framework

import (
	"context"
	"sync"

	"github.com/multiversx/mx-bridge-eth-go/clients/multiversx"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
)

type proxyWrapper struct {
	multiversx.Proxy

	mutHandlers          sync.RWMutex
	beforeTxSendHandlers []func(tx *transaction.FrontendTransaction)
}

// NewProxyWrapper will create a wrapper over the provided proxy
func NewProxyWrapper(proxy multiversx.Proxy) *proxyWrapper {
	return &proxyWrapper{
		Proxy: proxy,
	}
}

// SendTransaction is a wrapper over the send transaction original functionality
func (wrapper *proxyWrapper) SendTransaction(ctx context.Context, tx *transaction.FrontendTransaction) (string, error) {
	wrapper.callBeforeTransactionSendHandlers(tx)

	return wrapper.Proxy.SendTransaction(ctx, tx)
}

// SendTransactions is a wrapper over the send transactions original functionality
func (wrapper *proxyWrapper) SendTransactions(ctx context.Context, txs []*transaction.FrontendTransaction) ([]string, error) {
	for _, tx := range txs {
		wrapper.callBeforeTransactionSendHandlers(tx)
	}

	return wrapper.Proxy.SendTransactions(ctx, txs)
}

func (wrapper *proxyWrapper) callBeforeTransactionSendHandlers(tx *transaction.FrontendTransaction) {
	if tx == nil {
		return
	}

	wrapper.mutHandlers.RLock()
	for _, handler := range wrapper.beforeTxSendHandlers {
		handler(tx)
	}
	wrapper.mutHandlers.RUnlock()
}

// RegisterBeforeTransactionSendHandler will register the handler to be called before a transaction is being sent
func (wrapper *proxyWrapper) RegisterBeforeTransactionSendHandler(handler func(tx *transaction.FrontendTransaction)) {
	if handler == nil {
		return
	}

	wrapper.mutHandlers.Lock()
	wrapper.beforeTxSendHandlers = append(wrapper.beforeTxSendHandlers, handler)
	wrapper.mutHandlers.Unlock()
}

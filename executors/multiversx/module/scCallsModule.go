package module

import (
	"time"

	"github.com/multiversx/mx-bridge-eth-go/config"
	"github.com/multiversx/mx-bridge-eth-go/executors/multiversx"
	"github.com/multiversx/mx-bridge-eth-go/executors/multiversx/filters"
	"github.com/multiversx/mx-bridge-eth-go/parsers"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-crypto-go/signing"
	"github.com/multiversx/mx-chain-crypto-go/signing/ed25519"
	"github.com/multiversx/mx-chain-crypto-go/signing/ed25519/singlesig"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-sdk-go/core/polling"
	"github.com/multiversx/mx-sdk-go/interactors"
	"github.com/multiversx/mx-sdk-go/interactors/nonceHandlerV2"
)

var suite = ed25519.NewEd25519()
var keyGen = signing.NewKeyGenerator(suite)
var singleSigner = &singlesig.Ed25519Signer{}

type scCallsModule struct {
	cfg             config.ScCallsModuleConfig
	log             logger.Logger
	filter          multiversx.ScCallsExecuteFilter
	proxy           multiversx.Proxy
	nonceTxsHandler nonceTransactionsHandler
	txExecutor      multiversx.TransactionExecutor

	pollingHandlers []pollingHandler
	executors       []executor
}

// NewScCallsModule creates a starts a new scCallsModule instance
func NewScCallsModule(cfg config.ScCallsModuleConfig, proxy multiversx.Proxy, log logger.Logger, chCloseApp chan struct{}) (*scCallsModule, error) {
	if check.IfNil(proxy) {
		return nil, errNilProxy //TODO: add unit test for this
	}

	module := &scCallsModule{
		cfg: cfg,
		log: log,
		proxy: proxy,
	}

	err := module.createFilter()
	if err != nil {
		return nil, err
	}

	err = module.createNonceTxHandler()
	if err != nil {
		return nil, err
	}

	err = module.createTransactionExecutor(chCloseApp)
	if err != nil {
		return nil, err
	}

	err = module.createScCallsExecutor()
	if err != nil {
		return nil, err
	}

	err = module.createRefundExecutor()
	if err != nil {
		return nil, err
	}

	return module, nil
}

func (module *scCallsModule) createFilter() error {
	var err error
	module.filter, err = filters.NewPendingOperationFilter(module.cfg.Filter, module.log)

	return err
}

func (module *scCallsModule) createNonceTxHandler() error {
	argNonceHandler := nonceHandlerV2.ArgsNonceTransactionsHandlerV2{
		Proxy:            module.proxy,
		IntervalToResend: time.Second * time.Duration(module.cfg.General.IntervalToResendTxsInSeconds),
	}

	var err error
	module.nonceTxsHandler, err = nonceHandlerV2.NewNonceTransactionHandlerV2(argNonceHandler)

	return err
}

func (module *scCallsModule) createTransactionExecutor(chCloseApp chan struct{}) error {
	wallet := interactors.NewWallet()
	multiversXPrivateKeyBytes, err := wallet.LoadPrivateKeyFromPemFile(module.cfg.General.PrivateKeyFile)
	if err != nil {
		return err
	}

	privateKey, err := keyGen.PrivateKeyFromByteArray(multiversXPrivateKeyBytes)
	if err != nil {
		return err
	}

	argsTxExecutor := multiversx.ArgsTransactionExecutor{
		Proxy:             module.proxy,
		Log:               module.log,
		NonceTxHandler:    module.nonceTxsHandler,
		PrivateKey:        privateKey,
		SingleSigner:      singleSigner,
		TransactionChecks: module.cfg.TransactionChecks,
		CloseAppChan:      chCloseApp,
	}

	module.txExecutor, err = multiversx.NewTransactionExecutor(argsTxExecutor)

	return err
}

func (module *scCallsModule) createScCallsExecutor() error {
	argsExecutor := multiversx.ArgsScCallExecutor{
		ScProxyBech32Addresses: module.cfg.General.ScProxyBech32Addresses,
		TransactionExecutor:    module.txExecutor,
		Proxy:                  module.proxy,
		Codec:                  &parsers.MultiversxCodec{},
		Filter:                 module.filter,
		Log:                    module.log,
		ExecutorConfig:         module.cfg.ScCallsExecutor,
	}

	executorInstance, err := multiversx.NewScCallExecutor(argsExecutor)
	if err != nil {
		return err
	}
	module.executors = append(module.executors, executorInstance)

	argsPollingHandler := polling.ArgsPollingHandler{
		Log:              module.log,
		Name:             "MultiversX SC calls",
		PollingInterval:  time.Duration(module.cfg.ScCallsExecutor.PollingIntervalInMillis) * time.Millisecond,
		PollingWhenError: time.Duration(module.cfg.ScCallsExecutor.PollingIntervalInMillis) * time.Millisecond,
		Executor:         executorInstance,
	}

	pollingHandlerInstance, err := polling.NewPollingHandler(argsPollingHandler)
	if err != nil {
		return err
	}

	err = pollingHandlerInstance.StartProcessingLoop()
	if err != nil {
		return err
	}
	module.pollingHandlers = append(module.pollingHandlers, pollingHandlerInstance)

	return nil
}

func (module *scCallsModule) createRefundExecutor() error {
	argsExecutor := multiversx.ArgsRefundExecutor{
		ScProxyBech32Addresses: module.cfg.General.ScProxyBech32Addresses,
		TransactionExecutor:    module.txExecutor,
		Proxy:                  module.proxy,
		Codec:                  &parsers.MultiversxCodec{},
		Filter:                 module.filter,
		Log:                    module.log,
		GasToExecute:           module.cfg.RefundExecutor.GasToExecute,
	}

	executorInstance, err := multiversx.NewRefundExecutor(argsExecutor)
	if err != nil {
		return err
	}
	module.executors = append(module.executors, executorInstance)

	argsPollingHandler := polling.ArgsPollingHandler{
		Log:              module.log,
		Name:             "MultiversX refund executor",
		PollingInterval:  time.Duration(module.cfg.RefundExecutor.PollingIntervalInMillis) * time.Millisecond,
		PollingWhenError: time.Duration(module.cfg.RefundExecutor.PollingIntervalInMillis) * time.Millisecond,
		Executor:         executorInstance,
	}

	pollingHandlerInstance, err := polling.NewPollingHandler(argsPollingHandler)
	if err != nil {
		return err
	}

	err = pollingHandlerInstance.StartProcessingLoop()
	if err != nil {
		return err
	}
	module.pollingHandlers = append(module.pollingHandlers, pollingHandlerInstance)

	return nil
}

// GetNumSentTransaction returns the total sent transactions
func (module *scCallsModule) GetNumSentTransaction() uint32 {
	return module.txExecutor.GetNumSentTransaction()
}

// Close closes any components started
func (module *scCallsModule) Close() error {
	var lastError error

	for _, handlers := range module.pollingHandlers {
		err := handlers.Close()
		if err != nil {
			lastError = err
		}
	}

	err := module.nonceTxsHandler.Close()
	if err != nil {
		lastError = err
	}

	return lastError
}
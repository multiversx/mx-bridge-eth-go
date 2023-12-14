package multiversx

import (
	"context"
	"fmt"
	"math/big"
	"reflect"
	"sync"
	"time"

	"github.com/multiversx/mx-bridge-eth-go/bridges/ethMultiversX"
	"github.com/multiversx/mx-bridge-eth-go/clients"
	"github.com/multiversx/mx-bridge-eth-go/config"
	bridgeCore "github.com/multiversx/mx-bridge-eth-go/core"
	"github.com/multiversx/mx-bridge-eth-go/core/converters"
	"github.com/multiversx/mx-chain-core-go/core/check"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-crypto-go/signing/ed25519/singlesig"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-sdk-go/builders"
	"github.com/multiversx/mx-sdk-go/core"
	"github.com/multiversx/mx-sdk-go/data"
	"github.com/multiversx/mx-sdk-go/interactors/nonceHandlerV2"
)

const (
	proposeTransferFuncName  = "proposeMultiTransferEsdtBatch"
	proposeSetStatusFuncName = "proposeEsdtSafeSetCurrentTransactionBatchStatus"
	signFuncName             = "sign"
	performActionFuncName    = "performAction"
	minAllowedDelta          = 1

	multiversXDataGetterLogId = "MultiversXEth-MultiversXDataGetter"
)

// ClientArgs represents the argument for the NewClient constructor function
type ClientArgs struct {
	GasMapConfig                 config.MultiversXGasMapConfig
	Proxy                        Proxy
	Log                          logger.Logger
	RelayerPrivateKey            crypto.PrivateKey
	MultisigContractAddress      core.AddressHandler
	IntervalToResendTxsInSeconds uint64
	TokensMapper                 TokensMapper
	RoleProvider                 roleProvider
	StatusHandler                bridgeCore.StatusHandler
	AllowDelta                   uint64
}

// client represents the MultiversX Client implementation
type client struct {
	*mxClientDataGetter
	txHandler                 txHandler
	tokensMapper              TokensMapper
	relayerPublicKey          crypto.PublicKey
	relayerAddress            core.AddressHandler
	multisigContractAddress   core.AddressHandler
	log                       logger.Logger
	gasMapConfig              config.MultiversXGasMapConfig
	addressPublicKeyConverter bridgeCore.AddressConverter
	statusHandler             bridgeCore.StatusHandler
	allowDelta                uint64

	lastNonce                uint64
	retriesAvailabilityCheck uint64
	mut                      sync.RWMutex
}

// NewClient returns a new MultiversX Client instance
func NewClient(args ClientArgs) (*client, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	argNonceHandler := nonceHandlerV2.ArgsNonceTransactionsHandlerV2{
		Proxy:            args.Proxy,
		IntervalToResend: time.Second * time.Duration(args.IntervalToResendTxsInSeconds),
		Creator:          &nonceHandlerV2.AddressNonceHandlerCreator{},
	}
	nonceTxsHandler, err := nonceHandlerV2.NewNonceTransactionHandlerV2(argNonceHandler)
	if err != nil {
		return nil, err
	}

	publicKey := args.RelayerPrivateKey.GeneratePublic()
	publicKeyBytes, err := publicKey.ToByteArray()
	if err != nil {
		return nil, err
	}

	relayerAddress := data.NewAddressFromBytes(publicKeyBytes)

	argsMXClientDataGetter := ArgsMXClientDataGetter{
		MultisigContractAddress: args.MultisigContractAddress,
		RelayerAddress:          relayerAddress,
		Proxy:                   args.Proxy,
		Log:                     bridgeCore.NewLoggerWithIdentifier(logger.GetOrCreate(multiversXDataGetterLogId), multiversXDataGetterLogId),
	}
	getter, err := NewMXClientDataGetter(argsMXClientDataGetter)
	if err != nil {
		return nil, err
	}

	addressConverter, err := converters.NewAddressConverter()
	if err != nil {
		return nil, clients.ErrNilAddressConverter
	}

	c := &client{
		txHandler: &transactionHandler{
			proxy:                   args.Proxy,
			relayerAddress:          relayerAddress,
			multisigAddressAsBech32: args.MultisigContractAddress.AddressAsBech32String(),
			nonceTxHandler:          nonceTxsHandler,
			relayerPrivateKey:       args.RelayerPrivateKey,
			singleSigner:            &singlesig.Ed25519Signer{},
			roleProvider:            args.RoleProvider,
		},
		mxClientDataGetter:        getter,
		relayerPublicKey:          publicKey,
		relayerAddress:            relayerAddress,
		multisigContractAddress:   args.MultisigContractAddress,
		log:                       args.Log,
		gasMapConfig:              args.GasMapConfig,
		addressPublicKeyConverter: addressConverter,
		tokensMapper:              args.TokensMapper,
		statusHandler:             args.StatusHandler,
		allowDelta:                args.AllowDelta,
	}

	c.log.Info("NewMultiversXClient",
		"relayer address", relayerAddress.AddressAsBech32String(),
		"safe contract address", args.MultisigContractAddress.AddressAsBech32String())

	return c, nil
}

func checkArgs(args ClientArgs) error {
	if check.IfNil(args.Proxy) {
		return errNilProxy
	}
	if check.IfNil(args.RelayerPrivateKey) {
		return clients.ErrNilPrivateKey
	}
	if check.IfNil(args.MultisigContractAddress) {
		return fmt.Errorf("%w for the MultisigContractAddress argument", errNilAddressHandler)
	}
	if check.IfNil(args.Log) {
		return clients.ErrNilLogger
	}
	if check.IfNil(args.TokensMapper) {
		return clients.ErrNilTokensMapper
	}
	if check.IfNil(args.RoleProvider) {
		return errNilRoleProvider
	}
	if check.IfNil(args.StatusHandler) {
		return clients.ErrNilStatusHandler
	}
	if args.AllowDelta < minAllowedDelta {
		return fmt.Errorf("%w for args.AllowedDelta, got: %d, minimum: %d",
			clients.ErrInvalidValue, args.AllowDelta, minAllowedDelta)
	}
	err := checkGasMapValues(args.GasMapConfig)
	if err != nil {
		return err
	}
	return nil
}

func checkGasMapValues(gasMap config.MultiversXGasMapConfig) error {
	gasMapValue := reflect.ValueOf(gasMap)
	typeOfGasMapValue := gasMapValue.Type()

	for i := 0; i < gasMapValue.NumField(); i++ {
		fieldVal := gasMapValue.Field(i).Uint()
		if fieldVal == 0 {
			return fmt.Errorf("%w for field %s", errInvalidGasValue, typeOfGasMapValue.Field(i).Name)
		}
	}

	return nil
}

// GetPending returns the pending batch
func (c *client) GetPending(ctx context.Context) (*clients.TransferBatch, error) {
	c.log.Info("getting pending batch...")
	responseData, err := c.GetCurrentBatchAsDataBytes(ctx)
	if err != nil {
		return nil, err
	}

	if emptyResponse(responseData) {
		return nil, ErrNoPendingBatchAvailable
	}

	return c.createPendingBatchFromResponse(ctx, responseData)
}

func emptyResponse(response [][]byte) bool {
	return len(response) == 0 || (len(response) == 1 && len(response[0]) == 0)
}

func (c *client) createPendingBatchFromResponse(ctx context.Context, responseData [][]byte) (*clients.TransferBatch, error) {
	numFieldsForTransaction := 6
	dataLen := len(responseData)
	haveCorrectNumberOfArgs := (dataLen-1)%numFieldsForTransaction == 0 && dataLen > 1
	if !haveCorrectNumberOfArgs {
		return nil, fmt.Errorf("%w, got %d argument(s)", errInvalidNumberOfArguments, dataLen)
	}

	batchID, err := parseUInt64FromByteSlice(responseData[0])
	if err != nil {
		return nil, fmt.Errorf("%w while parsing batch ID", err)
	}

	batch := &clients.TransferBatch{
		ID: batchID,
	}

	cachedTokens := make(map[string][]byte)
	transferIndex := 0
	for i := 1; i < dataLen; i += numFieldsForTransaction {
		// blockNonce is the i-th element, let's ignore it for now
		depositNonce, errParse := parseUInt64FromByteSlice(responseData[i+1])
		if errParse != nil {
			return nil, fmt.Errorf("%w while parsing the deposit nonce, transfer index %d", errParse, transferIndex)
		}

		amount := big.NewInt(0).SetBytes(responseData[i+5])
		deposit := &clients.DepositTransfer{
			Nonce:            depositNonce,
			FromBytes:        responseData[i+2],
			DisplayableFrom:  c.addressPublicKeyConverter.ToBech32String(responseData[i+2]),
			ToBytes:          responseData[i+3],
			DisplayableTo:    c.addressPublicKeyConverter.ToHexStringWithPrefix(responseData[i+3]),
			TokenBytes:       responseData[i+4],
			DisplayableToken: string(responseData[i+4]),
			Amount:           amount,
		}

		storedConvertedTokenBytes, exists := cachedTokens[deposit.DisplayableToken]
		if !exists {
			deposit.ConvertedTokenBytes, err = c.tokensMapper.ConvertToken(ctx, deposit.TokenBytes)
			if err != nil {
				return nil, fmt.Errorf("%w while converting token bytes, transfer index %d", err, transferIndex)
			}
			cachedTokens[deposit.DisplayableToken] = deposit.ConvertedTokenBytes
		} else {
			deposit.ConvertedTokenBytes = storedConvertedTokenBytes
		}

		batch.Deposits = append(batch.Deposits, deposit)
		transferIndex++
	}

	batch.Statuses = make([]byte, len(batch.Deposits))

	c.log.Debug("created batch " + batch.String())

	return batch, nil
}

func (c *client) createCommonTxDataBuilder(funcName string, id int64) builders.TxDataBuilder {
	return builders.NewTxDataBuilder().Function(funcName).ArgInt64(id)
}

// ProposeSetStatus will trigger the proposal of the ESDT safe set current transaction batch status operation
func (c *client) ProposeSetStatus(ctx context.Context, batch *clients.TransferBatch) (string, error) {
	if batch == nil {
		return "", clients.ErrNilBatch
	}

	err := c.checkIsPaused(ctx)
	if err != nil {
		return "", err
	}

	txBuilder := c.createCommonTxDataBuilder(proposeSetStatusFuncName, int64(batch.ID))
	for _, stat := range batch.Statuses {
		txBuilder.ArgBytes([]byte{stat})
	}

	gasLimit := c.gasMapConfig.ProposeStatusBase + uint64(len(batch.Deposits))*c.gasMapConfig.ProposeStatusForEach
	hash, err := c.txHandler.SendTransactionReturnHash(ctx, txBuilder, gasLimit)
	if err == nil {
		c.log.Info("proposed set statuses"+batch.String(), "transaction hash", hash)
	}

	return hash, err
}

// ProposeTransfer will trigger the propose transfer operation
func (c *client) ProposeTransfer(ctx context.Context, batch *clients.TransferBatch) (string, error) {
	if batch == nil {
		return "", clients.ErrNilBatch
	}

	err := c.checkIsPaused(ctx)
	if err != nil {
		return "", err
	}

	txBuilder := c.createCommonTxDataBuilder(proposeTransferFuncName, int64(batch.ID))

	for _, dt := range batch.Deposits {
		txBuilder.ArgBytes(dt.FromBytes).
			ArgBytes(dt.ToBytes).
			ArgBytes(dt.ConvertedTokenBytes).
			ArgBigInt(dt.Amount).
			ArgInt64(int64(dt.Nonce))
		if len(dt.Data) > 0 {
			txBuilder.ArgBytes(dt.Data).
				ArgInt64(int64(dt.ExtraGasLimit))
		}
	}

	gasLimit := c.gasMapConfig.ProposeTransferBase + uint64(len(batch.Deposits))*c.gasMapConfig.ProposeTransferForEach
	hash, err := c.txHandler.SendTransactionReturnHash(ctx, txBuilder, gasLimit)
	if err == nil {
		c.log.Info("proposed transfer"+batch.String(), "transaction hash", hash)
	}

	return hash, err
}

// Sign will trigger the execution of a sign operation
func (c *client) Sign(ctx context.Context, actionID uint64) (string, error) {
	err := c.checkIsPaused(ctx)
	if err != nil {
		return "", err
	}

	txBuilder := c.createCommonTxDataBuilder(signFuncName, int64(actionID))

	hash, err := c.txHandler.SendTransactionReturnHash(ctx, txBuilder, c.gasMapConfig.Sign)
	if err == nil {
		c.log.Info("signed", "action ID", actionID, "transaction hash", hash)
	}

	return hash, err
}

// PerformAction will trigger the execution of the provided action ID
func (c *client) PerformAction(ctx context.Context, actionID uint64, batch *clients.TransferBatch) (string, error) {
	if batch == nil {
		return "", clients.ErrNilBatch
	}

	err := c.checkIsPaused(ctx)
	if err != nil {
		return "", err
	}

	txBuilder := c.createCommonTxDataBuilder(performActionFuncName, int64(actionID))

	gasLimit := c.gasMapConfig.PerformActionBase + uint64(len(batch.Statuses))*c.gasMapConfig.PerformActionForEach
	hash, err := c.txHandler.SendTransactionReturnHash(ctx, txBuilder, gasLimit)

	if err == nil {
		c.log.Info("performed action", "actionID", actionID, "transaction hash", hash)
	}

	return hash, err
}

func (c *client) checkIsPaused(ctx context.Context) error {
	isPaused, err := c.IsPaused(ctx)
	if err != nil {
		return fmt.Errorf("%w in client.ExecuteTransfer", err)
	}
	if isPaused {
		return fmt.Errorf("%w in client.ExecuteTransfer", clients.ErrMultisigContractPaused)
	}

	return nil
}

// CheckClientAvailability will check the client availability and will set the metric accordingly
func (c *client) CheckClientAvailability(ctx context.Context) error {
	c.mut.Lock()
	defer c.mut.Unlock()

	currentNonce, err := c.GetCurrentNonce(ctx)
	if err != nil {
		c.setStatusForAvailabilityCheck(ethmultiversx.Unavailable, err.Error(), currentNonce)

		return err
	}

	if currentNonce != c.lastNonce {
		c.retriesAvailabilityCheck = 0
		c.lastNonce = currentNonce
	}

	// if we reached this point we will need to increment the retries counter
	defer c.incrementRetriesAvailabilityCheck()

	if c.retriesAvailabilityCheck > c.allowDelta {
		message := fmt.Sprintf("nonce %d fetched for %d times in a row", currentNonce, c.retriesAvailabilityCheck)
		c.setStatusForAvailabilityCheck(ethmultiversx.Unavailable, message, currentNonce)

		return nil
	}

	c.setStatusForAvailabilityCheck(ethmultiversx.Available, "", currentNonce)

	return nil
}

func (c *client) incrementRetriesAvailabilityCheck() {
	c.retriesAvailabilityCheck++
}

func (c *client) setStatusForAvailabilityCheck(status ethmultiversx.ClientStatus, message string, nonce uint64) {
	c.statusHandler.SetStringMetric(bridgeCore.MetricMultiversXClientStatus, status.String())
	c.statusHandler.SetStringMetric(bridgeCore.MetricLastMultiversXClientError, message)
	c.statusHandler.SetIntMetric(bridgeCore.MetricLastBlockNonce, int(nonce))
}

// Close will close any started go routines. It returns nil.
func (c *client) Close() error {
	return c.txHandler.Close()
}

// IsInterfaceNil returns true if there is no value under the interface
func (c *client) IsInterfaceNil() bool {
	return c == nil
}

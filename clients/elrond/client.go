package elrond

import (
	"context"
	"fmt"
	"math/big"
	"reflect"
	"time"

	"github.com/ElrondNetwork/elrond-eth-bridge/clients"
	"github.com/ElrondNetwork/elrond-eth-bridge/config"
	bridgeCore "github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ElrondNetwork/elrond-eth-bridge/core/converters"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	"github.com/ElrondNetwork/elrond-go-crypto/signing/ed25519/singlesig"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/builders"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/core"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/data"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/interactors"
)

const (
	proposeTransferFuncName  = "proposeMultiTransferEsdtBatch"
	proposeSetStatusFuncName = "proposeEsdtSafeSetCurrentTransactionBatchStatus"
	signFuncName             = "sign"
	performActionFuncName    = "performAction"
	minRetriesOnQuorum       = 1
)

// ClientArgs represents the argument for the NewClient constructor function
type ClientArgs struct {
	GasMapConfig                 config.ElrondGasMapConfig
	Proxy                        ElrondProxy
	Log                          logger.Logger
	RelayerPrivateKey            crypto.PrivateKey
	MultisigContractAddress      core.AddressHandler
	IntervalToResendTxsInSeconds uint64
	TokensMapper                 TokensMapper
	MaxRetriesOnQuorumReached    uint64
}

// client represents the Elrond Client implementation
type client struct {
	*elrondClientDataGetter
	txHandler                 txHandler
	tokensMapper              TokensMapper
	relayerPublicKey          crypto.PublicKey
	relayerAddress            core.AddressHandler
	multisigContractAddress   core.AddressHandler
	log                       logger.Logger
	gasMapConfig              config.ElrondGasMapConfig
	addressPublicKeyConverter bridgeCore.AddressConverter
	maxRetriesOnQuorumReached uint64
}

// NewClient returns a new Elrond Client instance
func NewClient(args ClientArgs) (*client, error) {
	if check.IfNil(args.Proxy) {
		return nil, errNilProxy
	}
	if check.IfNil(args.RelayerPrivateKey) {
		return nil, errNilPrivateKey
	}
	if check.IfNil(args.MultisigContractAddress) {
		return nil, fmt.Errorf("%w for the MultisigContractAddress argument", errNilAddressHandler)
	}
	if check.IfNil(args.Log) {
		return nil, errNilLogger
	}
	if check.IfNil(args.TokensMapper) {
		return nil, errNilTokensMapper
	}
	if args.MaxRetriesOnQuorumReached < minRetriesOnQuorum {
		return nil, fmt.Errorf("%w for args.MaxRetriesOnQuorumReached, got: %d, minimum: %d",
			errInvalidValue, args.MaxRetriesOnQuorumReached, minRetriesOnQuorum)
	}

	err := checkGasMapValues(args.GasMapConfig)
	if err != nil {
		return nil, err
	}

	nonceTxsHandler, err := interactors.NewNonceTransactionHandler(args.Proxy, time.Second*time.Duration(args.IntervalToResendTxsInSeconds))
	if err != nil {
		return nil, err
	}

	publicKey := args.RelayerPrivateKey.GeneratePublic()
	publicKeyBytes, err := publicKey.ToByteArray()
	if err != nil {
		return nil, err
	}

	relayerAddress := data.NewAddressFromBytes(publicKeyBytes)

	argsDataGetter := ArgsDataGetter{
		MultisigContractAddress: args.MultisigContractAddress,
		RelayerAddress:          relayerAddress,
		Proxy:                   args.Proxy,
	}
	getter, err := NewDataGetter(argsDataGetter)
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
		},
		elrondClientDataGetter:    getter,
		relayerPublicKey:          publicKey,
		relayerAddress:            relayerAddress,
		multisigContractAddress:   args.MultisigContractAddress,
		log:                       args.Log,
		gasMapConfig:              args.GasMapConfig,
		addressPublicKeyConverter: addressConverter,
		tokensMapper:              args.TokensMapper,
		maxRetriesOnQuorumReached: args.MaxRetriesOnQuorumReached,
	}

	c.log.Info("NewElrondClient",
		"relayer address", relayerAddress.AddressAsBech32String(),
		"safe contract address", args.MultisigContractAddress.AddressAsBech32String())

	return c, nil
}

func checkGasMapValues(gasMap config.ElrondGasMapConfig) error {
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

		deposit.ConvertedTokenBytes, err = c.tokensMapper.ConvertToken(ctx, deposit.TokenBytes)
		if err != nil {
			return nil, fmt.Errorf("%w while converting token bytes, transfer index %d", err, transferIndex)
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
		return "", errNilBatch
	}

	txBuilder := c.createCommonTxDataBuilder(proposeSetStatusFuncName, int64(batch.ID))
	for _, stat := range batch.Statuses {
		txBuilder.ArgBytes([]byte{stat})
	}

	hash, err := c.txHandler.SendTransactionReturnHash(ctx, txBuilder, c.gasMapConfig.ProposeStatus)
	if err == nil {
		c.log.Info("proposed set statuses"+batch.String(), "transaction hash", hash)
	}

	return hash, err
}

// ResolveNewDeposits will try to add new statuses if the pending batch gets modified
func (c *client) ResolveNewDeposits(ctx context.Context, batch *clients.TransferBatch) error {
	if batch == nil {
		return errNilBatch
	}

	newBatch, err := c.GetPending(ctx)
	if err != nil {
		return fmt.Errorf("%w while getting new batch in ResolveNewDeposits method", err)
	}

	batch.ResolveNewDeposits(len(newBatch.Statuses))

	return nil
}

// ProposeTransfer will trigger the propose transfer operation
func (c *client) ProposeTransfer(ctx context.Context, batch *clients.TransferBatch) (string, error) {
	if batch == nil {
		return "", errNilBatch
	}

	txBuilder := c.createCommonTxDataBuilder(proposeTransferFuncName, int64(batch.ID))

	for _, dt := range batch.Deposits {
		txBuilder.ArgBytes(dt.FromBytes).
			ArgBytes(dt.ToBytes).
			ArgBytes(dt.ConvertedTokenBytes).
			ArgBigInt(dt.Amount).
			ArgInt64(int64(dt.Nonce))
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
		return "", errNilBatch
	}

	txBuilder := c.createCommonTxDataBuilder(performActionFuncName, int64(actionID))

	gasLimit := c.gasMapConfig.PerformActionBase + uint64(len(batch.Statuses))*c.gasMapConfig.PerformActionForEach
	hash, err := c.txHandler.SendTransactionReturnHash(ctx, txBuilder, gasLimit)

	if err == nil {
		c.log.Info("performed action", "actionID", actionID, "transaction hash", hash)
	}

	return hash, err
}

// GetMaxNumberOfRetriesOnQuorumReached returns the maximum number of retries allowed on quorum reached
func (c *client) GetMaxNumberOfRetriesOnQuorumReached() uint64 {
	return c.maxRetriesOnQuorumReached
}

// Close will close any started go routines. It returns nil.
func (c *client) Close() error {
	return c.txHandler.Close()
}

// IsInterfaceNil returns true if there is no value under the interface
func (c *client) IsInterfaceNil() bool {
	return c == nil
}

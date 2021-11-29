package elrond

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"reflect"
	"time"

	"github.com/ElrondNetwork/elrond-eth-bridge/clients"
	"github.com/ElrondNetwork/elrond-eth-bridge/config"
	elrondCore "github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/core/pubkeyConverter"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/core"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/data"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/interactors"
)

const (
	hexPrefix = "0x"
)

// ClientArgs represents the argument for the NewClient constructor function
type ClientArgs struct {
	GasMapConfig                 config.ElrondGasMapConfig
	Proxy                        ElrondProxy
	Log                          logger.Logger
	RelayerPrivateKey            crypto.PrivateKey
	MultisigContractAddress      core.AddressHandler
	IntervalToResendTxsInSeconds uint64
}

// client represents the Elrond Client implementation
type client struct {
	dataGetter                *dataGetter
	proxy                     ElrondProxy
	relayerPrivateKey         crypto.PrivateKey
	relayerPublicKey          crypto.PublicKey
	relayerAddress            core.AddressHandler
	multisigContractAddress   core.AddressHandler
	nonceTxHandler            NonceTransactionsHandler
	log                       logger.Logger
	gasMapConfig              config.ElrondGasMapConfig
	addressPublicKeyConverter elrondCore.PubkeyConverter
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

	addressPubKeyConverter, err := pubkeyConverter.NewBech32PubkeyConverter(core.AddressBytesLen, args.Log)
	if err != nil {
		return nil, err
	}

	c := &client{
		dataGetter:                getter,
		relayerPrivateKey:         args.RelayerPrivateKey,
		relayerPublicKey:          publicKey,
		relayerAddress:            relayerAddress,
		proxy:                     args.Proxy,
		multisigContractAddress:   args.MultisigContractAddress,
		log:                       args.Log,
		nonceTxHandler:            nonceTxsHandler,
		gasMapConfig:              args.GasMapConfig,
		addressPublicKeyConverter: addressPubKeyConverter,
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
	responseData, err := c.dataGetter.GetCurrentBatchAsDataBytes(ctx)
	if err != nil {
		return nil, err
	}

	if emptyResponse(responseData) {
		return nil, ErrNoPendingBatchAvailable
	}

	return c.createPendingBatchFromResponse(responseData)
}

func emptyResponse(response [][]byte) bool {
	return len(response) == 0 || (len(response) == 1 && len(response[0]) == 0)
}

func (c *client) createPendingBatchFromResponse(responseData [][]byte) (*clients.TransferBatch, error) {
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
			DisplayableFrom:  c.addressPublicKeyConverter.Encode(responseData[i+2]),
			ToBytes:          responseData[i+3],
			DisplayableTo:    fmt.Sprintf("%s%s", hexPrefix, hex.EncodeToString(responseData[i+3])),
			TokenBytes:       responseData[i+4],
			DisplayableToken: c.addressPublicKeyConverter.Encode(responseData[i+4]),
			Amount:           amount,
		}

		batch.Deposits = append(batch.Deposits, deposit)
		transferIndex++
	}

	c.log.Debug("created batch " + batch.String())

	return batch, nil
}

// TODO(next PR): remove comment and implement the rest of the needed functionality
//// ProposeSetStatus will trigger the proposal of the ESDT safe set current transaction batch status operation
//func (c *client) ProposeSetStatus(ctx context.Context, batch *bridge.Batch) {
//	builder := newBuilder(c.log).
//		Func("proposeEsdtSafeSetCurrentTransactionBatchStatus").
//		BatchId(batch.Id)
//
//	newBatch, err := c.GetPending(ctx)
//	if err != nil {
//		c.log.Error("Elrond: get pending batch failed in ProposeSetStatus", "error", err)
//		return
//	}
//
//	batch.ResolveNewDeposits(len(newBatch.Statuses))
//
//	for _, stat := range batch.Statuses {
//		builder = builder.Int(big.NewInt(int64(stat)))
//	}
//
//	hash, err := c.sendTransaction(builder, c.gasMapConfig.ProposeStatus)
//	if err != nil {
//		c.log.Error("Elrond: send transaction failed", "error", err)
//		return
//	}
//
//	c.log.Info("Elrond: Proposed status update", "hash", hash)
//}
//
//// ProposeTransfer will trigger the propose transfer operation
//func (c *client) ProposeTransfer(_ context.Context, batch *bridge.Batch) (string, error) {
//	builder := newBuilder(c.log).
//		Func("proposeMultiTransferEsdtBatch").
//		BatchId(batch.Id)
//
//	for _, tx := range batch.Transactions {
//		builder = builder.
//			Address([]byte(tx.To)).
//			HexString(c.GetTokenId(tx.TokenAddress)).
//			BigInt(tx.Amount)
//	}
//
//	batchData, errMarshal := json.Marshal(batch)
//	if errMarshal != nil {
//		c.log.Warn("Elrond: error not critical while serializing transaction", "error", errMarshal)
//	}
//
//	gasLimit := c.gasMapConfig.ProposeTransferBase + uint64(len(batch.Transactions))*c.gasMapConfig.ProposeTransferForEach
//	hash, err := c.sendTransaction(builder, gasLimit)
//	if err == nil {
//		c.log.Info("Elrond: Proposed transfer for batch ", batch.Id, "with hash", hash, "batch data", string(batchData))
//	} else {
//		c.log.Error("Elrond: Propose transfer errored", "batch data", string(batchData), "error", err)
//	}
//
//	return hash, err
//}
//
//// WasProposedTransfer returns true if the transfer action proposed was triggered
//func (c *client) WasProposedTransfer(_ context.Context, batch *bridge.Batch) bool {
//	valueRequest := newValueBuilder(c.bridgeAddress, c.address.AddressAsBech32String(), c.log).
//		Func("wasTransferActionProposed").
//		BatchId(batch.Id).
//		WithTx(batch, c.GetTokenId).
//		Build()
//
//	return c.executeBoolQuery(valueRequest)
//}
//
//// GetActionIdForProposeTransfer returns the action ID for the propose transfer operation
//func (c *client) GetActionIdForProposeTransfer(_ context.Context, batch *bridge.Batch) bridge.ActionId {
//	valueRequest := newValueBuilder(c.bridgeAddress, c.address.AddressAsBech32String(), c.log).
//		Func("getActionIdForTransferBatch").
//		BatchId(batch.Id).
//		WithTx(batch, c.GetTokenId).
//		Build()
//
//	response, err := c.executeUintQuery(valueRequest)
//	if err != nil {
//		c.log.Error(err.Error())
//		return bridge.NewActionId(0)
//	}
//
//	actionId := bridge.NewActionId(int64(response))
//
//	c.log.Info("Elrond: fetched actionId for transfer batch", "actionId", actionId, "batch", batch.Id)
//
//	return actionId
//}
//
//// WasProposedSetStatus returns true if the proposed set status was triggered
//func (c *client) WasProposedSetStatus(ctx context.Context, batch *bridge.Batch) bool {
//	valueRequest := newValueBuilder(c.bridgeAddress, c.address.AddressAsBech32String(), c.log).
//		Func("wasSetCurrentTransactionBatchStatusActionProposed").
//		BatchId(batch.Id)
//
//	newBatch, err := c.GetPending(ctx)
//	if err != nil {
//		c.log.Error("Elrond: get pending batch failed in WasProposedSetStatus", "error", err)
//		return false
//	}
//	batch.ResolveNewDeposits(len(newBatch.Statuses))
//
//	for _, stat := range batch.Statuses {
//		valueRequest = valueRequest.BigInt(big.NewInt(int64(stat)))
//	}
//
//	return c.executeBoolQuery(valueRequest.Build())
//}
//
//// GetTransactionsStatuses will return the transactions statuses from the batch ID
//func (c *client) GetTransactionsStatuses(_ context.Context, batchId bridge.BatchId) ([]uint8, error) {
//	if batchId == nil {
//		return nil, ErrNilBatchId
//	}
//
//	valueRequest := newValueBuilder(c.bridgeAddress, c.address.AddressAsBech32String(), c.log).
//		Func("getStatusesAfterExecution").
//		BatchId(batchId)
//
//	values, err := c.executeQuery(valueRequest.Build())
//	if err != nil {
//		return nil, err
//	}
//
//	if len(values) == 0 {
//		return nil, fmt.Errorf("%w for batch ID %v", ErrNoStatusForBatchID, batchId)
//	}
//	isFinished := c.convertToBool(values[0])
//	if !isFinished {
//		return nil, fmt.Errorf("%w for batch ID %v", ErrBatchNotFinished, batchId)
//	}
//
//	results := make([]byte, len(values)-1)
//	for i := 1; i < len(values); i++ {
//		results[i-1], err = getStatusFromBuff(values[i])
//		if err != nil {
//			return nil, fmt.Errorf("%w for result index %d", err, i)
//		}
//	}
//
//	if len(results) == 0 {
//		return nil, fmt.Errorf("%w status is finished, no results are given", ErrMalformedBatchResponse)
//	}
//
//	c.log.Debug("Elrond: got transaction status", "batchID", batchId, "status", results)
//
//	return results, nil
//}
//
//func getStatusFromBuff(buff []byte) (byte, error) {
//	if len(buff) == 0 {
//		return 0, ErrMalformedBatchResponse
//	}
//
//	return buff[len(buff)-1], nil
//}
//
//// GetActionIdForSetStatusOnPendingTransfer returns the action ID for setting the status on the pending transfer batch
//func (c *client) GetActionIdForSetStatusOnPendingTransfer(ctx context.Context, batch *bridge.Batch) bridge.ActionId {
//	valueRequest := newValueBuilder(c.bridgeAddress, c.address.AddressAsBech32String(), c.log).
//		Func("getActionIdForSetCurrentTransactionBatchStatus").
//		BatchId(batch.Id)
//
//	newBatch, err := c.GetPending(ctx)
//	if err != nil {
//		c.log.Error("Elrond: get pending batch failed in WasProposedSetStatus", "error", err)
//		return bridge.NewActionId(0)
//	}
//	batch.ResolveNewDeposits(len(newBatch.Statuses))
//
//	for _, stat := range batch.Statuses {
//		valueRequest = valueRequest.BigInt(big.NewInt(int64(stat)))
//	}
//
//	response, err := c.executeUintQuery(valueRequest.Build())
//	if err != nil {
//		c.log.Error(err.Error())
//		return bridge.NewActionId(0)
//	}
//
//	c.log.Debug("Elrond: got actionID", "actionID", response)
//
//	return bridge.NewActionId(int64(response))
//}
//
//// WasExecuted returns true if the provided actionId was executed or not
//func (c *client) WasExecuted(_ context.Context, actionId bridge.ActionId, _ bridge.BatchId) bool {
//	valueRequest := newValueBuilder(c.bridgeAddress, c.address.AddressAsBech32String(), c.log).
//		Func("wasActionExecuted").
//		ActionId(actionId).
//		Build()
//
//	result := c.executeBoolQuery(valueRequest)
//
//	if result {
//		c.log.Info(fmt.Sprintf("Elrond: ActionId %v was executed", actionId))
//	}
//
//	return result
//}
//
//// Sign will trigger the execution of a sign operation
//func (c *client) Sign(_ context.Context, actionId bridge.ActionId, batch *bridge.Batch) (string, error) {
//	builder := newBuilder(c.log).
//		Func("sign").
//		ActionId(actionId)
//
//	hash, err := c.sendTransaction(builder, c.gasMapConfig.Sign)
//
//	batchData, err := json.Marshal(batch)
//	if err != nil {
//		c.log.Warn("Elrond: error not critical while serializing transaction", "error", err)
//	}
//
//	if err == nil {
//		c.log.Info("Elrond: Signed", "hash", hash, "batch data", string(batchData))
//	} else {
//		c.log.Error("Elrond: Sign failed", "batch data", string(batchData), "error", err)
//	}
//
//	return hash, err
//}
//
//// Execute will trigger the execution of the provided action ID
//func (c *client) Execute(_ context.Context, actionId bridge.ActionId, batch *bridge.Batch, _ bridge.SignaturesHolder) (string, error) {
//	builder := newBuilder(c.log).
//		Func("performAction").
//		ActionId(actionId)
//
//	gasLimit := c.gasMapConfig.PerformActionBase + uint64(len(batch.Transactions))*c.gasMapConfig.PerformActionForEach
//	hash, err := c.sendTransaction(builder, gasLimit)
//
//	batchData, err := json.Marshal(batch)
//	if err != nil {
//		c.log.Warn("Elrond: error not critical while serializing transaction", "error", err)
//	}
//
//	if err == nil {
//		c.log.Info("Elrond: Executed action", "actionID", actionId, "batch data", string(batchData), "hash", hash)
//	} else {
//		c.log.Info("Elrond: Execution failed for action",
//			"actionID", actionId,
//			"batch data", string(batchData),
//			"hash", hash,
//			"error", err)
//	}
//
//	return hash, err
//}
//
//// SignersCount returns the signers count
//func (c *client) SignersCount(_ context.Context, _ *bridge.Batch, actionId bridge.ActionId, _ bridge.SignaturesHolder) uint {
//	valueRequest := newValueBuilder(c.bridgeAddress, c.address.AddressAsBech32String(), c.log).
//		Func("getActionSignerCount").
//		ActionId(actionId).
//		Build()
//
//	count, _ := c.executeUintQuery(valueRequest)
//	return uint(count)
//}
//
//// GetTokenId returns the token ID for the erc 20 address
//func (c *client) GetTokenId(address string) string {
//	if strings.Index(address, hexPrefix) == 0 {
//		address = address[len(hexPrefix):]
//	}
//
//	valueRequest := newValueBuilder(c.bridgeAddress, c.address.AddressAsBech32String(), c.log).
//		Func("getTokenIdForErc20Address").
//		HexString(address).
//		Build()
//
//	tokenId, err := c.executeStringQuery(valueRequest)
//	if err != nil {
//		c.log.Error(err.Error())
//	}
//
//	c.log.Debug("Elrond: get token ID", "address", address, "tokenID", tokenId)
//
//	return tokenId
//}
//
//// GetErc20Address returns the corresponding ERC20 address
//func (c *client) GetErc20Address(tokenId string) string {
//	valueRequest := newValueBuilder(c.bridgeAddress, c.address.AddressAsBech32String(), c.log).
//		Func("getErc20AddressForTokenId").
//		HexString(tokenId).
//		Build()
//
//	address, err := c.executeStringQuery(valueRequest)
//	if err != nil {
//		c.log.Error(err.Error())
//	}
//
//	c.log.Debug("Elrond: get erc20 address", "tokenID", tokenId, "address", address)
//
//	return address
//}
//
//// GetHexWalletAddress returns the wallet address as a hex string
//func (c *client) GetHexWalletAddress() string {
//	return hex.EncodeToString(c.address.AddressBytes())
//}

// Close will close any started go routines. It returns nil.
func (c *client) Close() error {
	return c.nonceTxHandler.Close()
}

// IsInterfaceNil returns true if there is no value under the interface
func (c *client) IsInterfaceNil() bool {
	return c == nil
}

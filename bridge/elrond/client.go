package elrond

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridge"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/core/pubkeyConverter"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	"github.com/ElrondNetwork/elrond-go-crypto/signing/ed25519/singlesig"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/builders"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/core"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/data"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/interactors"
)

const (
	hexPrefix = "0x"
)

var txSingleSigner = &singlesig.Ed25519Signer{}

// QueryResponseErr represents the query response error DTO struct
type QueryResponseErr struct {
	code      string
	message   string
	function  string
	arguments []string
	address   string
}

func (e QueryResponseErr) Error() string {
	return fmt.Sprintf("got response code %q and message %q while querying function %q with arguments %v "+
		"and address %v", e.code, e.message, e.function, e.arguments, e.address)
}

// client represents the Elrond Client implementation
type client struct {
	proxy          bridge.ElrondProxy
	bridgeAddress  core.AddressHandler
	privateKey     crypto.PrivateKey
	address        core.AddressHandler
	nonceTxHandler NonceTransactionsHandler
	log            logger.Logger
	gasMapConfig   bridge.ElrondGasMapConfig
}

// ClientArgs represents the argument for the NewClient constructor function
type ClientArgs struct {
	Config     bridge.ElrondConfig
	Proxy      bridge.ElrondProxy
	PrivateKey crypto.PrivateKey
	Address    core.AddressHandler
}

// NewClient returns a new Elrond Client instance
func NewClient(args ClientArgs) (*client, error) {
	if check.IfNil(args.Proxy) {
		return nil, ErrNilProxy
	}
	if check.IfNil(args.PrivateKey) {
		return nil, ErrNilPrivateKey
	}
	if check.IfNil(args.Address) {
		return nil, ErrNilAddressHandler
	}
	err := checkGasMapValues(args.Config.GasMap)
	if err != nil {
		return nil, err
	}

	// TODO inject this
	nonceTxsHandler, err := interactors.NewNonceTransactionHandler(args.Proxy, time.Second*time.Duration(args.Config.IntervalToResendTxsInSeconds))
	if err != nil {
		return nil, err
	}

	bridgeAddress, err := data.NewAddressFromBech32String(args.Config.BridgeAddress)
	if err != nil {
		return nil, fmt.Errorf("%w for args.Config.BridgeAddress", err)
	}

	c := &client{
		proxy:          args.Proxy,
		bridgeAddress:  bridgeAddress,
		privateKey:     args.PrivateKey,
		address:        args.Address,
		log:            logger.GetOrCreate("ElrondClient"),
		nonceTxHandler: nonceTxsHandler,
		gasMapConfig:   args.Config.GasMap,
	}

	c.log.Info("Elrond: NewClient", "address", c.address.AddressAsBech32String())

	return c, nil
}

func checkGasMapValues(gasMap bridge.ElrondGasMapConfig) error {
	gasMapValue := reflect.ValueOf(gasMap)
	typeOfGasMapValue := gasMapValue.Type()

	for i := 0; i < gasMapValue.NumField(); i++ {
		fieldVal := gasMapValue.Field(i).Uint()
		if fieldVal == 0 {
			return fmt.Errorf("%w for field %s", ErrInvalidGasValue, typeOfGasMapValue.Field(i).Name)
		}
	}

	return nil
}

// GetPending returns the pending batch
func (c *client) GetPending(_ context.Context) *bridge.Batch {
	c.log.Info("Elrond: Getting pending batch")
	responseData, err := c.getCurrentBatch()
	if err != nil {
		c.log.Error("Elrond: Failed to get the current batch", "error", err.Error())
		return nil
	}

	if emptyResponse(responseData) {
		return nil
	}

	addrPkConv, _ := pubkeyConverter.NewBech32PubkeyConverter(32, c.log)
	numArgs := 6
	idxAmount := 5
	var transactions []*bridge.DepositTransaction
	for i := 1; i < len(responseData); i += numArgs {
		if len(responseData) < i+idxAmount {
			c.log.Warn("Elrond: got an unexpected number of arguments", "index", i, "total args", len(responseData))
			break
		}

		amount := new(big.Int).SetBytes(responseData[i+idxAmount])
		blockNonce, errParse := parseIntFromByteSlice(responseData[i])
		if errParse != nil {
			c.log.Error("Elrond: parse error", "error", errParse.Error())
			return nil
		}
		depositNonce, errParse := parseIntFromByteSlice(responseData[i+1])
		if errParse != nil {
			c.log.Error("Elrond: parse error", "error", errParse.Error())
			return nil
		}

		to := fmt.Sprintf("0x%s", hex.EncodeToString(responseData[i+3]))
		tx := &bridge.DepositTransaction{
			From:          addrPkConv.Encode(responseData[i+2]),
			To:            to,
			DisplayableTo: to,
			TokenAddress:  hex.EncodeToString(responseData[i+4]),
			Amount:        amount,
			DepositNonce:  bridge.Nonce(depositNonce),
			BlockNonce:    bridge.Nonce(blockNonce),
			Status:        0,
			Error:         nil,
		}
		c.log.Trace("created deposit transaction: " + tx.String())
		transactions = append(transactions, tx)
	}

	batchID, err := parseIntFromByteSlice(responseData[0])
	if err != nil {
		c.log.Error("Elrond: parse error", "error", err.Error())
		return nil
	}

	return &bridge.Batch{
		ID:           bridge.BatchID(batchID),
		Transactions: transactions,
	}
}

func parseIntFromByteSlice(buff []byte) (int64, error) {
	if len(buff) == 0 {
		return 0, nil
	}

	val, err := strconv.ParseInt(hex.EncodeToString(buff), 16, 64)
	if err != nil {
		return 0, err
	}

	return val, nil
}

// ProposeSetStatus will trigger the proposal of the ESDT safe set current transaction batch status operation
func (c *client) ProposeSetStatus(_ context.Context, batch *bridge.Batch) {
	builder := builders.NewTxDataBuilder(c.log).
		Function("proposeEsdtSafeSetCurrentTransactionBatchStatus").
		ArgInt64(batch.ID.Int64())

	for _, tx := range batch.Transactions {
		builder = builder.ArgBigInt(big.NewInt(int64(tx.Status)))
	}

	hash, err := c.sendTransaction(builder, c.gasMapConfig.ProposeStatus)
	if err != nil {
		c.log.Error("Elrond: send transaction failed", "error", err.Error())
		return
	}

	c.log.Info("Elrond: Proposed status update", "hash", hash)
}

// ProposeTransfer will trigger the propose transfer operation
func (c *client) ProposeTransfer(_ context.Context, batch *bridge.Batch) (string, error) {
	builder := builders.NewTxDataBuilder(c.log).
		Function("proposeMultiTransferEsdtBatch").
		ArgInt64(batch.ID.Int64())

	for _, tx := range batch.Transactions {
		builder = builder.
			ArgAddress(data.NewAddressFromBytes([]byte(tx.To))).
			ArgHexString(c.GetTokenId(tx.TokenAddress)).
			ArgBigInt(tx.Amount)
	}

	gasLimit := c.gasMapConfig.ProposeTransferBase + uint64(len(batch.Transactions))*c.gasMapConfig.ProposeTransferForEach
	hash, err := c.sendTransaction(builder, gasLimit)

	if err == nil {
		c.log.Info("Elrond: Proposed transfer for batchID ", batch.ID, " with hash ", hash)
	} else {
		c.log.Error("Elrond: Propose transfer errored", "error", err.Error())
	}

	return hash, err
}

// WasProposedTransfer returns true if the transfer action proposed was triggered
func (c *client) WasProposedTransfer(_ context.Context, batch *bridge.Batch) bool {
	builder := builders.NewTxDataBuilder(c.log).
		Address(c.bridgeAddress).
		CallerAddress(c.address).
		Function("wasTransferActionProposed").
		ArgInt64(batch.ID.Int64())

	withTx(builder, batch, c.GetTokenId)

	valueRequest, err := builder.ToVmValueRequest()
	if err != nil {
		c.log.Error("error in client.WasProposedTransfer builder error", "error", err)
		return false
	}

	return c.executeBoolQuery(valueRequest)
}

// GetActionIdForProposeTransfer returns the action ID for the propose transfer operation
func (c *client) GetActionIdForProposeTransfer(_ context.Context, batch *bridge.Batch) bridge.ActionID {
	builder := builders.NewTxDataBuilder(c.log).
		Address(c.bridgeAddress).
		CallerAddress(c.address).
		Function("getActionIdForTransferBatch").
		ArgInt64(batch.ID.Int64())

	withTx(builder, batch, c.GetTokenId)

	valueRequest, err := builder.ToVmValueRequest()
	if err != nil {
		c.log.Error("error in client.GetActionIdForProposeTransfer builder error", "error", err)
		return 0
	}

	response, err := c.executeUintQuery(valueRequest)
	if err != nil {
		c.log.Error(err.Error())
		return 0
	}

	actionID := bridge.ActionID(response)

	c.log.Info("Elrond: fetched actionId for transfer batch", "actionID", actionID, "batchID", batch.ID)

	return actionID
}

// WasProposedSetStatus returns true if the proposed set status was triggered
func (c *client) WasProposedSetStatus(_ context.Context, batch *bridge.Batch) bool {
	builder := builders.NewTxDataBuilder(c.log).
		Address(c.bridgeAddress).
		CallerAddress(c.address).
		Function("wasSetCurrentTransactionBatchStatusActionProposed").
		ArgInt64(batch.ID.Int64())

	for _, tx := range batch.Transactions {
		builder = builder.ArgBigInt(big.NewInt(int64(tx.Status)))
	}

	valueRequest, err := builder.ToVmValueRequest()
	if err != nil {
		c.log.Error("error in client.WasProposedSetStatus builder error", "error", err)
		return false
	}

	return c.executeBoolQuery(valueRequest)
}

// GetTransactionsStatuses will return the transactions statuses from the batch ID
func (c *client) GetTransactionsStatuses(_ context.Context, batchID bridge.BatchID) ([]uint8, error) {
	builder := builders.NewTxDataBuilder(c.log).
		Address(c.bridgeAddress).
		CallerAddress(c.address).
		Function("getStatusesAfterExecution").
		ArgInt64(batchID.Int64())

	valueRequest, err := builder.ToVmValueRequest()
	if err != nil {
		return nil, err
	}

	values, err := c.executeQuery(valueRequest)
	if err != nil {
		return nil, err
	}

	if len(values) == 0 {
		return nil, fmt.Errorf("%w for batchID %v", ErrNoStatusForBatchID, batchID)
	}
	isFinished := c.convertToBool(values[0])
	if !isFinished {
		return nil, fmt.Errorf("%w for batchID %v", ErrBatchNotFinished, batchID)
	}

	results := make([]byte, len(values)-1)
	for i := 1; i < len(values); i++ {
		results[i-1], err = getStatusFromBuff(values[i])
		if err != nil {
			return nil, fmt.Errorf("%w for result index %d", err, i)
		}
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("%w status is finished, no results are given", ErrMalformedBatchResponse)
	}

	return results, nil
}

func getStatusFromBuff(buff []byte) (byte, error) {
	if len(buff) == 0 {
		return 0, ErrMalformedBatchResponse
	}

	return buff[len(buff)-1], nil
}

// GetActionIdForSetStatusOnPendingTransfer returns the action ID for setting the status on the pending transfer batch
func (c *client) GetActionIdForSetStatusOnPendingTransfer(_ context.Context, batch *bridge.Batch) bridge.ActionID {
	builder := builders.NewTxDataBuilder(c.log).
		Address(c.bridgeAddress).
		CallerAddress(c.address).
		Function("getActionIdForSetCurrentTransactionBatchStatus").
		ArgInt64(batch.ID.Int64())

	for _, tx := range batch.Transactions {
		builder = builder.ArgBigInt(big.NewInt(int64(tx.Status)))
	}

	valueRequest, err := builder.ToVmValueRequest()
	if err != nil {
		c.log.Error("error in client.GetActionIdForSetStatusOnPendingTransfer builder error", "error", err)
		return 0
	}

	response, err := c.executeUintQuery(valueRequest)
	if err != nil {
		c.log.Error(err.Error())
		return 0
	}

	return bridge.ActionID(response)
}

// WasExecuted returns true if the provided actionId was executed or not
func (c *client) WasExecuted(_ context.Context, actionID bridge.ActionID, _ bridge.BatchID) bool {
	builder := builders.NewTxDataBuilder(c.log).
		Address(c.bridgeAddress).
		CallerAddress(c.address).
		Function("wasActionExecuted").
		ArgInt64(actionID.Int64())

	valueRequest, err := builder.ToVmValueRequest()
	if err != nil {
		c.log.Error("error in client.WasExecuted builder error", "error", err)
		return false
	}

	result := c.executeBoolQuery(valueRequest)

	if result {
		c.log.Info(fmt.Sprintf("Elrond: ActionID %v was executed", actionID))
	}

	return result
}

// Sign will trigger the execution of a sign operation
func (c *client) Sign(_ context.Context, actionID bridge.ActionID, _ *bridge.Batch) (string, error) {
	builder := builders.NewTxDataBuilder(c.log).
		Function("sign").
		ArgInt64(actionID.Int64())

	hash, err := c.sendTransaction(builder, c.gasMapConfig.Sign)

	if err == nil {
		c.log.Info("Elrond: Signed", "hash", hash)
	} else {
		c.log.Error("Elrond: Sign failed", "error", err.Error())
	}

	return hash, err
}

// Execute will trigger the execution of the provided action ID
func (c *client) Execute(_ context.Context, actionID bridge.ActionID, batch *bridge.Batch, _ bridge.SignaturesHolder) (string, error) {
	builder := builders.NewTxDataBuilder(c.log).
		Function("performAction").
		ArgInt64(actionID.Int64())

	gasLimit := c.gasMapConfig.PerformActionBase + uint64(len(batch.Transactions))*c.gasMapConfig.PerformActionForEach
	hash, err := c.sendTransaction(builder, gasLimit)

	if err == nil {
		c.log.Info("Elrond: Executed action", "actionID", actionID, "hash", hash)
	} else {
		c.log.Info("Elrond: Execution failed for action", "actionID", actionID, "hash", hash, "error", err.Error())
	}

	return hash, err
}

// SignersCount returns the signers count
func (c *client) SignersCount(_ *bridge.Batch, actionID bridge.ActionID, _ bridge.SignaturesHolder) uint {
	builder := builders.NewTxDataBuilder(c.log).
		Address(c.bridgeAddress).
		CallerAddress(c.address).
		Function("getActionSignerCount").
		ArgInt64(actionID.Int64())

	valueRequest, err := builder.ToVmValueRequest()
	if err != nil {
		c.log.Error("error in client.SignersCount builder error", "error", err)
		return 0
	}

	count, _ := c.executeUintQuery(valueRequest)
	return uint(count)
}

// GetTokenId returns the token ID for the erc 20 address
func (c *client) GetTokenId(address string) string {
	if strings.Index(address, hexPrefix) == 0 {
		address = address[len(hexPrefix):]
	}

	builder := builders.NewTxDataBuilder(c.log).
		Address(c.bridgeAddress).
		CallerAddress(c.address).
		Function("getTokenIdForErc20Address").
		ArgHexString(address)

	valueRequest, err := builder.ToVmValueRequest()
	if err != nil {
		c.log.Error("error in client.GetTokenId builder error", "error", err)
		return ""
	}

	tokenId, err := c.executeStringQuery(valueRequest)
	if err != nil {
		c.log.Error(err.Error())
	}

	return tokenId
}

// GetErc20Address returns the corresponding ERC20 address
func (c *client) GetErc20Address(tokenId string) string {
	builder := builders.NewTxDataBuilder(c.log).
		Address(c.bridgeAddress).
		CallerAddress(c.address).
		Function("getErc20AddressForTokenId").
		ArgHexString(tokenId)

	valueRequest, err := builder.ToVmValueRequest()
	if err != nil {
		c.log.Error("error in client.GetErc20Address builder error", "error", err)
		return ""
	}

	address, err := c.executeStringQuery(valueRequest)
	if err != nil {
		c.log.Error("error in client.GetErc20Address", "error", err)
	}

	return address
}

// GetHexWalletAddress returns the wallet address as a hex string
func (c *client) GetHexWalletAddress() string {
	return hex.EncodeToString(c.address.AddressBytes())
}

func (c *client) executeQuery(valueRequest *data.VmValueRequest) ([][]byte, error) {
	response, err := c.proxy.ExecuteVMQuery(valueRequest)
	if err != nil {
		return nil, err
	}

	if response.Data.ReturnCode != "ok" {
		return nil, QueryResponseErr{
			code:      response.Data.ReturnCode,
			message:   response.Data.ReturnMessage,
			function:  valueRequest.FuncName,
			arguments: valueRequest.Args,
			address:   valueRequest.Address,
		}
	}

	return response.Data.ReturnData, nil
}

func (c *client) executeBoolQuery(valueRequest *data.VmValueRequest) bool {
	responseData, err := c.executeQuery(valueRequest)
	if err != nil {
		c.log.Error(err.Error())
		return false
	}

	if len(responseData) == 0 {
		return false
	}

	return c.convertToBool(responseData[0])
}

func (c *client) convertToBool(buff []byte) bool {
	if len(buff) == 0 {
		return false
	}

	result, err := strconv.ParseBool(fmt.Sprintf("%d", buff[0]))
	if err != nil {
		c.log.Error(err.Error())
		return false
	}

	return result
}

func (c *client) executeUintQuery(valueRequest *data.VmValueRequest) (uint64, error) {
	responseData, err := c.executeQuery(valueRequest)
	if err != nil {
		return 0, err
	}

	if len(responseData[0]) == 0 {
		return 0, err
	}

	result, err := strconv.ParseUint(hex.EncodeToString(responseData[0]), 16, 64)
	if err != nil {
		return 0, err
	}

	return result, nil
}

func (c *client) executeStringQuery(valueRequest *data.VmValueRequest) (string, error) {
	responseData, err := c.executeQuery(valueRequest)
	if err != nil {
		return "", err
	}

	if len(responseData[0]) == 0 {
		return "", err
	}

	return fmt.Sprintf("%x", responseData[0]), nil
}

func (c *client) signTransaction(builder builders.TxDataBuilder, cost uint64) (*data.Transaction, error) {
	networkConfig, err := c.proxy.GetNetworkConfig()
	if err != nil {
		return nil, err
	}

	nonce, err := c.nonceTxHandler.GetNonce(c.address)
	if err != nil {
		return nil, err
	}

	dataBytes, err := builder.ToDataBytes()
	if err != nil {
		return nil, err
	}

	tx := &data.Transaction{
		ChainID:  networkConfig.ChainID,
		Version:  networkConfig.MinTransactionVersion,
		GasLimit: cost,
		GasPrice: networkConfig.MinGasPrice,
		Nonce:    nonce,
		Data:     dataBytes,
		SndAddr:  c.address.AddressAsBech32String(),
		RcvAddr:  c.bridgeAddress.AddressAsBech32String(),
		Value:    "0",
	}

	err = c.signTransactionWithPrivateKey(tx)
	if err != nil {
		return nil, err
	}

	return tx, nil
}

// signTransactionWithPrivateKey signs a transaction with the client's private key
// TODO use the transaction interactor for signing and sending transactions
func (c *client) signTransactionWithPrivateKey(tx *data.Transaction) error {
	tx.Signature = ""
	bytes, err := json.Marshal(&tx)
	if err != nil {
		return err
	}
	signature, err := txSingleSigner.Sign(c.privateKey, bytes)
	if err != nil {
		return err
	}
	tx.Signature = hex.EncodeToString(signature)

	return nil
}

func (c *client) sendTransaction(builder builders.TxDataBuilder, cost uint64) (string, error) {
	tx, err := c.signTransaction(builder, cost)
	if err != nil {
		return "", err
	}

	return c.nonceTxHandler.SendTransaction(tx)
}

func (c *client) getCurrentBatch() ([][]byte, error) {
	builder := builders.NewTxDataBuilder(c.log).
		Address(c.bridgeAddress).
		CallerAddress(c.address).
		Function("getCurrentTxBatch")

	valueRequest, err := builder.ToVmValueRequest()
	if err != nil {
		return nil, err
	}

	return c.executeQuery(valueRequest)
}

// ExecuteVmQueryOnBridgeContract is able to execute queries on the defined bridge contract
func (c *client) ExecuteVmQueryOnBridgeContract(function string, params ...[]byte) ([][]byte, error) {
	builder := builders.NewTxDataBuilder(c.log).
		Address(c.bridgeAddress).
		CallerAddress(c.address).
		Function(function)

	for _, param := range params {
		builder.ArgHexString(hex.EncodeToString(param))
	}

	valueRequest, err := builder.ToVmValueRequest()
	if err != nil {
		return nil, err
	}

	return c.executeQuery(valueRequest)
}

func withTx(builder builders.TxDataBuilder, batch *bridge.Batch, mapper func(string) string) {
	for _, tx := range batch.Transactions {
		builder = builder.
			ArgAddress(data.NewAddressFromBytes([]byte(tx.To))).
			ArgHexString(mapper(tx.TokenAddress)).
			ArgBigInt(tx.Amount)
	}
}

// Close will close any started go routines. It returns nil.
func (c *client) Close() error {
	return c.nonceTxHandler.Close()
}

// IsInterfaceNil returns true if there is no value under the interface
func (c *client) IsInterfaceNil() bool {
	return c == nil
}

func emptyResponse(response [][]byte) bool {
	return len(response) == 0 || (len(response) == 1 && len(response[0]) == 0)
}

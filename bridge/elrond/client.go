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
	bridgeAddress  string
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

	_, err = data.NewAddressFromBech32String(args.Config.BridgeAddress)
	if err != nil {
		return nil, fmt.Errorf("%w for args.Config.BridgeAddress", err)
	}

	c := &client{
		proxy:          args.Proxy,
		bridgeAddress:  args.Config.BridgeAddress,
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
func (c *client) GetPending(_ context.Context) (*bridge.Batch, error) {
	c.log.Info("Elrond: Getting pending batch")
	responseData, err := c.getCurrentBatch()
	if err != nil {
		return nil, err
	}

	if emptyResponse(responseData) {
		return nil, nil
	}

	addrPkConv, _ := pubkeyConverter.NewBech32PubkeyConverter(32, c.log)
	numArgs := 6
	idxAmount := 5
	var transactions []*bridge.DepositTransaction
	for i := 1; i < len(responseData); i += numArgs {
		if len(responseData) < i+idxAmount {
			return nil, fmt.Errorf("Elrond: got an unexpected number of arguments, index %d, total args: %d", i, len(responseData))
		}

		amount := new(big.Int).SetBytes(responseData[i+idxAmount])
		blockNonce, errParse := parseIntFromByteSlice(responseData[i])
		if errParse != nil {
			return nil, fmt.Errorf("%w in client.GetPending, parseIntFromByteSlice(responseData[i])", err)
		}
		depositNonce, errParse := parseIntFromByteSlice(responseData[i+1])
		if errParse != nil {
			return nil, fmt.Errorf("%w in client.GetPending, parseIntFromByteSlice(responseData[i+1])", err)
		}

		tx := &bridge.DepositTransaction{
			From:          addrPkConv.Encode(responseData[i+2]),
			To:            fmt.Sprintf("0x%s", hex.EncodeToString(responseData[i+3])),
			DisplayableTo: fmt.Sprintf("0x%s", hex.EncodeToString(responseData[i+3])),
			TokenAddress:  hex.EncodeToString(responseData[i+4]),
			Amount:        amount,
			DepositNonce:  bridge.NewNonce(depositNonce),
			BlockNonce:    bridge.NewNonce(blockNonce),
			Error:         nil,
		}
		c.log.Debug("Elrond: created deposit transaction: " + tx.String())
		transactions = append(transactions, tx)
	}

	batchId, err := parseIntFromByteSlice(responseData[0])
	if err != nil {
		return nil, fmt.Errorf("%w in client.GetPending, parseIntFromByteSlice(responseData[0])", err)
	}

	c.log.Debug("Elrond: created batch", "batchID", batchId, "num transactions", len(transactions))

	return &bridge.Batch{
		Id:           bridge.NewBatchId(batchId),
		Transactions: transactions,
		Statuses:     make([]byte, len(transactions)),
	}, nil
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
func (c *client) ProposeSetStatus(ctx context.Context, batch *bridge.Batch) {
	builder := newBuilder(c.log).
		Func("proposeEsdtSafeSetCurrentTransactionBatchStatus").
		BatchId(batch.Id)

	newBatch, err := c.GetPending(ctx)
	if err != nil {
		c.log.Error("Elrond: get pending batch failed in ProposeSetStatus", "error", err)
		return
	}

	batch.ResolveNewDeposits(len(newBatch.Statuses))

	for _, stat := range batch.Statuses {
		builder = builder.Int(big.NewInt(int64(stat)))
	}

	hash, err := c.sendTransaction(builder, c.gasMapConfig.ProposeStatus)
	if err != nil {
		c.log.Error("Elrond: send transaction failed", "error", err)
		return
	}

	c.log.Info("Elrond: Proposed status update", "hash", hash)
}

// ProposeTransfer will trigger the propose transfer operation
func (c *client) ProposeTransfer(_ context.Context, batch *bridge.Batch) (string, error) {
	builder := newBuilder(c.log).
		Func("proposeMultiTransferEsdtBatch").
		BatchId(batch.Id)

	for _, tx := range batch.Transactions {
		builder = builder.
			Address([]byte(tx.To)).
			HexString(c.GetTokenId(tx.TokenAddress)).
			BigInt(tx.Amount)
	}

	batchData, errMarshal := json.Marshal(batch)
	if errMarshal != nil {
		c.log.Warn("Elrond: error not critical while serializing transaction", "error", errMarshal)
	}

	gasLimit := c.gasMapConfig.ProposeTransferBase + uint64(len(batch.Transactions))*c.gasMapConfig.ProposeTransferForEach
	hash, err := c.sendTransaction(builder, gasLimit)
	if err == nil {
		c.log.Info("Elrond: Proposed transfer for batch ", batch.Id, "with hash", hash, "batch data", string(batchData))
	} else {
		c.log.Error("Elrond: Propose transfer errored", "batch data", string(batchData), "error", err)
	}

	return hash, err
}

// WasProposedTransfer returns true if the transfer action proposed was triggered
func (c *client) WasProposedTransfer(_ context.Context, batch *bridge.Batch) bool {
	valueRequest := newValueBuilder(c.bridgeAddress, c.address.AddressAsBech32String(), c.log).
		Func("wasTransferActionProposed").
		BatchId(batch.Id).
		WithTx(batch, c.GetTokenId).
		Build()

	return c.executeBoolQuery(valueRequest)
}

// GetActionIdForProposeTransfer returns the action ID for the propose transfer operation
func (c *client) GetActionIdForProposeTransfer(_ context.Context, batch *bridge.Batch) bridge.ActionId {
	valueRequest := newValueBuilder(c.bridgeAddress, c.address.AddressAsBech32String(), c.log).
		Func("getActionIdForTransferBatch").
		BatchId(batch.Id).
		WithTx(batch, c.GetTokenId).
		Build()

	response, err := c.executeUintQuery(valueRequest)
	if err != nil {
		c.log.Error(err.Error())
		return bridge.NewActionId(0)
	}

	actionId := bridge.NewActionId(int64(response))

	c.log.Info("Elrond: fetched actionId for transfer batch", "actionId", actionId, "batch", batch.Id)

	return actionId
}

// WasProposedSetStatus returns true if the proposed set status was triggered
func (c *client) WasProposedSetStatus(ctx context.Context, batch *bridge.Batch) bool {
	valueRequest := newValueBuilder(c.bridgeAddress, c.address.AddressAsBech32String(), c.log).
		Func("wasSetCurrentTransactionBatchStatusActionProposed").
		BatchId(batch.Id)

	newBatch, err := c.GetPending(ctx)
	if err != nil {
		c.log.Error("Elrond: get pending batch failed in WasProposedSetStatus", "error", err)
		return false
	}
	batch.ResolveNewDeposits(len(newBatch.Statuses))

	for _, stat := range batch.Statuses {
		valueRequest = valueRequest.BigInt(big.NewInt(int64(stat)))
	}

	return c.executeBoolQuery(valueRequest.Build())
}

// GetTransactionsStatuses will return the transactions statuses from the batch ID
func (c *client) GetTransactionsStatuses(_ context.Context, batchId bridge.BatchId) ([]uint8, error) {
	if batchId == nil {
		return nil, ErrNilBatchId
	}

	valueRequest := newValueBuilder(c.bridgeAddress, c.address.AddressAsBech32String(), c.log).
		Func("getStatusesAfterExecution").
		BatchId(batchId)

	values, err := c.executeQuery(valueRequest.Build())
	if err != nil {
		return nil, err
	}

	if len(values) == 0 {
		return nil, fmt.Errorf("%w for batch ID %v", ErrNoStatusForBatchID, batchId)
	}
	isFinished := c.convertToBool(values[0])
	if !isFinished {
		return nil, fmt.Errorf("%w for batch ID %v", ErrBatchNotFinished, batchId)
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

	c.log.Debug("Elrond: got transaction status", "batchID", batchId, "status", results)

	return results, nil
}

func getStatusFromBuff(buff []byte) (byte, error) {
	if len(buff) == 0 {
		return 0, ErrMalformedBatchResponse
	}

	return buff[len(buff)-1], nil
}

// GetActionIdForSetStatusOnPendingTransfer returns the action ID for setting the status on the pending transfer batch
func (c *client) GetActionIdForSetStatusOnPendingTransfer(ctx context.Context, batch *bridge.Batch) bridge.ActionId {
	valueRequest := newValueBuilder(c.bridgeAddress, c.address.AddressAsBech32String(), c.log).
		Func("getActionIdForSetCurrentTransactionBatchStatus").
		BatchId(batch.Id)

	newBatch, err := c.GetPending(ctx)
	if err != nil {
		c.log.Error("Elrond: get pending batch failed in WasProposedSetStatus", "error", err)
		return bridge.NewActionId(0)
	}
	batch.ResolveNewDeposits(len(newBatch.Statuses))

	for _, stat := range batch.Statuses {
		valueRequest = valueRequest.BigInt(big.NewInt(int64(stat)))
	}

	response, err := c.executeUintQuery(valueRequest.Build())
	if err != nil {
		c.log.Error(err.Error())
		return bridge.NewActionId(0)
	}

	c.log.Debug("Elrond: got actionID", "actionID", response)

	return bridge.NewActionId(int64(response))
}

// WasExecuted returns true if the provided actionId was executed or not
func (c *client) WasExecuted(_ context.Context, actionId bridge.ActionId, _ bridge.BatchId) bool {
	valueRequest := newValueBuilder(c.bridgeAddress, c.address.AddressAsBech32String(), c.log).
		Func("wasActionExecuted").
		ActionId(actionId).
		Build()

	result := c.executeBoolQuery(valueRequest)

	if result {
		c.log.Info(fmt.Sprintf("Elrond: ActionId %v was executed", actionId))
	}

	return result
}

// Sign will trigger the execution of a sign operation
func (c *client) Sign(_ context.Context, actionId bridge.ActionId, batch *bridge.Batch) (string, error) {
	builder := newBuilder(c.log).
		Func("sign").
		ActionId(actionId)

	hash, err := c.sendTransaction(builder, c.gasMapConfig.Sign)

	batchData, err := json.Marshal(batch)
	if err != nil {
		c.log.Warn("Elrond: error not critical while serializing transaction", "error", err)
	}

	if err == nil {
		c.log.Info("Elrond: Signed", "hash", hash, "batch data", string(batchData))
	} else {
		c.log.Error("Elrond: Sign failed", "batch data", string(batchData), "error", err)
	}

	return hash, err
}

// Execute will trigger the execution of the provided action ID
func (c *client) Execute(_ context.Context, actionId bridge.ActionId, batch *bridge.Batch, _ bridge.SignaturesHolder) (string, error) {
	builder := newBuilder(c.log).
		Func("performAction").
		ActionId(actionId)

	gasLimit := c.gasMapConfig.PerformActionBase + uint64(len(batch.Transactions))*c.gasMapConfig.PerformActionForEach
	hash, err := c.sendTransaction(builder, gasLimit)

	batchData, err := json.Marshal(batch)
	if err != nil {
		c.log.Warn("Elrond: error not critical while serializing transaction", "error", err)
	}

	if err == nil {
		c.log.Info("Elrond: Executed action", "actionID", actionId, "batch data", string(batchData), "hash", hash)
	} else {
		c.log.Info("Elrond: Execution failed for action",
			"actionID", actionId,
			"batch data", string(batchData),
			"hash", hash,
			"error", err)
	}

	return hash, err
}

// SignersCount returns the signers count
func (c *client) SignersCount(_ context.Context, _ *bridge.Batch, actionId bridge.ActionId, _ bridge.SignaturesHolder) uint {
	valueRequest := newValueBuilder(c.bridgeAddress, c.address.AddressAsBech32String(), c.log).
		Func("getActionSignerCount").
		ActionId(actionId).
		Build()

	count, _ := c.executeUintQuery(valueRequest)
	return uint(count)
}

// GetTokenId returns the token ID for the erc 20 address
func (c *client) GetTokenId(address string) string {
	if strings.Index(address, hexPrefix) == 0 {
		address = address[len(hexPrefix):]
	}

	valueRequest := newValueBuilder(c.bridgeAddress, c.address.AddressAsBech32String(), c.log).
		Func("getTokenIdForErc20Address").
		HexString(address).
		Build()

	tokenId, err := c.executeStringQuery(valueRequest)
	if err != nil {
		c.log.Error(err.Error())
	}

	c.log.Debug("Elrond: get token ID", "address", address, "tokenID", tokenId)

	return tokenId
}

// GetErc20Address returns the corresponding ERC20 address
func (c *client) GetErc20Address(tokenId string) string {
	valueRequest := newValueBuilder(c.bridgeAddress, c.address.AddressAsBech32String(), c.log).
		Func("getErc20AddressForTokenId").
		HexString(tokenId).
		Build()

	address, err := c.executeStringQuery(valueRequest)
	if err != nil {
		c.log.Error(err.Error())
	}

	c.log.Debug("Elrond: get erc20 address", "tokenID", tokenId, "address", address)

	return address
}

// GetHexWalletAddress returns the wallet address as a hex string
func (c *client) GetHexWalletAddress() string {
	return hex.EncodeToString(c.address.AddressBytes())
}

func (c *client) executeQuery(valueRequest *data.VmValueRequest) ([][]byte, error) {
	response, err := c.proxy.ExecuteVMQuery(context.Background(), valueRequest)
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

func (c *client) signTransaction(builder *txDataBuilder, cost uint64) (*data.Transaction, error) {
	networkConfig, err := c.proxy.GetNetworkConfig(context.Background())
	if err != nil {
		return nil, err
	}

	nonce, err := c.nonceTxHandler.GetNonce(context.Background(), c.address)
	if err != nil {
		return nil, err
	}

	tx := &data.Transaction{
		ChainID:  networkConfig.ChainID,
		Version:  networkConfig.MinTransactionVersion,
		GasLimit: cost,
		GasPrice: networkConfig.MinGasPrice,
		Nonce:    nonce,
		Data:     builder.ToBytes(),
		SndAddr:  c.address.AddressAsBech32String(),
		RcvAddr:  c.bridgeAddress,
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

func (c *client) sendTransaction(builder *txDataBuilder, cost uint64) (string, error) {
	tx, err := c.signTransaction(builder, cost)
	if err != nil {
		return "", err
	}

	return c.nonceTxHandler.SendTransaction(context.Background(), tx)
}

func (c *client) getCurrentBatch() ([][]byte, error) {
	valueRequest := newValueBuilder(c.bridgeAddress, c.address.AddressAsBech32String(), c.log).
		Func("getCurrentTxBatch").
		Build()

	return c.executeQuery(valueRequest)
}

// ExecuteVmQueryOnBridgeContract is able to execute queries on the defined bridge contract
func (c *client) ExecuteVmQueryOnBridgeContract(function string, params ...[]byte) ([][]byte, error) {
	valueRequestBuilderInstance := newValueBuilder(c.bridgeAddress, c.address.AddressAsBech32String(), c.log).
		Func(function)
	for _, param := range params {
		valueRequestBuilderInstance.HexString(hex.EncodeToString(param))
	}

	valueRequest := valueRequestBuilderInstance.Build()

	return c.executeQuery(valueRequest)
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

// Builders

type valueRequestBuilder struct {
	address    string
	funcName   string
	callerAddr string
	args       []string
	log        logger.Logger
}

func newValueBuilder(address, callerAddr string, log logger.Logger) *valueRequestBuilder {
	return &valueRequestBuilder{
		address:    address,
		callerAddr: callerAddr,
		args:       []string{},
		log:        log,
	}
}

func (builder *valueRequestBuilder) Build() *data.VmValueRequest {
	return &data.VmValueRequest{
		Address:    builder.address,
		FuncName:   builder.funcName,
		CallerAddr: builder.callerAddr,
		Args:       builder.args,
	}
}

func (builder *valueRequestBuilder) Func(functionName string) *valueRequestBuilder {
	builder.funcName = functionName

	return builder
}

func (builder *valueRequestBuilder) Nonce(nonce bridge.Nonce) *valueRequestBuilder {
	return builder.BigInt(nonce)
}

func (builder *valueRequestBuilder) BatchId(batchId bridge.BatchId) *valueRequestBuilder {
	return builder.BigInt(batchId)
}

func (builder *valueRequestBuilder) ActionId(actionId bridge.ActionId) *valueRequestBuilder {
	return builder.BigInt(actionId)
}

func (builder *valueRequestBuilder) BigInt(value *big.Int) *valueRequestBuilder {
	builder.args = append(builder.args, intToHex(value))

	return builder
}

func (builder *valueRequestBuilder) HexString(value string) *valueRequestBuilder {
	builder.args = append(builder.args, value)

	return builder
}

func (builder *valueRequestBuilder) Address(value string) *valueRequestBuilder {
	builder.args = append(builder.args, hex.EncodeToString([]byte(value)))

	return builder
}

func (builder *valueRequestBuilder) WithTx(batch *bridge.Batch, mapper func(string) string) *valueRequestBuilder {
	for _, tx := range batch.Transactions {
		builder = builder.
			Address(tx.To).
			HexString(mapper(tx.TokenAddress)).
			BigInt(tx.Amount)
	}

	return builder
}

type txDataBuilder struct {
	function  string
	elements  []string
	separator string
	log       logger.Logger
}

func newBuilder(log logger.Logger) *txDataBuilder {
	return &txDataBuilder{
		function:  "",
		elements:  make([]string, 0),
		separator: "@",
		log:       log,
	}
}

func (builder *txDataBuilder) Func(function string) *txDataBuilder {
	builder.function = function

	return builder
}

func (builder *txDataBuilder) ActionId(value bridge.ActionId) *txDataBuilder {
	return builder.Int(value)
}

func (builder *txDataBuilder) BatchId(value bridge.BatchId) *txDataBuilder {
	return builder.Int(value)
}

func (builder *txDataBuilder) Nonce(nonce bridge.Nonce) *txDataBuilder {
	return builder.Int(nonce)
}

func (builder *txDataBuilder) Int(value *big.Int) *txDataBuilder {
	builder.elements = append(builder.elements, intToHex(value))

	return builder
}

func (builder *txDataBuilder) BigInt(value *big.Int) *txDataBuilder {
	builder.elements = append(builder.elements, hex.EncodeToString(value.Bytes()))

	return builder
}

func (builder *txDataBuilder) Address(bytes []byte) *txDataBuilder {
	builder.elements = append(builder.elements, hex.EncodeToString(bytes))

	return builder
}

func (builder *txDataBuilder) HexString(value string) *txDataBuilder {
	builder.elements = append(builder.elements, value)

	return builder
}

func (builder *txDataBuilder) ToString() string {
	result := builder.function
	for _, element := range builder.elements {
		result = result + builder.separator + element
	}

	return result
}

func (builder *txDataBuilder) ToBytes() []byte {
	return []byte(builder.ToString())
}

func intToHex(value *big.Int) string {
	return hex.EncodeToString(value.Bytes())
}

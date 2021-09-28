package elrond

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridge"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/core/pubkeyConverter"
	"github.com/ElrondNetwork/elrond-go-crypto/signing"
	"github.com/ElrondNetwork/elrond-go-crypto/signing/ed25519"
	"github.com/ElrondNetwork/elrond-go-crypto/signing/ed25519/singlesig"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/core"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/data"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/interactors"
)

const (
	signCost              = 45_000_000
	proposeTransferCost   = 45_000_000
	proposeTransferTxCost = 25_000_000
	proposeStatusCost     = 60_000_000
	performActionCost     = 70_000_000
	performActionTxCost   = 30_000_000
	getNextTxBatchCost    = 260_000_000
)

const (
	// canProposeAndSign is the value for the role held by an active validator
	canProposeAndSign = 2
)

// QueryResponseErr represents the query response error DTO struct
type QueryResponseErr struct {
	code    string
	message string
}

func (e QueryResponseErr) Error() string {
	return fmt.Sprintf("Got response code %q and message %q", e.code, e.message)
}

// client represents the Elrond Client implementation
type client struct {
	proxy               bridge.ElrondProxy
	bridgeAddress       string
	privateKey          []byte
	address             core.AddressHandler
	nonce               uint64
	log                 logger.Logger
	cancelFunc          func()
	nonceUpdateInterval time.Duration
}

// ClientArgs represents the argument for the NewClient constructor function
type ClientArgs struct {
	Config bridge.Config
	Proxy  bridge.ElrondProxy
}

// NewClient returns a new Elrond Client instance
func NewClient(args ClientArgs) (*client, error) {
	log := logger.GetOrCreate("ElrondClient")

	if check.IfNil(args.Proxy) {
		return nil, ErrNilProxy
	}
	if args.Config.NonceUpdateInSeconds == 0 {
		return nil, ErrInvalidNonceUpdateInterval
	}
	wallet := interactors.NewWallet()

	privateKey, err := wallet.LoadPrivateKeyFromPemFile(args.Config.PrivateKey)
	if err != nil {
		return nil, err
	}

	address, err := wallet.GetAddressFromPrivateKey(privateKey)
	if err != nil {
		return nil, err
	}

	log.Info("Elrond: NewClient", "address", address.AddressAsBech32String())

	c := &client{
		proxy:               args.Proxy,
		bridgeAddress:       args.Config.BridgeAddress,
		privateKey:          privateKey,
		address:             address,
		log:                 log,
		nonceUpdateInterval: time.Second * time.Duration(args.Config.NonceUpdateInSeconds),
	}

	var ctx context.Context
	ctx, c.cancelFunc = context.WithCancel(context.Background())
	c.saveCurrentNonce()
	go c.poll(ctx)

	return c, nil
}

func (c *client) poll(ctx context.Context) {
	for {
		select {
		case <-time.After(c.nonceUpdateInterval):
			c.saveCurrentNonce()
		case <-ctx.Done():
			c.log.Debug("Client.poll function is closing...")
			return
		}
	}
}

func (c *client) saveCurrentNonce() {
	account, err := c.proxy.GetAccount(c.address)
	if err != nil {
		c.log.Debug("Elrond: error polling account", "address", c.address.AddressAsBech32String(), "error", err.Error())
		return
	}

	c.log.Debug("Elrond: polled account", "address", c.address.AddressAsBech32String(), "nonce", account.Nonce)
	atomic.StoreUint64(&c.nonce, account.Nonce)
}

// GetPending returns the pending batch
func (c *client) GetPending(context.Context) *bridge.Batch {
	c.log.Info("Elrond: Getting pending batch")
	responseData, err := c.getCurrentBatch()
	if err != nil {
		c.log.Error("Elrond: Error querying current batch", "error", err.Error())
		return nil
	}

	if emptyResponse(responseData) {
		_, err = c.getNextPendingBatch()
		if err != nil {
			c.log.Error("Elrond: Error retrieving next pending batch", "error", err.Error())
			return nil
		}
	}

	responseData, err = c.getCurrentBatch()
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
		blockNonce, errParse := strconv.ParseInt(hex.EncodeToString(responseData[i]), 16, 64)
		if errParse != nil {
			c.log.Error("Elrond: parse error", "error", errParse.Error())
			return nil
		}
		depositNonce, errParse := strconv.ParseInt(hex.EncodeToString(responseData[i+1]), 16, 64)
		if errParse != nil {
			c.log.Error("Elrond: parse error", "error", errParse.Error())
			return nil
		}

		tx := &bridge.DepositTransaction{
			To:           fmt.Sprintf("0x%s", hex.EncodeToString(responseData[i+3])),
			From:         addrPkConv.Encode(responseData[i+2]),
			TokenAddress: fmt.Sprintf("0x%s", hex.EncodeToString(responseData[i+4])),
			Amount:       amount,
			DepositNonce: bridge.NewNonce(depositNonce),
			BlockNonce:   bridge.NewNonce(blockNonce),
			Status:       0,
			Error:        nil,
		}
		transactions = append(transactions, tx)
	}

	batchId, err := strconv.ParseInt(hex.EncodeToString(responseData[0]), 16, 64)
	if err != nil {
		c.log.Error("Elrond: parse error", "error", err.Error())
		return nil
	}

	return &bridge.Batch{
		Id:           bridge.NewBatchId(batchId),
		Transactions: transactions,
	}
}

// ProposeSetStatus will trigger the proposal of the ESDT safe set current transaction batch status operation
func (c *client) ProposeSetStatus(_ context.Context, batch *bridge.Batch) {
	builder := newBuilder(c.log).
		Func("proposeEsdtSafeSetCurrentTransactionBatchStatus").
		BatchId(batch.Id)

	for _, tx := range batch.Transactions {
		builder = builder.Int(big.NewInt(int64(tx.Status)))
	}

	hash, err := c.sendTransaction(builder, proposeStatusCost)
	if err != nil {
		c.log.Error("Elrond: send transaction failed", "error", err.Error())
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
			Address(tx.To).
			HexString(c.GetTokenId(tx.TokenAddress[2:])).
			BigInt(tx.Amount)
	}

	hash, err := c.sendTransaction(builder, uint64(proposeTransferCost+len(batch.Transactions)*proposeTransferTxCost))

	if err == nil {
		c.log.Info("Elrond: Proposed transfer for batch ", batch.Id, " with hash ", hash)
	} else {
		c.log.Error("Elrond: Propose transfer errored", "error", err.Error())
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
func (c *client) WasProposedSetStatus(_ context.Context, batch *bridge.Batch) bool {
	valueRequest := newValueBuilder(c.bridgeAddress, c.address.AddressAsBech32String(), c.log).
		Func("wasSetCurrentTransactionBatchStatusActionProposed").
		BatchId(batch.Id)

	for _, tx := range batch.Transactions {
		valueRequest = valueRequest.BigInt(big.NewInt(int64(tx.Status)))
	}

	return c.executeBoolQuery(valueRequest.Build())
}

// GetActionIdForSetStatusOnPendingTransfer returns the action ID for setting the status on the pending transfer batch
func (c *client) GetActionIdForSetStatusOnPendingTransfer(_ context.Context, batch *bridge.Batch) bridge.ActionId {
	valueRequest := newValueBuilder(c.bridgeAddress, c.address.AddressAsBech32String(), c.log).
		Func("getActionIdForSetCurrentTransactionBatchStatus").
		BatchId(batch.Id)

	for _, tx := range batch.Transactions {
		valueRequest = valueRequest.BigInt(big.NewInt(int64(tx.Status)))
	}

	response, err := c.executeUintQuery(valueRequest.Build())
	if err != nil {
		c.log.Error(err.Error())
		return bridge.NewActionId(0)
	}

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
func (c *client) Sign(_ context.Context, actionId bridge.ActionId) (string, error) {
	builder := newBuilder(c.log).
		Func("sign").
		ActionId(actionId)

	hash, err := c.sendTransaction(builder, signCost)

	if err == nil {
		c.log.Info("Elrond: Signed", "hash", hash)
	} else {
		c.log.Error("Elrond: Sign failed", "error", err.Error())
	}

	return hash, err
}

// Execute will trigger the execution of the provided action ID
func (c *client) Execute(_ context.Context, actionId bridge.ActionId, batch *bridge.Batch) (string, error) {
	builder := newBuilder(c.log).
		Func("performAction").
		ActionId(actionId)

	hash, err := c.sendTransaction(builder, uint64(performActionCost+len(batch.Transactions)*performActionTxCost))

	if err == nil {
		c.log.Info("Elrond: Executed action", "action ID", actionId, "hash", hash)
	} else {
		c.log.Info("Elrond: Execution failed for action", "action ID", actionId, "hash", hash, "error", err.Error())
	}

	return hash, err
}

// SignersCount returns the signers count
func (c *client) SignersCount(_ context.Context, actionId bridge.ActionId) uint {
	valueRequest := newValueBuilder(c.bridgeAddress, c.address.AddressAsBech32String(), c.log).
		Func("getActionSignerCount").
		ActionId(actionId).
		Build()

	count, _ := c.executeUintQuery(valueRequest)
	return uint(count)
}

// GetTokenId returns the token ID for the erc 20 address
func (c *client) GetTokenId(address string) string {
	valueRequest := newValueBuilder(c.bridgeAddress, c.address.AddressAsBech32String(), c.log).
		Func("getTokenIdForErc20Address").
		HexString(address).
		Build()

	tokenId, err := c.executeStringQuery(valueRequest)
	if err != nil {
		c.log.Error(err.Error())
	}

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

	return address
}

// IsWhitelisted returns true if the address can propose or sign
func (c *client) IsWhitelisted(address string) bool {
	valueRequest := newValueBuilder(c.bridgeAddress, c.address.AddressAsBech32String(), c.log).
		Func("userRole").
		HexString(address).
		Build()

	role, err := c.executeUintQuery(valueRequest)
	if err != nil {
		c.log.Error(err.Error())
		return false
	}

	return role == canProposeAndSign
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
		return nil, QueryResponseErr{response.Data.ReturnCode, response.Data.ReturnMessage}
	}

	return response.Data.ReturnData, nil
}

func (c *client) executeBoolQuery(valueRequest *data.VmValueRequest) bool {
	responseData, err := c.executeQuery(valueRequest)
	if err != nil {
		c.log.Error(err.Error())
		return false
	}

	if len(responseData[0]) == 0 {
		return false
	}

	result, err := strconv.ParseBool(fmt.Sprintf("%d", responseData[0][0]))
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
	networkConfig, err := c.proxy.GetNetworkConfig()
	if err != nil {
		return nil, err
	}

	nonce := atomic.LoadUint64(&c.nonce)
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

	err = c.signTransactionWithPrivateKey(tx, c.privateKey)
	if err != nil {
		return nil, err
	}

	return tx, nil
}

// signTransactionWithPrivateKey signs a transaction with the provided private key
// TODO use the transaction interactor for signing and sending transactions
func (c *client) signTransactionWithPrivateKey(tx *data.Transaction, privateKey []byte) error {
	tx.Signature = ""
	txSingleSigner := &singlesig.Ed25519Signer{}
	suite := ed25519.NewEd25519()
	keyGen := signing.NewKeyGenerator(suite)
	txSignPrivKey, err := keyGen.PrivateKeyFromByteArray(privateKey)
	if err != nil {
		return err
	}
	bytes, err := json.Marshal(&tx)
	if err != nil {
		return err
	}
	signature, err := txSingleSigner.Sign(txSignPrivKey, bytes)
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

	hash, err := c.proxy.SendTransaction(tx)
	if err == nil {
		atomic.AddUint64(&c.nonce, 1)
	}

	return hash, err
}

func (c *client) getCurrentBatch() ([][]byte, error) {
	valueRequest := newValueBuilder(c.bridgeAddress, c.address.AddressAsBech32String(), c.log).
		Func("getCurrentTxBatch").
		Build()

	return c.executeQuery(valueRequest)
}

func (c *client) getNextPendingBatch() (string, error) {
	builder := newBuilder(c.log).
		Func("getNextTransactionBatch")

	return c.sendTransaction(builder, getNextTxBatchCost)
}

// Close will close any started go routines. It returns nil.
func (c *client) Close() error {
	c.cancelFunc()

	return nil
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
	pkConv, _ := pubkeyConverter.NewBech32PubkeyConverter(32, builder.log)
	buff, _ := pkConv.Decode(value)
	builder.args = append(builder.args, hex.EncodeToString(buff))

	return builder
}

func (builder *valueRequestBuilder) WithTx(batch *bridge.Batch, mapper func(string) string) *valueRequestBuilder {
	for _, tx := range batch.Transactions {
		builder = builder.
			Address(tx.To).
			HexString(mapper(tx.TokenAddress[2:])).
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

func (builder *txDataBuilder) Address(value string) *txDataBuilder {
	pkConv, _ := pubkeyConverter.NewBech32PubkeyConverter(32, builder.log)
	buff, _ := pkConv.Decode(value)
	builder.elements = append(builder.elements, hex.EncodeToString(buff))

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

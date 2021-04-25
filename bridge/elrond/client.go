package elrond

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
	"time"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridge"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-sdk/erdgo"
	"github.com/ElrondNetwork/elrond-sdk/erdgo/blockchain"
	"github.com/ElrondNetwork/elrond-sdk/erdgo/data"
)

const (
	Executed = 3
	Rejected = 4
)

type QueryResponseErr struct {
	code    string
	message string
}

func (e QueryResponseErr) Error() string {
	return fmt.Sprintf("Got response code %q and message %q", e.code, e.message)
}

type elrondProxy interface {
	GetNetworkConfig() (*data.NetworkConfig, error)
	SendTransaction(*data.Transaction) (string, error)
	GetTransactionInfoWithResults(hash string) (*data.TransactionInfo, error)
	RequestTransactionCost(tx *data.Transaction) (*data.TxCostResponseData, error)
	ExecuteVMQuery(vmRequest *data.VmValueRequest) (*data.VmValuesResponseData, error)
}

type Client struct {
	proxy         elrondProxy
	bridgeAddress string
	privateKey    []byte
	address       string
	nonce         uint64
	log           logger.Logger
}

func NewClient(config bridge.Config) (*Client, error) {
	log := logger.GetOrCreate("ElrondClient")

	proxy := blockchain.NewElrondProxy(config.NetworkAddress, nil)

	privateKey, err := erdgo.LoadPrivateKeyFromPemFile(config.PrivateKeyPath)
	if err != nil {
		return nil, err
	}

	addressString, err := erdgo.GetAddressFromPrivateKey(privateKey)
	if err != nil {
		return nil, err
	}
	log.Info(fmt.Sprintf("Address: %q", addressString))

	address, err := data.NewAddressFromBech32String(addressString)
	if err != nil {
		return nil, err
	}

	account, err := proxy.GetAccount(address)
	if err != nil {
		return nil, err
	}
	initialNonce := account.Nonce

	return &Client{
		proxy:         proxy,
		bridgeAddress: config.BridgeAddress,
		privateKey:    privateKey,
		address:       address.AddressAsBech32String(),
		nonce:         initialNonce,
		log:           log,
	}, nil
}

func (c *Client) GetPendingDepositTransaction(context.Context) *bridge.DepositTransaction {
	// getNextPendingTransaction
	// if none -> error
	return nil
}

func (c *Client) ProposeSetStatusSuccessOnPendingTransfer(context.Context) {
	builder := newBuilder().
		Func("proposeEsdtSafeSetCurrentTransactionStatus").
		Int(Executed)

	_, _ = c.sendTransaction(builder)
}

func (c *Client) ProposeSetStatusFailedOnPendingTransfer(context.Context) {
	builder := newBuilder().
		Func("proposeEsdtSafeSetCurrentTransactionStatus").
		Int(Executed)

	_, _ = c.sendTransaction(builder)
}

func (c *Client) ProposeTransfer(_ context.Context, tx *bridge.DepositTransaction) (string, error) {
	// proposeMultiTransferEsdtTransferEsdtToken(depositTx) -> ActionId
	// pub enum TransactionStatus {
	//    None,
	//    Pending,
	//    InProgress,
	//    Executed,
	//    Rejected,
	//}
	builder := newBuilder().
		Func("proposeMultiTransferEsdtTransferEsdtToken").
		Nonce(tx.DepositNonce)

	return c.sendTransaction(builder)
}

func (c *Client) WasProposedTransfer(_ context.Context, nonce bridge.Nonce) bool {
	valueRequest := newValueBuilder(c.bridgeAddress, c.address).
		Func("wasTransferActionProposed").
		Nonce(nonce).
		Build()

	return c.executeBoolQuery(valueRequest)
}

func (c *Client) GetActionIdForProposeTransfer(_ context.Context, nonce bridge.Nonce) bridge.ActionId {
	valueRequest := newValueBuilder(c.bridgeAddress, c.address).
		Func("getActionIdForEthTxNonce").
		Nonce(nonce).
		Build()

	response, err := c.executeUintQuery(valueRequest)
	if err != nil {
		c.log.Error(err.Error())
		return bridge.ActionId(0)
	}

	return bridge.ActionId(response)
}

func (c *Client) WasProposedSetStatusSuccessOnPendingTransfer(context.Context) bool {
	valueRequest := newValueBuilder(c.bridgeAddress, c.address).
		Func("wasSetCurrentTransactionStatusActionProposed").
		Build()

	return c.executeBoolQuery(valueRequest)
}

func (c *Client) WasProposedSetStatusFailedOnPendingTransfer(context.Context) bool {
	valueRequest := newValueBuilder(c.bridgeAddress, c.address).
		Func("wasSetCurrentTransactionStatusActionProposed").
		Build()

	return c.executeBoolQuery(valueRequest)
}

func (c *Client) GetActionIdForSetStatusOnPendingTransfer(context.Context) bridge.ActionId {
	valueRequest := newValueBuilder(c.bridgeAddress, c.address).
		Func("getActionIdForSetCurrentTransactionStatus").
		Build()

	response, err := c.executeUintQuery(valueRequest)
	if err != nil {
		c.log.Error(err.Error())
		return bridge.ActionId(0)
	}

	return bridge.ActionId(response)
}

func (c *Client) WasExecuted(_ context.Context, actionId bridge.ActionId) bool {
	valueRequest := newValueBuilder(c.bridgeAddress, c.address).
		Func("wasActionExecuted").
		ActionId(actionId).
		Build()

	return c.executeBoolQuery(valueRequest)
}

func (c *Client) Sign(_ context.Context, actionId bridge.ActionId) (string, error) {
	builder := newBuilder().
		Func("sign").
		ActionId(actionId)

	return c.sendTransaction(builder)
}

func (c *Client) Execute(_ context.Context, actionId bridge.ActionId) (string, error) {
	builder := newBuilder().
		Func("performAction").
		ActionId(actionId)

	return c.sendTransaction(builder)
}

func (c *Client) SignersCount(context.Context, bridge.ActionId) uint {
	// getActionSignerCount(actionId)
	return 0
}

// Helpers

func (c *Client) executeQuery(valueRequest *data.VmValueRequest) ([][]byte, error) {
	response, err := c.proxy.ExecuteVMQuery(valueRequest)
	if err != nil {
		return nil, err
	}

	if response.Data.ReturnCode != "ok" {
		return nil, QueryResponseErr{response.Data.ReturnCode, response.Data.ReturnMessage}
	}

	return response.Data.ReturnData, nil
}

func (c *Client) executeBoolQuery(valueRequest *data.VmValueRequest) bool {
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

func (c *Client) executeUintQuery(valueRequest *data.VmValueRequest) (uint64, error) {
	responseData, err := c.executeQuery(valueRequest)
	if err != nil {
		return 0, err
	}

	if len(responseData[0]) == 0 {
		return 0, err
	}

	result, err := strconv.ParseUint(fmt.Sprintf("%d", responseData[0][0]), 10, 0)
	if err != nil {
		return 0, err
	}

	return result, nil
}

func (c *Client) signTransaction(builder *txDataBuilder) (*data.Transaction, error) {
	networkConfig, err := c.proxy.GetNetworkConfig()
	if err != nil {
		return nil, err
	}

	tx := &data.Transaction{
		ChainID:  networkConfig.ChainID,
		Version:  networkConfig.MinTransactionVersion,
		GasLimit: networkConfig.MinGasLimit,
		GasPrice: networkConfig.MinGasPrice,
		Nonce:    c.nonce,
		Data:     builder.ToBytes(),
		SndAddr:  c.address,
		RcvAddr:  c.bridgeAddress,
		Value:    "0",
	}

	cost, err := c.proxy.RequestTransactionCost(tx)
	if err != nil {
		return nil, err
	}
	c.log.Info(fmt.Sprintf("Min gaslimit: %d", tx.GasLimit))
	if cost.TxCost > 0 {
		tx.GasLimit = cost.TxCost
	} else {
		tx.GasLimit = 200000000
	}
	c.log.Info(fmt.Sprintf("Response message %q", cost.RetMessage))
	c.log.Info(fmt.Sprintf("Calculated gaslimit: %d", tx.GasLimit))

	err = erdgo.SignTransaction(tx, c.privateKey)
	if err != nil {
		return nil, err
	}

	return tx, nil
}

func (c *Client) incrementNonce() {
	c.nonce++
}

func (c *Client) sendTransaction(builder *txDataBuilder) (string, error) {
	tx, err := c.signTransaction(builder)
	if err != nil {
		return "", err
	}

	hash, err := c.proxy.SendTransaction(tx)
	if err == nil {
		c.incrementNonce()
	}

	return hash, err
}

func (c *Client) printTransactionResults(hash string) {
	time.Sleep(10 * time.Second)
	info, err := c.proxy.GetTransactionInfoWithResults(hash)
	if err != nil {
		c.log.Error(err.Error())
		return
	}

	scResults := info.Data.Transaction.ScResults
	for i := 0; i < len(scResults); i++ {
		c.log.Info(scResults[i].ReturnMessage)
	}
}

// Builders

type valueRequestBuilder struct {
	address    string
	funcName   string
	callerAddr string
	args       []string
}

func newValueBuilder(address, callerAddr string) *valueRequestBuilder {
	return &valueRequestBuilder{
		address:    address,
		callerAddr: callerAddr,
		args:       []string{},
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
	return builder.Int(int(nonce))
}

func (builder *valueRequestBuilder) ActionId(actionId bridge.ActionId) *valueRequestBuilder {
	return builder.Int(int(actionId))
}

func (builder *valueRequestBuilder) Int(value int) *valueRequestBuilder {
	builder.args = append(builder.args, intToHex(value))

	return builder
}

type txDataBuilder struct {
	function  string
	elements  []string
	separator string
}

func newBuilder() *txDataBuilder {
	return &txDataBuilder{
		function:  "",
		elements:  make([]string, 0),
		separator: "@",
	}
}

func (builder *txDataBuilder) Func(function string) *txDataBuilder {
	builder.function = function

	return builder
}

func (builder *txDataBuilder) ActionId(value bridge.ActionId) *txDataBuilder {
	return builder.Int(int(value))
}

func (builder *txDataBuilder) Nonce(nonce bridge.Nonce) *txDataBuilder {
	return builder.Int(int(nonce))
}

func (builder *txDataBuilder) Int(value int) *txDataBuilder {
	builder.elements = append(builder.elements, intToHex(value))

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

func intToHex(value int) string {
	return hex.EncodeToString(big.NewInt(int64(value)).Bytes())
}

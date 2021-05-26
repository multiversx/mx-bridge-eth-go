package elrond

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strconv"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridge"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core/pubkeyConverter"
	"github.com/ElrondNetwork/elrond-sdk/erdgo"
	"github.com/ElrondNetwork/elrond-sdk/erdgo/blockchain"
	"github.com/ElrondNetwork/elrond-sdk/erdgo/data"
)

const (
	ExecutionCost = 1000000000
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

	privateKey, err := erdgo.LoadPrivateKeyFromPemFile(config.PrivateKey)
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
	responseData, err := c.getCurrentTx()
	if err != nil {
		c.log.Error(err.Error())
		return nil
	}

	if len(responseData) == 0 {
		_, err = c.getNextPendingTransaction()
		if err != nil {
			c.log.Info(err.Error())
			return nil
		}
	}

	responseData, err = c.getCurrentTx()
	if err != nil {
		c.log.Error(err.Error())
		return nil
	}

	if len(responseData) == 0 {
		return nil
	}

	to := fmt.Sprintf("0x%s", hex.EncodeToString(responseData[3][:20]))

	addrPkConv, _ := pubkeyConverter.NewBech32PubkeyConverter(32)
	from := addrPkConv.Encode(responseData[2])
	tokenAddress := fmt.Sprintf("0x%s", hex.EncodeToString(responseData[4]))
	amount, err := strconv.ParseInt(hex.EncodeToString(responseData[5]), 16, 64)
	if err != nil {
		c.log.Error(err.Error())
		return nil
	}
	depositNonce, err := strconv.ParseInt(hex.EncodeToString(responseData[1]), 16, 64)
	if err != nil {
		c.log.Error(err.Error())
		return nil
	}

	return &bridge.DepositTransaction{
		To:           to,
		From:         from,
		TokenAddress: tokenAddress,
		Amount:       big.NewInt(amount),
		DepositNonce: bridge.NewNonce(depositNonce),
	}
}

func (c *Client) ProposeSetStatus(_ context.Context, status uint8, _ bridge.Nonce) {
	builder := newBuilder().
		Func("proposeEsdtSafeSetCurrentTransactionStatus").
		Int(big.NewInt(int64(status)))

	_, _ = c.sendTransaction(builder, 0)
}

func (c *Client) ProposeTransfer(_ context.Context, tx *bridge.DepositTransaction) (string, error) {
	builder := newBuilder().
		Func("proposeMultiTransferEsdtTransferEsdtToken").
		Nonce(tx.DepositNonce).
		Address(tx.To).
		HexString(c.GetTokenId(tx.TokenAddress[2:])).
		BigInt(tx.Amount)

	return c.sendTransaction(builder, 0)
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
		return bridge.NewActionId(0)
	}

	return bridge.NewActionId(int64(response))
}

func (c *Client) WasProposedSetStatusOnPendingTransfer(_ context.Context, status uint8) bool {
	valueRequest := newValueBuilder(c.bridgeAddress, c.address).
		Func("wasSetCurrentTransactionStatusActionProposed").
		Int(big.NewInt(int64(status))).
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
		return bridge.NewActionId(0)
	}

	return bridge.NewActionId(int64(response))
}

func (c *Client) WasExecuted(_ context.Context, actionId bridge.ActionId, _ bridge.Nonce) bool {
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

	return c.sendTransaction(builder, 0)
}

func (c *Client) Execute(_ context.Context, actionId bridge.ActionId, _ bridge.Nonce) (string, error) {
	builder := newBuilder().
		Func("performAction").
		ActionId(actionId)

	return c.sendTransaction(builder, ExecutionCost)
}

func (c *Client) SignersCount(_ context.Context, actionId bridge.ActionId) uint {
	valueRequest := newValueBuilder(c.bridgeAddress, c.address).
		Func("getActionSignerCount").
		ActionId(actionId).
		Build()

	count, _ := c.executeUintQuery(valueRequest)
	return uint(count)
}

// Mapper

func (c *Client) GetTokenId(address string) string {
	valueRequest := newValueBuilder(c.bridgeAddress, c.address).
		Func("getTokenIdForErc20Address").
		HexString(address).
		Build()

	tokenId, err := c.executeStringQuery(valueRequest)
	if err != nil {
		c.log.Error(err.Error())
	}

	return tokenId
}

func (c *Client) GetErc20Address(tokenId string) string {
	valueRequest := newValueBuilder(c.bridgeAddress, c.address).
		Func("getErc20AddressForTokenId").
		HexString(tokenId).
		Build()

	address, err := c.executeStringQuery(valueRequest)
	if err != nil {
		c.log.Error(err.Error())
	}

	return address
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

	result, err := strconv.ParseUint(hex.EncodeToString(responseData[0]), 16, 64)
	if err != nil {
		return 0, err
	}

	return result, nil
}

func (c *Client) executeStringQuery(valueRequest *data.VmValueRequest) (string, error) {
	responseData, err := c.executeQuery(valueRequest)
	if err != nil {
		return "", err
	}

	if len(responseData[0]) == 0 {
		return "", err
	}

	return fmt.Sprintf("%x", responseData[0]), nil
}

func (c *Client) signTransaction(builder *txDataBuilder, cost uint64) (*data.Transaction, error) {
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

	if cost == 0 {
		reqCost, err := c.proxy.RequestTransactionCost(tx)
		if err != nil {
			return nil, err
		}
		if reqCost.RetMessage != "" {
			return nil, errors.New(reqCost.RetMessage)
		}
		cost = reqCost.TxCost
	}

	tx.GasLimit = cost

	err = erdgo.SignTransaction(tx, c.privateKey)
	if err != nil {
		return nil, err
	}

	return tx, nil
}

// TODO: most likely the client should not do this
func (c *Client) incrementNonce() {
	c.nonce++
}

func (c *Client) sendTransaction(builder *txDataBuilder, cost uint64) (string, error) {
	tx, err := c.signTransaction(builder, cost)
	if err != nil {
		return "", err
	}

	hash, err := c.proxy.SendTransaction(tx)
	if err == nil {
		c.incrementNonce()
	}

	return hash, err
}

func (c *Client) getCurrentTx() ([][]byte, error) {
	valueRequest := newValueBuilder(c.bridgeAddress, c.address).
		Func("getCurrentTx").
		Build()

	return c.executeQuery(valueRequest)
}

func (c *Client) getNextPendingTransaction() (string, error) {
	builder := newBuilder().
		Func("getNextPendingTransaction")

	return c.sendTransaction(builder, ExecutionCost)
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
	return builder.Int(nonce)
}

func (builder *valueRequestBuilder) ActionId(actionId bridge.ActionId) *valueRequestBuilder {
	return builder.Int(actionId)
}

func (builder *valueRequestBuilder) Int(value *big.Int) *valueRequestBuilder {
	builder.args = append(builder.args, intToHex(value))

	return builder
}

func (builder *valueRequestBuilder) HexString(value string) *valueRequestBuilder {
	builder.args = append(builder.args, value)

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
	pkConv, _ := pubkeyConverter.NewBech32PubkeyConverter(32)
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

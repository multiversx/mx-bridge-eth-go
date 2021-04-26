package eth

import (
	"context"
	"reflect"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"

	logger "github.com/ElrondNetwork/elrond-go-logger"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridge"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type EthContract interface {
	GetNextPendingTransaction(opts *bind.CallOpts) (Deposit, error)
}

type Client struct {
	contract EthContract
	log      logger.Logger
}

func NewClient(config bridge.Config) (*Client, error) {
	log := logger.GetOrCreate("EthClient")

	ethClient, err := ethclient.Dial(config.NetworkAddress)
	if err != nil {
		return nil, err
	}

	instance, err := NewContract(common.HexToAddress(config.BridgeAddress), ethClient)
	if err != nil {
		return nil, err
	}

	client := &Client{
		contract: instance,
		log:      log,
	}

	return client, nil
}

func (c *Client) GetPendingDepositTransaction(ctx context.Context) *bridge.DepositTransaction {
	deposit, err := c.contract.GetNextPendingTransaction(&bind.CallOpts{Context: ctx})
	if err != nil {
		c.log.Error(err.Error())
		return nil
	}

	var result *bridge.DepositTransaction
	if !reflect.DeepEqual(deposit.Depositor, common.Address{}) {
		result = &bridge.DepositTransaction{
			To:           string(deposit.Recipient),
			From:         deposit.Depositor.String(),
			TokenAddress: deposit.TokenAddress.String(),
			Amount:       deposit.Amount,
			DepositNonce: bridge.Nonce(deposit.Nonce.Uint64()),
		}
	}

	return result
}

func (c *Client) ProposeTransfer(context.Context, *bridge.DepositTransaction) (string, error) {
	return "", nil
}

func (c *Client) ProposeSetStatusSuccessOnPendingTransfer(context.Context) {
}

func (c *Client) ProposeSetStatusFailedOnPendingTransfer(context.Context) {
}

func (c *Client) WasProposedTransfer(context.Context, bridge.Nonce) bool {
	return false
}

func (c *Client) GetActionIdForProposeTransfer(context.Context, bridge.Nonce) bridge.ActionId {
	return bridge.ActionId(0)
}

func (c *Client) WasProposedSetStatusSuccessOnPendingTransfer(context.Context) bool {
	return false
}

func (c *Client) WasProposedSetStatusFailedOnPendingTransfer(context.Context) bool {
	return false
}

func (c *Client) GetActionIdForSetStatusOnPendingTransfer(context.Context) bridge.ActionId {
	return bridge.ActionId(0)
}

func (c *Client) WasExecuted(context.Context, bridge.ActionId) bool {
	return false
}

func (c *Client) Sign(context.Context, bridge.ActionId) (string, error) {
	return "", nil
}

func (c *Client) Execute(context.Context, bridge.ActionId) (string, error) {
	// finishCurrentPendingTransaction([])
	return "tx_hash", nil
}

func (c *Client) SignersCount(context.Context, bridge.ActionId) uint {
	return 0
}

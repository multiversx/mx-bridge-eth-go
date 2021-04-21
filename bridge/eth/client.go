package eth

import (
	"context"
	"strings"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridge"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

const safeAbiDefinition = `[{"anonymous": false,"inputs": [{"indexed": false,"internalType": "address","name": "tokenAddress","type": "address"},{"indexed": false,"internalType": "address","name": "depositor","type": "address"},{"indexed": false,"internalType": "uint256","name": "amount","type": "uint256"}],"name": "ERC20Deposited","type": "event"},{"inputs": [{"internalType": "address","name": "tokenAddress","type": "address"},{"internalType": "uint256","name": "amount","type": "uint256"}],"name": "deposit","outputs": [],"stateMutability": "nonpayable","type": "function"}]`

type Client struct {
	contractCaller ethereum.ContractCaller
	bridgeAddress  common.Address
	bridgeAbi      abi.ABI
}

func NewClient(config bridge.Config) (*Client, error) {
	ethClient, err := ethclient.Dial(config.NetworkAddress)
	if err != nil {
		return nil, err
	}

	bridgeAbi, err := abi.JSON(strings.NewReader(safeAbiDefinition))
	if err != nil {
		return nil, err
	}

	client := &Client{
		contractCaller: ethClient,
		bridgeAddress:  common.HexToAddress(config.BridgeAddress),
		bridgeAbi:      bridgeAbi,
	}

	return client, nil
}

func (c *Client) GetPendingDepositTransaction(context.Context) *bridge.DepositTransaction {
	// GetPendingDepositTransaction empty block -> status: 0
	// None: 0, Pending: 1
	return nil
}

func (c *Client) ProposeTransfer(context.Context, *bridge.DepositTransaction) {
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

func (c *Client) Sign(context.Context, bridge.ActionId) {
}

func (c *Client) Execute(context.Context, bridge.ActionId) (string, error) {
	return "tx_hash", nil
}

func (c *Client) SignersCount(context.Context, bridge.ActionId) uint {
	return 0
}

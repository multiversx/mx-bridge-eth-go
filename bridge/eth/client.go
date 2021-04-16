package eth

import (
	"context"
	"github.com/ElrondNetwork/elrond-eth-bridge/bridge"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"strings"
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
	return nil
}

func (c *Client) Propose(*bridge.DepositTransaction) {
}

func (c *Client) WasProposed(*bridge.DepositTransaction) bool {
	return false
}

func (c *Client) WasExecuted(*bridge.DepositTransaction) bool {
	return false
}

func (c *Client) Sign(*bridge.DepositTransaction) {
}

func (c *Client) Execute(*bridge.DepositTransaction) (string, error) {
	return "tx_hash", nil
}

func (c *Client) SignersCount(*bridge.DepositTransaction) uint {
	return 0
}

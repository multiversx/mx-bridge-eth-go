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

const safeAbiDefinition = `[{"inputs": [{"internalType": "address[]","name": "board","type": "address[]"},{"internalType": "uint256","name": "intialQuorum","type": "uint256"},{"internalType": "address","name": "erc20Safe","type": "address"}],"stateMutability": "nonpayable","type": "constructor"},{"anonymous": false,"inputs": [{"indexed": false,"internalType": "address","name": "newRelayer","type": "address"}],"name": "RelayerAdded","type": "event"},{"anonymous": false,"inputs": [{"indexed": true,"internalType": "bytes32","name": "role","type": "bytes32"},{"indexed": true,"internalType": "bytes32","name": "previousAdminRole","type": "bytes32"},{"indexed": true,"internalType": "bytes32","name": "newAdminRole","type": "bytes32"}],"name": "RoleAdminChanged","type": "event"},{"anonymous": false,"inputs": [{"indexed": true,"internalType": "bytes32","name": "role","type": "bytes32"},{"indexed": true,"internalType": "address","name": "account","type": "address"},{"indexed": true,"internalType": "address","name": "sender","type": "address"}],"name": "RoleGranted","type": "event"},{"anonymous": false,"inputs": [{"indexed": true,"internalType": "bytes32","name": "role","type": "bytes32"},{"indexed": true,"internalType": "address","name": "account","type": "address"},{"indexed": true,"internalType": "address","name": "sender","type": "address"}],"name": "RoleRevoked","type": "event"},{"inputs": [],"name": "DEFAULT_ADMIN_ROLE","outputs": [{"internalType": "bytes32","name": "","type": "bytes32"}],"stateMutability": "view","type": "function"},{"inputs": [],"name": "RELAYER_ROLE","outputs": [{"internalType": "bytes32","name": "","type": "bytes32"}],"stateMutability": "view","type": "function"},{"inputs": [],"name": "_quorum","outputs": [{"internalType": "uint256","name": "","type": "uint256"}],"stateMutability": "view","type": "function"},{"inputs": [{"internalType": "address","name": "newRelayerAddress","type": "address"}],"name": "addRelayer","outputs": [],"stateMutability": "nonpayable","type": "function"},{"inputs": [{"internalType": "enum DepositStatus","name": "status","type": "uint8"}],"name": "finishCurrentPendingTransaction","outputs": [],"stateMutability": "nonpayable","type": "function"},{"inputs": [],"name": "getNextPendingTransaction","outputs": [{"components": [{"internalType": "uint256","name": "nonce","type": "uint256"},{"internalType": "address","name": "tokenAddress","type": "address"},{"internalType": "uint256","name": "amount","type": "uint256"},{"internalType": "address","name": "depositor","type": "address"},{"internalType": "bytes","name": "recipient","type": "bytes"},{"internalType": "enum DepositStatus","name": "status","type": "uint8"}],"internalType": "struct Deposit","name": "","type": "tuple"}],"stateMutability": "view","type": "function"},{"inputs": [{"internalType": "bytes32","name": "role","type": "bytes32"}],"name": "getRoleAdmin","outputs": [{"internalType": "bytes32","name": "","type": "bytes32"}],"stateMutability": "view","type": "function"},{"inputs": [{"internalType": "bytes32","name": "role","type": "bytes32"},{"internalType": "address","name": "account","type": "address"}],"name": "grantRole","outputs": [],"stateMutability": "nonpayable","type": "function"},{"inputs": [{"internalType": "bytes32","name": "role","type": "bytes32"},{"internalType": "address","name": "account","type": "address"}],"name": "hasRole","outputs": [{"internalType": "bool","name": "","type": "bool"}],"stateMutability": "view","type": "function"},{"inputs": [{"internalType": "bytes32","name": "role","type": "bytes32"},{"internalType": "address","name": "account","type": "address"}],"name": "renounceRole","outputs": [],"stateMutability": "nonpayable","type": "function"},{"inputs": [{"internalType": "bytes32","name": "role","type": "bytes32"},{"internalType": "address","name": "account","type": "address"}],"name": "revokeRole","outputs": [],"stateMutability": "nonpayable","type": "function"},{"inputs": [{"internalType": "uint256","name": "newQorum","type": "uint256"}],"name": "setQuorum","outputs": [],"stateMutability": "nonpayable","type": "function"},{"inputs": [{"internalType": "bytes4","name": "interfaceId","type": "bytes4"}],"name": "supportsInterface","outputs": [{"internalType": "bool","name": "","type": "bool"}],"stateMutability": "view","type": "function"}]`

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
	// getNextPendingTransaction empty block -> status: 0
	// None: 0, Pending: 1
	return nil
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

package elrond

import (
	"context"
	"github.com/ElrondNetwork/elrond-eth-bridge/safe"
	"math/big"
)

type Client struct{}

func NewClient() (*Client, error) {
	return &Client{}, nil
}

func (c *Client) Bridge(*safe.DepositTransaction) {
	// TODO: send transaction to safe
}

func (c *Client) GetTransactions(context.Context, *big.Int, safe.SafeTxChan) {
	// TODO: follow the pattern in eth to get blocks -> transactions to the safe contract
}

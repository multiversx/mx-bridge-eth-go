package elrond

import (
	"context"
	"github.com/ElrondNetwork/elrond-eth-bridge/safe"
)

type Client struct{}

func (c *Client) GetTransactions(context.Context, uint64) safe.SafeTxChan {
	// TODO: follow the pattern in eth to get blocks -> transactions to the safe contract
	return make(safe.SafeTxChan)
}

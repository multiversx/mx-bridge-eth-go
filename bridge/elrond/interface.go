package elrond

import (
	"github.com/ElrondNetwork/elrond-sdk-erdgo/core"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/data"
)

// NonceTransactionsHandler represents the interface able to handle the current nonce and the transactions resend mechanism
type NonceTransactionsHandler interface {
	GetNonce(address core.AddressHandler) (uint64, error)
	SendTransaction(tx *data.Transaction) (string, error)
	Close() error
}

package batchProcessor

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/multiversx/mx-bridge-eth-go/clients"
	"math/big"
)

// ArgListsBatch is a struct that contains the batch data in a format that is easy to use
type ArgListsBatch struct {
	Tokens              []common.Address
	Recipients          []common.Address
	ConvertedTokenBytes [][]byte
	Amounts             []*big.Int
	Nonces              []*big.Int
}

// ExtractList will extract the batch data into a format that is easy to use
func ExtractList(batch *clients.TransferBatch) (*ArgListsBatch, error) {
	arg := ArgListsBatch{}

	for _, dt := range batch.Deposits {
		recipient := common.BytesToAddress(dt.ToBytes)
		arg.Recipients = append(arg.Recipients, recipient)

		token := common.BytesToAddress(dt.ConvertedTokenBytes)
		arg.Tokens = append(arg.Tokens, token)

		amount := big.NewInt(0).Set(dt.Amount)
		arg.Amounts = append(arg.Amounts, amount)

		nonce := big.NewInt(0).SetUint64(dt.Nonce)
		arg.Nonces = append(arg.Nonces, nonce)

		arg.ConvertedTokenBytes = append(arg.ConvertedTokenBytes, dt.ConvertedTokenBytes)
	}

	return &arg, nil
}
package batchProcessor

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/multiversx/mx-bridge-eth-go/clients"
)

// ArgListsBatch is a struct that contains the batch data in a format that is easy to use
type ArgListsBatch struct {
	EthTokens     []common.Address
	Recipients    []common.Address
	MvxTokenBytes [][]byte
	Amounts       []*big.Int
	Nonces        []*big.Int
}

// ExtractListMvxToEth will extract the batch data into a format that is easy to use
// The transfer is from MultiversX to Ethereum
func ExtractListMvxToEth(batch *clients.TransferBatch) (*ArgListsBatch, error) {
	arg := ArgListsBatch{}

	for _, dt := range batch.Deposits {
		recipient := common.BytesToAddress(dt.ToBytes)
		arg.Recipients = append(arg.Recipients, recipient)

		token := common.BytesToAddress(dt.DestinationTokenBytes)
		arg.EthTokens = append(arg.EthTokens, token)

		amount := big.NewInt(0).Set(dt.Amount)
		arg.Amounts = append(arg.Amounts, amount)

		nonce := big.NewInt(0).SetUint64(dt.Nonce)
		arg.Nonces = append(arg.Nonces, nonce)

		arg.MvxTokenBytes = append(arg.MvxTokenBytes, dt.SourceTokenBytes)
	}

	return &arg, nil
}

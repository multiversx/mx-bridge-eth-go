package batchProcessor

import (
	"math/big"

	bridgeCore "github.com/multiversx/mx-bridge-eth-go/core"
)

// Direction is the direction of the transfer
type Direction string

const (
	// FromMultiversX is the direction of the transfer
	FromMultiversX Direction = "FromMultiversX"
	// ToMultiversX is the direction of the transfer
	ToMultiversX Direction = "ToMultiversX"
)

// ArgListsBatch is a struct that contains the batch data in a format that is easy to use
type ArgListsBatch struct {
	EthTokens     [][]byte
	Recipients    [][]byte
	MvxTokenBytes [][]byte
	Amounts       []*big.Int
	Nonces        []*big.Int
	Direction     Direction
}

// ArgListsBatchSui is a struct that contains the batch data in a format that is easy to use for Sui
type ArgListsBatchSui struct {
	SuiTokens     [][]byte
	Recipients    []string
	MvxTokenBytes [][]byte
	Amounts       []uint64
	Nonces        []uint64
	Direction     Direction
}

// SuiTransferData is a struct that contains the transfer data to be signed
type SuiTransferData struct {
	Recipients [][]byte
	SuiTokens  [][]byte
	Amounts    []uint64
	Nonces     []uint64
	BatchId    uint64
}

// ExtractListMvxToEth will extract the batch data into a format that is easy to use
// The transfer is from MultiversX to Ethereum
func ExtractListMvxToEth(batch *bridgeCore.TransferBatch) *ArgListsBatch {
	arg := &ArgListsBatch{
		Direction: FromMultiversX,
	}

	for _, dt := range batch.Deposits {
		recipient := dt.ToBytes
		arg.Recipients = append(arg.Recipients, recipient)

		token := dt.DestinationTokenBytes
		arg.EthTokens = append(arg.EthTokens, token)

		amount := big.NewInt(0).Set(dt.Amount)
		arg.Amounts = append(arg.Amounts, amount)

		nonce := big.NewInt(0).SetUint64(dt.Nonce)
		arg.Nonces = append(arg.Nonces, nonce)

		arg.MvxTokenBytes = append(arg.MvxTokenBytes, dt.SourceTokenBytes)
	}

	return arg
}

// ExtractListEthToMvx will extract the batch data into a format that is easy to use
// The transfer is from Ehtereum to MultiversX
func ExtractListEthToMvx(batch *bridgeCore.TransferBatch) *ArgListsBatch {
	arg := &ArgListsBatch{
		Direction: ToMultiversX,
	}

	for _, dt := range batch.Deposits {
		recipient := dt.ToBytes
		arg.Recipients = append(arg.Recipients, recipient)

		token := dt.SourceTokenBytes
		arg.EthTokens = append(arg.EthTokens, token)

		amount := big.NewInt(0).Set(dt.Amount)
		arg.Amounts = append(arg.Amounts, amount)

		nonce := big.NewInt(0).SetUint64(dt.Nonce)
		arg.Nonces = append(arg.Nonces, nonce)

		arg.MvxTokenBytes = append(arg.MvxTokenBytes, dt.DestinationTokenBytes)
	}

	return arg
}

package testsCommon

import (
	"crypto/rand"

	erdgoCore "github.com/ElrondNetwork/elrond-sdk-erdgo/core"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/data"
	"github.com/ethereum/go-ethereum/common"
)

// CreateRandomEthereumAddress will create a random Ethereum address
func CreateRandomEthereumAddress() common.Address {
	buff := make([]byte, len(common.Address{}))
	_, _ = rand.Read(buff)

	return common.BytesToAddress(buff)
}

// CreateRandomElrondAddress will create a random Elrond address
func CreateRandomElrondAddress() erdgoCore.AddressHandler {
	buff := make([]byte, 32)
	_, _ = rand.Read(buff)

	return data.NewAddressFromBytes(buff)
}

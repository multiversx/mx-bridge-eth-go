package testsCommon

import (
	"crypto/rand"

	"github.com/ethereum/go-ethereum/common"
	sdkCore "github.com/multiversx/mx-sdk-go/core"
	"github.com/multiversx/mx-sdk-go/data"
)

// CreateRandomEthereumAddress will create a random Ethereum address
func CreateRandomEthereumAddress() common.Address {
	buff := make([]byte, len(common.Address{}))
	_, _ = rand.Read(buff)

	return common.BytesToAddress(buff)
}

// CreateRandomMultiversXAddress will create a random MultiversX address
func CreateRandomMultiversXAddress() sdkCore.AddressHandler {
	buff := make([]byte, 32)
	_, _ = rand.Read(buff)

	return data.NewAddressFromBytes(buff)
}

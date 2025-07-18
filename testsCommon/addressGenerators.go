package testsCommon

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

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

// CreateRandomMultiversXSCAddress will create a random MultiversX smart contract address
func CreateRandomMultiversXSCAddress() sdkCore.AddressHandler {
	buff := make([]byte, 22)
	_, _ = rand.Read(buff)

	firstPart := append(make([]byte, 8), []byte{5, 0}...)

	return data.NewAddressFromBytes(append(firstPart, buff...))
}

// CreateRandomSuiAddressBytes will create a random Sui smart contract (object) or user address bytes
func CreateRandomSuiAddressBytes() [32]byte {
	buff := make([]byte, 32)
	_, _ = rand.Read(buff)

	return [32]byte(buff)
}

// CreateRandomCoinId will create a random Sui coin ID
func CreateRandomCoinId() string {
	addrBytes := CreateRandomSuiAddressBytes()
	return fmt.Sprintf("0x%s::coin::COIN", hex.EncodeToString(addrBytes[:]))
}

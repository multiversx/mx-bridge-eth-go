package parsers

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/multiversx/mx-sdk-go/core"
)

// CallData defines the struct holding SC call data parameters
type CallData struct {
	Type      byte
	Function  string
	GasLimit  uint64
	Arguments []string
}

// ProxySCCompleteCallData defines the struct holding Proxy SC complete call data
type ProxySCCompleteCallData struct {
	CallData
	From   common.Address
	To     core.AddressHandler
	Token  string
	Amount *big.Int
	Nonce  uint64
}

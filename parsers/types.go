package parsers

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/multiversx/mx-chain-core-go/core/check"
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
	RawCallData []byte
	From        common.Address
	To          core.AddressHandler
	Token       string
	Amount      *big.Int
	Nonce       uint64
}

// String returns the human-readable string version of the call data
func (callData ProxySCCompleteCallData) String() string {
	toString := "<nil>"
	var err error
	if !check.IfNil(callData.To) {
		toString, err = callData.To.AddressAsBech32String()
		if err != nil {
			toString = "<err>"
		}
	}
	amountString := "<nil>"
	if callData.Amount != nil {
		amountString = callData.Amount.String()
	}

	return fmt.Sprintf("Eth address: %s, MvX address: %s, token: %s, amount: %s, nonce: %d, raw call data: %x",
		callData.From.String(),
		toString,
		callData.Token,
		amountString,
		callData.Nonce,
		callData.RawCallData,
	)
}

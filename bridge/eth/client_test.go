package eth

import (
	"context"
	"github.com/ElrondNetwork/elrond-eth-bridge/bridge"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"testing"
)

// verify Client implements interface
var (
	_ = bridge.Bridge(&Client{})
)

type testContractCaller struct{}

func (c *testContractCaller) CallContract(context.Context, ethereum.CallMsg, *big.Int) ([]byte, error) {
	return nil, nil
}

func TestGetTransaction(t *testing.T) {
	_ = Client{
		contractCaller: &testContractCaller{},
		bridgeAddress:  common.Address{},
		bridgeAbi:      abi.ABI{},
	}
}

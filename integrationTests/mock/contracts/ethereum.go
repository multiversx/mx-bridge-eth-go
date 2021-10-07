package contracts

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridge/eth"
	"github.com/ElrondNetwork/elrond-eth-bridge/integrationTests/mock"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

const abiStructJson = `[
{"inputs":[],"name":"getNextPendingBatch","outputs":[{"components":[{"internalType":"uint256","name":"nonce","type":"uint256"},{"internalType":"uint256","name":"timestamp","type":"uint256"},{"internalType":"uint256","name":"lastUpdatedBlockNumber","type":"uint256"},{"components":[{"internalType":"uint256","name":"nonce","type":"uint256"},{"internalType":"address","name":"tokenAddress","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"},{"internalType":"address","name":"depositor","type":"address"},{"internalType":"bytes","name":"recipient","type":"bytes"},{"internalType":"enumDepositStatus","name":"status","type":"uint8"}],"internalType":"structDeposit[]","name":"deposits","type":"tuple[]"}],"internalType":"structBatch","name":"","type":"tuple"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"bytes32","name":"role","type":"bytes32"}],"name":"getRoleAdmin","outputs":[{"internalType":"bytes32","name":"","type":"bytes32"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"bytes32","name":"role","type":"bytes32"},{"internalType":"address","name":"account","type":"address"}],"name":"grantRole","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"bytes32","name":"role","type":"bytes32"},{"internalType":"address","name":"account","type":"address"}],"name":"hasRole","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"quorum","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"relayerAddress","type":"address"}],"name":"removeRelayer","outputs":[],"stateMutability":"nonpayable","type":"function"}
]
`

// EthereumContract extends the contract implementation with the Elrond contract logic
type EthereumContract struct {
	*mock.Contract
	abiInstance abi.ABI
}

// NewEthereumContract defines the mocked Etherum contract functions
func NewEthereumContract(address string) (*EthereumContract, error) {
	ec := &EthereumContract{
		Contract: mock.NewContract(address),
	}

	abiInstance, err := abi.JSON(strings.NewReader(abiStructJson))
	if err != nil {
		return nil, err
	}
	ec.abiInstance = abiInstance

	ec.createContractFunctions()

	return ec, nil
}

func (ec *EthereumContract) createContractFunctions() {
	ec.AddHandler("eth_getCode", ec.getCode)
	ec.AddHandler("0xc073de1f", ec.getNextPendingBatch)
}

// should return ((uint256,uint256,uint256,(uint256,address,uint256,address,bytes,uint8)[]))
func (ec *EthereumContract) getNextPendingBatch(_ string, _ string, _ ...string) ([][]byte, error) {
	b := &eth.Batch{
		Nonce:                  big.NewInt(1),
		Timestamp:              big.NewInt(2),
		LastUpdatedBlockNumber: big.NewInt(3),
		Deposits: []eth.Deposit{
			{
				Nonce:        big.NewInt(4),
				TokenAddress: common.BytesToAddress([]byte("12345678901234567890")),
				Amount:       big.NewInt(5),
				Depositor:    common.BytesToAddress([]byte("23456789012345678901")),
				Recipient:    []byte("recipient1"),
				Status:       1,
			},
			{
				Nonce:        big.NewInt(6),
				TokenAddress: common.BytesToAddress([]byte("34567890123456789012")),
				Amount:       big.NewInt(7),
				Depositor:    common.BytesToAddress([]byte("45678901234567890123")),
				Recipient:    []byte("recipient2"),
				Status:       2,
			},
		},
	}

	method := ec.abiInstance.Methods["getNextPendingBatch"]
	buff, err := method.Outputs.Pack(b)

	return [][]byte{buff}, err
}

func (ec *EthereumContract) getCode(caller string, value string, arguments ...string) ([][]byte, error) {
	log.Warn("getCode", "caller", caller, "value", value, "arguments", fmt.Sprintf("%v", arguments))

	return [][]byte{make([]byte, 65535)}, nil
}

package contracts

import (
	"fmt"
	"math/big"
	"strings"
	"sync"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridge/eth"
	"github.com/ElrondNetwork/elrond-eth-bridge/integrationTests/mock"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/core"
	"github.com/ethereum/go-ethereum/accounts/abi"
)

// EthereumContract extends the contract implementation with the Elrond contract logic
type EthereumContract struct {
	*mock.Contract
	abiInstance      abi.ABI
	mutState         sync.Mutex
	nonce            int
	transferRequests []*transferRequest
}

// NewEthereumContract defines the mocked Etherum contract functions
func NewEthereumContract(address string) (*EthereumContract, error) {
	ec := &EthereumContract{
		Contract:         mock.NewContract(address),
		transferRequests: make([]*transferRequest, 0),
	}

	ethereumABIData := eth.BridgeMetaData.ABI
	abiInstance, err := abi.JSON(strings.NewReader(ethereumABIData))
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
	ec.mutState.Lock()
	defer ec.mutState.Unlock()

	b := &eth.Batch{
		Nonce:                  big.NewInt(int64(ec.nonce)),
		Timestamp:              big.NewInt(1),
		LastUpdatedBlockNumber: big.NewInt(2),
		Deposits:               make([]eth.Deposit, 0),
	}
	ec.nonce++

	for _, tr := range ec.transferRequests {
		b.Deposits = append(b.Deposits, tr.toEthDepositInfo())
	}

	log.Debug("EthereumContract.getNextPendingBatch prepared deposit info", "num deposits", len(ec.transferRequests))
	ec.transferRequests = make([]*transferRequest, 0)

	method := ec.abiInstance.Methods["getNextPendingBatch"]
	buff, err := method.Outputs.Pack(b)

	return [][]byte{buff}, err
}

// AddTransferRequest will store the transfer request up until the next getNextPendingBatch function call
func (ec *EthereumContract) AddTransferRequest(tokenAddress []byte, depositor []byte, to core.AddressHandler, amount *big.Int) {
	ec.mutState.Lock()
	defer ec.mutState.Unlock()

	tr := &transferRequest{
		depositor:    depositor,
		to:           to,
		tokenAddress: tokenAddress,
		amount:       amount,
		nonce:        ec.nonce,
	}
	ec.nonce++

	ec.transferRequests = append(ec.transferRequests, tr)

	log.Debug("EthereumContract.AddTransferRequest",
		"token address", tokenAddress, "depositor", depositor, "to", to.AddressAsBech32String(),
		"amount", amount.String())
}

func (ec *EthereumContract) getCode(caller string, value string, arguments ...string) ([][]byte, error) {
	log.Warn("getCode", "caller", caller, "value", value, "arguments", fmt.Sprintf("%v", arguments))

	return [][]byte{make([]byte, 65535)}, nil
}

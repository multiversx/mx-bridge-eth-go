package contracts

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/ElrondNetwork/elrond-eth-bridge/integrationTests/mock"
	logger "github.com/ElrondNetwork/elrond-go-logger"
)

var log = logger.GetOrCreate("integrationTests/mock/constracts")

const canProposeAndSign = 2

// ElrondContract extends the contract implementation with the Elrond contract logic
type ElrondContract struct {
	*mock.Contract
	mutWhitelisted       sync.RWMutex
	whitelistedAddresses map[string]struct{}
}

// NewElrondContract defines the mocked Elrond contract functions
func NewElrondContract(address string) *ElrondContract {
	ec := &ElrondContract{
		Contract:             mock.NewContract(address),
		whitelistedAddresses: make(map[string]struct{}),
	}

	ec.createContractFunctions()

	return ec
}

func (ec *ElrondContract) createContractFunctions() {
	ec.AddHandler("getCurrentTxBatch", ec.getCurrentTxBatch)
	ec.AddHandler("userRole", ec.userRole)
}

func (ec *ElrondContract) getCurrentTxBatch(caller string, value string, arguments ...string) ([][]byte, error) {
	log.Warn("getCurrentTxBatch", "caller", caller, "value", value, "arguments", fmt.Sprintf("%v", arguments))

	return make([][]byte, 0), nil
}

func (ec *ElrondContract) userRole(caller string, value string, arguments ...string) ([][]byte, error) {
	log.Warn("userRole", "caller", caller, "value", value, "arguments", fmt.Sprintf("%v", arguments))

	ec.mutWhitelisted.RLock()
	_, isWhiteListed := ec.whitelistedAddresses[caller]
	ec.mutWhitelisted.RUnlock()

	userRole := big.NewInt(0).Bytes()
	if isWhiteListed {
		userRole = big.NewInt(canProposeAndSign).Bytes()
	}

	return [][]byte{userRole}, nil
}

// WhiteListAddress will mark the provided address as whitelisted (so it can propose and sign)
func (ec *ElrondContract) WhiteListAddress(address string) {
	ec.mutWhitelisted.Lock()
	ec.whitelistedAddresses[address] = struct{}{}
	ec.mutWhitelisted.Unlock()
}

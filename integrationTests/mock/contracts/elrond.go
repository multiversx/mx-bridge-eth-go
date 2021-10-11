package contracts

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"sync"

	"github.com/ElrondNetwork/elrond-eth-bridge/integrationTests/mock"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	logger "github.com/ElrondNetwork/elrond-go-logger"
)

var log = logger.GetOrCreate("integrationTests/mock/contracts")

const canProposeAndSign = 2

// ElrondContract extends the contract implementation with the Elrond contract logic
type ElrondContract struct {
	*mock.Contract
	mutWhitelisted       sync.RWMutex
	whitelistedAddresses map[string]struct{}
	tokensHandler        TokensHandler
}

// NewElrondContract defines the mocked Elrond contract functions
func NewElrondContract(address string, th TokensHandler) (*ElrondContract, error) {
	if check.IfNil(th) {
		return nil, ErrNilTokenHandler
	}

	ec := &ElrondContract{
		Contract:             mock.NewContract(address),
		whitelistedAddresses: make(map[string]struct{}),
		tokensHandler:        th,
	}

	ec.createContractFunctions()

	return ec, nil
}

func (ec *ElrondContract) createContractFunctions() {
	ec.AddHandler("getCurrentTxBatch", ec.getCurrentTxBatch)
	ec.AddHandler("userRole", ec.userRole)
	ec.AddHandler("getTokenIdForErc20Address", ec.getTokenIdForErc20Address)
}

func (ec *ElrondContract) getCurrentTxBatch(caller string, value string, arguments ...string) ([][]byte, error) {
	log.Debug("getCurrentTxBatch", "caller", caller, "value", value, "arguments", arguments)

	return make([][]byte, 0), nil
}

func (ec *ElrondContract) userRole(caller string, value string, arguments ...string) ([][]byte, error) {
	log.Debug("userRole", "caller", caller, "value", value, "arguments", arguments)

	ec.mutWhitelisted.RLock()
	_, isWhiteListed := ec.whitelistedAddresses[caller]
	ec.mutWhitelisted.RUnlock()

	userRole := big.NewInt(0).Bytes()
	if isWhiteListed {
		userRole = big.NewInt(canProposeAndSign).Bytes()
	}

	return [][]byte{userRole}, nil
}

func (ec *ElrondContract) getTokenIdForErc20Address(caller string, value string, arguments ...string) ([][]byte, error) {
	log.Debug("getTokenIdForErc20Address", "caller", caller, "value", value, "arguments", arguments)

	if len(arguments) != 1 {
		return nil, fmt.Errorf("%w - expected 1 argument", ErrInvalidNumberOfArguments)
	}

	address, err := hex.DecodeString(arguments[0])
	if err != nil {
		return nil, err
	}

	ticker, err := ec.tokensHandler.GetTickerFromEthAddress(address)
	if err != nil {
		return nil, err
	}

	return [][]byte{[]byte(ticker)}, nil
}

// WhiteListAddress will mark the provided address as whitelisted (so it can propose and sign)
func (ec *ElrondContract) WhiteListAddress(address string) {
	ec.mutWhitelisted.Lock()
	ec.whitelistedAddresses[address] = struct{}{}
	ec.mutWhitelisted.Unlock()
}

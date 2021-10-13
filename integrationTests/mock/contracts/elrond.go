package contracts

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"sync"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridge"
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
	actionMapper         map[*big.Int]*big.Int
	currentTxBatch       *bridge.Batch
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
		actionMapper:         make(map[*big.Int]*big.Int),
	}

	ec.createContractFunctions()

	return ec, nil
}

func (ec *ElrondContract) createContractFunctions() {
	ec.AddHandler("getCurrentTxBatch", ec.getCurrentTxBatch)
	ec.AddHandler("userRole", ec.userRole)
	ec.AddHandler("getTokenIdForErc20Address", ec.getTokenIdForErc20Address)
	ec.AddHandler("wasTransferActionProposed", ec.wasTransferActionProposed)
	ec.AddHandler("proposeMultiTransferEsdtBatch", ec.proposeMultiTransferEsdtBatch)
	ec.AddHandler("getActionSignerCount", ec.getActionSignerCount)
	ec.AddHandler("proposeEsdtSafeSetCurrentTransactionBatchStatus", ec.proposeEsdtSafeSetCurrentTransactionBatchStatus)
}

func (ec *ElrondContract) getCurrentTxBatch(caller string, value string, arguments ...string) ([][]byte, error) {
	log.Debug("getCurrentTxBatch", "caller", caller, "value", value, "arguments", arguments)

	return make([][]byte, 0), nil // ec.currentTxBatch
}

func (ec *ElrondContract) wasTransferActionProposed(caller string, value string, arguments ...string) ([][]byte, error) {
	log.Warn("wasTransferActionProposed", "caller", caller, "value", value, "arguments", fmt.Sprintf("%v", arguments))
	batchId, _ := strconv.ParseInt(arguments[0], 10, 64)
	ret := byte(1) // we assume we have the action in map
	if _, ok := ec.actionMapper[big.NewInt(batchId)]; !ok {
		ret = 0
	}
	return [][]byte{{ret}}, nil
}

func (ec *ElrondContract) proposeMultiTransferEsdtBatch(caller string, value string, arguments ...string) ([][]byte, error) {
	log.Warn("proposeMultiTransferEsdtBatch ", "caller", caller, "value", value, "arguments", fmt.Sprintf("%v", arguments))
	var args []string
	err := json.Unmarshal([]byte(arguments[0]), &args)
	if err != nil {
		log.Error("ElrondContract: Error unmarshal arguments", "error", err.Error())
		return make([][]byte, 0), nil
	}
	buf, _ := base64.URLEncoding.DecodeString(args[0])
	batchId := int64(buf[0])
	// every 3rd elements starting from 2nd
	// is empty string for separation so we skip any of them
	for i := 2; i < len(args); i += 3 {
		token, _ := base64.URLEncoding.DecodeString(args[i])
		amountInBytes, _ := base64.URLEncoding.DecodeString(args[i+1])
		log.Debug("token: ", token, "amount: ", amountInBytes[0])
	}
	log.Debug(strings.Join(args, ":"))

	ret := byte(1) // we assume we have the action in map
	if _, ok := ec.actionMapper[big.NewInt(batchId)]; !ok {
		ret = 0
		ec.actionMapper[big.NewInt(batchId)] = big.NewInt(0)
	}
	return [][]byte{{ret}}, nil
}

func (ec *ElrondContract) proposeEsdtSafeSetCurrentTransactionBatchStatus(caller string, value string, arguments ...string) ([][]byte, error) {
	log.Warn("proposeEsdtSafeSetCurrentTransactionBatchStatus ", "caller", caller, "value", value, "arguments", fmt.Sprintf("%v", arguments))

	ret := byte(1)
	return [][]byte{{ret}}, nil
}

func (ec *ElrondContract) getActionSignerCount(caller string, value string, arguments ...string) ([][]byte, error) {
	log.Warn("getActionSignerCount", "caller", caller, "value", value, "arguments", fmt.Sprintf("%v", arguments))
	batchId := int64(-1)
	_, err := fmt.Sscan(arguments[0], &batchId)
	if err != nil {
		log.Error("ElrondContract: Error parsing batchId", "error", err.Error())
	}
	ret := []byte{0}
	if _, ok := ec.actionMapper[big.NewInt(batchId)]; !ok {
		ret = ec.actionMapper[big.NewInt(batchId)].Bytes()
	}
	return [][]byte{ret}, nil
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

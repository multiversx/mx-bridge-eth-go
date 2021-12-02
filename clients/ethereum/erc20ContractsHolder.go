package ethereum

import (
	"context"
	"fmt"
	"math/big"
	"sync"

	"github.com/ElrondNetwork/elrond-eth-bridge/clients/ethereum/contract"
	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethCommon "github.com/ethereum/go-ethereum/common"
)

type ArgsErc20SafeContractsHolder struct {
	ethClient              bind.ContractBackend
	ethClientStatusHandler core.StatusHandler
}

// Erc20SafeContractsHolder represents the Erc20ContractsHolder implementation
type Erc20SafeContractsHolder struct {
	mut                    sync.RWMutex
	contracts              map[ethCommon.Address]*erc20ContractWrapper
	ethClient              bind.ContractBackend
	ethClientStatusHandler core.StatusHandler
}

// NewErc20SafeContractsHolder returns a new Erc20SafeContractsHolder instance
func NewErc20SafeContractsHolder(args ArgsErc20SafeContractsHolder) (*Erc20SafeContractsHolder, error) {
	if check.IfNilReflect(args.ethClient) {
		return nil, errNilEthClient
	}
	if check.IfNil(args.ethClientStatusHandler) {
		return nil, errNilStatusHandler
	}
	return &Erc20SafeContractsHolder{
		contracts:              make(map[ethCommon.Address]*erc20ContractWrapper),
		ethClient:              args.ethClient,
		ethClientStatusHandler: args.ethClientStatusHandler,
	}, nil
}

// BalanceOf returns the ERC20 balance of the provided address
// if the ERC20 contract does not exists in the map of contract wrappers, it will create and add it first
func (h *Erc20SafeContractsHolder) BalanceOf(ctx context.Context, erc20Address ethCommon.Address, address ethCommon.Address) (*big.Int, error) {
	h.mut.Lock()
	defer h.mut.Unlock()

	var wrapper *erc20ContractWrapper
	if wrapper, exists := h.contracts[erc20Address]; !exists {
		contractInstance, err := contract.NewGenericErc20(erc20Address, h.ethClient)
		if err != nil {
			return nil, fmt.Errorf("%w for %s", err, erc20Address.String())
		}
		args := ArgsErc20ContractWrapper{
			StatusHandler: h.ethClientStatusHandler,
			Erc20Contract: contractInstance,
		}
		wrapper, err = NewErc20ContractWrapper(args)
		if err != nil {
			return nil, err
		}

		h.contracts[erc20Address] = wrapper
	}
	h.ethClientStatusHandler.AddIntMetric(core.MetricNumEthClientRequests, 1)
	return wrapper.erc20Contract.BalanceOf(&bind.CallOpts{Context: ctx}, address)
}

// IsInterfaceNil returns true if there is no value under the interface
func (h *Erc20SafeContractsHolder) IsInterfaceNil() bool {
	return h == nil
}

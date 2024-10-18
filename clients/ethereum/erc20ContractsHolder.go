package ethereum

import (
	"context"
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/multiversx/mx-bridge-eth-go/clients"
	"github.com/multiversx/mx-bridge-eth-go/clients/ethereum/contract"
	"github.com/multiversx/mx-bridge-eth-go/clients/ethereum/wrappers"
	"github.com/multiversx/mx-bridge-eth-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
)

// ArgsErc20SafeContractsHolder is the argument DTO used in the NewErc20SafeContractsHolder function
type ArgsErc20SafeContractsHolder struct {
	EthClient              bind.ContractBackend
	EthClientStatusHandler core.StatusHandler
}

// erc20SafeContractsHolder represents the Erc20ContractsHolder implementation
type erc20SafeContractsHolder struct {
	mut                    sync.RWMutex
	contracts              map[ethCommon.Address]erc20ContractWrapper
	ethClient              bind.ContractBackend
	ethClientStatusHandler core.StatusHandler
}

// NewErc20SafeContractsHolder returns a new erc20SafeContractsHolder instance
func NewErc20SafeContractsHolder(args ArgsErc20SafeContractsHolder) (*erc20SafeContractsHolder, error) {
	if check.IfNilReflect(args.EthClient) {
		return nil, errNilEthClient
	}
	if check.IfNil(args.EthClientStatusHandler) {
		return nil, clients.ErrNilStatusHandler
	}
	return &erc20SafeContractsHolder{
		contracts:              make(map[ethCommon.Address]erc20ContractWrapper),
		ethClient:              args.EthClient,
		ethClientStatusHandler: args.EthClientStatusHandler,
	}, nil
}

// BalanceOf returns the ERC20 balance of the provided address
// if the ERC20 contract does not exists in the map of contract wrappers, it will create and add it first
func (h *erc20SafeContractsHolder) BalanceOf(ctx context.Context, erc20Address ethCommon.Address, address ethCommon.Address) (*big.Int, error) {
	h.mut.Lock()
	defer h.mut.Unlock()

	wrapper, exists := h.contracts[erc20Address]
	if !exists {
		contractInstance, err := contract.NewGenericERC20(erc20Address, h.ethClient)
		if err != nil {
			return nil, fmt.Errorf("%w for %s", err, erc20Address.String())
		}
		args := wrappers.ArgsErc20ContractWrapper{
			StatusHandler: h.ethClientStatusHandler,
			Erc20Contract: contractInstance,
		}
		wrapper, err = wrappers.NewErc20ContractWrapper(args)
		if err != nil {
			return nil, err
		}

		h.contracts[erc20Address] = wrapper
	}

	return wrapper.BalanceOf(ctx, address)
}

// IsInterfaceNil returns true if there is no value under the interface
func (h *erc20SafeContractsHolder) IsInterfaceNil() bool {
	return h == nil
}

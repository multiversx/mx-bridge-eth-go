package wrappers

import (
	"context"
	"math/big"

	"github.com/ElrondNetwork/elrond-eth-bridge/clients"
	"github.com/ElrondNetwork/elrond-eth-bridge/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

// ArgsErc20ContractWrapper is the DTO used to construct an erc20ContractWrapper instance
type ArgsErc20ContractWrapper struct {
	StatusHandler core.StatusHandler
	Erc20Contract genericErc20Contract
}

type erc20ContractWrapper struct {
	statusHandler core.StatusHandler
	erc20Contract genericErc20Contract
}

// NewErc20ContractWrapper creates a new instance of type erc20ContractWrapper
func NewErc20ContractWrapper(args ArgsErc20ContractWrapper) (*erc20ContractWrapper, error) {
	if check.IfNilReflect(args.Erc20Contract) {
		return nil, errNilErc20Contract
	}
	if check.IfNil(args.StatusHandler) {
		return nil, clients.ErrNilStatusHandler
	}

	return &erc20ContractWrapper{
		statusHandler: args.StatusHandler,
		erc20Contract: args.Erc20Contract,
	}, nil
}

// BalanceOf returns the ERC20 balance of the provided address
func (wrapper *erc20ContractWrapper) BalanceOf(ctx context.Context, account common.Address) (*big.Int, error) {
	wrapper.statusHandler.AddIntMetric(core.MetricNumEthClientRequests, 1)
	return wrapper.erc20Contract.BalanceOf(&bind.CallOpts{Context: ctx}, account)
}

// IsInterfaceNil returns true if there is no value under the interface
func (wrapper *erc20ContractWrapper) IsInterfaceNil() bool {
	return wrapper == nil
}

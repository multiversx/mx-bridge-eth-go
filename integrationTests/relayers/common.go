package relayers

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	bridgeTests "github.com/multiversx/mx-bridge-eth-go/testsCommon/bridge"
)

func createMockErc20ContractsHolder(tokens []common.Address, safeContractEthAddress common.Address, availableBalances []*big.Int) *bridgeTests.ERC20ContractsHolderStub {
	return &bridgeTests.ERC20ContractsHolderStub{
		BalanceOfCalled: func(ctx context.Context, erc20Address common.Address, address common.Address) (*big.Int, error) {
			for i, tk := range tokens {
				if tk != erc20Address {
					continue
				}

				if address == safeContractEthAddress {
					return availableBalances[i], nil
				}

				return big.NewInt(0), nil
			}

			return nil, fmt.Errorf("unregistered token %s", erc20Address.Hex())
		},
	}
}

func availableTokensMapToSlices(erc20Map map[common.Address]*big.Int) ([]common.Address, []*big.Int) {
	tokens := make([]common.Address, 0, len(erc20Map))
	availableBalances := make([]*big.Int, 0, len(erc20Map))

	for addr, val := range erc20Map {
		tokens = append(tokens, addr)
		availableBalances = append(availableBalances, val)
	}

	return tokens, availableBalances
}

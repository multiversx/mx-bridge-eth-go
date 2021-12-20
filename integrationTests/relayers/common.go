package relayers

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ElrondNetwork/elrond-eth-bridge/testsCommon/bridgeV2"
	"github.com/ethereum/go-ethereum/common"
)

func createMockErc20ContractsHolder(tokens []common.Address, safeContractEthAddress common.Address, availableBalances []*big.Int) *bridgeV2.ERC20ContractsHolderStub {
	return &bridgeV2.ERC20ContractsHolderStub{
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

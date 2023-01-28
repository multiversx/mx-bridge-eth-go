package mappers

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/multiversx/mx-bridge-eth-go/clients"
	"github.com/multiversx/mx-chain-core-go/core/check"
)

type multiversXToErc20 struct {
	dg DataGetter
}

// NewMultiversXToErc20Mapper returns a new instance of multiversXToErc20
func NewMultiversXToErc20Mapper(dg DataGetter) (*multiversXToErc20, error) {
	if check.IfNil(dg) {
		return nil, clients.ErrNilDataGetter
	}

	return &multiversXToErc20{
		dg: dg,
	}, nil
}

// ConvertToken will return erd token id given a specific erc20 address
func (mapper *multiversXToErc20) ConvertToken(ctx context.Context, sourceBytes []byte) ([]byte, error) {

	response, err := mapper.dg.GetERC20AddressForTokenId(ctx, sourceBytes)
	if err != nil {
		return nil, err
	}

	if len(response) == 0 {
		return nil, fmt.Errorf("%w for provided %s", errUnknownToken, hex.EncodeToString(sourceBytes))
	}

	return response[0], nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (mapper *multiversXToErc20) IsInterfaceNil() bool {
	return mapper == nil
}

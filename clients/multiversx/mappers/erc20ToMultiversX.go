package mappers

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/multiversx/mx-bridge-eth-go/clients"
	"github.com/multiversx/mx-chain-core-go/core/check"
)

type erc20ToMultiversX struct {
	dg DataGetter
}

// NewErc20ToMultiversXMapper returns a new instance of erc20ToMultiversX
func NewErc20ToMultiversXMapper(dg DataGetter) (*erc20ToMultiversX, error) {
	if check.IfNil(dg) {
		return nil, clients.ErrNilDataGetter
	}

	return &erc20ToMultiversX{
		dg: dg,
	}, nil
}

// ConvertToken will return erd token id given a specific erc20 address
func (mapper *erc20ToMultiversX) ConvertToken(ctx context.Context, sourceBytes []byte) ([]byte, error) {

	response, err := mapper.dg.GetTokenIdForErc20Address(ctx, sourceBytes)
	if err != nil {
		return nil, err
	}

	if len(response) == 0 {
		return nil, fmt.Errorf("%w for provided %s", errUnknownToken, hex.EncodeToString(sourceBytes))
	}

	return response[0], nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (mapper *erc20ToMultiversX) IsInterfaceNil() bool {
	return mapper == nil
}

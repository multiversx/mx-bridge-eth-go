package mappers

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/ElrondNetwork/elrond-eth-bridge/clients"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
)

type erc20ToElrond struct {
	dg DataGetter
}

// NewErc20ToElrondMapper returns a new instance of erc20ToElrond
func NewErc20ToElrondMapper(dg DataGetter) (*erc20ToElrond, error) {
	if check.IfNil(dg) {
		return nil, clients.ErrNilDataGetter
	}

	return &erc20ToElrond{
		dg: dg,
	}, nil
}

// ConvertToken will return erd token id given a specific erc20 address
func (mapper *erc20ToElrond) ConvertToken(ctx context.Context, sourceBytes []byte) ([]byte, error) {

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
func (mapper *erc20ToElrond) IsInterfaceNil() bool {
	return mapper == nil
}

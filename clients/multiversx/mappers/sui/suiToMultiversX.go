package sui

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/multiversx/mx-bridge-eth-go/clients"
	"github.com/multiversx/mx-bridge-eth-go/clients/multiversx/mappers"
	"github.com/multiversx/mx-chain-core-go/core/check"
)

type suiToMultiversX struct {
	dg mappers.DataGetter
}

// NewSuiToMultiversXMapper returns a new instance of erc20ToMultiversX
func NewSuiToMultiversXMapper(dg mappers.DataGetter) (*suiToMultiversX, error) {
	if check.IfNil(dg) {
		return nil, clients.ErrNilDataGetter
	}

	return &suiToMultiversX{
		dg: dg,
	}, nil
}

// ConvertToken will return erd token id given a specific sui coin type
func (mapper *suiToMultiversX) ConvertToken(ctx context.Context, sourceBytes []byte) ([]byte, error) {

	response, err := mapper.dg.GetTokenIdForSuiCoin(ctx, sourceBytes)
	if err != nil {
		return nil, err
	}

	if len(response) == 0 {
		return nil, fmt.Errorf("%w for provided %s", mappers.ErrUnknownToken, hex.EncodeToString(sourceBytes))
	}

	return response[0], nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (mapper *suiToMultiversX) IsInterfaceNil() bool {
	return mapper == nil
}

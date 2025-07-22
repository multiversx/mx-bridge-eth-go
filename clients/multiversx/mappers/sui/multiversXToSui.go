package sui

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/multiversx/mx-bridge-eth-go/clients"
	"github.com/multiversx/mx-bridge-eth-go/clients/multiversx/mappers"
	"github.com/multiversx/mx-chain-core-go/core/check"
)

type multiversXToSui struct {
	dg mappers.DataGetter
}

// NewMultiversXToSuiMapper returns a new instance of multiversXToSui
func NewMultiversXToSuiMapper(dg mappers.DataGetter) (*multiversXToSui, error) {
	if check.IfNil(dg) {
		return nil, clients.ErrNilDataGetter
	}

	return &multiversXToSui{
		dg: dg,
	}, nil
}

// ConvertToken will return erd token id given a specific sui coin type
func (mapper *multiversXToSui) ConvertToken(ctx context.Context, sourceBytes []byte) ([]byte, error) {

	response, err := mapper.dg.GetSuiCoinForTokenId(ctx, sourceBytes)
	if err != nil {
		return nil, err
	}

	if len(response) == 0 {
		return nil, fmt.Errorf("%w for provided %s", mappers.ErrUnknownToken, hex.EncodeToString(sourceBytes))
	}

	return response[0], nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (mapper *multiversXToSui) IsInterfaceNil() bool {
	return mapper == nil
}

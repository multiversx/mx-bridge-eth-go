package mappers

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
)

type elrondToErc20 struct {
	dg DataGetter
}

func NewElrondToErc20Mapper(dg DataGetter) (*elrondToErc20, error) {
	if check.IfNil(dg) {
		return nil, errNilDataGetter
	}

	return &elrondToErc20{
		dg: dg,
	}, nil
}

func (mapper *elrondToErc20) ConvertToken(ctx context.Context, sourceBytes []byte) ([]byte, error) {

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
func (mapper *elrondToErc20) IsInterfaceNil() bool {
	return mapper == nil
}

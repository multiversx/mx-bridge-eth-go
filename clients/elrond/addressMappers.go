package elrond

import (
	"context"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
)

type Erc20ToElrond struct {
	dg DataGetter
}

func NewErc20ToElrondMapper(dg DataGetter) (*Erc20ToElrond, error) {
	if check.IfNil(dg) {
		return nil, errNilDataGetter
	}

	return &Erc20ToElrond{
		dg: dg,
	}, nil
}

func (mapper *Erc20ToElrond) ConvertToken(ctx context.Context, sourceBytes []byte) ([]byte, error) {

	response, err := mapper.dg.GetTokenIdForErc20Address(ctx, sourceBytes)
	if err != nil {
		return nil, err
	}

	elrondAddress := []byte{}
	if len(response) > 0 {
		elrondAddress = response[0]
	}

	return elrondAddress, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (mapper *Erc20ToElrond) IsInterfaceNil() bool {
	return mapper == nil
}

type ElrondToErc20 struct {
	dg DataGetter
}

func NewElrondToErc20Mapper(dg DataGetter) (*ElrondToErc20, error) {
	if check.IfNil(dg) {
		return nil, errNilDataGetter
	}

	return &ElrondToErc20{
		dg: dg,
	}, nil
}

func (mapper *ElrondToErc20) ConvertToken(ctx context.Context, sourceBytes []byte) ([]byte, error) {

	response, err := mapper.dg.GetERC20AddressForTokenId(ctx, sourceBytes)
	if err != nil {
		return nil, err
	}

	erc20Address := []byte{}
	if len(response) > 0 {
		erc20Address = response[0]
	}

	return erc20Address, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (mapper *ElrondToErc20) IsInterfaceNil() bool {
	return mapper == nil
}

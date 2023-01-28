package mappers

import "context"

// DataGetter defines the interface able to handle get requests for MultiversX blockchain
type DataGetter interface {
	GetTokenIdForErc20Address(ctx context.Context, erc20Address []byte) ([][]byte, error)
	GetERC20AddressForTokenId(ctx context.Context, tokenId []byte) ([][]byte, error)
	IsInterfaceNil() bool
}

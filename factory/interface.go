package factory

import (
	"context"

	erdgoCore "github.com/ElrondNetwork/elrond-sdk-erdgo/core"
)

type dataGetter interface {
	GetTokenIdForErc20Address(ctx context.Context, erc20Address []byte) ([][]byte, error)
	GetERC20AddressForTokenId(ctx context.Context, tokenId []byte) ([][]byte, error)
	GetAllStakedRelayers(ctx context.Context) ([][]byte, error)
	IsInterfaceNil() bool
}

// ElrondRoleProvider defines the operations for the Elrond role provider
type ElrondRoleProvider interface {
	Execute(ctx context.Context) error
	IsWhitelisted(address erdgoCore.AddressHandler) bool
	IsInterfaceNil() bool
}

// EthereumRoleProvider defines the operations for the Ethereum role provider
type EthereumRoleProvider interface {
	Execute(ctx context.Context) error
	VerifyEthSignature(signature []byte, messageHash []byte) error
	IsInterfaceNil() bool
}

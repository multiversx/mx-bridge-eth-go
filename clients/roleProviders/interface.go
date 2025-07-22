package roleproviders

import (
	"context"

	"github.com/block-vision/sui-go-sdk/models"
	"github.com/ethereum/go-ethereum/common"
)

// DataGetter defines the interface able to handle get requests for MultiversX blockchain
type DataGetter interface {
	GetAllStakedRelayers(ctx context.Context) ([][]byte, error)
	IsInterfaceNil() bool
}

// EthereumChainInteractor defines an Ethereum client able to respond to requests
type EthereumChainInteractor interface {
	GetRelayers(ctx context.Context) ([]common.Address, error)
	IsInterfaceNil() bool
}

// SuiDataGetter defines the interface able to handle get requests for Sui blockchain
type SuiDataGetter interface {
	GetRelayers(ctx context.Context) ([]models.SuiAddress, error)
	IsInterfaceNil() bool
}

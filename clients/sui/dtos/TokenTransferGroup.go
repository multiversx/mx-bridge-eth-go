package dtos

import (
	"github.com/block-vision/sui-go-sdk/models"
)

type TokenTransferGroup struct {
	Recipients []models.SuiAddressBytes
	Amounts    []uint64
	Tokens     [][]byte
	Nonces     []uint64
	Signatures [][96]byte
}

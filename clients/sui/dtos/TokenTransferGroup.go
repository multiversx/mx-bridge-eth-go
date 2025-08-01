package dtos

type TokenTransferGroup struct {
	Recipients [][]byte
	Amounts    []uint64
	Nonces     []uint64
}

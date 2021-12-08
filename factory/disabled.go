package factory

// TODO remove this file after completion of all flows
type disabledSignatureHolder struct{}

// Signatures -
func (d *disabledSignatureHolder) Signatures(_ []byte) [][]byte {
	return make([][]byte, 0)
}

// IsInterfaceNil -
func (d *disabledSignatureHolder) IsInterfaceNil() bool {
	return d == nil
}

package disabled

type disabledSignaturesHolder struct {
}

// NewDisabledSignaturesHolder will return a disabled signature holder instance
func NewDisabledSignaturesHolder() *disabledSignaturesHolder {
	return &disabledSignaturesHolder{}
}

// Signatures returns an empty slice
func (disabled *disabledSignaturesHolder) Signatures(_ []byte) [][]byte {
	return make([][]byte, 0)
}

// ClearStoredSignatures does nothing
func (disabled *disabledSignaturesHolder) ClearStoredSignatures() {
}

// IsInterfaceNil returns true if there is no value under the interface
func (disabled *disabledSignaturesHolder) IsInterfaceNil() bool {
	return disabled == nil
}

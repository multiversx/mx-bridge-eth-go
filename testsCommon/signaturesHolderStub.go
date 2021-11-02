package testsCommon

// SignaturesHolderStub -
type SignaturesHolderStub struct {
	SignaturesCalled func(messageHash []byte) [][]byte
}

// Signatures -
func (stub *SignaturesHolderStub) Signatures(messageHash []byte) [][]byte {
	if stub.SignaturesCalled != nil {
		return stub.SignaturesCalled(messageHash)
	}

	return make([][]byte, 0)
}

// IsInterfaceNil -
func (stub *SignaturesHolderStub) IsInterfaceNil() bool {
	return stub == nil
}

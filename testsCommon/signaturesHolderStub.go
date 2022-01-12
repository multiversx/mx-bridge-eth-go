package testsCommon

// SignaturesHolderStub -
type SignaturesHolderStub struct {
	SignaturesCalled            func(messageHash []byte) [][]byte
	ClearStoredSignaturesCalled func()
}

// Signatures -
func (stub *SignaturesHolderStub) Signatures(messageHash []byte) [][]byte {
	if stub.SignaturesCalled != nil {
		return stub.SignaturesCalled(messageHash)
	}

	return make([][]byte, 0)
}

// ClearStoredSignatures -
func (stub *SignaturesHolderStub) ClearStoredSignatures() {
	if stub.ClearStoredSignaturesCalled != nil {
		stub.ClearStoredSignaturesCalled()
	}
}

// IsInterfaceNil -
func (stub *SignaturesHolderStub) IsInterfaceNil() bool {
	return stub == nil
}

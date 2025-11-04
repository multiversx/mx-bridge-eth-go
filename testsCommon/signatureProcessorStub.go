package testsCommon

// SignatureProcessorStub -
type SignatureProcessorStub struct {
	VerifySignatureCalled func(signature []byte, messageHash []byte) error
}

// VerifySignature -
func (sps *SignatureProcessorStub) VerifySignature(signature []byte, messageHash []byte) error {
	if sps.VerifySignatureCalled != nil {
		return sps.VerifySignatureCalled(signature, messageHash)
	}

	return nil
}

// IsInterfaceNil -
func (sps *SignatureProcessorStub) IsInterfaceNil() bool {
	return sps == nil
}

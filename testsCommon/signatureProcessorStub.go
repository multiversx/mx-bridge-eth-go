package testsCommon

// SignatureProcessorStub -
type SignatureProcessorStub struct {
	VerifyEthSignatureCalled func(signature []byte, messageHash []byte) error
}

// VerifyEthSignature -
func (sps *SignatureProcessorStub) VerifyEthSignature(signature []byte, messageHash []byte) error {
	if sps.VerifyEthSignatureCalled != nil {
		return sps.VerifyEthSignatureCalled(signature, messageHash)
	}

	return nil
}

// IsInterfaceNil -
func (sps *SignatureProcessorStub) IsInterfaceNil() bool {
	return sps == nil
}

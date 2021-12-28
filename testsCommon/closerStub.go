package testsCommon

// CloserStub -
type CloserStub struct {
	CloseCalled func() error
}

// Close -
func (stub *CloserStub) Close() error {
	if stub.CloseCalled != nil {
		return stub.CloseCalled()
	}

	return nil
}

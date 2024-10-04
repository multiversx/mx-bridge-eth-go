package testsCommon

type AddressHandlerStub struct {
	AddressAsBech32StringCalled func() (string, error)
	AddressBytesCalled          func() []byte
	AddressSliceCalled          func() [32]byte
	IsValidCalled               func() bool
	PrettyCalled                func() string
}

// AddressAsBech32String -
func (stub *AddressHandlerStub) AddressAsBech32String() (string, error) {
	if stub.AddressAsBech32StringCalled != nil {
		return stub.AddressAsBech32StringCalled()
	}

	return "", nil
}

// AddressBytes -
func (stub *AddressHandlerStub) AddressBytes() []byte {
	if stub.AddressBytesCalled != nil {
		return stub.AddressBytesCalled()
	}

	return nil
}

// AddressSlice -
func (stub *AddressHandlerStub) AddressSlice() [32]byte {
	if stub.AddressSliceCalled != nil {
		return stub.AddressSliceCalled()
	}

	return [32]byte{}
}

// IsValid -
func (stub *AddressHandlerStub) IsValid() bool {
	if stub.IsValidCalled != nil {
		return stub.IsValidCalled()
	}

	return false
}

// Pretty -
func (stub *AddressHandlerStub) Pretty() string {
	if stub.PrettyCalled != nil {
		return stub.PrettyCalled()
	}

	return ""
}

// IsInterfaceNil -
func (stub *AddressHandlerStub) IsInterfaceNil() bool {
	return stub == nil
}

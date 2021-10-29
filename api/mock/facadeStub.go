package mock

// FacadeStub is the mock implementation of a node router handler
type FacadeStub struct {
	RestApiInterfaceCalled func() string
	PprofEnabledCalled     func() bool
}

// RestApiInterface -
func (f *FacadeStub) RestApiInterface() string {
	if f.RestApiInterfaceCalled != nil {
		return f.RestApiInterfaceCalled()
	}
	return "localhost:8080"
}

// PprofEnabled -
func (f *FacadeStub) PprofEnabled() bool {
	if f.PprofEnabledCalled != nil {
		f.PprofEnabledCalled()
	}
	return false
}

// IsInterfaceNil returns true if there is no value under the interface
func (f *FacadeStub) IsInterfaceNil() bool {
	return f == nil
}

package mock

type FacadeStub struct{}

// RestApiInterface -
func (f *FacadeStub) RestApiInterface() string {
	return "localhost:8080"
}

// PprofEnabled -
func (f *FacadeStub) PprofEnabled() bool {
	return false
}

// IsInterfaceNil returns true if there is no value under the interface
func (f *FacadeStub) IsInterfaceNil() bool {
	return f == nil
}

package api

// FacadeHandler defines all the methods that a facade should implement
type FacadeHandler interface {
	RestApiInterface() string
	PprofEnabled() bool
	IsInterfaceNil() bool
}

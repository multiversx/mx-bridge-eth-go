package roleProvider

// ChainInteractor defines a client able to respond to VM queries
type ChainInteractor interface {
	ExecuteVmQueryOnBridgeContract(function string, params ...[]byte) ([][]byte, error)
	IsInterfaceNil() bool
}

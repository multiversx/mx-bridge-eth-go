package roleProvider

// ChainClient defines a chain client able to respond to VM queries
type ChainClient interface {
	ExecuteVmQueryOnBridgeContract(function string, params ...[]byte) ([][]byte, error)
	IsInterfaceNil() bool
}

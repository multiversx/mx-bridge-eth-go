package bridge

// Config represents the Elrond Config parameters
type Config struct {
	NetworkAddress       string
	BridgeAddress        string
	PrivateKey           string
	NonceUpdateInSeconds uint
	GasLimit             uint64
}

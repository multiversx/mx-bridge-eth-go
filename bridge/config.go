package bridge

// Config represents the Elrond Config parameters
type Config struct {
	NetworkAddress               string
	BridgeAddress                string
	PrivateKey                   string
	IntervalToResendTxsInSeconds uint64
	GasLimit                     uint64
}

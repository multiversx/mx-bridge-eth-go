package parsers

// CallData defines the struct holding SC call data parameters
type CallData struct {
	Type      byte
	Function  string
	GasLimit  uint64
	Arguments []string
}

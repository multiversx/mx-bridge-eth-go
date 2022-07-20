package clients

import "strings"

// Chain defines all the chains supported
type Chain string

const (
	Elrond   Chain = "Elrond"
	Ethereum Chain = "Ethereum"
	Bsc      Chain = "Bsc"
)

func (c Chain) ToLower() string {
	return strings.ToLower(string(c))
}

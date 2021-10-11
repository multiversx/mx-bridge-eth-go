package contracts

// TokensHandler represents a token handler able to manage all tokens in this framework
type TokensHandler interface {
	AddNewToken(ethAddress []byte, ticker string)
	GetTickerFromEthAddress(ethAddress []byte) (string, error)
	IsInterfaceNil() bool
}

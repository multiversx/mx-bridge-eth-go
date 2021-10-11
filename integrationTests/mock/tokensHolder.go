package mock

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"sync"
)

type token struct {
	ethAddress []byte
	ticker     string
}

type tokensHolder struct {
	mutTokens sync.RWMutex
	tokens    []*token
}

// NewTokensHolder creates a new instance of tokensHolder able to keep track of all tokens in this framework
func NewTokensHolder() *tokensHolder {
	return &tokensHolder{
		tokens: make([]*token, 0),
	}
}

// AddNewToken adds a new token to this registry
func (th *tokensHolder) AddNewToken(ethAddress []byte, ticker string) {
	t := &token{
		ethAddress: ethAddress,
		ticker:     ticker,
	}

	th.mutTokens.Lock()
	th.tokens = append(th.tokens, t)
	th.mutTokens.Unlock()
}

// GetTickerFromEthAddress returns the ticker, if existing, or an error otherwise
func (th *tokensHolder) GetTickerFromEthAddress(ethAddress []byte) (string, error) {
	th.mutTokens.RLock()
	defer th.mutTokens.RUnlock()

	for _, t := range th.tokens {
		if bytes.Equal(t.ethAddress, ethAddress) {
			return t.ticker, nil
		}
	}

	return "", fmt.Errorf("token not register for address %s", hex.EncodeToString(ethAddress))
}

// IsInterfaceNil returns true if there is no value under the interface
func (th *tokensHolder) IsInterfaceNil() bool {
	return th == nil
}

//go:build slow

package slowTests

import (
	"sync"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

type tokenData struct {
	abstractTokenIdentifier string

	mvxUniversalTokenTicker     string
	mvxChainSpecificTokenTicker string
	ethTokenName                string
	ethTokenSymbol              string

	mvxUniversalToken     string
	mvxChainSpecificToken string
	ethErc20Address       common.Address
	ethErc20Contract      erc20Contract
}

type tokensRegistry struct {
	testing.TB
	mut    sync.RWMutex
	tokens map[string]*tokenData
}

func newTokenRegistry(tb testing.TB) *tokensRegistry {
	return &tokensRegistry{
		TB:     tb,
		tokens: make(map[string]*tokenData, 100),
	}
}

func (registry *tokensRegistry) addToken(params issueTokenParams) {
	registry.mut.Lock()
	defer registry.mut.Unlock()

	_, found := registry.tokens[params.abstractTokenIdentifier]
	require.False(registry, found, "can not register more than one instance of the same abstract token identifier %s", params.abstractTokenIdentifier)

	newToken := &tokenData{
		abstractTokenIdentifier:     params.abstractTokenIdentifier,
		mvxUniversalTokenTicker:     params.mvxUniversalTokenTicker,
		mvxChainSpecificTokenTicker: params.mvxChainSpecificTokenDisplayName,
		ethTokenName:                params.ethTokenName,
		ethTokenSymbol:              params.ethTokenSymbol,
	}

	registry.tokens[params.abstractTokenIdentifier] = newToken
}

func (registry *tokensRegistry) registerUniversalToken(abstractTokenIdentifier string, mvxUniversalToken string) {
	registry.mut.Lock()
	defer registry.mut.Unlock()

	data, found := registry.tokens[abstractTokenIdentifier]
	require.True(registry, found, "abstract token identifier not registered %s", abstractTokenIdentifier)

	data.mvxUniversalToken = mvxUniversalToken
}

func (registry *tokensRegistry) registerChainSpecificToken(abstractTokenIdentifier string, mvxChainSpecificToken string) {
	registry.mut.Lock()
	defer registry.mut.Unlock()

	data, found := registry.tokens[abstractTokenIdentifier]
	require.True(registry, found, "abstract token identifier not registered %s", abstractTokenIdentifier)

	data.mvxChainSpecificToken = mvxChainSpecificToken
}

func (registry *tokensRegistry) registerEthAddressAndContract(
	abstractTokenIdentifier string,
	ethErc20Address common.Address,
	ethErc20Contract erc20Contract,
) {
	registry.mut.Lock()
	defer registry.mut.Unlock()

	data, found := registry.tokens[abstractTokenIdentifier]
	require.True(registry, found, "abstract token identifier not registered %s", abstractTokenIdentifier)

	data.ethErc20Address = ethErc20Address
	data.ethErc20Contract = ethErc20Contract
}

func (registry *tokensRegistry) getTokenData(abstractTokenIdentifier string) *tokenData {
	registry.mut.RLock()
	defer registry.mut.RUnlock()

	return registry.tokens[abstractTokenIdentifier]
}

package chain

import (
	"fmt"
	"strings"
)

const (
	evmCompatibleChainToMultiversXNameTemplate  = "%sToMultiversX"
	multiversXToEvmCompatibleChainNameTemplate  = "MultiversXTo%s"
	baseLogIdTemplate                           = "%sMultiversX-Base"
	multiversXClientLogIdTemplate               = "%sMultiversX-MultiversXClient"
	multiversXDataGetterLogIdTemplate           = "%sMultiversX-MultiversXDataGetter"
	evmCompatibleChainClientLogIdTemplate       = "%sMultiversX-%sClient"
	multiversXRoleProviderLogIdTemplate         = "%sMultiversX-MultiversXRoleProvider"
	evmCompatibleChainRoleProviderLogIdTemplate = "%sMultiversX-%sRoleProvider"
	broadcasterLogIdTemplate                    = "%sMultiversX-Broadcaster"
)

// Chain defines all the chain supported
type Chain string

const (
	// MultiversX is the string representation of the MultiversX chain
	MultiversX Chain = "msx"

	// Ethereum is the string representation of the Ethereum chain
	Ethereum Chain = "Ethereum"

	// Bsc is the string representation of the Binance smart chain
	Bsc Chain = "Bsc"

	// Polygon is the string representation of the Polygon chain
	Polygon Chain = "Polygon"
)

// ToLower returns the lowercase string of chain
func (c Chain) ToLower() string {
	return strings.ToLower(string(c))
}

// EvmCompatibleChainToMultiversXName returns the string using chain value and evmCompatibleChainToMultiversXNameTemplate
func (c Chain) EvmCompatibleChainToMultiversXName() string {
	return fmt.Sprintf(evmCompatibleChainToMultiversXNameTemplate, c)
}

// MultiversXToEvmCompatibleChainName returns the string using chain value and multiversXToEvmCompatibleChainNameTemplate
func (c Chain) MultiversXToEvmCompatibleChainName() string {
	return fmt.Sprintf(multiversXToEvmCompatibleChainNameTemplate, c)
}

// BaseLogId returns the string using chain value and baseLogIdTemplate
func (c Chain) BaseLogId() string {
	return fmt.Sprintf(baseLogIdTemplate, c)
}

// MultiversXClientLogId returns the string using chain value and multiversXClientLogIdTemplate
func (c Chain) MultiversXClientLogId() string {
	return fmt.Sprintf(multiversXClientLogIdTemplate, c)
}

// MultiversXDataGetterLogId returns the string using chain value and multiversXDataGetterLogIdTemplate
func (c Chain) MultiversXDataGetterLogId() string {
	return fmt.Sprintf(multiversXDataGetterLogIdTemplate, c)
}

// EvmCompatibleChainClientLogId returns the string using chain value and evmCompatibleChainClientLogIdTemplate
func (c Chain) EvmCompatibleChainClientLogId() string {
	return fmt.Sprintf(evmCompatibleChainClientLogIdTemplate, c, c)
}

// MultiversXRoleProviderLogId returns the string using chain value and multiversXRoleProviderLogIdTemplate
func (c Chain) MultiversXRoleProviderLogId() string {
	return fmt.Sprintf(multiversXRoleProviderLogIdTemplate, c)
}

// EvmCompatibleChainRoleProviderLogId returns the string using chain value and evmCompatibleChainRoleProviderLogIdTemplate
func (c Chain) EvmCompatibleChainRoleProviderLogId() string {
	return fmt.Sprintf(evmCompatibleChainRoleProviderLogIdTemplate, c, c)
}

// BroadcasterLogId returns the string using chain value and broadcasterLogIdTemplate
func (c Chain) BroadcasterLogId() string {
	return fmt.Sprintf(broadcasterLogIdTemplate, c)
}

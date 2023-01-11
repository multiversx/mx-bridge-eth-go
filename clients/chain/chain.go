package chain

import (
	"fmt"
	"strings"
)

const (
	evmCompatibleChainToElrondNameTemplate      = "%sToElrond"
	elrondToEvmCompatibleChainNameTemplate      = "ElrondTo%s"
	baseLogIdTemplate                           = "%sElrond-Base"
	elrondClientLogIdTemplate                   = "%sElrond-ElrondClient"
	elrondDataGetterLogIdTemplate               = "%sElrond-ElrondDataGetter"
	evmCompatibleChainClientLogIdTemplate       = "%sElrond-%sClient"
	elrondRoleProviderLogIdTemplate             = "%sElrond-ElrondRoleProvider"
	evmCompatibleChainRoleProviderLogIdTemplate = "%sElrond-%sRoleProvider"
	broadcasterLogIdTemplate                    = "%sElrond-Broadcaster"
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
)

// ToLower returns the lowercase string of chain
func (c Chain) ToLower() string {
	return strings.ToLower(string(c))
}

// EvmCompatibleChainToElrondName returns the string using chain value and evmCompatibleChainToElrondNameTemplate
func (c Chain) EvmCompatibleChainToElrondName() string {
	return fmt.Sprintf(evmCompatibleChainToElrondNameTemplate, c)
}

// ElrondToEvmCompatibleChainName returns the string using chain value and elrondToEvmCompatibleChainNameTemplate
func (c Chain) ElrondToEvmCompatibleChainName() string {
	return fmt.Sprintf(elrondToEvmCompatibleChainNameTemplate, c)
}

// BaseLogId returns the string using chain value and baseLogIdTemplate
func (c Chain) BaseLogId() string {
	return fmt.Sprintf(baseLogIdTemplate, c)
}

// ElrondClientLogId returns the string using chain value and elrondClientLogIdTemplate
func (c Chain) ElrondClientLogId() string {
	return fmt.Sprintf(elrondClientLogIdTemplate, c)
}

// ElrondDataGetterLogId returns the string using chain value and elrondDataGetterLogIdTemplate
func (c Chain) ElrondDataGetterLogId() string {
	return fmt.Sprintf(elrondDataGetterLogIdTemplate, c)
}

// EvmCompatibleChainClientLogId returns the string using chain value and evmCompatibleChainClientLogIdTemplate
func (c Chain) EvmCompatibleChainClientLogId() string {
	return fmt.Sprintf(evmCompatibleChainClientLogIdTemplate, c, c)
}

// ElrondRoleProviderLogId returns the string using chain value and elrondRoleProviderLogIdTemplate
func (c Chain) ElrondRoleProviderLogId() string {
	return fmt.Sprintf(elrondRoleProviderLogIdTemplate, c)
}

// EvmCompatibleChainRoleProviderLogId returns the string using chain value and evmCompatibleChainRoleProviderLogIdTemplate
func (c Chain) EvmCompatibleChainRoleProviderLogId() string {
	return fmt.Sprintf(evmCompatibleChainRoleProviderLogIdTemplate, c, c)
}

// BroadcasterLogId returns the string using chain value and broadcasterLogIdTemplate
func (c Chain) BroadcasterLogId() string {
	return fmt.Sprintf(broadcasterLogIdTemplate, c)
}

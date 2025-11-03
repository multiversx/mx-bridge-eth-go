package chain

import (
	"fmt"
	"strings"
)

const (
	peerChainToMultiversXNameTemplate   = "%sToMultiversX"
	multiversXToPeerChainNameTemplate   = "MultiversXTo%s"
	baseLogIdTemplate                   = "%sMultiversX-Base"
	multiversXClientLogIdTemplate       = "%sMultiversX-MultiversXClient"
	multiversXDataGetterLogIdTemplate   = "%sMultiversX-MultiversXDataGetter"
	peerChainDataGetterLogIdTemplate    = "%sMultiversX-%sDataGetter"
	peerChainClientLogIdTemplate        = "%sMultiversX-%sClient"
	multiversXRoleProviderLogIdTemplate = "%sMultiversX-MultiversXRoleProvider"
	peerChainRoleProviderLogIdTemplate  = "%sMultiversX-%sRoleProvider"
	broadcasterLogIdTemplate            = "%sMultiversX-Broadcaster"
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

	// Sui is the string representation of the Sui chain
	Sui Chain = "Sui"
)

// ToLower returns the lowercase string of chain
func (c Chain) ToLower() string {
	return strings.ToLower(string(c))
}

// PeerChainToMultiversXName returns the string using chain value and peerChainToMultiversXNameTemplate
func (c Chain) PeerChainToMultiversXName() string {
	return fmt.Sprintf(peerChainToMultiversXNameTemplate, c)
}

// MultiversXToPeerChainName returns the string using chain value and multiversXToPeerChainNameTemplate
func (c Chain) MultiversXToPeerChainName() string {
	return fmt.Sprintf(multiversXToPeerChainNameTemplate, c)
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

// PeerChainDataGetterLogId returns the string using chain value and peerChainDataGetterLogIdTemplate
func (c Chain) PeerChainDataGetterLogId() string {
	return fmt.Sprintf(peerChainDataGetterLogIdTemplate, c, c)
}

// PeerChainClientLogId returns the string using chain value and peerChainClientLogIdTemplate
func (c Chain) PeerChainClientLogId() string {
	return fmt.Sprintf(peerChainClientLogIdTemplate, c, c)
}

// MultiversXRoleProviderLogId returns the string using chain value and multiversXRoleProviderLogIdTemplate
func (c Chain) MultiversXRoleProviderLogId() string {
	return fmt.Sprintf(multiversXRoleProviderLogIdTemplate, c)
}

// PeerChainRoleProviderLogId returns the string using chain value and peerChainRoleProviderLogIdTemplate
func (c Chain) PeerChainRoleProviderLogId() string {
	return fmt.Sprintf(peerChainRoleProviderLogIdTemplate, c, c)
}

// BroadcasterLogId returns the string using chain value and broadcasterLogIdTemplate
func (c Chain) BroadcasterLogId() string {
	return fmt.Sprintf(broadcasterLogIdTemplate, c)
}

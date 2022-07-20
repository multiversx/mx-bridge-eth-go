package chain

import (
	"fmt"
	"strings"
)

const (
	ethToElrondNameTemplate         = "%sToElrond"
	elrondToEthNameTemplate         = "ElrondTo%s"
	baseLogIdTemplate               = "%sElrond-Base"
	elrondClientLogIdTemplate       = "%sElrond-ElrondClient"
	elrondDataGetterLogIdTemplate   = "%sElrond-ElrondDataGetter"
	ethClientLogIdTemplate          = "%sElrond-%sClient"
	elrondRoleProviderLogIdTemplate = "%sElrond-ElrondRoleProvider"
	ethRoleProviderLogIdTemplate    = "%sElrond-%sRoleProvider"
	broadcasterLogIdTemplate        = "%sElrond-Broadcaster"
)

// Chain defines all the chain supported
type Chain string

const (
	Elrond   Chain = "Elrond"
	Ethereum Chain = "Ethereum"
	Bsc      Chain = "Bsc"
)

// ToLower returns the lowercase string of chain
func (c Chain) ToLower() string {
	return strings.ToLower(string(c))
}

// EthToElrondName return the string using chain value and ethToElrondNameTemplate
func (c Chain) EthToElrondName() string {
	return fmt.Sprintf(ethToElrondNameTemplate, c)
}

// ElrondToEthName return the string using chain value and elrondToEthNameTemplate
func (c Chain) ElrondToEthName() string {
	return fmt.Sprintf(elrondToEthNameTemplate, c)
}

// BaseLogId return the string using chain value and baseLogIdTemplate
func (c Chain) BaseLogId() string {
	return fmt.Sprintf(baseLogIdTemplate, c)
}

// ElrondClientLogId return the string using chain value and elrondClientLogIdTemplate
func (c Chain) ElrondClientLogId() string {
	return fmt.Sprintf(elrondClientLogIdTemplate, c)
}

// ElrondDataGetterLogId return the string using chain value and elrondDataGetterLogIdTemplate
func (c Chain) ElrondDataGetterLogId() string {
	return fmt.Sprintf(elrondDataGetterLogIdTemplate, c)
}

// EthClientLogId return the string using chain value and ethClientLogIdTemplate
func (c Chain) EthClientLogId() string {
	return fmt.Sprintf(ethClientLogIdTemplate, c, c)
}

// ElrondRoleProviderLogId return the string using chain value and elrondRoleProviderLogIdTemplate
func (c Chain) ElrondRoleProviderLogId() string {
	return fmt.Sprintf(elrondRoleProviderLogIdTemplate, c)
}

// EthRoleProviderLogId return the string using chain value and ethRoleProviderLogIdTemplate
func (c Chain) EthRoleProviderLogId() string {
	return fmt.Sprintf(ethRoleProviderLogIdTemplate, c, c)
}

// BroadcasterLogId return the string using chain value and broadcasterLogIdTemplate
func (c Chain) BroadcasterLogId() string {
	return fmt.Sprintf(broadcasterLogIdTemplate, c)
}

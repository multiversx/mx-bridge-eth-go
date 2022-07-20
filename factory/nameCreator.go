package factory

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-eth-bridge/clients"
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

func ethToElrondName(chain clients.Chain) string {
	s := fmt.Sprintf(ethToElrondNameTemplate, chain)
	return s
}

func elrondToEthName(chain clients.Chain) string {
	return fmt.Sprintf(elrondToEthNameTemplate, chain)
}

func baseLogId(chain clients.Chain) string {
	return fmt.Sprintf(baseLogIdTemplate, chain)
}

func elrondClientLogId(chain clients.Chain) string {
	return fmt.Sprintf(elrondClientLogIdTemplate, chain)
}

func elrondDataGetterLogId(chain clients.Chain) string {
	return fmt.Sprintf(elrondDataGetterLogIdTemplate, chain)
}

func ethClientLogId(chain clients.Chain) string {
	return fmt.Sprintf(ethClientLogIdTemplate, chain, chain)
}

func elrondRoleProviderLogId(chain clients.Chain) string {
	return fmt.Sprintf(elrondRoleProviderLogIdTemplate, chain)
}

func ethRoleProviderLogId(chain clients.Chain) string {
	return fmt.Sprintf(ethRoleProviderLogIdTemplate, chain, chain)
}

func broadcasterLogId(chain clients.Chain) string {
	return fmt.Sprintf(broadcasterLogIdTemplate, chain)
}

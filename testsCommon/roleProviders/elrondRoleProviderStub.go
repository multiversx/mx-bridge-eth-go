package roleProviders

import "github.com/ElrondNetwork/elrond-sdk-erdgo/core"

// ElrondRoleProviderStub -
type ElrondRoleProviderStub struct {
	IsWhitelistedCalled func(address core.AddressHandler) bool
}

// IsWhitelisted -
func (stub *ElrondRoleProviderStub) IsWhitelisted(address core.AddressHandler) bool {
	if stub.IsWhitelistedCalled != nil {
		return stub.IsWhitelistedCalled(address)
	}

	return true
}

// IsInterfaceNil -
func (stub *ElrondRoleProviderStub) IsInterfaceNil() bool {
	return stub == nil
}

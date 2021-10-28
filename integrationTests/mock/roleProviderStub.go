package mock

import "github.com/ElrondNetwork/elrond-sdk-erdgo/core"

// RoleProviderStub -
type RoleProviderStub struct {
	IsWhitelistedCalled func(address core.AddressHandler) bool
}

// IsWhitelisted -
func (rps *RoleProviderStub) IsWhitelisted(address core.AddressHandler) bool {
	if rps.IsWhitelistedCalled != nil {
		return rps.IsWhitelistedCalled(address)
	}

	return false
}

// IsInterfaceNil -
func (rps *RoleProviderStub) IsInterfaceNil() bool {
	return rps == nil
}

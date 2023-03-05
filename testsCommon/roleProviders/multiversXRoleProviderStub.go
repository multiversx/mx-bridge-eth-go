package roleproviders

import "github.com/multiversx/mx-sdk-go/core"

// MultiversXRoleProviderStub -
type MultiversXRoleProviderStub struct {
	IsWhitelistedCalled func(address core.AddressHandler) bool
}

// IsWhitelisted -
func (stub *MultiversXRoleProviderStub) IsWhitelisted(address core.AddressHandler) bool {
	if stub.IsWhitelistedCalled != nil {
		return stub.IsWhitelistedCalled(address)
	}

	return true
}

// IsInterfaceNil -
func (stub *MultiversXRoleProviderStub) IsInterfaceNil() bool {
	return stub == nil
}

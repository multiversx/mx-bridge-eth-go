package mock

// RoleProviderStub -
type RoleProviderStub struct {
	IsWhitelistedCalled func(s string) bool
}

// IsWhitelisted -
func (rps *RoleProviderStub) IsWhitelisted(s string) bool {
	if rps.IsWhitelistedCalled != nil {
		return rps.IsWhitelistedCalled(s)
	}

	return false
}

// IsInterfaceNil -
func (rps *RoleProviderStub) IsInterfaceNil() bool {
	return rps == nil
}

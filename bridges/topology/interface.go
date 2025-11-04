package topology

// PublicKeysProvider defines the behavior of a provider able to return all public keys allowed to operate on the relayers network
type PublicKeysProvider interface {
	SortedPublicKeys() [][]byte
	IsInterfaceNil() bool
}

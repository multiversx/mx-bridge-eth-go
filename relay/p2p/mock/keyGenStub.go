package mock

import crypto "github.com/ElrondNetwork/elrond-go-crypto"

// KeyGenStub mocks a key generation implementation
type KeyGenStub struct {
	GeneratePairStub            func() (crypto.PrivateKey, crypto.PublicKey)
	PrivateKeyFromByteArrayStub func(b []byte) (crypto.PrivateKey, error)
	PublicKeyFromByteArrayStub  func(b []byte) (crypto.PublicKey, error)
	CheckPublicKeyValidStub     func(b []byte) error
	SuiteStub                   func() crypto.Suite
}

// GeneratePair generates a pair of private and public keys
func (keyGen *KeyGenStub) GeneratePair() (crypto.PrivateKey, crypto.PublicKey) {
	if keyGen.GeneratePairStub != nil {
		return keyGen.GeneratePairStub()
	}

	return nil, &PublicKeyStub{}
}

// PrivateKeyFromByteArray generates the private key from it's byte array representation
func (keyGen *KeyGenStub) PrivateKeyFromByteArray(b []byte) (crypto.PrivateKey, error) {
	if keyGen.PrivateKeyFromByteArrayStub != nil {
		return keyGen.PrivateKeyFromByteArrayStub(b)
	}

	return nil, nil
}

// PublicKeyFromByteArray generates a public key from it's byte array representation
func (keyGen *KeyGenStub) PublicKeyFromByteArray(b []byte) (crypto.PublicKey, error) {
	if keyGen.PublicKeyFromByteArrayStub != nil {
		return keyGen.PublicKeyFromByteArrayStub(b)
	}

	return &PublicKeyStub{}, nil
}

// CheckPublicKeyValid verifies the validity of the public key
func (keyGen *KeyGenStub) CheckPublicKeyValid(b []byte) error {
	if keyGen.CheckPublicKeyValidStub != nil {
		return keyGen.CheckPublicKeyValidStub(b)
	}

	return nil
}

// Suite -
func (keyGen *KeyGenStub) Suite() crypto.Suite {
	if keyGen.SuiteStub != nil {
		return keyGen.SuiteStub()
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (keyGen *KeyGenStub) IsInterfaceNil() bool {
	return keyGen == nil
}

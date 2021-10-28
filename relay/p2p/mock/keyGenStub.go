package mock

import crypto "github.com/ElrondNetwork/elrond-go-crypto"

// KeyGenStub -
type KeyGenStub struct {
	GeneratePairStub            func() (crypto.PrivateKey, crypto.PublicKey)
	PrivateKeyFromByteArrayStub func(b []byte) (crypto.PrivateKey, error)
	PublicKeyFromByteArrayStub  func(b []byte) (crypto.PublicKey, error)
	CheckPublicKeyValidStub     func(b []byte) error
	SuiteStub                   func() crypto.Suite
}

// GeneratePair -
func (keyGen *KeyGenStub) GeneratePair() (crypto.PrivateKey, crypto.PublicKey) {
	if keyGen.GeneratePairStub != nil {
		return keyGen.GeneratePairStub()
	}

	return nil, &PublicKeyStub{}
}

// PrivateKeyFromByteArray -
func (keyGen *KeyGenStub) PrivateKeyFromByteArray(b []byte) (crypto.PrivateKey, error) {
	if keyGen.PrivateKeyFromByteArrayStub != nil {
		return keyGen.PrivateKeyFromByteArrayStub(b)
	}

	return nil, nil
}

// PublicKeyFromByteArray -
func (keyGen *KeyGenStub) PublicKeyFromByteArray(b []byte) (crypto.PublicKey, error) {
	if keyGen.PublicKeyFromByteArrayStub != nil {
		return keyGen.PublicKeyFromByteArrayStub(b)
	}

	return &PublicKeyStub{}, nil
}

// CheckPublicKeyValid -
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

// IsInterfaceNil -
func (keyGen *KeyGenStub) IsInterfaceNil() bool {
	return keyGen == nil
}

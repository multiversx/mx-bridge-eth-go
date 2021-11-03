package crypto

import crypto "github.com/ElrondNetwork/elrond-go-crypto"

// SingleSignerStub -
type SingleSignerStub struct {
	SignCalled   func(private crypto.PrivateKey, msg []byte) ([]byte, error)
	VerifyCalled func(public crypto.PublicKey, msg []byte, sig []byte) error
}

// Sign -
func (sss *SingleSignerStub) Sign(private crypto.PrivateKey, msg []byte) ([]byte, error) {
	if sss.SignCalled != nil {
		return sss.SignCalled(private, msg)
	}

	return make([]byte, 0), nil
}

// Verify -
func (sss *SingleSignerStub) Verify(public crypto.PublicKey, msg []byte, sig []byte) error {
	if sss.VerifyCalled != nil {
		return sss.VerifyCalled(public, msg, sig)
	}

	return nil
}

// IsInterfaceNil -
func (sss *SingleSignerStub) IsInterfaceNil() bool {
	return sss == nil
}

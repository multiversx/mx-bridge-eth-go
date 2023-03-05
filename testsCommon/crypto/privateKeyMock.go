package crypto

import (
	"crypto/rand"

	"github.com/multiversx/mx-chain-core-go/hashing/sha256"
	crypto "github.com/multiversx/mx-chain-crypto-go"
)

// PublicKeyMock -
type PublicKeyMock struct {
	pubKey []byte
}

type privateKeyMock struct {
	privKey []byte
}

// NewPrivateKeyMock -
func NewPrivateKeyMock() *privateKeyMock {
	buff := make([]byte, 32)
	_, _ = rand.Read(buff)

	return &privateKeyMock{
		privKey: buff,
	}
}

// ToByteArray -
func (sspk *PublicKeyMock) ToByteArray() ([]byte, error) {
	return sspk.pubKey, nil
}

// Suite -
func (sspk *PublicKeyMock) Suite() crypto.Suite {
	return nil
}

// Point -
func (sspk *PublicKeyMock) Point() crypto.Point {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (sspk *PublicKeyMock) IsInterfaceNil() bool {
	return sspk == nil
}

// ToByteArray -
func (sk *privateKeyMock) ToByteArray() ([]byte, error) {
	return sk.privKey, nil
}

// GeneratePublic -
func (sk *privateKeyMock) GeneratePublic() crypto.PublicKey {
	return &PublicKeyMock{
		pubKey: sha256.NewSha256().Compute(string(sk.privKey)),
	}
}

// Suite -
func (sk *privateKeyMock) Suite() crypto.Suite {
	return nil
}

// Scalar -
func (sk *privateKeyMock) Scalar() crypto.Scalar {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (sk *privateKeyMock) IsInterfaceNil() bool {
	return sk == nil
}

package crypto

import crypto "github.com/ElrondNetwork/elrond-go-crypto"

// PublicKeyStub -
type PublicKeyStub struct {
	ToByteArrayCalled func() ([]byte, error)
	SuiteCalled       func() crypto.Suite
	PointCalled       func() crypto.Point
}

// ToByteArray -
func (pks *PublicKeyStub) ToByteArray() ([]byte, error) {
	if pks.ToByteArrayCalled != nil {
		return pks.ToByteArrayCalled()
	}

	return make([]byte, 0), nil
}

// Suite -
func (pks *PublicKeyStub) Suite() crypto.Suite {
	if pks.SuiteCalled != nil {
		return pks.SuiteCalled()
	}

	return nil
}

// Point -
func (pks *PublicKeyStub) Point() crypto.Point {
	if pks.PointCalled != nil {
		return pks.PointCalled()
	}

	return nil
}

// IsInterfaceNil -
func (pks *PublicKeyStub) IsInterfaceNil() bool {
	return pks == nil
}

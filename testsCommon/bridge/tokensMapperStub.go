package bridge

import "context"

// TokensMapperStub -
type TokensMapperStub struct {
	ConvertTokenCalled func(ctx context.Context, sourceBytes []byte) ([]byte, error)
}

// ConvertToken -
func (stub *TokensMapperStub) ConvertToken(ctx context.Context, sourceBytes []byte) ([]byte, error) {
	if stub.ConvertTokenCalled != nil {
		return stub.ConvertTokenCalled(ctx, sourceBytes)
	}

	return make([]byte, 0), nil
}

// IsInterfaceNil -
func (stub *TokensMapperStub) IsInterfaceNil() bool {
	return stub == nil
}

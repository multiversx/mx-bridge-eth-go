package integrationTests

import "context"

type httpClientWrapper interface {
	GetHTTP(ctx context.Context, endpoint string) ([]byte, int, error)
	PostHTTP(ctx context.Context, endpoint string, data []byte) ([]byte, int, error)
	IsInterfaceNil() bool
}

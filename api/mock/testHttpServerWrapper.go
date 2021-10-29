package mock

import (
	"context"
	"net/http/httptest"
)

// TestHttpServerWrapper -
type TestHttpServerWrapper struct {
	*httptest.Server
}

// ListenAndServe does nothing, used to implement server interface
func (serv *TestHttpServerWrapper) ListenAndServe() error {
	return nil
}

// Shutdown closes the test http server
func (serv *TestHttpServerWrapper) Shutdown(_ context.Context) error {
	serv.Close()

	return nil
}

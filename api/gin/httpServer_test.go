package gin

import (
	"context"
	"net/http"
	"testing"

	apiErrors "github.com/ElrondNetwork/elrond-eth-bridge/api/errors"
	testsServer "github.com/ElrondNetwork/elrond-eth-bridge/testsCommon/server"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/stretchr/testify/assert"
)

func TestNewHttpServer(t *testing.T) {
	t.Parallel()

	t.Run("nil server should error", func(t *testing.T) {
		t.Parallel()

		hs, err := NewHttpServer(nil)
		assert.Equal(t, apiErrors.ErrNilHttpServer, err)
		assert.True(t, check.IfNil(hs))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		hs, err := NewHttpServer(&testsServer.ServerStub{})
		assert.Nil(t, err)
		assert.False(t, check.IfNil(hs))
	})
}

func TestNewHttpServer_Start(t *testing.T) {
	t.Parallel()

	t.Run("ListenAndServe returns closed server", func(t *testing.T) {
		t.Parallel()

		s := &testsServer.ServerStub{
			ListenAndServeCalled: func() error {
				return http.ErrServerClosed
			},
		}

		hs, _ := NewHttpServer(s)
		assert.False(t, check.IfNil(hs))

		hs.Start()
	})
	t.Run("ListenAndServe returns other error", func(t *testing.T) {
		t.Parallel()

		s := &testsServer.ServerStub{
			ListenAndServeCalled: func() error {
				return http.ErrContentLength
			},
		}

		hs, _ := NewHttpServer(s)
		assert.False(t, check.IfNil(hs))

		hs.Start()
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		s := &testsServer.ServerStub{
			ShutdownCalled: func(ctx context.Context) error {
				return expectedErr
			},
		}
		hs, _ := NewHttpServer(s)
		assert.False(t, check.IfNil(hs))

		hs.Start()

		err := hs.Close()
		assert.Equal(t, expectedErr, err)
	})
}

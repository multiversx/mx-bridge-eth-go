package gin

import (
	"net/http"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-eth-bridge/api/errors"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/stretchr/testify/assert"
)

func TestNewHttpServer(t *testing.T) {
	t.Parallel()

	t.Run("nil server should error", func(t *testing.T) {
		t.Parallel()

		hs, err := NewHttpServer(nil)
		assert.Equal(t, errors.ErrNilHttpServer, err)
		assert.True(t, check.IfNil(hs))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		hs, err := NewHttpServer(&http.Server{})
		assert.Nil(t, err)
		assert.False(t, check.IfNil(hs))
	})
}

func TestNewHttpServer_Start(t *testing.T) {
	t.Parallel()

	t.Run("closed server", func(t *testing.T) {
		t.Parallel()

		defer func() {
			r := recover()
			if r != nil {
				assert.Fail(t, "should not panic")
			}
		}()

		server := &http.Server{}
		_ = server.Close()

		hs, _ := NewHttpServer(server)
		assert.False(t, check.IfNil(hs))

		hs.Start()
	})
	t.Run("access denied", func(t *testing.T) {
		t.Parallel()

		defer func() {
			r := recover()
			if r != nil {
				assert.Fail(t, "should not panic")
			}
		}()

		server := &http.Server{}

		hs, _ := NewHttpServer(server)
		assert.False(t, check.IfNil(hs))

		hs.Start()
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		defer func() {
			r := recover()
			if r != nil {
				assert.Fail(t, "should not panic")
			}
		}()

		server := &http.Server{
			Addr: "127.0.0.1:8080",
		}
		hs, _ := NewHttpServer(server)
		assert.False(t, check.IfNil(hs))

		go hs.Start()
		time.Sleep(3 * time.Second)
		err := hs.Close()
		assert.Nil(t, err)
	})
}

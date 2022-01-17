package p2p

import (
	"errors"
	"testing"
	"time"

	elrondCore "github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/stretchr/testify/assert"
)

func TestNewPeerDenialEvaluator(t *testing.T) {
	t.Parallel()

	t.Run("nil blackListIDsCache should error", func(t *testing.T) {
		t.Parallel()

		pde, err := NewPeerDenialEvaluator(nil, &mock.TimeCacheStub{})
		assert.Nil(t, pde)
		assert.Equal(t, ErrNilBlackListIDsCache, err)
	})
	t.Run("nil blackListedPublicKeysCache should error", func(t *testing.T) {
		t.Parallel()

		pde, err := NewPeerDenialEvaluator(&mock.PeerBlackListHandlerStub{}, nil)
		assert.Nil(t, pde)
		assert.Equal(t, ErrNilBlackListedPublicKeysCache, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		pde, err := NewPeerDenialEvaluator(&mock.PeerBlackListHandlerStub{}, &mock.TimeCacheStub{})
		assert.False(t, check.IfNil(pde))
		assert.Nil(t, err)
	})
}

func Test_peerDenialEvaluator_IsDenied(t *testing.T) {
	t.Parallel()

	t.Run("should work, missing from both", func(t *testing.T) {
		t.Parallel()

		pde, err := NewPeerDenialEvaluator(&mock.PeerBlackListHandlerStub{}, &mock.TimeCacheStub{})
		assert.False(t, check.IfNil(pde))
		assert.Nil(t, err)

		assert.False(t, pde.IsDenied(pid))
	})
	t.Run("should work, found in blackListIDsCache", func(t *testing.T) {
		t.Parallel()

		bhStub := &mock.PeerBlackListHandlerStub{
			HasCalled: func(pid elrondCore.PeerID) bool {
				return true
			},
		}
		pde, err := NewPeerDenialEvaluator(bhStub, &mock.TimeCacheStub{})
		assert.False(t, check.IfNil(pde))
		assert.Nil(t, err)

		assert.True(t, pde.IsDenied(pid))
	})
	t.Run("should work, found in blackListedPublicKeysCache", func(t *testing.T) {
		t.Parallel()

		bhStub := &mock.TimeCacheStub{
			HasCalled: func(key string) bool {
				return true
			},
		}
		pde, err := NewPeerDenialEvaluator(&mock.PeerBlackListHandlerStub{}, bhStub)
		assert.False(t, check.IfNil(pde))
		assert.Nil(t, err)

		assert.True(t, pde.IsDenied(pid))
	})
}

func Test_peerDenialEvaluator_UpsertPeerID(t *testing.T) {
	t.Parallel()

	t.Run("UpsertCalled returns err", func(t *testing.T) {
		wasCalled := false
		expectedErr := errors.New("expected error")
		bhStub := &mock.PeerBlackListHandlerStub{
			UpsertCalled: func(pid elrondCore.PeerID, span time.Duration) error {
				wasCalled = true
				return expectedErr
			},
		}
		pde, err := NewPeerDenialEvaluator(bhStub, &mock.TimeCacheStub{})
		assert.False(t, check.IfNil(pde))
		assert.Nil(t, err)

		err = pde.UpsertPeerID(pid, time.Second)
		assert.Equal(t, expectedErr, err)
		assert.True(t, wasCalled)
	})
	t.Run("should work", func(t *testing.T) {
		wasCalled := false
		bhStub := &mock.PeerBlackListHandlerStub{
			UpsertCalled: func(pid elrondCore.PeerID, span time.Duration) error {
				wasCalled = true
				return nil
			},
		}
		pde, err := NewPeerDenialEvaluator(bhStub, &mock.TimeCacheStub{})
		assert.False(t, check.IfNil(pde))
		assert.Nil(t, err)

		err = pde.UpsertPeerID(pid, time.Second)
		assert.Nil(t, err)
		assert.True(t, wasCalled)
	})

}

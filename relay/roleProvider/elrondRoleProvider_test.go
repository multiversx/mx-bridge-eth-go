package roleProvider

import (
	"encoding/hex"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-eth-bridge/relay/roleProvider/mock"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/stretchr/testify/assert"
)

func createMockArgs() ArgsElrondRoleProvider {
	return ArgsElrondRoleProvider{
		ChainClient:     &mock.ChainClientStub{},
		UsePolling:      true,
		PollingInterval: time.Second,
		Log:             logger.GetOrCreate("test"),
	}
}

func TestNewElrondRoleProvider(t *testing.T) {
	t.Parallel()

	t.Run("nil chain client should error", func(t *testing.T) {
		args := createMockArgs()
		args.ChainClient = nil

		erp, err := NewElrondRoleProvider(args)
		assert.True(t, check.IfNil(erp))
		assert.Equal(t, ErrNilChainClient, err)
	})
	t.Run("nil logger should error", func(t *testing.T) {
		args := createMockArgs()
		args.Log = nil

		erp, err := NewElrondRoleProvider(args)
		assert.True(t, check.IfNil(erp))
		assert.Equal(t, ErrNilLogger, err)
	})
	t.Run("nil invalid polling should error", func(t *testing.T) {
		args := createMockArgs()
		args.PollingInterval = time.Second - time.Nanosecond

		erp, err := NewElrondRoleProvider(args)
		assert.True(t, check.IfNil(erp))
		assert.True(t, errors.Is(err, ErrInvalidValue))
	})
	t.Run("should work", func(t *testing.T) {
		args := createMockArgs()

		erp, err := NewElrondRoleProvider(args)
		assert.False(t, check.IfNil(erp))
		assert.Nil(t, err)

		_ = erp.Close()
	})
	t.Run("should work with invalid time but disabled polling", func(t *testing.T) {
		args := createMockArgs()
		args.UsePolling = false
		args.PollingInterval = time.Second - time.Nanosecond

		erp, err := NewElrondRoleProvider(args)
		assert.False(t, check.IfNil(erp))
		assert.Nil(t, err)

		_ = erp.Close()
	})
}

func TestElrondRoleProvider_QueryWhiteListed(t *testing.T) {
	t.Parallel()

	t.Run("query without polling errors each time will retry", func(t *testing.T) {
		expectedErr := errors.New("expected error")
		wgErrorPollingIntervalSet := &sync.WaitGroup{}
		wgErrorPollingIntervalSet.Add(1)

		numCalls := uint32(0)
		args := createMockArgs()
		args.ChainClient = &mock.ChainClientStub{
			ExecuteVmQueryOnBridgeContractCalled: func(function string, params ...[]byte) ([][]byte, error) {
				wgErrorPollingIntervalSet.Wait()

				atomic.AddUint32(&numCalls, 1)
				return nil, expectedErr
			},
		}

		args.UsePolling = false

		erp, _ := NewElrondRoleProvider(args)
		erp.pollingWhenError = time.Millisecond * 100
		wgErrorPollingIntervalSet.Done()

		time.Sleep(time.Millisecond * 350)
		assert.True(t, erp.loopStatus.IsSet())

		assert.Equal(t, uint32(4), atomic.LoadUint32(&numCalls))
		_ = erp.Close()

		time.Sleep(time.Second)
		assert.False(t, erp.loopStatus.IsSet())

		assert.Equal(t, uint32(4), atomic.LoadUint32(&numCalls))
	})
	t.Run("query with polling errors each time will retry", func(t *testing.T) {
		expectedErr := errors.New("expected error")
		wgErrorPollingIntervalSet := &sync.WaitGroup{}
		wgErrorPollingIntervalSet.Add(1)

		numCalls := uint32(0)
		args := createMockArgs()
		args.ChainClient = &mock.ChainClientStub{
			ExecuteVmQueryOnBridgeContractCalled: func(function string, params ...[]byte) ([][]byte, error) {
				wgErrorPollingIntervalSet.Wait()

				atomic.AddUint32(&numCalls, 1)
				return nil, expectedErr
			},
		}

		erp, _ := NewElrondRoleProvider(args)
		erp.pollingWhenError = time.Millisecond * 100
		wgErrorPollingIntervalSet.Done()

		time.Sleep(time.Millisecond * 350)
		assert.True(t, erp.loopStatus.IsSet())

		assert.Equal(t, uint32(4), atomic.LoadUint32(&numCalls))
		_ = erp.Close()

		time.Sleep(time.Second)
		assert.False(t, erp.loopStatus.IsSet())

		assert.Equal(t, uint32(4), atomic.LoadUint32(&numCalls))
	})
	t.Run("query with large polling interval should work", func(t *testing.T) {
		whitelistedAddresses := [][]byte{
			[]byte("address 1"),
			[]byte("address 2"),
		}

		args := createMockArgs()
		args.PollingInterval = time.Hour
		args.ChainClient = &mock.ChainClientStub{
			ExecuteVmQueryOnBridgeContractCalled: func(function string, params ...[]byte) ([][]byte, error) {
				return whitelistedAddresses, nil
			},
		}

		erp, _ := NewElrondRoleProvider(args)
		time.Sleep(time.Second)
		_ = erp.Close()

		for _, addr := range whitelistedAddresses {
			assert.True(t, erp.IsWhitelisted(hex.EncodeToString(addr)))
		}

		assert.False(t, erp.IsWhitelisted("random address"))
		erp.mut.RLock()
		assert.Equal(t, 2, len(erp.whitelistedAddresses))
		erp.mut.RUnlock()
	})
	t.Run("query without polling should query only once", func(t *testing.T) {
		whitelistedAddresses := [][]byte{
			[]byte("address 1"),
			[]byte("address 2"),
		}

		args := createMockArgs()
		args.UsePolling = false
		args.PollingInterval = time.Millisecond
		numCalls := uint32(0)
		args.ChainClient = &mock.ChainClientStub{
			ExecuteVmQueryOnBridgeContractCalled: func(function string, params ...[]byte) ([][]byte, error) {
				atomic.AddUint32(&numCalls, 1)
				return whitelistedAddresses, nil
			},
		}

		erp, _ := NewElrondRoleProvider(args)
		time.Sleep(time.Second)
		assert.False(t, erp.loopStatus.IsSet())

		_ = erp.Close()

		for _, addr := range whitelistedAddresses {
			assert.True(t, erp.IsWhitelisted(hex.EncodeToString(addr)))
		}

		assert.False(t, erp.IsWhitelisted("random address"))
		erp.mut.RLock()
		assert.Equal(t, 2, len(erp.whitelistedAddresses))
		erp.mut.RUnlock()
		assert.Equal(t, uint32(1), atomic.LoadUint32(&numCalls))
	})

}

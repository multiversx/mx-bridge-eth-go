package roleProvider

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-eth-bridge/relay/roleProvider/mock"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/data"
	"github.com/stretchr/testify/assert"
)

func createMockArgs() ArgsElrondRoleProvider {
	return ArgsElrondRoleProvider{
		ChainInteractor: &mock.ChainInteractorStub{},
		PollingInterval: time.Second,
		Log:             logger.GetOrCreate("test"),
	}
}

func TestNewElrondRoleProvider(t *testing.T) {
	t.Parallel()

	t.Run("nil chain interactor should error", func(t *testing.T) {
		args := createMockArgs()
		args.ChainInteractor = nil

		erp, err := NewElrondRoleProvider(args)
		assert.True(t, check.IfNil(erp))
		assert.Equal(t, ErrNilChainInteractor, err)
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
}

func TestElrondRoleProvider_QueryPollingErrorsEachTimeWillRetry(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	wgErrorPollingIntervalSet := &sync.WaitGroup{}
	wgErrorPollingIntervalSet.Add(1)

	numCalls := uint32(0)
	args := createMockArgs()
	args.ChainInteractor = &mock.ChainInteractorStub{
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
}

func TestElrondProvider_QueryWithLargePollingIntervalShouldWork(t *testing.T) {
	t.Parallel()

	whitelistedAddresses := [][]byte{
		[]byte("address 1"),
		[]byte("address 2"),
	}

	args := createMockArgs()
	args.PollingInterval = time.Hour
	args.ChainInteractor = &mock.ChainInteractorStub{
		ExecuteVmQueryOnBridgeContractCalled: func(function string, params ...[]byte) ([][]byte, error) {
			return whitelistedAddresses, nil
		},
	}

	erp, _ := NewElrondRoleProvider(args)
	time.Sleep(time.Second)
	_ = erp.Close()

	for _, addr := range whitelistedAddresses {
		addressHandler := data.NewAddressFromBytes(addr)
		assert.True(t, erp.IsWhitelisted(addressHandler))
	}

	randomAddress := data.NewAddressFromBytes([]byte("random address"))
	assert.False(t, erp.IsWhitelisted(randomAddress))
	assert.False(t, erp.IsWhitelisted(nil))
	erp.mut.RLock()
	assert.Equal(t, 2, len(erp.whitelistedAddresses))
	erp.mut.RUnlock()
}

func TestElrondProvider_QueryWithSmallPollingIntervalShouldWork(t *testing.T) {
	t.Parallel()

	whitelistedAddresses := [][]byte{
		[]byte("address 1"),
		[]byte("address 2"),
	}

	args := createMockArgs()
	args.PollingInterval = time.Second
	numCalls := uint32(0)
	args.ChainInteractor = &mock.ChainInteractorStub{
		ExecuteVmQueryOnBridgeContractCalled: func(function string, params ...[]byte) ([][]byte, error) {
			atomic.AddUint32(&numCalls, 1)
			return whitelistedAddresses, nil
		},
	}

	erp, _ := NewElrondRoleProvider(args)
	time.Sleep(time.Millisecond * 3500)
	_ = erp.Close()

	assert.Equal(t, uint32(4), atomic.LoadUint32(&numCalls))

	for _, addr := range whitelistedAddresses {
		addressHandler := data.NewAddressFromBytes(addr)
		assert.True(t, erp.IsWhitelisted(addressHandler))
	}

	randomAddress := data.NewAddressFromBytes([]byte("random address"))
	assert.False(t, erp.IsWhitelisted(randomAddress))
	assert.False(t, erp.IsWhitelisted(nil))
	erp.mut.RLock()
	assert.Equal(t, 2, len(erp.whitelistedAddresses))
	erp.mut.RUnlock()
}

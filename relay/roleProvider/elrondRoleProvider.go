package roleProvider

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core/atomic"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/core"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/data"
)

const getAllStakedRelayersFunctionName = "getAllStakedRelayers"
const pollingIntervalInCaseOfError = time.Second * 5
const minimumPollingInterval = time.Second

// ArgsElrondRoleProvider is the argument for the elrond role provider constructor
type ArgsElrondRoleProvider struct {
	ChainInteractor ChainInteractor
	Log             logger.Logger
	PollingInterval time.Duration
}

type elrondRoleProvider struct {
	chainInteractor      ChainInteractor
	log                  logger.Logger
	whitelistedAddresses map[string]struct{}
	cancel               func()
	pollingInterval      time.Duration
	pollingWhenError     time.Duration
	mut                  sync.RWMutex
	loopStatus           atomic.Flag
}

// NewElrondRoleProvider creates a new elrond role provider instance able to fetch the whitelisted addresses
func NewElrondRoleProvider(args ArgsElrondRoleProvider) (*elrondRoleProvider, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	erp := &elrondRoleProvider{
		chainInteractor:      args.ChainInteractor,
		log:                  args.Log,
		pollingInterval:      args.PollingInterval,
		whitelistedAddresses: make(map[string]struct{}),
		pollingWhenError:     pollingIntervalInCaseOfError,
	}

	ctx, cancel := context.WithCancel(context.Background())
	erp.cancel = cancel

	go erp.requestsLoop(ctx)

	return erp, nil
}

func checkArgs(args ArgsElrondRoleProvider) error {
	if check.IfNil(args.ChainInteractor) {
		return ErrNilChainInteractor
	}
	if check.IfNil(args.Log) {
		return ErrNilLogger
	}
	if args.PollingInterval < minimumPollingInterval {
		return fmt.Errorf("%w for PollingInterval", ErrInvalidValue)
	}

	return nil
}

func (erp *elrondRoleProvider) requestsLoop(ctx context.Context) {
	erp.loopStatus.Set()
	defer erp.loopStatus.Unset()

	for {
		pollingChan := time.After(erp.pollingInterval)

		results, err := erp.chainInteractor.ExecuteVmQueryOnBridgeContract(getAllStakedRelayersFunctionName)
		if err != nil {
			erp.log.Error("error in elrondRoleProvider.requestsLoop",
				"error", err, "retrying after", erp.pollingWhenError)
			pollingChan = time.After(erp.pollingWhenError)
		} else {
			erp.processResults(results)
		}

		select {
		case <-pollingChan:
		case <-ctx.Done():
			erp.log.Debug("role provider main requests loop is closing...")
			return
		}
	}
}

func (erp *elrondRoleProvider) processResults(results [][]byte) {
	currentList := make([]string, 0, len(results))

	erp.mut.Lock()
	erp.whitelistedAddresses = make(map[string]struct{})

	for _, result := range results {
		address := data.NewAddressFromBytes(result)
		currentList = append(currentList, address.AddressAsBech32String())

		erp.whitelistedAddresses[string(address.AddressBytes())] = struct{}{}
	}
	erp.mut.Unlock()

	erp.log.Debug("fetched whitelisted addresses:\n" + strings.Join(currentList, "\n"))
}

// IsWhitelisted returns true if the non-nil address provided is whitelisted or not
func (erp *elrondRoleProvider) IsWhitelisted(address core.AddressHandler) bool {
	if check.IfNil(address) {
		return false
	}

	erp.mut.RLock()
	defer erp.mut.RUnlock()

	_, exists := erp.whitelistedAddresses[string(address.AddressBytes())]

	return exists
}

// Close will close any containing members and clean any go routines associated
func (erp *elrondRoleProvider) Close() error {
	erp.cancel()

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (erp *elrondRoleProvider) IsInterfaceNil() bool {
	return erp == nil
}

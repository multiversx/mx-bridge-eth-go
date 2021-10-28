package roleProvider

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core/atomic"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/data"
)

const getAllStakedRelayersFunctionName = "getAllStakedRelayers"
const pollingIntervalInCaseOfError = time.Second * 5
const initialStartupDelay = time.Nanosecond
const minimumPollingInterval = time.Second

// ArgsElrondRoleProvider is the argument for the elrond role provider constructor
type ArgsElrondRoleProvider struct {
	ChainClient     ChainClient
	UsePolling      bool
	PollingInterval time.Duration
	Log             logger.Logger
}

type elrondRoleProvider struct {
	chainClient          ChainClient
	log                  logger.Logger
	usePolling           bool
	pollingInterval      time.Duration
	pollingWhenError     time.Duration
	mut                  sync.RWMutex
	whitelistedAddresses map[string]struct{}
	cancel               func()
	loopStatus           atomic.Flag
}

// NewElrondRoleProvider creates a new elrond role provider instance able to fetch the whitelisted addresses
func NewElrondRoleProvider(args ArgsElrondRoleProvider) (*elrondRoleProvider, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	erp := &elrondRoleProvider{
		chainClient:          args.ChainClient,
		log:                  args.Log,
		usePolling:           args.UsePolling,
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
	if check.IfNil(args.ChainClient) {
		return ErrNilChainClient
	}
	if check.IfNil(args.Log) {
		return ErrNilLogger
	}
	if args.PollingInterval < minimumPollingInterval && args.UsePolling {
		return fmt.Errorf("%w for PollingInterval", ErrInvalidValue)
	}

	return nil
}

func (erp *elrondRoleProvider) requestsLoop(ctx context.Context) {
	erp.loopStatus.Set()
	defer erp.loopStatus.Unset()

	pollingChan := time.After(initialStartupDelay)
	for {
		select {
		case <-pollingChan:
		case <-ctx.Done():
			erp.log.Debug("role provider main requests loop is closing...")
			return
		}

		results, err := erp.chainClient.ExecuteVmQueryOnBridgeContract(getAllStakedRelayersFunctionName)
		if err != nil {
			erp.log.Error("error in elrondRoleProvider.requestsLoop",
				"error", err, "retrying after", erp.pollingWhenError)
			pollingChan = time.After(erp.pollingWhenError)
			continue
		}

		erp.processResults(results)

		if !erp.usePolling {
			return
		}

		pollingChan = time.After(erp.pollingInterval)
	}
}

func (erp *elrondRoleProvider) processResults(results [][]byte) {
	currentList := make([]string, 0, len(results))
	newAddresses := make([]string, 0)
	previousAddresses := erp.whitelistedAddresses
	erp.mut.Lock()
	erp.whitelistedAddresses = make(map[string]struct{})

	for _, result := range results {
		address := data.NewAddressFromBytes(result)
		currentList = append(currentList, address.AddressAsBech32String())
		hexAddress := hex.EncodeToString(result)
		_, alreadyExists := previousAddresses[hexAddress]
		if !alreadyExists {
			newAddresses = append(newAddresses, address.AddressAsBech32String())
		} else {
			delete(previousAddresses, hexAddress)
		}

		erp.whitelistedAddresses[hexAddress] = struct{}{}
	}
	erp.mut.Unlock()
	message := "fetched whitelisted addresses:"
	if len(newAddresses) > 0 {
		message += "\n\tnew joiners:\n\t" + strings.Join(newAddresses, "\n\t")
	}
	if len(previousAddresses) > 0 {
		message += "\n\tleavers:\n"
		for k := range previousAddresses {
			message += "\t" + k + "\n"
		}
	}
	if message == "fetched whitelisted addresses:" {
		message += " no changes!"
	}
	erp.log.Debug(message)
}

// IsWhitelisted returns true if the hex address provided is whitelisted
func (erp *elrondRoleProvider) IsWhitelisted(hexAddress string) bool {
	erp.mut.RLock()
	defer erp.mut.RUnlock()

	_, exists := erp.whitelistedAddresses[hexAddress]

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

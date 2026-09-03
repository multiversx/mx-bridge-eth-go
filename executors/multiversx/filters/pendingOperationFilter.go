package filters

import (
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/multiversx/mx-bridge-eth-go/config"
	"github.com/multiversx/mx-bridge-eth-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-sdk-go/data"
)

const (
	wildcardString   = "*"
	emptyString      = ""
	ethAddressPrefix = "0x"
)

var ethWildcardString = ""

func init() {
	var ethAddressWildcard = common.Address{}
	ethAddressWildcard.SetBytes([]byte(wildcardString))
	ethWildcardString = ethAddressWildcard.String()
}

type pendingOperationFilter struct {
	allowedEthAddresses []string
	deniedEthAddresses  []string
	allowedMvxAddresses []string
	deniedMvxAddresses  []string
	allowedTokens       []string
	deniedTokens        []string
}

// NewPendingOperationFilter creates a new instance of type pendingOperationFilter
func NewPendingOperationFilter(cfg config.PendingOperationsFilterConfig, log logger.Logger) (*pendingOperationFilter, error) {
	if check.IfNil(log) {
		return nil, errNilLogger
	}
	if len(cfg.AllowedMvxAddresses)+len(cfg.AllowedEthAddresses)+len(cfg.AllowedTokens) == 0 {
		return nil, errNoItemsAllowed
	}

	filter := &pendingOperationFilter{}
	err := filter.parseConfigs(cfg)
	if err != nil {
		return nil, err
	}

	err = filter.checkLists()
	if err != nil {
		return nil, err
	}

	log.Info("NewPendingOperationFilter config options",
		"DeniedEthAddresses", strings.Join(filter.deniedEthAddresses, ", "),
		"DeniedMvxAddresses", strings.Join(filter.deniedMvxAddresses, ", "),
		"DeniedTokens", strings.Join(filter.deniedTokens, ", "),
		"AllowedEthAddresses", strings.Join(filter.allowedEthAddresses, ", "),
		"AllowedMvxAddresses", strings.Join(filter.allowedMvxAddresses, ", "),
		"AllowedTokens", strings.Join(filter.allowedTokens, ", "),
	)

	return filter, nil
}

func (filter *pendingOperationFilter) parseConfigs(cfg config.PendingOperationsFilterConfig) error {
	var err error

	// denied lists do not support wildcard items
	filter.deniedEthAddresses, err = parseList(cfg.DeniedEthAddresses, wildcardString)
	if err != nil {
		return fmt.Errorf("%w in list DeniedEthAddresses", err)
	}

	filter.deniedMvxAddresses, err = parseList(cfg.DeniedMvxAddresses, wildcardString)
	if err != nil {
		return fmt.Errorf("%w in list DeniedMvxAddresses", err)
	}

	filter.deniedTokens, err = parseList(cfg.DeniedTokens, wildcardString)
	if err != nil {
		return fmt.Errorf("%w in list DeniedTokens", err)
	}

	// allowed lists do not support empty items
	filter.allowedEthAddresses, err = parseList(cfg.AllowedEthAddresses, emptyString)
	if err != nil {
		return fmt.Errorf("%w in list AllowedEthAddresses", err)
	}

	filter.allowedMvxAddresses, err = parseList(cfg.AllowedMvxAddresses, emptyString)
	if err != nil {
		return fmt.Errorf("%w in list AllowedMvxAddresses", err)
	}

	filter.allowedTokens, err = parseList(cfg.AllowedTokens, emptyString)
	if err != nil {
		return fmt.Errorf("%w in list AllowedTokens", err)
	}

	return nil
}

func parseList(list []string, unsupportedMarker string) ([]string, error) {
	newList := make([]string, 0, len(list))
	for index, item := range list {
		item = strings.ToLower(item)
		item = strings.Trim(item, "\r\n \t")
		if item == unsupportedMarker {
			return nil, fmt.Errorf("%w %s on item at index %d", errUnsupportedMarker, unsupportedMarker, index)
		}

		newList = append(newList, item)
	}

	return newList, nil
}

func (filter *pendingOperationFilter) checkLists() error {
	err := filter.checkList(filter.allowedEthAddresses, checkEthItemValid)
	if err != nil {
		return fmt.Errorf("%w in list AllowedEthAddresses", err)
	}

	err = filter.checkList(filter.deniedEthAddresses, checkEthItemValid)
	if err != nil {
		return fmt.Errorf("%w in list DeniedEthAddresses", err)
	}

	err = filter.checkList(filter.allowedMvxAddresses, checkMvxItemValid)
	if err != nil {
		return fmt.Errorf("%w in list AllowedMvxAddresses", err)
	}

	err = filter.checkList(filter.deniedMvxAddresses, checkMvxItemValid)
	if err != nil {
		return fmt.Errorf("%w in list DeniedMvxAddresses", err)
	}

	return nil
}

func (filter *pendingOperationFilter) checkList(list []string, checkItem func(item string) error) error {
	for index, item := range list {
		if item == wildcardString {
			continue
		}

		err := checkItem(item)
		if err != nil {
			return fmt.Errorf("%w on item at index %d", err, index)
		}
	}

	return nil
}

func checkMvxItemValid(item string) error {
	_, errNewAddr := data.NewAddressFromBech32String(item)
	return errNewAddr
}

func checkEthItemValid(item string) error {
	if !strings.HasPrefix(item, ethAddressPrefix) {
		return fmt.Errorf("%w (missing %s prefix)", errMissingEthPrefix, ethAddressPrefix)
	}

	return nil
}

// ShouldExecute returns true if the To, From or token are not denied and allowed
func (filter *pendingOperationFilter) ShouldExecute(callData core.ProxySCCompleteCallData) bool {
	if check.IfNil(callData.To) {
		return false
	}

	toAddress, err := callData.To.AddressAsBech32String()
	if err != nil {
		return false
	}

	isSpecificallyDenied := filter.stringExistsInList(callData.From.String(), filter.deniedEthAddresses, ethWildcardString) ||
		filter.stringExistsInList(toAddress, filter.deniedMvxAddresses, wildcardString) ||
		filter.stringExistsInList(callData.Token, filter.deniedTokens, wildcardString)
	if isSpecificallyDenied {
		return false
	}

	isAllowed := filter.stringExistsInList(callData.From.String(), filter.allowedEthAddresses, ethWildcardString) ||
		filter.stringExistsInList(toAddress, filter.allowedMvxAddresses, wildcardString) ||
		filter.stringExistsInList(callData.Token, filter.allowedTokens, wildcardString)

	return isAllowed
}

func (filter *pendingOperationFilter) stringExistsInList(needle string, haystack []string, wildcardMarker string) bool {
	needle = strings.ToLower(needle)
	wildcardMarker = strings.ToLower(wildcardMarker)

	for _, item := range haystack {
		if item == wildcardMarker {
			return true
		}

		if item == needle {
			return true
		}
	}

	return false
}

// IsInterfaceNil returns true if there is no value under the interface
func (filter *pendingOperationFilter) IsInterfaceNil() bool {
	return filter == nil
}

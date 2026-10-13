// Package batchRedirect replaces the recipient of specific MultiversX deposits before their batch is signed and executed on Ethereum.
package batchRedirect

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	bridgeCore "github.com/multiversx/mx-bridge-eth-go/core"
)

// ErrRecoveryAddressNotSet signals that a redirect matched but no recovery address is configured
var ErrRecoveryAddressNotSet = errors.New("recovery address for redirected deposits is not configured")

// ErrUnexpectedDeposit signals a deposit to a redirected recipient, in a redirected batch, that does not match its redirect exactly
var ErrUnexpectedDeposit = errors.New("deposit to a redirected recipient does not match the expected redirect")

// ErrInvalidRecoveryAddress signals a configured recovery address that is malformed, zero or one of the redirected recipients
var ErrInvalidRecoveryAddress = errors.New("invalid recovery address")

// Redirect identifies one deposit whose recipient is replaced by the recovery address
type Redirect struct {
	BatchID      uint64
	DepositNonce uint64
	From         []byte
	OriginalTo   []byte
	SourceToken  string
	Amount       *big.Int
}

var (
	redirectedSender    = mustDecodeHex("30adcf1257a0c227befca62f5d073b90a6c9875ee7e80e48ec0c3bbd5f622d13")
	redirectedRecipient = mustDecodeHex("717fb1435f5fa0cc1c7e541e5d92ef71342fcd41")
	usdcUnit            = big.NewInt(0).Exp(big.NewInt(10), big.NewInt(18), nil)
)

// Redirects lists the deposits replaced by the recovery address
var Redirects = []Redirect{
	{BatchID: 5956, DepositNonce: 7717, From: redirectedSender, OriginalTo: redirectedRecipient, SourceToken: "BSCUSDC-887875", Amount: usdc(49_999)},
	{BatchID: 5956, DepositNonce: 7718, From: redirectedSender, OriginalTo: redirectedRecipient, SourceToken: "BSCUSDC-887875", Amount: usdc(30_440)},
}

// Redirector applies Redirects using the configured recovery address
type Redirector struct {
	recoveryAddress []byte
}

// NewRedirector validates the configured recovery address; an empty value is allowed and makes matching batches refused
func NewRedirector(configuredRecoveryAddress string) (*Redirector, error) {
	trimmed := strings.TrimPrefix(strings.TrimSpace(configuredRecoveryAddress), "0x")
	if len(trimmed) == 0 {
		return &Redirector{}, nil
	}

	address, err := hex.DecodeString(trimmed)
	if err != nil || len(address) != common.AddressLength {
		return nil, fmt.Errorf("%w: %q is not a 20-byte hex address", ErrInvalidRecoveryAddress, configuredRecoveryAddress)
	}
	if bytes.Equal(address, make([]byte, common.AddressLength)) {
		return nil, fmt.Errorf("%w: zero address", ErrInvalidRecoveryAddress)
	}
	for _, redirect := range Redirects {
		if bytes.Equal(address, redirect.OriginalTo) {
			return nil, fmt.Errorf("%w: it is a redirected recipient", ErrInvalidRecoveryAddress)
		}
	}

	return &Redirector{recoveryAddress: address}, nil
}

// RecoveryAddress returns the configured recovery address as a 0x-prefixed hex string, or empty when not configured
func (redirector *Redirector) RecoveryAddress() string {
	if redirector == nil || len(redirector.recoveryAddress) == 0 {
		return ""
	}

	return common.BytesToAddress(redirector.recoveryAddress).Hex()
}

// Apply returns a copy of the batch with every redirected deposit sent to the recovery address, and how many deposits were redirected.
// It refuses the batch when a redirect matches without a configured recovery address, or when a deposit in a redirected batch
// pays a redirected recipient without matching its redirect exactly.
func (redirector *Redirector) Apply(batch *bridgeCore.TransferBatch) (*bridgeCore.TransferBatch, int, error) {
	if batch == nil {
		return nil, 0, nil
	}

	deposits := make([]*bridgeCore.DepositTransfer, len(batch.Deposits))
	numRedirected := 0
	for index, deposit := range batch.Deposits {
		deposits[index] = deposit
		redirect, matched, err := find(batch.ID, deposit)
		if err != nil {
			return nil, 0, fmt.Errorf("%w: batch %d, deposit index %d, %s", err, batch.ID, index, deposit.String())
		}
		if !matched {
			continue
		}
		if redirector == nil || len(redirector.recoveryAddress) == 0 {
			return nil, 0, fmt.Errorf("%w: batch %d, deposit nonce %d", ErrRecoveryAddressNotSet, batch.ID, redirect.DepositNonce)
		}

		redirected := *deposit
		redirected.ToBytes = append([]byte(nil), redirector.recoveryAddress...)
		redirected.DisplayableTo = redirector.RecoveryAddress()
		deposits[index] = &redirected
		numRedirected++
	}
	if numRedirected == 0 {
		return batch, 0, nil
	}

	result := *batch
	result.Deposits = deposits
	return &result, numRedirected, nil
}

func find(batchID uint64, deposit *bridgeCore.DepositTransfer) (Redirect, bool, error) {
	for _, redirect := range Redirects {
		if redirect.matches(batchID, deposit) {
			return redirect, true, nil
		}
	}
	for _, redirect := range Redirects {
		if redirect.BatchID == batchID && bytes.Equal(redirect.OriginalTo, deposit.ToBytes) {
			return Redirect{}, false, ErrUnexpectedDeposit
		}
	}

	return Redirect{}, false, nil
}

func (redirect Redirect) matches(batchID uint64, deposit *bridgeCore.DepositTransfer) bool {
	return redirect.BatchID == batchID &&
		redirect.DepositNonce == deposit.Nonce &&
		bytes.Equal(redirect.From, deposit.FromBytes) &&
		bytes.Equal(redirect.OriginalTo, deposit.ToBytes) &&
		redirect.SourceToken == string(deposit.SourceTokenBytes) &&
		deposit.Amount != nil && redirect.Amount.Cmp(deposit.Amount) == 0
}

func usdc(units int64) *big.Int {
	return big.NewInt(0).Mul(big.NewInt(units), usdcUnit)
}

func mustDecodeHex(value string) []byte {
	decoded, err := hex.DecodeString(value)
	if err != nil {
		panic(err)
	}

	return decoded
}

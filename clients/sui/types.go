package sui

import (
	"encoding/json"
	"fmt"
	"github.com/block-vision/sui-go-sdk/models"
)

// Batch represents a batch of deposits from safe contract
type Batch struct {
	Nonce                  uint64
	TimestampMs            uint64
	LastUpdatedTimestampMs uint64
	DepositsCount          uint16
}

// Deposit represents a single deposit in the batch
type Deposit struct {
	Nonce          uint64
	TokenTypeBytes []byte
	Amount         uint64
	Sender         models.SuiAddressBytes
	Recipient      []byte
}

// TokenTransferGroup holds the data for a group of transfers involving a specific token
type TokenTransferGroup struct {
	Tokens     [][]byte
	Recipients []models.SuiAddressBytes
	Amounts    []uint64
	Nonces     []uint64
	Signatures [][96]byte
}

// InspectResult represents the result of inspecting a transaction block
type InspectResult struct {
	ReturnValues []ReturnValue `json:"returnValues"`
}

// ReturnValue represents a single return value with its type
type ReturnValue struct {
	Bytes []byte
	Type  string
}

// UnmarshalJSON custom unmarshaler for ReturnValue
func (rv *ReturnValue) UnmarshalJSON(data []byte) error {
	var parts []json.RawMessage
	err := json.Unmarshal(data, &parts)
	if err != nil {
		return fmt.Errorf("split into parts: %w", err)
	}
	if len(parts) != 2 {
		return fmt.Errorf("expected 2 elements, got %d", len(parts))
	}

	err = json.Unmarshal(parts[0], &rv.Bytes)
	if err != nil {
		return fmt.Errorf("decode byte array: %w", err)
	}

	err = json.Unmarshal(parts[1], &rv.Type)
	if err != nil {
		return fmt.Errorf("decode type string: %w", err)
	}

	return nil
}

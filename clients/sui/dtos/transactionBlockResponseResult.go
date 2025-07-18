package dtos

import (
	"encoding/json"
	"fmt"
)

// InspectResult represents the result of inspecting a transaction block
type InspectResult struct {
	ReturnValues []ReturnValue `json:"returnValues"`
}

// ReturnValue represents a single return value with its type
type ReturnValue struct {
	Bytes []byte
	Type  string
}

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

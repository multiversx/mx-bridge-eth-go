package ethereum

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
)

// DepositInfo is the deposit info list
type DepositInfo struct {
	DepositNonce            uint64         `json:"DepositNonce"`
	Token                   string         `json:"Token"`
	ContractAddressString   string         `json:"ContractAddress"`
	Decimals                byte           `json:"Decimals"`
	ContractAddress         common.Address `json:"-"`
	Amount                  *big.Int       `json:"-"`
	AmountString            string         `json:"Amount"`
	DenominatedAmount       *big.Float     `json:"-"`
	DenominatedAmountString string         `json:"DenominatedAmount"`
}

// BatchInfo is the batch info list
type BatchInfo struct {
	OldSafeContractAddress string         `json:"OldSafeContractAddress"`
	NewSafeContractAddress string         `json:"NewSafeContractAddress"`
	BatchID                uint64         `json:"BatchID"`
	MessageHash            common.Hash    `json:"MessageHash"`
	DepositsInfo           []*DepositInfo `json:"DepositsInfo"`
}

// SignatureInfo is the struct holding signature info
type SignatureInfo struct {
	Address     string `json:"Address"`
	MessageHash string `json:"MessageHash"`
	Signature   string `json:"Signature"`
}

// FloatWrapper is a wrapper of the big.Float that supports specifying if the value is maximum
type FloatWrapper struct {
	*big.Float
	IsMax bool
}

var maxValues = []string{"all", "max", "*"}

// TokensBalancesDisplayString will convert the deposit balances into a human-readable string
func TokensBalancesDisplayString(batchInfo *BatchInfo) string {
	maxTokenLen := 0
	maxIntegerValueLen := 0
	integerIndex := 0
	tokenIntegerSpace := make(map[string]int)
	decimalSeparator := "." // src/math/big/ftoa.go L302
	for _, deposit := range batchInfo.DepositsInfo {
		if len(deposit.Token) > maxTokenLen {
			maxTokenLen = len(deposit.Token)
		}

		valueParts := strings.Split(deposit.DenominatedAmountString, decimalSeparator)
		integerPart := valueParts[integerIndex]
		if len(integerPart) > maxIntegerValueLen {
			maxIntegerValueLen = len(valueParts[integerIndex])
		}
		tokenIntegerSpace[deposit.Token] = len(valueParts[integerIndex])
	}

	tokens := make([]string, 0, len(batchInfo.DepositsInfo))
	for _, deposit := range batchInfo.DepositsInfo {
		spaceRequired := strings.Repeat(" ", maxTokenLen-len(deposit.Token)+maxIntegerValueLen-tokenIntegerSpace[deposit.Token])
		tokenInfo := fmt.Sprintf(" %s: %s%s", deposit.Token, spaceRequired, deposit.DenominatedAmountString)

		tokens = append(tokens, tokenInfo)
	}

	return strings.Join(tokens, "\n")
}

// ConvertPartialMigrationStringToMap converts the partial migration string to its map representation
func ConvertPartialMigrationStringToMap(partialMigration string) (map[string]*FloatWrapper, error) {
	partsSeparator := ","
	tokenAmountSeparator := ":"
	parts := strings.Split(partialMigration, partsSeparator)

	partialMap := make(map[string]*FloatWrapper)
	for _, part := range parts {
		part = strings.Trim(part, " \t\n")
		splt := strings.Split(part, tokenAmountSeparator)
		if len(splt) != 2 {
			return nil, fmt.Errorf("%w at token %s, invalid format", errInvalidPartialMigrationString, part)
		}

		token := splt[0]
		if isMaxValueString(splt[1]) {
			partialMap[token] = &FloatWrapper{
				Float: big.NewFloat(0),
				IsMax: true,
			}

			continue
		}

		amount, ok := big.NewFloat(0).SetString(splt[1])
		if !ok {
			return nil, fmt.Errorf("%w at token %s, not a number", errInvalidPartialMigrationString, part)
		}

		if partialMap[token] == nil {
			partialMap[token] = &FloatWrapper{
				Float: big.NewFloat(0).Set(amount),
				IsMax: false,
			}
			continue
		}

		if partialMap[token].IsMax {
			// do not attempt to add something to an already max float
			continue
		}

		partialMap[token].Add(partialMap[token].Float, amount)
	}

	return partialMap, nil
}

func isMaxValueString(value string) bool {
	value = strings.ToLower(value)

	for _, maxValue := range maxValues {
		if value == maxValue {
			return true
		}
	}

	return false
}

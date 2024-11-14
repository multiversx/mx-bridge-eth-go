package ethereum

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTokensBalancesDisplayString(t *testing.T) {
	t.Parallel()

	batchInfo := &BatchInfo{
		DepositsInfo: []*DepositInfo{
			{
				Token:                   "ETHUSDC-220753",
				DenominatedAmountString: "5900401.957669",
			},
			{
				Token:                   "ETHUTK-8cdf7a",
				DenominatedAmountString: "224564287.8192652",
			},
			{
				Token:                   "ETHUSDT-9c73c6",
				DenominatedAmountString: "542188.933704",
			},
			{
				Token:                   "ETHBUSD-450923",
				DenominatedAmountString: "22294.352736330155",
			},
			{
				Token:                   "ETHHMT-18538a",
				DenominatedAmountString: "435",
			},
			{
				Token:                   "ETHCGG-ee4e0c",
				DenominatedAmountString: "1594290.750967581",
			},
			{
				Token:                   "ETHINFRA-60a3bf2",
				DenominatedAmountString: "141172.59598039952",
			},
			{
				Token:                   "ETHWBTC-74e282",
				DenominatedAmountString: "39.46386326",
			},
			{
				Token:                   "ETHWETH-e1c126",
				DenominatedAmountString: "664.1972941951753",
			},
			{
				Token:                   "ETHWSDAI-572803",
				DenominatedAmountString: "5431.516086574386",
			},
			{
				Token:                   "ETHWDAI-bd65f9",
				DenominatedAmountString: "118591.44846500318",
			},
			{
				Token:                   "ETHUMB-291202",
				DenominatedAmountString: "4065258.3239772925",
			},
		},
	}

	expectedString :=
		` ETHUSDC-220753:     5900401.957669
 ETHUTK-8cdf7a:    224564287.8192652
 ETHUSDT-9c73c6:      542188.933704
 ETHBUSD-450923:       22294.352736330155
 ETHHMT-18538a:          435
 ETHCGG-ee4e0c:      1594290.750967581
 ETHINFRA-60a3bf2:    141172.59598039952
 ETHWBTC-74e282:          39.46386326
 ETHWETH-e1c126:         664.1972941951753
 ETHWSDAI-572803:       5431.516086574386
 ETHWDAI-bd65f9:      118591.44846500318
 ETHUMB-291202:      4065258.3239772925`

	assert.Equal(t, expectedString, TokensBalancesDisplayString(batchInfo))
}

func TestConvertPartialMigrationStringToMap(t *testing.T) {
	t.Parallel()

	t.Run("invalid part should error", func(t *testing.T) {
		t.Parallel()

		str := "k,f"
		results, err := ConvertPartialMigrationStringToMap(str)
		assert.Nil(t, results)
		assert.ErrorIs(t, err, errInvalidPartialMigrationString)
		assert.Contains(t, err.Error(), "at token k, invalid format")

		str = "k:1:2,f"
		results, err = ConvertPartialMigrationStringToMap(str)
		assert.Nil(t, results)
		assert.ErrorIs(t, err, errInvalidPartialMigrationString)
		assert.Contains(t, err.Error(), "at token k:1:2, invalid format")
	})
	t.Run("amount is empty should error", func(t *testing.T) {
		t.Parallel()

		str := "k:,f:1"
		results, err := ConvertPartialMigrationStringToMap(str)
		assert.Nil(t, results)
		assert.ErrorIs(t, err, errInvalidPartialMigrationString)
		assert.Contains(t, err.Error(), "at token k:, not a number")

		str = "k:1d2,f"
		results, err = ConvertPartialMigrationStringToMap(str)
		assert.Nil(t, results)
		assert.ErrorIs(t, err, errInvalidPartialMigrationString)
		assert.Contains(t, err.Error(), "at token k:1d2, not a number")
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		str := "k:1,f:1,k:2.2,g:0.001,h:0"
		results, err := ConvertPartialMigrationStringToMap(str)
		assert.Nil(t, err)

		expectedResults := map[string]*FloatWrapper{
			"k": {Float: big.NewFloat(3.2)},
			"f": {Float: big.NewFloat(1)},
			"g": {Float: big.NewFloat(0.001)},
			"h": {Float: big.NewFloat(0)},
		}

		assert.Nil(t, err)
		assert.Equal(t, expectedResults, results)
	})
	t.Run("should work with maximum available", func(t *testing.T) {
		t.Parallel()

		str := "k:1,l:0,m:*,n:AlL,o:MaX"
		results, err := ConvertPartialMigrationStringToMap(str)
		assert.Nil(t, err)

		expectedResults := map[string]*FloatWrapper{
			"k": {Float: big.NewFloat(1)},
			"l": {Float: big.NewFloat(0), IsMax: false},
			"m": {Float: big.NewFloat(0), IsMax: true},
			"n": {Float: big.NewFloat(0), IsMax: true},
			"o": {Float: big.NewFloat(0), IsMax: true},
		}

		assert.Nil(t, err)
		assert.Equal(t, expectedResults, results)
	})
}

package ethereum

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadAllSignatures(t *testing.T) {
	t.Parallel()

	dirPath := "testdata"
	sigs := LoadAllSignatures(log, dirPath)

	expectedSigs := []SignatureInfo{
		{
			Address:     "0x3FE464Ac5aa562F7948322F92020F2b668D543d8",
			MessageHash: "0xc5b805c73d01e35e10a27a4cab86f096c976f0910ae23f5c6b307a823f0c49fb",
			Signature:   "74a91b07c796d1fcb18517994f4b71fe5f1c10317e95c609eabac9e7dbfc517c3e9c402585774a7129e4b5bbfade40647afc52bb38cb2a4b63163cbe2577eee201",
		},
		{
			Address:     "0xA6504Cc508889bbDBd4B748aFf6EA6b5D0d2684c",
			MessageHash: "0xc5b805c73d01e35e10a27a4cab86f096c976f0910ae23f5c6b307a823f0c49fb",
			Signature:   "111222333",
		},
	}

	assert.Equal(t, expectedSigs, sigs)
}

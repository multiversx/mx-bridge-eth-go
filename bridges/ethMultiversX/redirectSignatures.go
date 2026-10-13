package ethmultiversx

import (
	"encoding/hex"
	"sort"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/multiversx/mx-bridge-eth-go/core/batchProcessor"
)

type redirectSignatureLog struct {
	batchID        uint64
	redirectedHash common.Hash
	originalHash   common.Hash
	logged         map[string]struct{}
	lastSummary    string
}

func (executor *bridgeExecutor) trackRedirectSignatures() error {
	original := executor.unredirectedBatch
	if original == nil || original.ID != executor.batch.ID {
		executor.redirectSignatures = nil
		return nil
	}

	originalHash, err := executor.ethereumClient.GenerateMessageHash(batchProcessor.ExtractListMvxToEth(original), original.ID)
	if err != nil {
		return err
	}

	current := executor.redirectSignatures
	if current == nil || current.batchID != original.ID || current.redirectedHash != executor.msgHash {
		executor.redirectSignatures = &redirectSignatureLog{
			batchID:        original.ID,
			redirectedHash: executor.msgHash,
			originalHash:   originalHash,
			logged:         make(map[string]struct{}),
		}
	}

	executor.log.Info("signing the redirected batch; signatures on the original hash come from relayers without this release",
		"batch ID", original.ID, "redirected hash", executor.msgHash, "original hash", originalHash)
	return nil
}

func (executor *bridgeExecutor) logRedirectSignatures() {
	tracked := executor.redirectSignatures
	if tracked == nil {
		return
	}

	redirectedSigners := executor.logSignaturesFor(tracked, tracked.redirectedHash, false)
	originalSigners := executor.logSignaturesFor(tracked, tracked.originalHash, true)

	summary := strings.Join(redirectedSigners, " ") + "|" + strings.Join(originalSigners, " ")
	if summary == tracked.lastSummary {
		return
	}
	tracked.lastSummary = summary
	executor.log.Info("redirected batch signature stats",
		"batch ID", tracked.batchID,
		"redirected-hash signers", len(redirectedSigners),
		"original-hash signers", len(originalSigners),
		"redirected", strings.Join(redirectedSigners, " "),
		"original", strings.Join(originalSigners, " "))
}

func (executor *bridgeExecutor) logSignaturesFor(tracked *redirectSignatureLog, hash common.Hash, isOriginalHash bool) []string {
	signers := make([]string, 0)
	for _, signature := range executor.sigsHolder.Signatures(hash.Bytes()) {
		signer := recoverSigner(hash, signature)
		signers = append(signers, signer)

		key := hash.Hex() + signer
		if _, alreadyLogged := tracked.logged[key]; alreadyLogged {
			continue
		}
		tracked.logged[key] = struct{}{}

		if isOriginalHash {
			executor.log.Warn("received a signature on the ORIGINAL batch hash: this relayer does not run the redirect release",
				"batch ID", tracked.batchID, "signer", signer, "signature", hex.EncodeToString(signature))
			continue
		}
		executor.log.Info("received a signature on the redirected batch hash",
			"batch ID", tracked.batchID, "signer", signer, "signature", hex.EncodeToString(signature))
	}
	sort.Strings(signers)

	return signers
}

func recoverSigner(hash common.Hash, signature []byte) string {
	publicKey, err := crypto.SigToPub(hash.Bytes(), signature)
	if err != nil {
		return "unrecoverable"
	}

	return crypto.PubkeyToAddress(*publicKey).Hex()
}

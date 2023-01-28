package p2p

import (
	"time"

	chainCore "github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/process"
)

type peerDenialEvaluator struct {
	blackListIDsCache          process.PeerBlackListCacher
	blackListedPublicKeysCache process.TimeCacher
}

// NewPeerDenialEvaluator creates a new instance of peerDenialEvaluator
func NewPeerDenialEvaluator(blackListIDsCache process.PeerBlackListCacher, blackListedPublicKeysCache process.TimeCacher) (*peerDenialEvaluator, error) {
	if check.IfNil(blackListIDsCache) {
		return nil, ErrNilBlackListIDsCache
	}
	if check.IfNil(blackListedPublicKeysCache) {
		return nil, ErrNilBlackListedPublicKeysCache
	}

	return &peerDenialEvaluator{
		blackListIDsCache:          blackListIDsCache,
		blackListedPublicKeysCache: blackListedPublicKeysCache,
	}, nil
}

// IsDenied returns true if the provided peer id is denied to access the network
// It also checks if the public key is denied
func (p *peerDenialEvaluator) IsDenied(pid chainCore.PeerID) bool {
	if p.blackListIDsCache.Has(pid) {
		return true
	}

	return p.blackListedPublicKeysCache.Has(string(pid.Bytes()))
}

// UpsertPeerID will update or insert the provided peer id in the corresponding time cache
func (p *peerDenialEvaluator) UpsertPeerID(pid chainCore.PeerID, duration time.Duration) error {
	return p.blackListIDsCache.Upsert(pid, duration)
}

// IsInterfaceNil returns true if there is no value under the interface
func (p *peerDenialEvaluator) IsInterfaceNil() bool {
	return p == nil
}

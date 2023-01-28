package testsCommon

import "github.com/multiversx/mx-bridge-eth-go/core"

// BroadcasterStub -
type BroadcasterStub struct {
	BroadcastSignatureCalled func(signature []byte, messageHash []byte)
	BroadcastJoinTopicCalled func()
	SortedPublicKeysCalled   func() [][]byte
	RegisterOnTopicsCalled   func() error
	AddBroadcastClientCalled func(client core.BroadcastClient) error
	CloseCalled              func() error
}

// BroadcastSignature -
func (bs *BroadcasterStub) BroadcastSignature(signature []byte, messageHash []byte) {
	if bs.BroadcastSignatureCalled != nil {
		bs.BroadcastSignatureCalled(signature, messageHash)
	}
}

// BroadcastJoinTopic -
func (bs *BroadcasterStub) BroadcastJoinTopic() {
	if bs.BroadcastJoinTopicCalled != nil {
		bs.BroadcastJoinTopicCalled()
	}
}

// SortedPublicKeys -
func (bs *BroadcasterStub) SortedPublicKeys() [][]byte {
	if bs.SortedPublicKeysCalled != nil {
		return bs.SortedPublicKeysCalled()
	}

	return make([][]byte, 0)
}

// RegisterOnTopics -
func (bs *BroadcasterStub) RegisterOnTopics() error {
	if bs.RegisterOnTopicsCalled != nil {
		return bs.RegisterOnTopicsCalled()
	}

	return nil
}

// AddBroadcastClient -
func (bs *BroadcasterStub) AddBroadcastClient(client core.BroadcastClient) error {
	if bs.AddBroadcastClientCalled != nil {
		return bs.AddBroadcastClientCalled(client)
	}

	return nil
}

// Close -
func (bs *BroadcasterStub) Close() error {
	if bs.CloseCalled() != nil {
		return bs.CloseCalled()
	}

	return nil
}

// IsInterfaceNil -
func (bs *BroadcasterStub) IsInterfaceNil() bool {
	return bs == nil
}

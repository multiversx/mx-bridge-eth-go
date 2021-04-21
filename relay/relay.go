package relay

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridge"
	"github.com/ElrondNetwork/elrond-eth-bridge/bridge/elrond"
	"github.com/ElrondNetwork/elrond-eth-bridge/bridge/eth"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	factoryMarshalizer "github.com/ElrondNetwork/elrond-go/marshal/factory"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p"
)

const (
	ActionsTopicName = "actions/1"
	JoinedAction     = "joined"

	PrivateTopicName = "private/1"

	Timeout = 30 * time.Second
)

type Peers []core.PeerID

type Timer interface {
	after(d time.Duration) <-chan time.Time
	nowUnix() int64
}

type defaultTimer struct{}

func (s *defaultTimer) after(d time.Duration) <-chan time.Time {
	return time.After(d)
}

func (s *defaultTimer) nowUnix() int64 {
	return time.Now().Unix()
}

type NetMessenger interface {
	ID() core.PeerID
	Bootstrap() error
	Addresses() []string
	RegisterMessageProcessor(string, p2p.MessageProcessor) error
	HasTopic(name string) bool
	CreateTopic(name string, createChannelForTopic bool) error
	Broadcast(topic string, buff []byte)
	SendToConnectedPeer(topic string, buff []byte, peerID core.PeerID) error
	Close() error
}

type Relay struct {
	mu sync.Mutex

	peers     Peers
	messenger NetMessenger
	timer     Timer
	log       logger.Logger

	ethBridge    bridge.Bridge
	elrondBridge bridge.Bridge
}

func NewRelay(config *Config, name string) (*Relay, error) {
	ethBridge, err := eth.NewClient(config.Eth)
	if err != nil {
		return nil, err
	}

	elrondBridge, err := elrond.NewClient(config.Elrond)
	if err != nil {
		return nil, err
	}

	messenger, err := buildNetMessenger(config.P2P)
	if err != nil {
		return nil, err
	}

	return &Relay{
		peers:     make(Peers, 0),
		messenger: messenger,
		timer:     &defaultTimer{},
		log:       logger.GetOrCreate(name),

		ethBridge:    ethBridge,
		elrondBridge: elrondBridge,
	}, nil
}

func (r *Relay) Start(ctx context.Context) error {
	if err := r.init(ctx); err != nil {
		return nil
	}
	r.join(ctx)

	monitorEth := NewMonitor(r.ethBridge, r.elrondBridge, r.timer, r, "EthToElrond")
	go monitorEth.Start(ctx)
	monitorElrond := NewMonitor(r.elrondBridge, r.ethBridge, r.timer, r, "ElrondToEth")
	go monitorElrond.Start(ctx)

	<-ctx.Done()
	if err := r.Stop(); err != nil {
		return err
	}

	return nil
}

func (r *Relay) Stop() error {
	return r.messenger.Close()
}

// TopologyProvider

func (r *Relay) PeerCount() int {
	return len(r.peers)
}

func (r *Relay) AmITheLeader() bool {
	if len(r.peers) == 0 {
		return false
	} else {
		numberOfPeers := int64(len(r.peers))
		index := (r.timer.nowUnix() / int64(Timeout.Seconds())) % numberOfPeers

		return r.peers[index] == r.messenger.ID()
	}
}

// MessageProcessor

func (r *Relay) ProcessReceivedMessage(message p2p.MessageP2P, _ core.PeerID) error {
	r.log.Info(fmt.Sprintf("Got message on topic %q", message.Topic()))

	switch message.Topic() {
	case ActionsTopicName:
		r.log.Info(fmt.Sprintf("Action: %q\n", string(message.Data())))
		switch string(message.Data()) {
		case JoinedAction:
			r.addPeer(message.Peer())
			if err := r.broadcastTopology(message.Peer()); err != nil {
				r.log.Error(err.Error())
			}
		}
	case PrivateTopicName:
		if err := r.setTopology(message.Data()); err != nil {
			r.log.Error(err.Error())
		}
	}

	return nil
}

func (r *Relay) IsInterfaceNil() bool {
	return r == nil
}

func (r *Relay) broadcastTopology(toPeer core.PeerID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.peers) == 1 && r.peers[0] == r.messenger.ID() {
		return nil
	}

	var data bytes.Buffer
	enc := gob.NewEncoder(&data)
	if err := enc.Encode(r.peers); err != nil {
		return err
	}

	if err := r.messenger.SendToConnectedPeer(PrivateTopicName, data.Bytes(), toPeer); err != nil {
		return err
	}

	return nil
}

func (r *Relay) setTopology(data []byte) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// TODO: ignore if peers are already set
	if len(r.peers) > 1 {
		// ignore this call if we already have peers
		// TODO: find a better way here
		return nil
	}

	dec := gob.NewDecoder(bytes.NewReader(data))
	var topology Peers
	if err := dec.Decode(&topology); err != nil {
		return err
	}
	r.peers = topology

	return nil
}

func (r *Relay) addPeer(peerID core.PeerID) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// TODO: account for peers that rejoin
	if len(r.peers) == 0 || r.peers[len(r.peers)-1] < peerID {
		r.peers = append(r.peers, peerID)
		return
	}

	// TODO: can optimize via binary search
	for index, peer := range r.peers {
		if peer > peerID {
			r.peers = append(r.peers, "")
			copy(r.peers[index+1:], r.peers[index:])
			r.peers[index] = peerID
			break
		}
	}
}

// Helpers

func (r *Relay) init(ctx context.Context) error {
	if err := r.messenger.Bootstrap(); err != nil {
		return err
	}

	select {
	case <-r.timer.after(10 * time.Second):
		r.log.Info(fmt.Sprint(r.messenger.Addresses()))

		if err := r.registerTopicProcessors(); err != nil {
			return nil
		}
	case <-ctx.Done():
		return nil
	}

	return nil
}

func (r *Relay) join(ctx context.Context) {
	rand.Seed(time.Now().UnixNano())
	v := rand.Intn(5)

	select {
	case <-r.timer.after(time.Duration(v) * time.Second):
		r.messenger.Broadcast(ActionsTopicName, []byte(JoinedAction))
	case <-ctx.Done():
	}
}

func (r *Relay) registerTopicProcessors() error {
	topics := []string{ActionsTopicName, PrivateTopicName}
	for _, topic := range topics {
		if !r.messenger.HasTopic(topic) {
			if err := r.messenger.CreateTopic(topic, true); err != nil {
				return err
			}
		}

		r.log.Info(fmt.Sprintf("Registered on topic %q", topic))
		if err := r.messenger.RegisterMessageProcessor(topic, r); err != nil {
			return err
		}
	}

	return nil
}

func buildNetMessenger(cfg ConfigP2P) (NetMessenger, error) {
	internalMarshalizer, err := factoryMarshalizer.NewMarshalizer("gogo protobuf")
	if err != nil {
		panic(err)
	}

	nodeConfig := config.NodeConfig{
		Port:                       cfg.Port,
		Seed:                       cfg.Seed,
		MaximumExpectedPeerCount:   0,
		ThresholdMinConnectedPeers: 0,
	}
	peerDiscoveryConfig := config.KadDhtPeerDiscoveryConfig{
		Enabled:                          true,
		RefreshIntervalInSec:             5,
		ProtocolID:                       "/erd/relay/1.0.0",
		InitialPeerList:                  cfg.InitialPeerList,
		BucketSize:                       0,
		RoutingTableRefreshIntervalInSec: 300,
	}

	p2pConfig := config.P2PConfig{
		Node:                nodeConfig,
		KadDhtPeerDiscovery: peerDiscoveryConfig,
		Sharding: config.ShardingConfig{
			TargetPeerCount:         0,
			MaxIntraShardValidators: 0,
			MaxCrossShardValidators: 0,
			MaxIntraShardObservers:  0,
			MaxCrossShardObservers:  0,
			Type:                    "NilListSharder",
		},
	}

	args := libp2p.ArgsNetworkMessenger{
		Marshalizer:   internalMarshalizer,
		ListenAddress: libp2p.ListenAddrWithIp4AndTcp,
		P2pConfig:     p2pConfig,
		SyncTimer:     &libp2p.LocalSyncTimer{},
	}

	messenger, err := libp2p.NewNetworkMessenger(args)
	if err != nil {
		panic(err)
	}

	return messenger, nil
}

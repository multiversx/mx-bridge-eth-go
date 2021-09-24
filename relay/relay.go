package relay

import (
	"bytes"
	"context"
	"encoding/gob"

	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridge/eth"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/blockchain"

	"github.com/ElrondNetwork/elrond-go/ntp"

	"github.com/ElrondNetwork/elrond-eth-bridge/bridge"
	"github.com/ElrondNetwork/elrond-eth-bridge/bridge/elrond"
	"github.com/ElrondNetwork/elrond-go-core/core"
	factoryMarshalizer "github.com/ElrondNetwork/elrond-go-core/marshal/factory"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/epochStart/bootstrap/disabled"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/libp2p"
)

const (
	joinTopicName            = "join/1"
	privateTopicName         = "private/1"
	signTopicName            = "sign/1"
	timeout                  = 40 * time.Second
	defaultTopicIdentifier   = "default"
	p2pPeerNetworkDiscoverer = "optimized"
)

type Peers []core.PeerID

type Signatures map[core.PeerID][]byte

type Timer interface {
	After(d time.Duration) <-chan time.Time
	NowUnix() int64
	Start()
	Close() error
}

type defaultTimer struct {
	ntpSyncTimer ntp.SyncTimer
}

func NewDefaultTimer() *defaultTimer {
	return &defaultTimer{
		ntpSyncTimer: ntp.NewSyncTime(config.NTPConfig{SyncPeriodSeconds: 3600}, nil),
	}
}

func (s *defaultTimer) After(d time.Duration) <-chan time.Time {
	return time.After(d)
}

func (s *defaultTimer) NowUnix() int64 {
	return s.ntpSyncTimer.CurrentTime().Unix()
}

func (s *defaultTimer) Start() {
	s.ntpSyncTimer.StartSyncingTime()
}

func (s *defaultTimer) Close() error {
	return s.ntpSyncTimer.Close()
}

type NetMessenger interface {
	ID() core.PeerID
	Bootstrap() error
	Addresses() []string
	RegisterMessageProcessor(topic string, identifier string, processor p2p.MessageProcessor) error
	HasTopic(name string) bool
	CreateTopic(name string, createChannelForTopic bool) error
	Broadcast(topic string, buff []byte)
	SendToConnectedPeer(topic string, buff []byte, peerID core.PeerID) error
	Close() error
}

type Relay struct {
	mu sync.Mutex

	peers      Peers
	messenger  NetMessenger
	timer      Timer
	log        logger.Logger
	signatures Signatures

	ethBridge    bridge.Bridge
	elrondBridge bridge.Bridge

	roleProvider                bridge.RoleProvider
	elrondWalletAddressProvider bridge.WalletAddressProvider
	quorumProvider              bridge.QuorumProvider
}

func NewRelay(config *Config, name string) (*Relay, error) {
	relay := &Relay{}

	proxy := blockchain.NewElrondProxy(config.Elrond.NetworkAddress, nil)
	clientArgs := elrond.ClientArgs{
		Config: config.Elrond,
		Proxy:  proxy,
	}
	elrondBridge, err := elrond.NewClient(clientArgs)
	if err != nil {
		return nil, err
	}
	relay.elrondBridge = elrondBridge
	relay.roleProvider = elrondBridge
	relay.elrondWalletAddressProvider = elrondBridge

	ethBridge, err := eth.NewClient(config.Eth, relay, elrondBridge)
	if err != nil {
		return nil, err
	}
	relay.ethBridge = ethBridge
	relay.quorumProvider = ethBridge

	messenger, err := buildNetMessenger(config.P2P)
	if err != nil {
		return nil, err
	}
	relay.messenger = messenger

	relay.peers = make(Peers, 0)
	relay.timer = NewDefaultTimer()
	relay.log = logger.GetOrCreate(name)
	relay.signatures = make(map[core.PeerID][]byte)
	return relay, nil
}

func (r *Relay) Start(ctx context.Context) error {
	if err := r.init(ctx); err != nil {
		return nil
	}
	r.join(ctx)

	r.timer.Start()

	monitorEth := NewMonitor(r.ethBridge, r.elrondBridge, r.timer, r, r.quorumProvider, "EthToElrond")
	go monitorEth.Start(ctx)
	monitorElrond := NewMonitor(r.elrondBridge, r.ethBridge, r.timer, r, r.quorumProvider, "ElrondToEth")
	go monitorElrond.Start(ctx)

	<-ctx.Done()
	if err := r.Stop(); err != nil {
		return err
	}

	return nil
}

func (r *Relay) Stop() error {
	if err := r.timer.Close(); err != nil {
		r.log.Error(err.Error())
	}
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
		index := (r.timer.NowUnix() / int64(timeout.Seconds())) % numberOfPeers

		return r.peers[index] == r.messenger.ID()
	}
}

func (r *Relay) Clean() {
	r.signatures = make(Signatures)
}

// MessageProcessor

func (r *Relay) ProcessReceivedMessage(message p2p.MessageP2P, _ core.PeerID) error {
	r.log.Info(fmt.Sprintf("Got message on topic %q", message.Topic()))

	switch message.Topic() {
	case joinTopicName:
		elrondPublicAddress := string(message.Data())
		if !r.roleProvider.IsWhitelisted(elrondPublicAddress) {
			r.log.Error(fmt.Sprintf("A peer with address %q tryed to join but is not whitelisted", elrondPublicAddress))
			return nil
		}

		r.addPeer(message.Peer())
		if err := r.broadcastTopology(message.Peer()); err != nil {
			r.log.Error(err.Error())
		}
	case signTopicName:
		r.addSignatureForPeer(message.Peer(), message.Data())
	case privateTopicName:
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

	if err := r.messenger.SendToConnectedPeer(privateTopicName, data.Bytes(), toPeer); err != nil {
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

// Broadcaster

func (r *Relay) Signatures() [][]byte {
	result := make([][]byte, 0)

	for _, signature := range r.signatures {
		result = append(result, signature)
	}
	return result
}

func (r *Relay) SendSignature(signature []byte) {
	r.messenger.Broadcast(signTopicName, signature)
}

// Helpers

func (r *Relay) init(ctx context.Context) error {
	if err := r.messenger.Bootstrap(); err != nil {
		return err
	}

	select {
	case <-r.timer.After(10 * time.Second):
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
	case <-r.timer.After(time.Duration(v) * time.Second):
		r.log.Debug(fmt.Sprintf("Joining with address %s", r.elrondWalletAddressProvider.GetHexWalletAddress()))
		r.messenger.Broadcast(joinTopicName, []byte(r.elrondWalletAddressProvider.GetHexWalletAddress()))
	case <-ctx.Done():
	}
}

func (r *Relay) addSignatureForPeer(peerID core.PeerID, signature []byte) {
	r.signatures[peerID] = signature
}

func (r *Relay) registerTopicProcessors() error {
	topics := []string{joinTopicName, privateTopicName, signTopicName}
	for _, topic := range topics {
		if !r.messenger.HasTopic(topic) {
			if err := r.messenger.CreateTopic(topic, true); err != nil {
				return err
			}
		}

		r.log.Info(fmt.Sprintf("Registered on topic %q", topic))
		if err := r.messenger.RegisterMessageProcessor(topic, defaultTopicIdentifier, r); err != nil {
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
		ProtocolID:                       cfg.ProtocolID,
		InitialPeerList:                  cfg.InitialPeerList,
		BucketSize:                       0,
		RoutingTableRefreshIntervalInSec: 300,
		Type:                             p2pPeerNetworkDiscoverer,
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
		Marshalizer:          internalMarshalizer,
		ListenAddress:        libp2p.ListenAddrWithIp4AndTcp,
		P2pConfig:            p2pConfig,
		SyncTimer:            &libp2p.LocalSyncTimer{},
		PreferredPeersHolder: disabled.NewPreferredPeersHolder(),
		NodeOperationMode:    p2p.NormalOperation,
	}

	messenger, err := libp2p.NewNetworkMessenger(args)
	if err != nil {
		panic(err)
	}

	return messenger, nil
}

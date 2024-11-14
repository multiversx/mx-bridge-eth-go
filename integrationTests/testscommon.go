package integrationTests

import (
	"fmt"

	"github.com/multiversx/mx-bridge-eth-go/core"
	"github.com/multiversx/mx-bridge-eth-go/testsCommon/crypto"
	"github.com/multiversx/mx-chain-communication-go/p2p/libp2p"
	"github.com/multiversx/mx-chain-core-go/hashing/blake2b"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-crypto-go/signing"
	"github.com/multiversx/mx-chain-crypto-go/signing/ed25519"
	"github.com/multiversx/mx-chain-crypto-go/signing/ed25519/singlesig"
	"github.com/multiversx/mx-chain-go/p2p"
	p2pConfig "github.com/multiversx/mx-chain-go/p2p/config"
	"github.com/multiversx/mx-chain-go/testscommon/p2pmocks"
	logger "github.com/multiversx/mx-chain-logger-go"
)

// Log -
var Log = logger.GetOrCreate("integrationtests/broadcaster")
var suite = ed25519.NewEd25519()

// TestKeyGenerator -
var TestKeyGenerator = signing.NewKeyGenerator(suite)

// TestSingleSigner -
var TestSingleSigner = &singlesig.Ed25519Signer{}

// TestMarshalizer -
var TestMarshalizer = &marshal.JsonMarshalizer{}

// TestHasher -
var TestHasher = blake2b.NewBlake2b()

// Connectable defines the operations for a struct to become connectable by other struct
// In other words, all instances that implement this interface are able to connect with each other
type Connectable interface {
	ConnectTo(connectable Connectable) error
	GetConnectableAddress() string
	IsInterfaceNil() bool
}

// Broadcaster defines a component able to communicate with other such instances and manage signatures and other state related data
type Broadcaster interface {
	BroadcastSignature(signature []byte, messageHash []byte)
	BroadcastJoinTopic()
	SortedPublicKeys() [][]byte
	AddBroadcastClient(client core.BroadcastClient) error
	Close() error
	IsInterfaceNil() bool
}

// ConnectNodes will try to connect all provided connectable instances in a full mesh fashion
func ConnectNodes(nodes []Connectable) {
	encounteredErrors := make([]error, 0)

	for i := 0; i < len(nodes)-1; i++ {
		for j := i + 1; j < len(nodes); j++ {
			src := nodes[i]
			dst := nodes[j]
			err := src.ConnectTo(dst)
			if err != nil {
				encounteredErrors = append(encounteredErrors,
					fmt.Errorf("%w while %s was connecting to %s", err, src.GetConnectableAddress(), dst.GetConnectableAddress()))
			}
		}
	}

	printEncounteredErrors(encounteredErrors)
}

func printEncounteredErrors(encounteredErrors []error) {
	if len(encounteredErrors) == 0 {
		return
	}

	printArguments := make([]interface{}, 0, len(encounteredErrors)*2)
	for i, err := range encounteredErrors {
		if err == nil {
			continue
		}

		printArguments = append(printArguments, fmt.Sprintf("err%d", i))
		printArguments = append(printArguments, err.Error())
	}

	Log.Warn("errors encountered while connecting hosts", printArguments...)
}

// CreateMessengerWithNoDiscovery creates a new libp2p messenger with no peer discovery
func CreateMessengerWithNoDiscovery() p2p.Messenger {
	p2pCfg := p2pConfig.P2PConfig{
		Node: p2pConfig.NodeConfig{
			Port: "0",
			Transports: p2pConfig.P2PTransportConfig{
				TCP: p2pConfig.P2PTCPTransport{
					ListenAddress: "/ip4/127.0.0.1/tcp/%d",
				},
			},
			ResourceLimiter: p2pConfig.P2PResourceLimiterConfig{
				Type:                   "default autoscale",
				ManualSystemMemoryInMB: 0,
				ManualMaximumFD:        0,
			},
		},
		KadDhtPeerDiscovery: p2pConfig.KadDhtPeerDiscoveryConfig{
			Enabled:    false,
			ProtocolID: "/erd/relay/1.0.0",
		},
		Sharding: p2pConfig.ShardingConfig{
			Type: p2p.NilListSharder,
		},
	}

	return CreateMessengerFromConfig(p2pCfg)
}

// CreateMessengerFromConfig creates a new libp2p messenger with provided configuration
func CreateMessengerFromConfig(p2pConfig p2pConfig.P2PConfig) p2p.Messenger {
	arg := libp2p.ArgsNetworkMessenger{
		Marshaller:            &marshal.JsonMarshalizer{},
		P2pConfig:             p2pConfig,
		SyncTimer:             &libp2p.LocalSyncTimer{},
		PreferredPeersHolder:  &p2pmocks.PeersHolderStub{},
		ConnectionWatcherType: "disabled",
		PeersRatingHandler:    &p2pmocks.PeersRatingHandlerStub{},
		P2pPrivateKey:         crypto.NewPrivateKeyMock(),
		P2pSingleSigner:       &crypto.SingleSignerStub{},
		P2pKeyGenerator:       &crypto.KeyGenStub{},
		Logger:                Log,
	}

	libP2PMes, err := libp2p.NewNetworkMessenger(arg)
	Log.LogIfError(err)

	return libP2PMes
}

// CreateLinkedMessengers will create the specified number of messengers and will connect them all between them
func CreateLinkedMessengers(numMessengers int) []p2p.Messenger {
	connectables := make([]Connectable, 0, numMessengers)
	messengers := make([]p2p.Messenger, 0, numMessengers)
	for i := 0; i < numMessengers; i++ {
		mes := CreateMessengerWithNoDiscovery()
		messengers = append(messengers, mes)

		connectable := &messengerWrapper{
			Messenger: mes,
		}
		connectables = append(connectables, connectable)
	}

	ConnectNodes(connectables)

	return messengers
}

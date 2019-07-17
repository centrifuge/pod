// +build unit

package messenger

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	pb "github.com/centrifuge/centrifuge-protobufs/gen/go/protocol"
	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testlogging"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/identity/ideth"
	"github.com/centrifuge/go-centrifuge/jobs/jobsv1"
	"github.com/centrifuge/go-centrifuge/p2p/common"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/storage/leveldb"
	"github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ipfs/go-cid"
	ds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-ipfs-addr"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-crypto"
	"github.com/libp2p/go-libp2p-host"
	"github.com/libp2p/go-libp2p-kad-dht"
	libp2pPeer "github.com/libp2p/go-libp2p-peer"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	"github.com/libp2p/go-libp2p-peerstore/pstoremem"
	"github.com/libp2p/go-libp2p-protocol"
	ma "github.com/multiformats/go-multiaddr"
	mh "github.com/multiformats/go-multihash"
	"github.com/stretchr/testify/assert"
)

const MessengerDummyProtocol protocol.ID = "/messeger/dummy/0.0.1"
const MessengerDummyProtocol2 protocol.ID = "/messeger/dummy/0.0.2"

var mockedHandler = func(ctx context.Context, peer libp2pPeer.ID, protoc protocol.ID, msg *pb.P2PEnvelope) (envelope *pb.P2PEnvelope, e error) {
	dataEnv, err := p2pcommon.ResolveDataEnvelope(msg)
	if err != nil {
		return nil, err
	}
	if p2pcommon.MessageTypeSendAnchoredDoc.Equals(dataEnv.Header.Type) {
		dataEnv.Header.Type = p2pcommon.MessageTypeSendAnchoredDocRep.String()
	} else if p2pcommon.MessageTypeRequestSignature.Equals(dataEnv.Header.Type) {
		dataEnv.Header.Type = p2pcommon.MessageTypeRequestSignatureRep.String()
	} else if p2pcommon.MessageTypeSendAnchoredDocRep.Equals(dataEnv.Header.Type) {
		dataEnv.Header.Type = p2pcommon.MessageTypeSendAnchoredDoc.String()
	} else if p2pcommon.MessageTypeError.Equals(dataEnv.Header.Type) {
		return nil, errors.New("dummy error")
	} else if p2pcommon.MessageTypeRequestSignatureRep.Equals(dataEnv.Header.Type) {
		return nil, nil
	} else if p2pcommon.MessageTypeInvalid.Equals(dataEnv.Header.Type) {
		return nil, errors.New("invalid data")
	}

	return p2pcommon.PrepareP2PEnvelope(ctx, uint32(0), p2pcommon.MessageTypeFromString(dataEnv.Header.Type), dataEnv)
}

var cfg config.Service

func TestMain(m *testing.M) {
	ctx := make(map[string]interface{})
	ethClient := &ethereum.MockEthClient{}
	ethClient.On("GetEthClient").Return(nil)
	ctx[ethereum.BootstrappedEthereumClient] = ethClient
	ibootstappers := []bootstrap.TestBootstrapper{
		&testlogging.TestLoggingBootstrapper{},
		&config.Bootstrapper{},
		&leveldb.Bootstrapper{},
		jobsv1.Bootstrapper{},
		&queue.Bootstrapper{},
		&ideth.Bootstrapper{},
		&configstore.Bootstrapper{},
		&queue.Bootstrapper{},
		&anchors.Bootstrapper{},
		documents.Bootstrapper{},
	}
	bootstrap.RunTestBootstrappers(ibootstappers, ctx)
	cfg = ctx[config.BootstrappedConfigStorage].(config.Service)
	result := m.Run()
	bootstrap.RunTestTeardown(ibootstappers)
	os.Exit(result)
}

// Using a single test for all cases to use the same hosts in a synchronous way
func TestHandleNewMessage(t *testing.T) {
	cfg, err := cfg.GetConfig()
	assert.NoError(t, err)
	cfg = updateKeys(cfg)
	ctx, canc := context.WithCancel(context.Background())
	c := testingconfig.CreateTenantContextWithContext(t, ctx, cfg)
	r := rand.Reader
	p1 := 35000
	p2 := 35001
	p3 := 35002
	h1 := createRandomHost(t, p1, r)
	h2 := createRandomHost(t, p2, r)
	h3 := createRandomHost(t, p3, r)
	// set h2 as the bootnode for h1
	_ = runDHT(c, h1, []string{fmt.Sprintf("/ip4/127.0.0.1/tcp/%d/ipfs/%s", p2, h2.ID().Pretty())})

	m1 := NewP2PMessenger(c, h1, 5*time.Second, mockedHandler)
	m2 := NewP2PMessenger(c, h2, 5*time.Second, mockedHandler)

	m1.Init(MessengerDummyProtocol)
	m2.Init(MessengerDummyProtocol)
	m2.Init(MessengerDummyProtocol2)

	// 1. happy path
	// from h1 to h2 (with a message size ~ MessageSizeMax, has to be less because of the length bytes)
	p2pEnv, err := p2pcommon.PrepareP2PEnvelope(c, uint32(0), p2pcommon.MessageTypeRequestSignature, &p2ppb.Envelope{Body: utils.RandomSlice(MessageSizeMax - 400)})
	assert.NoError(t, err)
	msg, err := m1.SendMessage(c, h2.ID(), p2pEnv, MessengerDummyProtocol)
	assert.NoError(t, err)
	dataEnv, err := p2pcommon.ResolveDataEnvelope(msg)
	assert.NoError(t, err)
	assert.True(t, p2pcommon.MessageTypeRequestSignatureRep.Equals(dataEnv.Header.Type))

	// from h1 to h2 different protocol - intentionally setting reply-response in opposite for differentiation
	p2pEnv, err = p2pcommon.PrepareP2PEnvelope(c, uint32(0), p2pcommon.MessageTypeSendAnchoredDocRep, &p2ppb.Envelope{Body: utils.RandomSlice(3)})
	assert.NoError(t, err)
	msg, err = m1.SendMessage(c, h2.ID(), p2pEnv, MessengerDummyProtocol2)
	assert.NoError(t, err)
	dataEnv, err = p2pcommon.ResolveDataEnvelope(msg)
	assert.NoError(t, err)
	assert.True(t, p2pcommon.MessageTypeSendAnchoredDoc.Equals(dataEnv.Header.Type))

	// from h2 to h1
	p2pEnv, err = p2pcommon.PrepareP2PEnvelope(c, uint32(0), p2pcommon.MessageTypeSendAnchoredDoc, &p2ppb.Envelope{Body: utils.RandomSlice(3)})
	assert.NoError(t, err)
	msg, err = m2.SendMessage(c, h1.ID(), p2pEnv, MessengerDummyProtocol)
	assert.NoError(t, err)
	dataEnv, err = p2pcommon.ResolveDataEnvelope(msg)
	assert.NoError(t, err)
	assert.True(t, p2pcommon.MessageTypeSendAnchoredDocRep.Equals(dataEnv.Header.Type))

	// 2. unrecognized  protocol
	p2pEnv, err = p2pcommon.PrepareP2PEnvelope(c, uint32(0), p2pcommon.MessageTypeRequestSignature, &p2ppb.Envelope{Body: utils.RandomSlice(3)})
	assert.NoError(t, err)
	msg, err = m1.SendMessage(c, h2.ID(), p2pEnv, "wrong")
	if assert.Error(t, err) {
		assert.Equal(t, "protocol not supported", err.Error())
	}

	// 3. unrecognized message type - stream would be reset by the peer
	p2pEnv, err = p2pcommon.PrepareP2PEnvelope(c, uint32(0), p2pcommon.MessageTypeInvalid, &p2ppb.Envelope{Body: utils.RandomSlice(3)})
	assert.NoError(t, err)
	msg, err = m1.SendMessage(c, h2.ID(), p2pEnv, MessengerDummyProtocol)
	if assert.Error(t, err) {
		assert.Equal(t, "stream reset", err.Error())
	}

	// 4. handler error
	p2pEnv, err = p2pcommon.PrepareP2PEnvelope(c, uint32(0), p2pcommon.MessageTypeError, &p2ppb.Envelope{Body: utils.RandomSlice(3)})
	assert.NoError(t, err)
	msg, err = m1.SendMessage(c, h2.ID(), p2pEnv, MessengerDummyProtocol)
	if assert.Error(t, err) {
		assert.Equal(t, "stream reset", err.Error())
	}

	// 5. can't find host - h3
	p2pEnv, err = p2pcommon.PrepareP2PEnvelope(c, uint32(0), p2pcommon.MessageTypeRequestSignature, &p2ppb.Envelope{Body: utils.RandomSlice(3)})
	assert.NoError(t, err)
	msg, err = m1.SendMessage(c, h3.ID(), p2pEnv, MessengerDummyProtocol)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "dial attempt failed: no good addresses")
	}

	// 6. handler nil response
	p2pEnv, err = p2pcommon.PrepareP2PEnvelope(c, uint32(0), p2pcommon.MessageTypeRequestSignatureRep, &p2ppb.Envelope{Body: utils.RandomSlice(3)})
	assert.NoError(t, err)
	msg, err = m1.SendMessage(c, h2.ID(), p2pEnv, MessengerDummyProtocol)
	if assert.Error(t, err) {
		assert.Equal(t, "timed out reading response", err.Error())
	}

	// 7. message size more than the max
	// from h1 to h2 (with a message size > MessageSizeMax)
	p2pEnv, err = p2pcommon.PrepareP2PEnvelope(c, uint32(0), p2pcommon.MessageTypeRequestSignature, &p2ppb.Envelope{Body: utils.RandomSlice(MessageSizeMax)})
	assert.NoError(t, err)
	msg, err = m1.SendMessage(c, h2.ID(), p2pEnv, MessengerDummyProtocol)
	if assert.Error(t, err) {
		assert.Equal(t, "stream reset", err.Error())
	}
	canc()
}

func createRandomHost(t *testing.T, port int, r io.Reader) host.Host {
	priv, pub, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	assert.NoError(t, err)
	h1, err := makeBasicHost(priv, pub, "", port)
	assert.NoError(t, err)
	return h1
}

// makeBasicHost creates a LibP2P host with a peer ID listening on the given port
func makeBasicHost(priv crypto.PrivKey, pub crypto.PubKey, externalIP string, listenPort int) (host.Host, error) {
	// Obtain Peer ID from public key
	// We should be using the following method to get the ID, but looks like is not compatible with
	// secio when adding the pub and pvt keys, fail as id+pub/pvt key is checked to match and method defaults to
	// IDFromPublicKey(pk)
	//pid, err := peer.IDFromEd25519PublicKey(pub)
	pid, err := libp2pPeer.IDFromPublicKey(pub)
	if err != nil {
		return nil, err
	}

	// Create a peerstore
	ps := pstore.NewPeerstore(
		pstoremem.NewKeyBook(),
		pstoremem.NewAddrBook(),
		pstoremem.NewPeerMetadata())

	// Add the keys to the peerstore
	// for this peer ID.
	err = ps.AddPubKey(pid, pub)
	if err != nil {
		log.Infof("Could not enable encryption: %v\n", err)
		return nil, err
	}

	err = ps.AddPrivKey(pid, priv)
	if err != nil {
		log.Infof("Could not enable encryption: %v\n", err)
		return nil, err
	}

	var extMultiAddr ma.Multiaddr
	if externalIP == "" {
		log.Warning("External IP not defined, Peers might not be able to resolve this node if behind NAT\n")
	} else {
		extMultiAddr, err = ma.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d", externalIP, listenPort))
		if err != nil {
			return nil, errors.New("failed to create multiaddr: %v", err)
		}
	}

	addressFactory := func(addrs []ma.Multiaddr) []ma.Multiaddr {
		if extMultiAddr != nil {
			// We currently support a single protocol and transport, if we add more to support then we will need to adapt this code
			addrs = []ma.Multiaddr{extMultiAddr}
		}
		return addrs
	}

	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", listenPort)),
		libp2p.Identity(priv),
		libp2p.DefaultMuxers,
		libp2p.AddrsFactory(addressFactory),
	}

	bhost, err := libp2p.New(context.Background(), opts...)
	if err != nil {
		return nil, err
	}

	hostAddr, err := ma.NewMultiaddr(fmt.Sprintf("/ipfs/%s", bhost.ID().Pretty()))
	if err != nil {
		return nil, errors.New("failed to get addr: %v", err)
	}

	log.Infof("P2P Server at: %s %s\n", hostAddr.String(), bhost.Addrs())
	return bhost, nil
}

func runDHT(ctx context.Context, h host.Host, bootstrapPeers []string) error {
	// Run it as a Bootstrap Node
	dhtClient := dht.NewDHT(ctx, h, ds.NewMapDatastore())
	log.Infof("Bootstrapping %s\n", bootstrapPeers)

	for _, addr := range bootstrapPeers {
		iaddr, _ := ipfsaddr.ParseString(addr)
		pinfo, _ := pstore.InfoFromP2pAddr(iaddr.Multiaddr())
		if err := h.Connect(ctx, *pinfo); err != nil {
			log.Info("Bootstrapping to peer failed: ", err)
		}
	}

	// Using the sha256 of our "topic" as our rendezvous value
	cidPref, _ := cid.NewPrefixV1(cid.Raw, mh.SHA2_256).Sum([]byte("centrifuge-dht"))

	// First, announce ourselves as participating in this topic
	log.Info("Announcing ourselves...")
	tctx, cancel := context.WithTimeout(ctx, time.Second*10)
	if err := dhtClient.Provide(tctx, cidPref, true); err != nil {
		// Important to keep this as Non-Fatal error, otherwise it will fail for a node that behaves as well as bootstrap one
		log.Infof("Error: %s\n", err.Error())
	}
	cancel()

	// Now, look for others who have announced
	log.Info("Searching for other peers ...")
	tctx, cancel = context.WithTimeout(ctx, time.Second*10)
	peers, err := dhtClient.FindProviders(tctx, cidPref)
	if err != nil {
		log.Error(err)
	}
	cancel()
	log.Infof("Found %d peers!\n", len(peers))

	// Now connect to them, so they are added to the PeerStore
	for _, pe := range peers {
		log.Infof("Peer %s %s\n", pe.ID.Pretty(), pe.Addrs)

		if pe.ID == h.ID() {
			// No sense connecting to ourselves
			continue
		}

		tctx, cancel := context.WithTimeout(ctx, time.Second*5)
		if err := h.Connect(tctx, pe); err != nil {
			log.Info("Failed to connect to peer: ", err)
		}
		cancel()
	}

	log.Info("Bootstrapping and discovery complete!")
	return nil
}

func updateKeys(c config.Configuration) config.Configuration {
	n := c.(*configstore.NodeConfig)
	n.MainIdentity.SigningKeyPair.Pub = "../../build/resources/signingKey.pub.pem"
	n.MainIdentity.SigningKeyPair.Pvt = "../../build/resources/signingKey.key.pem"
	return c
}

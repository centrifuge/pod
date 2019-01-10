package p2p

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/centrifuge/go-centrifuge/identity"

	"github.com/centrifuge/go-centrifuge/config/configstore"

	"github.com/centrifuge/go-centrifuge/p2p/common"
	"github.com/centrifuge/go-centrifuge/p2p/receiver"

	"github.com/libp2p/go-libp2p-protocol"

	"github.com/centrifuge/go-centrifuge/errors"

	cented25519 "github.com/centrifuge/go-centrifuge/crypto/ed25519"
	"github.com/ipfs/go-cid"
	ds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-ipfs-addr"
	logging "github.com/ipfs/go-log"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-crypto"
	"github.com/libp2p/go-libp2p-host"
	"github.com/libp2p/go-libp2p-kad-dht"
	libp2pPeer "github.com/libp2p/go-libp2p-peer"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	ma "github.com/multiformats/go-multiaddr"
	mh "github.com/multiformats/go-multihash"
)

var log = logging.Logger("p2p-server")

// peer implements node.Server
type peer struct {
	config         configstore.Service
	host           host.Host
	handlerCreator func() *receiver.Handler
	mes            messenger
}

// Name returns the P2PServer
func (*peer) Name() string {
	return "P2PServer"
}

// Start starts the DHT and libp2p host
func (s *peer) Start(ctx context.Context, wg *sync.WaitGroup, startupErr chan<- error) {
	defer wg.Done()

	nc, err := s.config.GetConfig()
	if err != nil {
		startupErr <- err
		return
	}

	if nc.GetP2PPort() == 0 {
		startupErr <- errors.New("please provide a port to bind on")
		return
	}

	// Make a host that listens on the given multiaddress
	// first obtain the keys configured
	priv, pub, err := s.createSigningKey(nc.GetSigningKeyPair())
	if err != nil {
		startupErr <- err
		return
	}
	s.host, err = makeBasicHost(priv, pub, nc.GetP2PExternalIP(), nc.GetP2PPort())
	if err != nil {
		startupErr <- err
		return
	}

	s.mes = newP2PMessenger(ctx, s.host, nc.P2PConnectionTimeout, s.handlerCreator().HandleInterceptor)
	tcs, err := s.config.GetAllTenants()
	if err != nil {
		startupErr <- err
		return
	}
	var protocols []protocol.ID
	for _, t := range tcs {
		CID, err := identity.ToCentID(t.IdentityID)
		if err != nil {
			startupErr <- err
			return
		}
		protocols = append(protocols, p2pcommon.ProtocolForCID(CID))
	}
	s.mes.init(protocols...)

	// Start DHT and properly ignore errors :)
	_ = runDHT(ctx, s.host, nc.GetBootstrapPeers())
	<-ctx.Done()

}

func (s *peer) createSigningKey(pubKey, privKey string) (priv crypto.PrivKey, pub crypto.PubKey, err error) {
	// Create the signing key for the host
	publicKey, privateKey, err := cented25519.GetSigningKeyPair(pubKey, privKey)
	if err != nil {
		return nil, nil, errors.New("failed to get keys: %v", err)
	}

	var key []byte
	key = append(key, privateKey...)
	key = append(key, publicKey...)

	priv, err = crypto.UnmarshalEd25519PrivateKey(key)
	if err != nil {
		return nil, nil, err
	}

	pub = priv.GetPublic()
	return priv, pub, nil
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
	ps := pstore.NewPeerstore()

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

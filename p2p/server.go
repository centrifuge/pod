package p2p

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	"github.com/centrifuge/go-centrifuge/documents"
	cented25519 "github.com/centrifuge/go-centrifuge/keytools/ed25519"
	"github.com/ipfs/go-cid"
	ds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-ipfs-addr"
	logging "github.com/ipfs/go-log"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-crypto"
	"github.com/libp2p/go-libp2p-host"
	"github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p-peer"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	ma "github.com/multiformats/go-multiaddr"
	mh "github.com/multiformats/go-multihash"
	"github.com/paralin/go-libp2p-grpc"
)

var log = logging.Logger("p2p-server")

// Config defines methods that are required for the package p2p.
type Config interface {
	GetP2PExternalIP() string
	GetP2PPort() int
	GetBootstrapPeers() []string
	GetP2PConnectionTimeout() time.Duration
	GetNetworkID() uint32
	GetIdentityID() ([]byte, error)
	GetSigningKeyPair() (pub, priv string)
}

// p2pServer implements api.Server
type p2pServer struct {
	config   Config
	host     host.Host
	registry *documents.ServiceRegistry
	protocol *p2pgrpc.GRPCProtocol
	handler  p2ppb.P2PServiceServer
}

// Name returns the P2PServer
func (*p2pServer) Name() string {
	return "P2PServer"
}

// Start starts the DHT and GRPC server for p2p communications
func (s *p2pServer) Start(ctx context.Context, wg *sync.WaitGroup, startupErr chan<- error) {
	defer wg.Done()

	if s.config.GetP2PPort() == 0 {
		startupErr <- errors.New("please provide a port to bind on")
		return
	}

	// Make a host that listens on the given multiaddress
	var err error
	s.host, err = s.makeBasicHost(s.config.GetP2PPort())
	if err != nil {
		startupErr <- err
		return
	}

	// Set the grpc protocol handler on it
	s.protocol = p2pgrpc.NewGRPCProtocol(ctx, s.host)
	p2ppb.RegisterP2PServiceServer(s.protocol.GetGRPCServer(), s.handler)

	serveErr := make(chan error)
	go func() {
		err := s.protocol.Serve()
		serveErr <- err
	}()

	s.host.Peerstore().AddAddr(s.host.ID(), s.host.Addrs()[0], pstore.TempAddrTTL)

	// Start DHT
	s.runDHT(ctx, s.host)

	for {
		select {
		case err := <-serveErr:
			log.Infof("GRPC server error: %v", err)
			s.protocol.GetGRPCServer().GracefulStop()
			log.Info("GRPC server stopped")
			return
		case <-ctx.Done():
			log.Info("Shutting down GRPC server")
			s.protocol.GetGRPCServer().GracefulStop()
			log.Info("GRPC server stopped")
			return
		}
	}
}

func (s *p2pServer) runDHT(ctx context.Context, h host.Host) error {
	// Run it as a Bootstrap Node
	dhtClient := dht.NewDHT(ctx, h, ds.NewMapDatastore())

	bootstrapPeers := s.config.GetBootstrapPeers()
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

// makeBasicHost creates a LibP2P host with a peer ID listening on the given port
func (s *p2pServer) makeBasicHost(listenPort int) (host.Host, error) {
	priv, pub, err := s.createSigningKey()
	if err != nil {
		return nil, err
	}

	// Obtain Peer ID from public key
	// We should be using the following method to get the ID, but looks like is not compatible with
	// secio when adding the pub and pvt keys, fail as id+pub/pvt key is checked to match and method defaults to
	// IDFromPublicKey(pk)
	//pid, err := peer.IDFromEd25519PublicKey(pub)
	pid, err := peer.IDFromPublicKey(pub)
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

	externalIP := s.config.GetP2PExternalIP()
	var extMultiAddr ma.Multiaddr
	if externalIP == "" {
		log.Warning("External IP not defined, Peers might not be able to resolve this node if behind NAT\n")
	} else {
		extMultiAddr, err = ma.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d", externalIP, listenPort))
		if err != nil {
			return nil, fmt.Errorf("failed to create multiaddr: %v", err)
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
		return nil, fmt.Errorf("failed to get addr: %v", err)
	}

	log.Infof("P2P Server at: %s %s\n", hostAddr.String(), bhost.Addrs())
	return bhost, nil
}

func (s *p2pServer) createSigningKey() (priv crypto.PrivKey, pub crypto.PubKey, err error) {
	// Create the signing key for the host
	publicKey, privateKey, err := cented25519.GetSigningKeyPair(s.config.GetSigningKeyPair())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get keys: %v", err)
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

// PLEASE DO NOT call any config.* stuff here as it creates dependencies that can't be injected easily when testing
package p2p

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	"github.com/centrifuge/go-centrifuge/config"
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
	cented25519 "github.com/centrifuge/go-centrifuge/keytools/ed25519"
)

var log = logging.Logger("cent-p2p-server")
var HostInstance host.Host
var GRPCProtoInstance p2pgrpc.GRPCProtocol

type Config interface {
	GetP2PExternalIP()  string
	GetP2PPort()        int
	GetBootstrapPeers() []string
}

type CentP2PServer struct {
	config Config
}

func NewCentP2PServer(config Config) *CentP2PServer {
	return &CentP2PServer{config}
}

func (*CentP2PServer) Name() string {
	return "CentP2PServer"
}

func (c *CentP2PServer) Start(ctx context.Context, wg *sync.WaitGroup, startupErr chan<- error) {
	defer wg.Done()

	if c.config.GetP2PPort() == 0 {
		startupErr <- errors.New("please provide a port to bind on")
		return
	}

	// Make a host that listens on the given multiaddress
	hostInstance, err := c.makeBasicHost(c.config.GetP2PPort())
	if err != nil {
		startupErr <- err
		return
	}
	HostInstance = hostInstance
	// Set the grpc protocol handler on it
	grpcProto := p2pgrpc.NewGRPCProtocol(context.Background(), hostInstance)
	GRPCProtoInstance = *grpcProto

	p2ppb.RegisterP2PServiceServer(grpcProto.GetGRPCServer(), &Handler{})
	errOut := make(chan error)
	go func(proto *p2pgrpc.GRPCProtocol, errOut chan<- error) {
		errOut <- proto.Serve()
	}(grpcProto, errOut)

	hostInstance.Peerstore().AddAddr(hostInstance.ID(), hostInstance.Addrs()[0], pstore.TempAddrTTL)

	// Start DHT
	c.runDHT(ctx, hostInstance)

	for {
		select {
		case err := <-errOut:
			log.Infof("failed to accept p2p grpc connections: %v\n", err)
			startupErr <- err
			return
		case <-ctx.Done():
			log.Info("Shutting down GRPC server")
			grpcProto.GetGRPCServer().Stop()
			log.Info("GRPC server stopped")
			return
		}
	}
}

func (c *CentP2PServer) runDHT(ctx context.Context, h host.Host) error {

	//dhtClient := dht.NewDHTClient(ctx, h, rdStore) // Just run it as a client, will not respond to discovery requests
	dhtClient := dht.NewDHT(ctx, h, ds.NewMapDatastore()) // Run it as a Bootstrap Node

	log.Infof("Bootstrapping %s\n", c.config.GetBootstrapPeers())
	for _, addr := range c.config.GetBootstrapPeers() {
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
	tctx, _ := context.WithTimeout(ctx, time.Second*10)
	if err := dhtClient.Provide(tctx, cidPref, true); err != nil {
		// Important to keep this as Non-Fatal error, otherwise it will fail for a node that behaves as well as bootstrap one
		log.Infof("Error: %s\n", err.Error())
	}

	// Now, look for others who have announced
	log.Info("Searching for other peers ...")
	tctx, _ = context.WithTimeout(ctx, time.Second*10)
	peers, err := dhtClient.FindProviders(tctx, cidPref)
	if err != nil {
		log.Error(err)
	}
	log.Infof("Found %d peers!\n", len(peers))
	for _, p1 := range peers {
		log.Infof("Peer %s %s\n", p1.ID.Pretty(), p1.Addrs)
	}

	// Now connect to them, so they are added to the PeerStore
	for _, pe := range peers {
		if pe.ID == h.ID() {
			// No sense connecting to ourselves
			continue
		}
		tctx, _ := context.WithTimeout(ctx, time.Second*5)
		if err := h.Connect(tctx, pe); err != nil {
			log.Info("Failed to connect to peer: ", err)
		}
	}

	log.Info("Bootstrapping and discovery complete!")
	return nil
}

// makeBasicHost creates a LibP2P host with a peer ID listening on the given port
func (c *CentP2PServer) makeBasicHost(listenPort int) (host.Host, error) {
	priv, pub, err := c.createSigningKey()
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

	externalIP := config.Config().GetP2PExternalIP()
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

func (c *CentP2PServer) createSigningKey() (priv crypto.PrivKey, pub crypto.PubKey, err error) {
	// Create the signing key for the host
	publicKey, privateKey, err := cented25519.GetSigningKeyPairFromConfig()
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

func GetHost() (h host.Host) {
	h = HostInstance
	if h == nil {
		log.Fatal("Host undefined")
	}
	return
}

func GetGRPCProto() (g *p2pgrpc.GRPCProtocol) {
	g = &GRPCProtoInstance
	if g == nil {
		log.Fatal("Grpc not instantiated")
	}
	return
}

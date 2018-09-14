package p2p

import (
	"context"
	"fmt"
	"time"

	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/p2p"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/config"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/keytools/ed25519"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/notification"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/p2p/p2phandler"
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

var log = logging.Logger("p2p")
var HostInstance host.Host
var GRPCProtoInstance p2pgrpc.GRPCProtocol

// makeBasicHost creates a LibP2P host with a random peer ID listening on the given port
func makeBasicHost(listenPort int) (host.Host, error) {
	// Get the signing key for the host.
	publicKey, privateKey := ed25519.GetSigningKeyPairFromConfig()
	var key []byte
	key = append(key, privateKey...)
	key = append(key, publicKey...)

	priv, err := crypto.UnmarshalEd25519PrivateKey(key)
	if err != nil {
		return nil, err
	}
	pub := priv.GetPublic()

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

	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", listenPort)),
		libp2p.Identity(priv),
		libp2p.DefaultMuxers,
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

func RunDHT(ctx context.Context, h host.Host) {

	//dhtClient := dht.NewDHTClient(ctx, h, rdStore) // Just run it as a client, will not respond to discovery requests
	dhtClient := dht.NewDHT(ctx, h, ds.NewMapDatastore()) // Run it as a Bootstrap Node

	bootstrapPeers := config.Config.GetBootstrapPeers()

	log.Infof("Bootstrapping %s\n", bootstrapPeers)
	for _, addr := range bootstrapPeers {
		iaddr, _ := ipfsaddr.ParseString(addr)

		pinfo, _ := pstore.InfoFromP2pAddr(iaddr.Multiaddr())

		if err := h.Connect(ctx, *pinfo); err != nil {
			log.Info("Bootstrapping to peer failed: ", err)
		}
	}

	// Using the sha256 of our "topic" as our rendezvous value
	c, _ := cid.NewPrefixV1(cid.Raw, mh.SHA2_256).Sum([]byte("centrifuge-dht"))

	// First, announce ourselves as participating in this topic
	log.Info("Announcing ourselves...")
	tctx, _ := context.WithTimeout(ctx, time.Second*10)
	if err := dhtClient.Provide(tctx, c, true); err != nil {
		// Important to keep this as Non-Fatal error, otherwise it will fail for a node that behaves as well as bootstrap one
		log.Infof("Error: %s\n", err.Error())
	}

	// Now, look for others who have announced
	log.Info("Searching for other peers ...")
	tctx, _ = context.WithTimeout(ctx, time.Second*10)
	peers, err := dhtClient.FindProviders(tctx, c)
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
}

func RunP2P() {
	// Parse options from the command line
	port := config.Config.GetP2PPort()
	if port == 0 {
		log.Fatal("Please provide a port to bind on")
	}

	// Make a host that listens on the given multiaddress
	hostInstance, err := makeBasicHost(port)
	if err != nil {
		log.Fatal(err)
	}
	HostInstance = hostInstance
	// Set the grpc protocol handler on it
	grpcProto := p2pgrpc.NewGRPCProtocol(context.Background(), hostInstance)
	GRPCProtoInstance = *grpcProto

	p2ppb.RegisterP2PServiceServer(grpcProto.GetGRPCServer(), &p2phandler.Handler{Notifier: &notification.WebhookSender{}})

	hostInstance.Peerstore().AddAddr(hostInstance.ID(), hostInstance.Addrs()[0], pstore.TempAddrTTL)

	// Start DHT
	RunDHT(context.Background(), hostInstance)

	select {}
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

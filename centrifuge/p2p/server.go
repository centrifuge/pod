package p2p

import (
	"context"
	"fmt"
	"log"

	golog "github.com/ipfs/go-log"
	"github.com/libp2p/go-libp2p-crypto"
	"github.com/libp2p/go-libp2p-host"
	"github.com/libp2p/go-libp2p-peer"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	"github.com/libp2p/go-libp2p-swarm"
	bhost "github.com/libp2p/go-libp2p/p2p/host/basic"
	ma "github.com/multiformats/go-multiaddr"
	gologging "github.com/whyrusleeping/go-logging"
	msmux "github.com/whyrusleeping/go-smux-multistream"
	yamux "github.com/whyrusleeping/go-smux-yamux"
	"github.com/paralin/go-libp2p-grpc"
	"github.com/spf13/viper"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/keytools"
	"github.com/ipfs/go-cid"
	mh "github.com/multiformats/go-multihash"
	"github.com/ipfs/go-ipfs-addr"
	"time"
	"github.com/libp2p/go-libp2p-kad-dht"
	ds "github.com/ipfs/go-datastore"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/p2p"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument/controller"
)

var	HostInstance host.Host
var GRPCProtoInstance p2pgrpc.GRPCProtocol

// makeBasicHost creates a LibP2P host with a random peer ID listening on the
// given multiaddress.
func makeBasicHost(listenPort int) (host.Host, error) {
	// Get the signing key for the host.
	publicKey, privateKey := keytools.GetSigningKeyPairFromConfig()
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

	// Create a multiaddress
	addr, err := ma.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", listenPort))
	if err != nil {
		return nil, err
	}

	// Create a peerstore
	ps := pstore.NewPeerstore()

	// Add the keys to the peerstore
	// for this peer ID.
	err = ps.AddPubKey(pid, pub)
	if err != nil {
		log.Printf("Could not enable encryption: %v\n", err)
		return nil, err
	}
	err = ps.AddPrivKey(pid, priv)
	if err != nil {
		log.Printf("Could not enable encryption: %v\n", err)
		return nil, err
	}

	// Set up stream multiplexer
	tpt := msmux.NewBlankTransport()
	tpt.AddTransport("/yamux/1.0.0", yamux.DefaultTransport)

	// Create swarm (implements libP2P Network)
	swrm, err := swarm.NewSwarmWithProtector(
		context.Background(),
		[]ma.Multiaddr{addr},
		pid,
		ps,
		nil,
		tpt,
		nil,
	)
	if err != nil {
		return nil, err
	}

	netw := (*swarm.Network)(swrm)
	basicHost := bhost.New(netw)

	// Build host multiaddress
	hostAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/ipfs/%s", basicHost.ID().Pretty()))

	// Now we can build a full multiaddress to reach this host
	// by encapsulating both addresses:
	fullAddr := addr.Encapsulate(hostAddr)
	log.Printf("I am %s\n", fullAddr)

	return basicHost, nil
}

func RunDHT(ctx context.Context, h host.Host) {

	//dhtClient := dht.NewDHTClient(ctx, h, rdStore) // Just run it as a client, will not respond to discovery requests
	dhtClient := dht.NewDHT(ctx, h, ds.NewMapDatastore()) // Run it as a Bootstrap Node

	// IPFS Bootstrap Peer nodes
	//"/ip4/172.16.0.102/tcp/38204/ipfs/QmNYcCDjtCRdYaYPpNkTSiQTpLLxRapMe1P3EGsmA2wK7D",
	//"/ip4/104.131.131.82/tcp/4001/ipfs/QmaCpDMGvV2BGHeYERUEnRQAwe3N8SzbUtfsmvsqQLuvuJ",
	//"/ip4/104.236.179.241/tcp/4001/ipfs/QmSoLPppuBtQSGwKDZT2M73ULpjvfd3aZ6ha4oFGL1KrGM",
	//"/ip4/104.236.76.40/tcp/4001/ipfs/QmSoLV4Bbm51jM9C4gDYZQ9Cy3U6aXMJDAbzgu2fzaDs64",
	//"/ip4/128.199.219.111/tcp/4001/ipfs/QmSoLSafTMBsPKadTEgaXctDQVcqN88CNLHXMkTNwMKPnu",
	//"/ip4/178.62.158.247/tcp/4001/ipfs/QmSoLer265NRgSp2LA3dPaeykiS1J6DifTC88f5uVQKNAd",
	bootstrapPeers := viper.GetStringSlice("p2p.bootstrapPeers")

	log.Printf("Bootstrapping %s\n", bootstrapPeers)
	for _, addr := range bootstrapPeers {
		iaddr, _ := ipfsaddr.ParseString(addr)

		pinfo, _ := pstore.InfoFromP2pAddr(iaddr.Multiaddr())

		if err := h.Connect(ctx, *pinfo); err != nil {
			log.Println("Bootstrapping to peer failed: ", err)
		}
	}

	// Using the sha256 of our "topic" as our rendezvous value
	c, _ := cid.NewPrefixV1(cid.Raw, mh.SHA2_256).Sum([]byte("centrifuge-dht"))

	// First, announce ourselves as participating in this topic
	log.Println("Announcing ourselves...")
	tctx, _ := context.WithTimeout(ctx,  time.Second*10)
	if err := dhtClient.Provide(tctx, c, true); err != nil {
		// Important to keep this as Non-Fatal error, otherwise it will fail for a node that behaves as well as bootstrap one
		log.Printf("Error: %s\n", err.Error())
	}

	// Now, look for others who have announced
	log.Println("Searching for other peers ...")
	tctx, _ = context.WithTimeout(ctx, time.Second*10)
	peers, err := dhtClient.FindProviders(tctx, c)
	if err != nil {
		panic(err)
	}
	log.Printf("Found %d peers!\n", len(peers))
	for _, p1 := range peers {
		log.Printf("Peer %s %s\n", p1.ID.Pretty(), p1.Addrs)
	}

	// Now connect to them, so they are added to the PeerStore
	for _, pe := range peers {
		if pe.ID == h.ID() {
			// No sense connecting to ourselves
			continue
		}

		tctx, _ := context.WithTimeout(ctx, time.Second*5)
		if err := h.Connect(tctx, pe); err != nil {
			log.Println("Failed to connect to peer: ", err)
		}
	}

	log.Println("Bootstrapping and discovery complete!")
}

func RunP2P() {
	// LibP2P code uses golog to log messages. They log with different
	// string IDs (i.e. "swarm"). We can control the verbosity level for
	// all loggers with:
	golog.SetAllLoggers(gologging.DEBUG) // Change to DEBUG for extra info

	// Parse options from the command line
	port := viper.GetInt("p2p.port")
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

	p2ppb.RegisterP2PServiceServer(grpcProto.GetGRPCServer(), &coredocumentcontroller.P2PService{})

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

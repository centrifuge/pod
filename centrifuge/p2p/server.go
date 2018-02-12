package p2p

import (
	"context"
	"fmt"
	"log"

	golog "github.com/ipfs/go-log"
	crypto "github.com/libp2p/go-libp2p-crypto"
	host "github.com/libp2p/go-libp2p-host"
	peer "github.com/libp2p/go-libp2p-peer"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	swarm "github.com/libp2p/go-libp2p-swarm"
	bhost "github.com/libp2p/go-libp2p/p2p/host/basic"
	ma "github.com/multiformats/go-multiaddr"
	gologging "github.com/whyrusleeping/go-logging"
	msmux "github.com/whyrusleeping/go-smux-multistream"
	yamux "github.com/whyrusleeping/go-smux-yamux"
	"github.com/paralin/go-libp2p-grpc"
	"github.com/spf13/viper"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/keytools"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/storage/invoicestorage"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/storage"
)

//go:generate protoc -I $PROTOBUF/src/ -I . -I ../ -I $GOPATH/src -I ../../vendor/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis -I ../../vendor/github.com/grpc-ecosystem/grpc-gateway --go_out=plugins=grpc:$GOPATH/src/ p2p.proto

var	HostInstance = make(chan host.Host)
var GRPCProtoInstance = make(chan *p2pgrpc.GRPCProtocol)


type P2PService struct {

}

func (srv *P2PService) TransmitInvoice(ctx context.Context, req *TransmitInvoiceDocument) (rep *TransmitReply, err error) {
	invoiceStorage := invoicestorage.StorageService{}
	invoiceStorage.SetStorageBackend(storage.GetStorage())

	fmt.Println("I RECEIVED A DOCUMENT")
	fmt.Println("I RECEIVED A DOCUMENT")
	fmt.Println("I RECEIVED A DOCUMENT")

	err = invoiceStorage.PutDocument(req.Invoice)
	if err != nil {
		return

	}
	err = invoiceStorage.ReceiveDocument(req.Invoice)
	rep = &TransmitReply{req.Invoice}
	return
}

// makeBasicHost creates a LibP2P host with a random peer ID listening on the
// given multiaddress.
func makeBasicHost(listenPort int) (host.Host, error) {
	// Get the signing key for the host.
	publicKey, privateKey := keytools.GetSigningKeysFromConfig()
	var key []byte
	key = append(key, privateKey...)
	key = append(key, publicKey...)

	priv, err := crypto.UnmarshalEd25519PrivateKey(key)
	if err != nil {
		return nil, err
	}
	pub := priv.GetPublic()

	// Obtain Peer ID from public key
	pid, err := peer.IDFromPublicKey(pub)
	if err != nil {
		return nil, err
	}

	// Create a multiaddress
	addr, err := ma.NewMultiaddr(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", listenPort))
	if err != nil {
		return nil, err
	}

	// Create a peerstore
	ps := pstore.NewPeerstore()

	// Add the keys to the peerstore
	// for this peer ID.
	ps.AddPrivKey(pid, priv)
	ps.AddPubKey(pid, pub)

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

func RunP2P() {
	// LibP2P code uses golog to log messages. They log with different
	// string IDs (i.e. "swarm"). We can control the verbosity level for
	// all loggers with:
	golog.SetAllLoggers(gologging.INFO) // Change to DEBUG for extra info

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
	HostInstance <- hostInstance

	// Set the grpc protocol handler on it
	grpcProto := p2pgrpc.NewGRPCProtocol(context.Background(), hostInstance)
	GRPCProtoInstance <- grpcProto
	RegisterP2PServiceServer(grpcProto.GetGRPCServer(), &P2PService{})

	select {}
}

// GetStorage is a singleton implementation returning the default database as configured
func GetHost() (h host.Host) {
	h = <- HostInstance
	if h == nil {
		log.Fatal("Host undefined")
	}
	return
}


func GetGRPCProto() (g *p2pgrpc.GRPCProtocol) {
	g = <- GRPCProtoInstance
	if g == nil {
		log.Fatal("Grpc not instantiated")
	}
	return
}
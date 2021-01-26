package p2p

import (
	"context"
	"fmt"
	"sync"
	"time"

	pb "github.com/centrifuge/centrifuge-protobufs/gen/go/protocol"
	"github.com/centrifuge/go-centrifuge/config"
	crypto2 "github.com/centrifuge/go-centrifuge/crypto"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	p2pcommon "github.com/centrifuge/go-centrifuge/p2p/common"
	ms "github.com/centrifuge/go-centrifuge/p2p/messenger"
	"github.com/centrifuge/go-centrifuge/p2p/receiver"
	logging "github.com/ipfs/go-log"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	libp2pPeer "github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/libp2p/go-libp2p-core/routing"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	secio "github.com/libp2p/go-libp2p-secio"
	"github.com/libp2p/go-tcp-transport"
	ma "github.com/multiformats/go-multiaddr"
)

var log = logging.Logger("p2p-server")

// messenger is an interface to wrap p2p messaging implementation
type messenger interface {

	// Init inits the messenger
	Init(protocols ...protocol.ID)

	// SendMessage sends a message through messenger
	SendMessage(ctx context.Context, p libp2pPeer.ID, pmes *pb.P2PEnvelope, protoc protocol.ID) (*pb.P2PEnvelope, error)
}

// peer implements node.Server
type peer struct {
	disablePeerStore bool
	config           config.Service
	idService        identity.Service
	host             host.Host
	handlerCreator   func() *receiver.Handler
	mes              messenger
	dht              *dht.IpfsDHT
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
	priv, pub, err := crypto2.ObtainP2PKeypair(nc.GetP2PKeyPair())
	if err != nil {
		startupErr <- err
		return
	}
	s.host, s.dht, err = makeBasicHost(ctx, priv, pub, nc.GetP2PExternalIP(), nc.GetP2PPort())
	if err != nil {
		startupErr <- err
		return
	}

	s.mes = ms.NewP2PMessenger(ctx, s.host, nc.GetP2PConnectionTimeout(), s.handlerCreator().HandleInterceptor)
	err = s.initProtocols()
	if err != nil {
		startupErr <- err
		return
	}

	// Start DHT and properly ignore errors :)
	_ = s.runDHT(ctx, nc.GetBootstrapPeers())

	if nc.IsDebugLogEnabled() {
		go func() {
			for {
				num := s.host.Peerstore().Peers()
				log.Debugf("for host %s the peers in the peerstore are", s.host.ID(), num)
				time.Sleep(2 * time.Second)
			}
		}()
	}

	<-ctx.Done()

}

func (s *peer) initProtocols() error {
	tcs, err := s.config.GetAccounts()
	if err != nil {
		return err
	}
	var protocols []protocol.ID
	for _, t := range tcs {
		accID := t.GetIdentityID()
		did, err := identity.NewDIDFromBytes(accID)
		if err != nil {
			return err
		}
		protocols = append(protocols, p2pcommon.ProtocolForDID(did))
	}
	s.mes.Init(protocols...)
	return nil
}

func (s *peer) InitProtocolForDID(did identity.DID) {
	p := p2pcommon.ProtocolForDID(did)
	s.mes.Init(p)
}

func (s *peer) runDHT(ctx context.Context, bootstrapPeers []string) error {
	log.Infof("Bootstrapping %s\n", bootstrapPeers)

	for _, addr := range bootstrapPeers {
		multiaddr, _ := ma.NewMultiaddr(addr)
		p, err := libp2pPeer.AddrInfoFromP2pAddr(multiaddr)
		if err != nil {
			log.Info(err)
			continue
		}

		s.host.Peerstore().AddAddrs(p.ID, p.Addrs, peerstore.PermanentAddrTTL)
		err = s.host.Connect(ctx, *p)
		if err != nil {
			log.Info("Bootstrapping to peer failed: ", err)
			continue
		}

		fmt.Printf("Connection to %s %s successful\n", p.ID, p.Addrs)
	}

	err := s.dht.Bootstrap(ctx)
	if err != nil {
		log.Errorf("Bootstrap Error: %s", err.Error())
		return err
	}

	log.Info("Bootstrapping and discovery complete!")
	return nil
}

// makeBasicHost creates a LibP2P host with a peer ID listening on the given port
func makeBasicHost(ctx context.Context, priv crypto.PrivKey, pub crypto.PubKey, externalIP string, listenPort int) (host.Host, *dht.IpfsDHT, error) {
	var err error
	var extMultiAddr ma.Multiaddr
	if externalIP == "" {
		log.Warn("External IP not defined, Peers might not be able to resolve this node if behind NAT\n")
	} else {
		extMultiAddr, err = ma.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d", externalIP, listenPort))
		if err != nil {
			return nil, nil, errors.New("failed to create multiaddr: %v", err)
		}
	}

	addressFactory := func(addrs []ma.Multiaddr) []ma.Multiaddr {
		if extMultiAddr != nil {
			// We currently support a single protocol and transport, if we add more to support then we will need to adapt this code
			addrs = []ma.Multiaddr{extMultiAddr}
		}
		return addrs
	}

	var idht *dht.IpfsDHT
	opts := []libp2p.Option{
		libp2p.Identity(priv),
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", listenPort)),
		libp2p.Security(secio.ID, secio.New),
		// support any other default transports (TCP)
		libp2p.Transport(tcp.NewTCPTransport),
		// Attempt to open ports using uPNP for NATed hosts.
		libp2p.NATPortMap(),
		// Let this host use the DHT to find other hosts
		libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
			idht, err = dht.New(ctx, h, dht.Mode(dht.ModeAutoServer))
			return idht, err
		}),
		libp2p.DefaultMuxers,
		libp2p.AddrsFactory(addressFactory),
	}

	bhost, err := libp2p.New(ctx, opts...)
	if err != nil {
		return nil, nil, err
	}

	log.Infof("P2P Server at: %s %s\n", bhost.ID(), bhost.Addrs())
	return bhost, idht, err
}

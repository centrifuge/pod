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
	"github.com/centrifuge/go-centrifuge/p2p/common"
	ms "github.com/centrifuge/go-centrifuge/p2p/messenger"
	"github.com/centrifuge/go-centrifuge/p2p/receiver"
	ds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-ipfs-addr"
	logging "github.com/ipfs/go-log"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	libp2pPeer "github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p-peerstore/pstoremem"
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
	s.host, err = makeBasicHost(ctx, priv, pub, nc.GetP2PExternalIP(), nc.GetP2PPort())
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
		DID, err := identity.NewDIDFromBytes(accID)
		if err != nil {
			return err
		}
		protocols = append(protocols, p2pcommon.ProtocolForDID(&DID))
	}
	s.mes.Init(protocols...)
	return nil
}

func (s *peer) InitProtocolForDID(DID *identity.DID) {
	p := p2pcommon.ProtocolForDID(DID)
	s.mes.Init(p)
}

func (s *peer) runDHT(ctx context.Context, bootstrapPeers []string) error {
	s.dht = dht.NewDHT(ctx, s.host, ds.NewMapDatastore())
	log.Infof("Bootstrapping %s\n", bootstrapPeers)

	for _, addr := range bootstrapPeers {
		iaddr, _ := ipfsaddr.ParseString(addr)
		pinfo, _ := libp2pPeer.AddrInfoFromP2pAddr(iaddr.Multiaddr())
		if err := s.host.Connect(ctx, *pinfo); err != nil {
			log.Info("Bootstrapping to peer failed: ", err)
		}
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
func makeBasicHost(ctx context.Context, priv crypto.PrivKey, pub crypto.PubKey, externalIP string, listenPort int) (host.Host, error) {
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
	ps := pstoremem.NewPeerstore()

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
		log.Warn("External IP not defined, Peers might not be able to resolve this node if behind NAT\n")
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
		libp2p.Peerstore(ps),
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", listenPort)),
		libp2p.Identity(priv),
		libp2p.DefaultMuxers,
		libp2p.AddrsFactory(addressFactory),
	}

	bhost, err := libp2p.New(ctx, opts...)
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

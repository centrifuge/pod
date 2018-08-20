package p2p

import (
	"fmt"

	"github.com/libp2p/go-libp2p-peer"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	ma "github.com/multiformats/go-multiaddr"

	"context"

	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/p2p"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

// Opens a client connection with libp2p
func OpenClient(target string) p2ppb.P2PServiceClient {
	log.Info("Opening connection to: %s", target)
	ipfsAddr, err := ma.NewMultiaddr(target)
	if err != nil {
		log.Fatal(err)
	}

	pid, err := ipfsAddr.ValueForProtocol(ma.P_IPFS)
	if err != nil {
		log.Fatal(err)
	}

	peerID, err := peer.IDB58Decode(pid)
	if err != nil {
		log.Fatal(err)
	}

	// Decapsulate the /ipfs/<peerID> part from the target
	// /ip4/<a.b.c.d>/ipfs/<peer> becomes /ip4/<a.b.c.d>
	targetPeerAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/ipfs/%s", peer.IDB58Encode(peerID)))
	targetAddr := ipfsAddr.Decapsulate(targetPeerAddr)

	hostInstance := GetHost()
	grpcProtoInstance := GetGRPCProto()

	// We have a peer ID and a targetAddr so we add it to the peer store
	// so LibP2P knows how to contact it
	hostInstance.Peerstore().AddAddr(peerID, targetAddr, pstore.PermanentAddrTTL)

	// make a new stream from host B to host A
	g, err := grpcProtoInstance.Dial(context.Background(), peerID, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}

	for {
		if g.GetState() == connectivity.Ready {
			break
		}
	}
	return p2ppb.NewP2PServiceClient(g)
}

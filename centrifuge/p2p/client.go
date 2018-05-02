package p2p

import (
	"fmt"
	"log"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/libp2p/go-libp2p-peer"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	"context"
	"google.golang.org/grpc"
	"sync"
	"google.golang.org/grpc/connectivity"
	cdgrpc "github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument/coredocumentgrpc"
)

// Opens a client connection with libp2p
func OpenClient (target string) cdgrpc.P2PServiceClient {
	log.Printf("Opening connection to: %s", target)
	ipfsaddr, err := ma.NewMultiaddr(target)
	if err != nil {
		log.Fatalln(err)
	}

	pid, err := ipfsaddr.ValueForProtocol(ma.P_IPFS)
	if err != nil {
		log.Fatalln(err)
	}

	peerid, err := peer.IDB58Decode(pid)
	if err != nil {
		log.Fatalln(err)
	}

	// Decapsulate the /ipfs/<peerID> part from the target
	// /ip4/<a.b.c.d>/ipfs/<peer> becomes /ip4/<a.b.c.d>
	targetPeerAddr, _ := ma.NewMultiaddr(
		fmt.Sprintf("/ipfs/%s", peer.IDB58Encode(peerid)))
	targetAddr := ipfsaddr.Decapsulate(targetPeerAddr)


	hostInstance := GetHost()
	grpcProtoInstance := GetGRPCProto()

	// We have a peer ID and a targetAddr so we add it to the peerstore
	// so LibP2P knows how to contact it
	hostInstance.Peerstore().AddAddr(peerid, targetAddr, pstore.PermanentAddrTTL)

	// make a new stream from host B to host A
	var wg sync.WaitGroup
	var grpcConn *grpc.ClientConn
	wg.Add(1)
	go func() {
		defer wg.Done()
		g, err := grpcProtoInstance.Dial(context.Background(), peerid, grpc.WithInsecure())
		if err != nil {
			log.Fatalln(err)
		}
		for {
			if g.GetState() == connectivity.Ready {
				break
			}
		}
		grpcConn = g
	}()
	wg.Wait()
	return cdgrpc.NewP2PServiceClient(grpcConn)
}
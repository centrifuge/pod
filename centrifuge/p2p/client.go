package p2p

import (
	"context"
	"fmt"

	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/coredocument"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/p2p"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/centerrors"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/code"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/config"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/identity"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/signatures"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/version"
	"github.com/libp2p/go-libp2p-peer"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	ma "github.com/multiformats/go-multiaddr"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

// Opens a client connection with libp2p
func OpenClient(target string) (p2ppb.P2PServiceClient, error) {
	log.Info("Opening connection to: %s", target)
	ipfsAddr, err := ma.NewMultiaddr(target)
	if err != nil {
		return nil, err
	}

	pid, err := ipfsAddr.ValueForProtocol(ma.P_IPFS)
	if err != nil {
		return nil, err
	}

	peerID, err := peer.IDB58Decode(pid)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	for {
		if g.GetState() == connectivity.Ready {
			break
		}
	}
	return p2ppb.NewP2PServiceClient(g), nil
}

// getSignatureForDocument requests the target node to sign the document
func getSignatureForDocument(ctx context.Context, doc coredocumentpb.CoreDocument, client p2ppb.P2PServiceClient) (*p2ppb.SignatureResponse, error) {
	header := p2ppb.CentrifugeHeader{
		NetworkIdentifier:  config.Config.GetNetworkID(),
		CentNodeVersion:    version.GetVersion().String(),
		SenderCentrifugeId: config.Config.GetIdentityId(),
	}

	req := &p2ppb.SignatureRequest{
		Header:   &header,
		Document: &doc,
	}

	resp, err := client.RequestDocumentSignature(ctx, req)
	if err != nil {
		return nil, centerrors.Wrap(err, "request for document signature failed")
	}

	compatible := version.CheckVersion(resp.CentNodeVersion)
	if !compatible {
		return nil, version.IncompatibleVersionError(resp.CentNodeVersion)
	}

	valid, err := signatures.ValidateSignature(resp.Signature, doc.SigningRoot)
	if err != nil {
		return nil, centerrors.Wrap(err, "failed to validate signature")
	}

	if !valid {
		return nil, centerrors.New(code.AuthenticationFailed, "signature invalid")
	}

	return resp, nil
}

type signatureResponseWrap struct {
	resp *p2ppb.SignatureResponse
	err  error
}

func getSignatureAsync(ctx context.Context, doc coredocumentpb.CoreDocument, client p2ppb.P2PServiceClient, out chan<- signatureResponseWrap) {
	resp, err := getSignatureForDocument(ctx, doc, client)
	out <- signatureResponseWrap{
		resp: resp,
		err:  err,
	}
}

// GetSignaturesForDocument requests peer nodes for the signature and verifies them
func GetSignaturesForDocument(ctx context.Context, doc *coredocumentpb.CoreDocument, idService identity.Service, centIDs []identity.CentId) error {
	if doc == nil {
		return centerrors.NilError(doc)
	}

	targets, err := identity.GetClientsP2PURLs(centIDs)
	if err != nil {
		return centerrors.Wrap(err, "failed to get P2P urls")
	}

	in := make(chan signatureResponseWrap)
	defer close(in)

	for _, target := range targets {
		client, err := OpenClient(target)
		if err != nil {
			return centerrors.Wrap(err, "failed to connect to target")
		}

		// for now going with context.background, once we have a timeout for request
		// we can use context.Timeout for that
		go getSignatureAsync(ctx, *doc, client, in)
	}

	var responses []signatureResponseWrap
	for range targets {
		responses = append(responses, <-in)
	}

	for _, resp := range responses {
		if resp.err != nil {
			return err
		}

		doc.Signatures = append(doc.Signatures, resp.resp.Signature)
	}

	return nil
}

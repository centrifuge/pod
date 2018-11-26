package p2p

import (
	"context"
	"fmt"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	"github.com/centrifuge/go-centrifuge/centerrors"
	"github.com/centrifuge/go-centrifuge/code"
	"github.com/centrifuge/go-centrifuge/coredocument"
	"github.com/centrifuge/go-centrifuge/header"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/version"
	"github.com/libp2p/go-libp2p-peer"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	ma "github.com/multiformats/go-multiaddr"
	"google.golang.org/grpc"
)

// Client defines methods that can be implemented by any type handling p2p communications.
type Client interface {
	OpenClient(target string) (p2ppb.P2PServiceClient, error)
	GetSignaturesForDocument(ctx *header.ContextHeader, identityService identity.Service, doc *coredocumentpb.CoreDocument) error
}

// OpenClient returns P2PServiceClient to contact the remote peer
func (s *p2pServer) OpenClient(target string) (p2ppb.P2PServiceClient, error) {
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

	// We have a peer ID and a targetAddr so we add it to the peer store
	// so LibP2P knows how to contact it
	s.host.Peerstore().AddAddr(peerID, targetAddr, pstore.PermanentAddrTTL)

	// make a new stream from host B to host A with timeout
	// Retrial is handled internally, connection request will be cancelled by the connection timeout context
	ctx, cancel := context.WithTimeout(context.Background(), s.config.GetP2PConnectionTimeout())
	defer cancel()
	g, err := s.protocol.Dial(ctx, peerID, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, fmt.Errorf("failed to dial peer [%s]: %v", peerID.Pretty(), err)
	}

	return p2ppb.NewP2PServiceClient(g), nil
}

// getSignatureForDocument requests the target node to sign the document
func (s *p2pServer) getSignatureForDocument(ctx context.Context, identityService identity.Service, doc coredocumentpb.CoreDocument, client p2ppb.P2PServiceClient, receiverCentID identity.CentID) (*p2ppb.SignatureResponse, error) {
	senderID, err := s.config.GetIdentityID()
	if err != nil {
		return nil, err
	}

	header := p2ppb.CentrifugeHeader{
		NetworkIdentifier:  s.config.GetNetworkID(),
		CentNodeVersion:    version.GetVersion().String(),
		SenderCentrifugeId: senderID,
	}

	req := &p2ppb.SignatureRequest{
		Header:   &header,
		Document: &doc,
	}

	log.Infof("Requesting signature from %s\n", receiverCentID)

	resp, err := client.RequestDocumentSignature(ctx, req)
	if err != nil {
		return nil, centerrors.Wrap(err, "request for document signature failed")
	}

	compatible := version.CheckVersion(resp.CentNodeVersion)
	if !compatible {
		return nil, version.IncompatibleVersionError(resp.CentNodeVersion)
	}

	err = identity.ValidateCentrifugeIDBytes(resp.Signature.EntityId, receiverCentID)
	if err != nil {
		return nil, centerrors.New(code.AuthenticationFailed, err.Error())
	}

	err = identityService.ValidateSignature(resp.Signature, doc.SigningRoot)
	if err != nil {
		return nil, centerrors.New(code.AuthenticationFailed, "signature invalid")
	}

	log.Infof("Signature successfully received from %s\n", receiverCentID)
	return resp, nil
}

type signatureResponseWrap struct {
	resp *p2ppb.SignatureResponse
	err  error
}

func (s *p2pServer) getSignatureAsync(ctx context.Context, identityService identity.Service, doc coredocumentpb.CoreDocument, client p2ppb.P2PServiceClient, receiverCentID identity.CentID, out chan<- signatureResponseWrap) {
	resp, err := s.getSignatureForDocument(ctx, identityService, doc, client, receiverCentID)
	out <- signatureResponseWrap{
		resp: resp,
		err:  err,
	}
}

// GetSignaturesForDocument requests peer nodes for the signature and verifies them
func (s *p2pServer) GetSignaturesForDocument(ctx *header.ContextHeader, identityService identity.Service, doc *coredocumentpb.CoreDocument) error {
	in := make(chan signatureResponseWrap)
	defer close(in)

	extCollaborators, err := coredocument.GetExternalCollaborators(ctx.Self().ID, doc)
	if err != nil {
		return centerrors.Wrap(err, "failed to get external collaborators")
	}

	var count int
	for _, collaborator := range extCollaborators {
		collaboratorID, err := identity.ToCentID(collaborator)
		if err != nil {
			return centerrors.Wrap(err, "failed to convert to CentID")
		}
		target, err := identityService.GetClientP2PURL(collaboratorID)

		if err != nil {
			return centerrors.Wrap(err, "failed to get P2P url")
		}

		client, err := s.OpenClient(target)
		if err != nil {
			log.Error(centerrors.Wrap(err, "failed to connect to target"))
			continue
		}

		// for now going with context.background, once we have a timeout for request
		// we can use context.Timeout for that
		count++
		go s.getSignatureAsync(ctx.Context(), identityService, *doc, client, collaboratorID, in)
	}

	var responses []signatureResponseWrap
	for i := 0; i < count; i++ {
		responses = append(responses, <-in)
	}

	for _, resp := range responses {
		if resp.err != nil {
			log.Error(resp.err)
			continue
		}

		doc.Signatures = append(doc.Signatures, resp.resp.Signature)
	}

	return nil
}

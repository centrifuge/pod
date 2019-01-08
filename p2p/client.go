package p2p

import (
	"context"
	"fmt"

	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/protocol"
	"github.com/golang/protobuf/proto"

	"github.com/centrifuge/go-centrifuge/contextutil"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/errors"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	"github.com/centrifuge/go-centrifuge/centerrors"
	"github.com/centrifuge/go-centrifuge/code"
	"github.com/centrifuge/go-centrifuge/coredocument"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/version"
	"github.com/libp2p/go-libp2p-peer"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	ma "github.com/multiformats/go-multiaddr"
)

// Client defines methods that can be implemented by any type handling p2p communications.
type Client interface {

	// GetSignaturesForDocument gets the signatures for document
	GetSignaturesForDocument(ctx context.Context, identityService identity.Service, doc *coredocumentpb.CoreDocument) error

	// after all signatures are collected the sender sends the document including the signatures
	SendAnchoredDocument(ctx context.Context, id identity.Identity, in *p2ppb.AnchorDocumentRequest) (*p2ppb.AnchorDocumentResponse, error)
}

func (s *centPeer) SendAnchoredDocument(ctx context.Context, id identity.Identity, in *p2ppb.AnchorDocumentRequest) (*p2ppb.AnchorDocumentResponse, error) {
	pid, err := s.getPeerID(id)
	if err != nil {
		return nil, err
	}
	marshalledRequest, err := proto.Marshal(in)
	if err != nil {
		return nil, err
	}

	// TODO [multi-tenancy] modify the protocol id here to include the centID of the receiving node
	recv, err := s.mes.sendMessage(ctx, pid, &protocolpb.P2PEnvelope{Type: protocolpb.MessageType_MESSAGE_TYPE_SEND_ANCHORED_DOC, Body: marshalledRequest}, CentrifugeProtocol)
	if err != nil {
		return nil, err
	}

	// handle client error
	if recv.Type == protocolpb.MessageType_MESSAGE_TYPE_ERROR {
		return nil, convertClientError(recv)
	}

	if recv.Type != protocolpb.MessageType_MESSAGE_TYPE_SEND_ANCHORED_DOC_REP {
		return nil, errors.New("the received send anchored document response is incorrect")
	}
	r := new(p2ppb.AnchorDocumentResponse)
	err = proto.Unmarshal(recv.Body, r)
	if err != nil {
		return nil, err
	}
	return r, nil
}

// OpenClient returns P2PServiceClient to contact the remote peer
func (s *centPeer) getPeerID(id identity.Identity) (peer.ID, error) {
	lastB58Key, err := id.CurrentP2PKey()
	if err != nil {
		return "", errors.New("error fetching p2p key: %v", err)
	}
	target := fmt.Sprintf("/ipfs/%s", lastB58Key)
	log.Info("Opening connection to: %s", target)
	ipfsAddr, err := ma.NewMultiaddr(target)
	if err != nil {
		return "", err
	}

	pid, err := ipfsAddr.ValueForProtocol(ma.P_IPFS)
	if err != nil {
		return "", err
	}

	peerID, err := peer.IDB58Decode(pid)
	if err != nil {
		return "", err
	}

	// Decapsulate the /ipfs/<peerID> part from the target
	// /ip4/<a.b.c.d>/ipfs/<peer> becomes /ip4/<a.b.c.d>
	targetPeerAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/ipfs/%s", pid))
	targetAddr := ipfsAddr.Decapsulate(targetPeerAddr)

	// We have a peer ID and a targetAddr so we add it to the peer store
	// so LibP2P knows how to contact it
	s.host.Peerstore().AddAddr(peerID, targetAddr, pstore.PermanentAddrTTL)

	return peerID, nil
}

// getSignatureForDocument requests the target node to sign the document
func (s *centPeer) getSignatureForDocument(ctx context.Context, identityService identity.Service, doc coredocumentpb.CoreDocument, receiverPeer peer.ID, receiverCentID identity.CentID) (*p2ppb.SignatureResponse, error) {
	senderID, err := s.config.GetIdentityID()
	if err != nil {
		return nil, err
	}

	req, err := s.createSignatureRequest(senderID, &doc)
	if err != nil {
		return nil, err
	}

	log.Infof("Requesting signature from %s\n", receiverCentID)
	recv, err := s.mes.sendMessage(ctx, receiverPeer, req, CentrifugeProtocol)
	if err != nil {
		return nil, err
	}

	// handle client error
	if recv.Type == protocolpb.MessageType_MESSAGE_TYPE_ERROR {
		return nil, convertClientError(recv)
	}

	if recv.Type != protocolpb.MessageType_MESSAGE_TYPE_REQUEST_SIGNATURE_REP {
		return nil, errors.New("the received request signature response is incorrect")
	}
	resp := new(p2ppb.SignatureResponse)
	err = proto.Unmarshal(recv.Body, resp)
	if err != nil {
		return nil, err
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

func (s *centPeer) getSignatureAsync(ctx context.Context, identityService identity.Service, doc coredocumentpb.CoreDocument, receiverPeer peer.ID, receiverCentID identity.CentID, out chan<- signatureResponseWrap) {
	resp, err := s.getSignatureForDocument(ctx, identityService, doc, receiverPeer, receiverCentID)
	out <- signatureResponseWrap{
		resp: resp,
		err:  err,
	}
}

// GetSignaturesForDocument requests peer nodes for the signature and verifies them
func (s *centPeer) GetSignaturesForDocument(ctx context.Context, identityService identity.Service, doc *coredocumentpb.CoreDocument) error {
	in := make(chan signatureResponseWrap)
	defer close(in)

	self, err := contextutil.Self(ctx)
	if err != nil {
		return centerrors.Wrap(err, "failed to get self")
	}

	extCollaborators, err := coredocument.GetExternalCollaborators(self.ID, doc)
	if err != nil {
		return centerrors.Wrap(err, "failed to get external collaborators")
	}

	var count int
	for _, collaborator := range extCollaborators {
		collaboratorID, err := identity.ToCentID(collaborator)
		if err != nil {
			return centerrors.Wrap(err, "failed to convert to CentID")
		}
		id, err := identityService.LookupIdentityForID(collaboratorID)
		if err != nil {
			return centerrors.Wrap(err, "error fetching collaborator identity")
		}

		receiverPeer, err := s.getPeerID(id)
		if err != nil {
			log.Error(centerrors.Wrap(err, "failed to connect to target"))
			continue
		}

		// for now going with context.background, once we have a timeout for request
		// we can use context.Timeout for that
		count++
		c, _ := context.WithTimeout(ctx, s.config.GetP2PConnectionTimeout())
		go s.getSignatureAsync(c, identityService, *doc, receiverPeer, collaboratorID, in)
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

func (s *centPeer) createSignatureRequest(senderID []byte, doc *coredocumentpb.CoreDocument) (*protocolpb.P2PEnvelope, error) {
	h := p2ppb.CentrifugeHeader{
		NetworkIdentifier:  s.config.GetNetworkID(),
		CentNodeVersion:    version.GetVersion().String(),
		SenderCentrifugeId: senderID,
	}
	req := &p2ppb.SignatureRequest{
		Header:   &h,
		Document: doc,
	}

	reqB, err := proto.Marshal(req)
	if err != nil {
		return nil, err
	}
	return &protocolpb.P2PEnvelope{Type: protocolpb.MessageType_MESSAGE_TYPE_REQUEST_SIGNATURE, Body: reqB}, nil
}

func convertClientError(recv *protocolpb.P2PEnvelope) error {
	resp := new(errorspb.Error)
	err := proto.Unmarshal(recv.Body, resp)
	if err != nil {
		return err
	}
	return errors.New(resp.Message)
}

package p2p

import (
	"context"
	"fmt"

	"github.com/centrifuge/go-centrifuge/p2p/common"

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
	libp2pPeer "github.com/libp2p/go-libp2p-peer"
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

func (s *peer) SendAnchoredDocument(ctx context.Context, id identity.Identity, in *p2ppb.AnchorDocumentRequest) (*p2ppb.AnchorDocumentResponse, error) {
	nc, err := s.config.GetConfig()
	if err != nil {
		return nil, err
	}

	peerCtx, _ := context.WithTimeout(ctx, nc.GetP2PConnectionTimeout())
	cid := id.CentID()
	tc, err := s.config.GetTenant(cid[:])
	if err == nil {
		// this is a local tenant
		h := s.handlerCreator()
		// the following context has to be different from the parent context since its initiating a local peer call
		localCtx, err := contextutil.NewCentrifugeContext(peerCtx, tc)
		if err != nil {
			return nil, err
		}
		return h.SendAnchoredDocument(localCtx, in, cid[:])
	}

	// this is a remote tenant
	pid, err := s.getPeerID(id)
	if err != nil {
		return nil, err
	}

	envelope, err := p2pcommon.PrepareP2PEnvelope(ctx, nc.NetworkID, p2pcommon.MessageTypeSendAnchoredDoc, in)
	if err != nil {
		return nil, err
	}

	recv, err := s.mes.sendMessage(
		ctx, pid,
		envelope,
		p2pcommon.ProtocolForCID(id.CentID()))
	if err != nil {
		return nil, err
	}

	recvEnvelope, err := p2pcommon.ResolveDataEnvelope(recv)
	if err != nil {
		return nil, err
	}

	// handle client error
	if p2pcommon.MessageTypeError.Equals(recvEnvelope.Header.Type) {
		return nil, convertClientError(recvEnvelope)
	}

	if !p2pcommon.MessageTypeSendAnchoredDocRep.Equals(recvEnvelope.Header.Type) {
		return nil, errors.New("the received send anchored document response is incorrect")
	}

	r := new(p2ppb.AnchorDocumentResponse)
	err = proto.Unmarshal(recvEnvelope.Body, r)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// OpenClient returns P2PServiceClient to contact the remote peer
func (s *peer) getPeerID(id identity.Identity) (libp2pPeer.ID, error) {
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

	peerID, err := libp2pPeer.IDB58Decode(pid)
	if err != nil {
		return "", err
	}

	if !s.disablePeerStore {
		// Decapsulate the /ipfs/<peerID> part from the target
		// /ip4/<a.b.c.d>/ipfs/<peer> becomes /ip4/<a.b.c.d>
		targetPeerAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/ipfs/%s", pid))
		targetAddr := ipfsAddr.Decapsulate(targetPeerAddr)
		// We have a peer ID and a targetAddr so we add it to the peer store
		// so LibP2P knows how to contact it
		s.host.Peerstore().AddAddr(peerID, targetAddr, pstore.PermanentAddrTTL)
	}

	return peerID, nil
}

// getSignatureForDocument requests the target node to sign the document
func (s *peer) getSignatureForDocument(ctx context.Context, identityService identity.Service, doc coredocumentpb.CoreDocument, receiverCentID identity.CentID) (*p2ppb.SignatureResponse, error) {
	nc, err := s.config.GetConfig()
	if err != nil {
		return nil, err
	}

	var resp *p2ppb.SignatureResponse
	var header *p2ppb.Header

	tc, err := s.config.GetTenant(receiverCentID[:])
	if err == nil {
		// this is a local tenant
		h := s.handlerCreator()
		// create a context with receiving tenant value
		localPeerCtx, err := contextutil.NewCentrifugeContext(ctx, tc)
		if err != nil {
			return nil, err
		}

		resp, err = h.RequestDocumentSignature(localPeerCtx, &p2ppb.SignatureRequest{Document: &doc})
		if err != nil {
			return nil, err
		}
		header = &p2ppb.Header{NodeVersion: version.GetVersion().String()}
	} else {
		// this is a remote tenant
		id, err := identityService.LookupIdentityForID(receiverCentID)
		if err != nil {
			return nil, err
		}

		receiverPeer, err := s.getPeerID(id)
		if err != nil {
			return nil, err
		}
		envelope, err := p2pcommon.PrepareP2PEnvelope(ctx, nc.NetworkID, p2pcommon.MessageTypeRequestSignature, &p2ppb.SignatureRequest{Document: &doc})
		if err != nil {
			return nil, err
		}
		log.Infof("Requesting signature from %s\n", receiverPeer)
		recv, err := s.mes.sendMessage(ctx, receiverPeer, envelope, p2pcommon.ProtocolForCID(receiverCentID))
		if err != nil {
			return nil, err
		}
		recvEnvelope, err := p2pcommon.ResolveDataEnvelope(recv)
		if err != nil {
			return nil, err
		}
		// handle client error
		if p2pcommon.MessageTypeError.Equals(recvEnvelope.Header.Type) {
			return nil, convertClientError(recvEnvelope)
		}
		if !p2pcommon.MessageTypeRequestSignatureRep.Equals(recvEnvelope.Header.Type) {
			return nil, errors.New("the received request signature response is incorrect")
		}
		resp = new(p2ppb.SignatureResponse)
		err = proto.Unmarshal(recvEnvelope.Body, resp)
		if err != nil {
			return nil, err
		}
		header = recvEnvelope.Header
	}

	err = validateSignatureResp(identityService, receiverCentID, &doc, header, resp)
	if err != nil {
		return nil, err
	}

	log.Infof("Signature successfully received from %s\n", receiverCentID)
	return resp, nil
}

type signatureResponseWrap struct {
	resp *p2ppb.SignatureResponse
	err  error
}

func (s *peer) getSignatureAsync(ctx context.Context, identityService identity.Service, doc *coredocumentpb.CoreDocument, receiverCentID identity.CentID, out chan<- signatureResponseWrap) {
	resp, err := s.getSignatureForDocument(ctx, identityService, *doc, receiverCentID)
	out <- signatureResponseWrap{
		resp: resp,
		err:  err,
	}
}

// GetSignaturesForDocument requests peer nodes for the signature and verifies them
func (s *peer) GetSignaturesForDocument(ctx context.Context, identityService identity.Service, doc *coredocumentpb.CoreDocument) error {
	in := make(chan signatureResponseWrap)
	defer close(in)

	nc, err := s.config.GetConfig()
	if err != nil {
		return err
	}

	self, err := contextutil.Self(ctx)
	if err != nil {
		return centerrors.Wrap(err, "failed to get self")
	}

	extCollaborators, err := coredocument.GetExternalCollaborators(self.ID, doc)
	if err != nil {
		return centerrors.Wrap(err, "failed to get external collaborators")
	}

	var count int
	peerCtx, _ := context.WithTimeout(ctx, nc.GetP2PConnectionTimeout())
	for _, collaborator := range extCollaborators {
		collaboratorID, err := identity.ToCentID(collaborator)
		if err != nil {
			return centerrors.Wrap(err, "failed to convert to CentID")
		}
		count++
		go s.getSignatureAsync(peerCtx, identityService, doc, collaboratorID, in)
	}

	var responses []signatureResponseWrap
	for i := 0; i < count; i++ {
		responses = append(responses, <-in)
	}

	for _, resp := range responses {
		if resp.err != nil {
			// this error is ignored since we would still anchor the document
			log.Error(resp.err)
			continue
		}

		doc.Signatures = append(doc.Signatures, resp.resp.Signature)
	}

	return nil
}

func convertClientError(recv *p2ppb.Envelope) error {
	resp := new(errorspb.Error)
	err := proto.Unmarshal(recv.Body, resp)
	if err != nil {
		return err
	}
	return errors.New(resp.Message)
}

func validateSignatureResp(identityService identity.Service, receiver identity.CentID, doc *coredocumentpb.CoreDocument, header *p2ppb.Header, resp *p2ppb.SignatureResponse) error {
	compatible := version.CheckVersion(header.NodeVersion)
	if !compatible {
		return version.IncompatibleVersionError(header.NodeVersion)
	}

	err := identity.ValidateCentrifugeIDBytes(resp.Signature.EntityId, receiver)
	if err != nil {
		return centerrors.New(code.AuthenticationFailed, err.Error())
	}

	err = identityService.ValidateSignature(resp.Signature, doc.SigningRoot)
	if err != nil {
		return centerrors.New(code.AuthenticationFailed, "signature invalid")
	}
	return nil
}

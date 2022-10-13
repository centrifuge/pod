package p2p

import (
	"context"
	"fmt"
	"time"

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	p2ppb "github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	keystoreType "github.com/centrifuge/chain-custom-types/pkg/keystore"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/crypto/ed25519"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	p2pcommon "github.com/centrifuge/go-centrifuge/p2p/common"
	"github.com/centrifuge/go-centrifuge/version"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/golang/protobuf/proto"
	libp2pPeer "github.com/libp2p/go-libp2p-core/peer"
	pstore "github.com/libp2p/go-libp2p-core/peerstore"
	ma "github.com/multiformats/go-multiaddr"
)

func (s *peer) SendAnchoredDocument(ctx context.Context, receiverID *types.AccountID, in *p2ppb.AnchorDocumentRequest) (*p2ppb.AnchorDocumentResponse, error) {
	nc, err := s.config.GetConfig()
	if err != nil {
		return nil, err
	}

	selfIdentity, err := contextutil.Identity(ctx)
	if err != nil {
		return nil, err
	}

	peerCtx, cancel := context.WithTimeout(ctx, nc.GetP2PConnectionTimeout())
	defer cancel()

	tc, err := s.config.GetAccount(receiverID[:])
	if err == nil {
		// this is a local account
		h := s.handlerCreator()
		// the following context has to be different from the parent context since its initiating a local peer call
		localCtx := contextutil.WithAccount(peerCtx, tc)
		return h.SendAnchoredDocument(localCtx, in, selfIdentity)
	}

	err = s.idService.ValidateAccount(ctx, receiverID)
	if err != nil {
		return nil, err
	}

	// this is a remote account
	pid, err := s.getPeerID(ctx, receiverID)
	if err != nil {
		return nil, err
	}

	envelope, err := p2pcommon.PrepareP2PEnvelope(ctx, nc.GetNetworkID(), p2pcommon.MessageTypeSendAnchoredDoc, in)
	if err != nil {
		return nil, err
	}

	recv, err := s.mes.SendMessage(
		ctx, pid,
		envelope,
		p2pcommon.ProtocolForIdentity(receiverID))
	if err != nil {
		return nil, err
	}

	recvEnvelope, err := p2pcommon.ResolveDataEnvelope(recv)
	if err != nil {
		return nil, err
	}

	// handle client error
	if p2pcommon.MessageTypeError.Equals(recvEnvelope.Header.Type) {
		return nil, p2pcommon.ConvertClientError(recvEnvelope)
	}

	if !p2pcommon.MessageTypeSendAnchoredDocRep.Equals(recvEnvelope.Header.Type) {
		return nil, errors.New("the received getDocument response is incorrect")
	}

	r := new(p2ppb.AnchorDocumentResponse)
	err = proto.Unmarshal(recvEnvelope.Body, r)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (s *peer) GetDocumentRequest(ctx context.Context, requesterID *types.AccountID, in *p2ppb.GetDocumentRequest) (*p2ppb.GetDocumentResponse, error) {
	nc, err := s.config.GetConfig()
	if err != nil {
		return nil, err
	}

	sender, err := contextutil.Identity(ctx)
	if err != nil {
		return nil, err
	}

	peerCtx, cancel := context.WithTimeout(ctx, nc.GetP2PConnectionTimeout())
	defer cancel()

	acc, err := s.config.GetAccount(requesterID.ToBytes())
	if err == nil {
		// this is a local account
		h := s.handlerCreator()
		// the following context has to be different from the parent context since its initiating a local peer call
		localCtx := contextutil.WithAccount(peerCtx, acc)

		return h.GetDocument(localCtx, in, sender)
	}

	err = s.idService.ValidateAccount(ctx, requesterID)
	if err != nil {
		return nil, err
	}

	// this is a remote account
	pid, err := s.getPeerID(ctx, requesterID)
	if err != nil {
		return nil, err
	}

	envelope, err := p2pcommon.PrepareP2PEnvelope(ctx, nc.GetNetworkID(), p2pcommon.MessageTypeGetDoc, in)
	if err != nil {
		return nil, err
	}

	recv, err := s.mes.SendMessage(
		ctx, pid,
		envelope,
		p2pcommon.ProtocolForIdentity(requesterID))
	if err != nil {
		return nil, err
	}

	recvEnvelope, err := p2pcommon.ResolveDataEnvelope(recv)
	if err != nil {
		return nil, err
	}

	// handle client error
	if p2pcommon.MessageTypeError.Equals(recvEnvelope.Header.Type) {
		return nil, p2pcommon.ConvertClientError(recvEnvelope)
	}

	if !p2pcommon.MessageTypeGetDocRep.Equals(recvEnvelope.Header.Type) {
		return nil, errors.New("the received get document response is incorrect")
	}

	r := new(p2ppb.GetDocumentResponse)
	err = proto.Unmarshal(recvEnvelope.Body, r)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// getPeerID returns peerID to contact the remote peer
func (s *peer) getPeerID(ctx context.Context, accountID *types.AccountID) (libp2pPeer.ID, error) {
	lastp2pKey, err := s.idService.GetLastKeyByPurpose(ctx, accountID, keystoreType.KeyPurposeP2PDiscovery)
	if err != nil {
		return "", errors.New("error fetching p2p key: %v", err)
	}

	if err = s.idService.ValidateKey(ctx, accountID, lastp2pKey[:], keystoreType.KeyPurposeP2PDiscovery, time.Now()); err != nil {
		return "", errors.New("invalid p2p key: %s", err)
	}

	p2pID, err := ed25519.PublicKeyToP2PKey(*lastp2pKey)

	if err != nil {
		return "", fmt.Errorf("couldn't create p2p key: %w", err)
	}

	target := fmt.Sprintf("/ipfs/%s", p2pID.Pretty())

	log.Infof("Opening connection to: %s\n", target)

	ipfsAddr, err := ma.NewMultiaddr(target)
	if err != nil {
		return "", err
	}

	pid, err := ipfsAddr.ValueForProtocol(ma.P_IPFS)
	if err != nil {
		return "", err
	}

	peerID, err := libp2pPeer.Decode(pid)
	if err != nil {
		return "", err
	}

	if !s.disablePeerStore {
		nc, err := s.config.GetConfig()
		if err != nil {
			return peerID, err
		}
		c, canc := context.WithTimeout(ctx, nc.GetP2PConnectionTimeout())
		defer canc()
		pinfo, err := s.dht.FindPeer(c, peerID)
		if err != nil {
			return peerID, err
		}

		// We have a peer ID and a targetAddr so we add it to the peer store
		// so LibP2P knows how to contact it (this call might be redundant)
		s.host.Peerstore().AddAddrs(peerID, pinfo.Addrs, pstore.PermanentAddrTTL)
	}

	return peerID, nil
}

// getSignatureForDocument requests the target node to sign the document
func (s *peer) getSignatureForDocument(ctx context.Context, model documents.Document, collaborator, sender *types.AccountID) (*p2ppb.SignatureResponse, error) {
	nc, err := s.config.GetConfig()
	if err != nil {
		return nil, err
	}
	cd, err := model.PackCoreDocument()
	if err != nil {
		return nil, errors.New("failed to pack core document: %v", err)
	}
	var resp *p2ppb.SignatureResponse
	var header *p2ppb.Header
	tc, err := s.config.GetAccount(collaborator[:])
	if err == nil {
		// this is a local account
		h := s.handlerCreator()
		// create a context with receiving account value
		localPeerCtx := contextutil.WithAccount(ctx, tc)

		resp, err = h.RequestDocumentSignature(localPeerCtx, &p2ppb.SignatureRequest{Document: cd}, sender)
		if err != nil {
			return nil, err
		}
		header = &p2ppb.Header{NodeVersion: version.GetVersion().String()}
	} else {
		// this is a remote account
		err = s.idService.ValidateAccount(ctx, collaborator)
		if err != nil {
			return nil, err
		}

		receiverPeer, err := s.getPeerID(ctx, collaborator)
		if err != nil {
			return nil, err
		}
		envelope, err := p2pcommon.PrepareP2PEnvelope(ctx, nc.GetNetworkID(), p2pcommon.MessageTypeRequestSignature, &p2ppb.SignatureRequest{Document: cd})
		if err != nil {
			return nil, err
		}
		log.Infof("Requesting signature from %s\n", receiverPeer)
		recv, err := s.mes.SendMessage(ctx, receiverPeer, envelope, p2pcommon.ProtocolForIdentity(collaborator))
		if err != nil {
			return nil, err
		}
		recvEnvelope, err := p2pcommon.ResolveDataEnvelope(recv)
		if err != nil {
			return nil, err
		}
		// handle client error
		if p2pcommon.MessageTypeError.Equals(recvEnvelope.Header.Type) {
			return nil, p2pcommon.ConvertClientError(recvEnvelope)
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

	err = s.validateSignatureResp(model, collaborator, header, resp)
	if err != nil {
		return nil, err
	}

	log.Infof("Signature successfully received from %s\n", collaborator.ToHexString())
	return resp, nil
}

type signatureResponseWrap struct {
	resp *p2ppb.SignatureResponse
	err  error
}

func (s *peer) getSignatureAsync(ctx context.Context, model documents.Document, collaborator, sender *types.AccountID, out chan<- signatureResponseWrap) {
	resp, err := s.getSignatureForDocument(ctx, model, collaborator, sender)
	out <- signatureResponseWrap{
		resp: resp,
		err:  err,
	}
}

// GetSignaturesForDocument requests peer nodes for the signature, verifies them, and returns those signatures.
func (s *peer) GetSignaturesForDocument(ctx context.Context, model documents.Document) (signatures []*coredocumentpb.Signature, signatureCollectionErrors []error, err error) {
	in := make(chan signatureResponseWrap)
	defer close(in)

	nc, err := s.config.GetConfig()
	if err != nil {
		return nil, nil, err
	}

	selfIdentity, err := contextutil.Identity(ctx)
	if err != nil {
		return nil, nil, errors.New("failed to get self ID")
	}

	cs, err := model.GetSignerCollaborators(selfIdentity)
	if err != nil {
		return nil, nil, errors.New("failed to get external collaborators")
	}

	var count int
	peerCtx, cancel := context.WithTimeout(ctx, nc.GetP2PConnectionTimeout())
	defer cancel()
	for _, c := range cs {
		count++
		go s.getSignatureAsync(peerCtx, model, c, selfIdentity, in)
	}

	var responses []signatureResponseWrap
	for i := 0; i < count; i++ {
		responses = append(responses, <-in)
	}

	for _, resp := range responses {
		if resp.err != nil {
			signatureCollectionErrors = append(signatureCollectionErrors, resp.err)
			continue
		}

		signatures = append(signatures, resp.resp.Signatures...)
	}

	return signatures, signatureCollectionErrors, nil
}

func (s *peer) validateSignatureResp(
	model documents.Document,
	receiver *types.AccountID,
	header *p2ppb.Header,
	resp *p2ppb.SignatureResponse,
) error {
	compatible := version.CheckVersion(header.NodeVersion)
	if !compatible {
		return version.IncompatibleVersionError(header.NodeVersion)
	}

	signingRoot, err := model.CalculateSigningRoot()
	if err != nil {
		return errors.New("failed to calculate signing root: %s", err.Error())
	}

	for _, sig := range resp.Signatures {
		signerAccountID, err := types.NewAccountID(sig.SignerId)
		if err != nil {
			return errors.New("invalid signer account ID")
		}

		if !signerAccountID.Equal(receiver) {
			return errors.New("invalid signature")
		}

		// TODO(cdamian): Get a proper context here
		ctx := context.Background()

		timestamp, err := model.Timestamp()

		if err != nil {
			return errors.New("couldn't retrieve document timestamp: %s", err)
		}

		err = s.idService.ValidateSignature(
			ctx,
			receiver,
			sig.PublicKey,
			sig.Signature,
			documents.ConsensusSignaturePayload(signingRoot, sig.TransitionValidated),
			timestamp,
		)

		if err != nil {
			return errors.New("signature invalid with err: %s", err)
		}
	}

	return nil
}

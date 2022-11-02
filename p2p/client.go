package p2p

import (
	"context"
	"fmt"
	"sync"
	"time"

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	p2ppb "github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	keystoreType "github.com/centrifuge/chain-custom-types/pkg/keystore"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	p2pcommon "github.com/centrifuge/go-centrifuge/p2p/common"
	"github.com/centrifuge/go-centrifuge/version"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/golang/protobuf/proto"
	libp2pPeer "github.com/libp2p/go-libp2p-core/peer"
	pstore "github.com/libp2p/go-libp2p-core/peerstore"
)

func (s *p2pPeer) SendAnchoredDocument(ctx context.Context, receiverID *types.AccountID, req *p2ppb.AnchorDocumentRequest) (*p2ppb.AnchorDocumentResponse, error) {
	sender, err := contextutil.Identity(ctx)
	if err != nil {
		log.Errorf("Couldn't get sender identity: %s", err)

		return nil, errors.ErrContextIdentityRetrieval
	}

	acc, err := s.cfgService.GetAccount(receiverID.ToBytes())

	if err == nil { // this is a local account
		peerCtx, cancel := context.WithTimeout(ctx, s.config.GetP2PConnectionTimeout())
		defer cancel()

		localCtx := contextutil.WithAccount(peerCtx, acc)

		// the following context has to be different from the parent context since its initiating a local peer call
		return s.handler.SendAnchoredDocument(localCtx, req, sender)
	}

	err = s.idService.ValidateAccount(receiverID)
	if err != nil {
		log.Errorf("Couldn't validate receiver account: %s", err)

		return nil, ErrInvalidReceiverAccount
	}

	// this is a remote account
	pid, err := s.getPeerID(ctx, receiverID)
	if err != nil {
		log.Errorf("Couldn't retrieve peer ID: %s", err)

		return nil, ErrPeerIDRetrieval
	}

	envelope, err := p2pcommon.PrepareP2PEnvelope(ctx, s.config.GetNetworkID(), p2pcommon.MessageTypeSendAnchoredDoc, req)
	if err != nil {
		log.Errorf("Couldn't prepare P2P envelope: %s", err)

		return nil, ErrP2PEnvelopePreparation
	}

	recv, err := s.mes.SendMessage(
		ctx,
		pid,
		envelope,
		p2pcommon.ProtocolForIdentity(receiverID),
	)

	if err != nil {
		log.Errorf("Couldn't send P2P message: %s", err)

		return nil, ErrP2PMessageSending
	}

	recvEnvelope, err := p2pcommon.ResolveDataEnvelope(recv)
	if err != nil {
		log.Errorf("Couldn't resolve data envelope: %s", err)

		return nil, ErrP2PDataEnvelopeResolving
	}

	// handle client error
	if p2pcommon.MessageTypeError.Equals(recvEnvelope.Header.Type) {
		clientErr := p2pcommon.ConvertClientError(recvEnvelope)

		log.Errorf("P2P client error: %s", clientErr)

		return nil, errors.NewTypedError(ErrP2PClient, clientErr)
	}

	if !p2pcommon.MessageTypeSendAnchoredDocRep.Equals(recvEnvelope.Header.Type) {
		log.Error("Incorrect response message type")

		return nil, ErrIncorrectResponseMessageType
	}

	r := new(p2ppb.AnchorDocumentResponse)
	err = proto.Unmarshal(recvEnvelope.Body, r)
	if err != nil {
		log.Errorf("Couldn't decode response: %s", err)

		return nil, ErrResponseDecodeError
	}

	return r, nil
}

func (s *p2pPeer) GetDocumentRequest(ctx context.Context, documentOwner *types.AccountID, req *p2ppb.GetDocumentRequest) (*p2ppb.GetDocumentResponse, error) {
	sender, err := contextutil.Identity(ctx)
	if err != nil {
		log.Errorf("Couldn't get sender identity: %s", err)

		return nil, errors.ErrContextIdentityRetrieval
	}

	acc, err := s.cfgService.GetAccount(documentOwner.ToBytes())
	if err == nil { // this is a local account
		peerCtx, cancel := context.WithTimeout(ctx, s.config.GetP2PConnectionTimeout())
		defer cancel()

		localCtx := contextutil.WithAccount(peerCtx, acc)

		return s.handler.GetDocument(localCtx, req, sender)
	}

	err = s.idService.ValidateAccount(documentOwner)
	if err != nil {
		log.Errorf("Couldn't validate requester account: %s", err)

		return nil, ErrInvalidRequesterAccount
	}

	// this is a remote account
	pid, err := s.getPeerID(ctx, documentOwner)
	if err != nil {
		log.Errorf("Couldn't retrieve peer ID: %s", err)

		return nil, ErrPeerIDRetrieval
	}

	envelope, err := p2pcommon.PrepareP2PEnvelope(ctx, s.config.GetNetworkID(), p2pcommon.MessageTypeGetDoc, req)
	if err != nil {
		log.Errorf("Couldn't prepare P2P envelope: %s", err)

		return nil, ErrP2PEnvelopePreparation
	}

	recv, err := s.mes.SendMessage(
		ctx,
		pid,
		envelope,
		p2pcommon.ProtocolForIdentity(documentOwner),
	)

	if err != nil {
		log.Errorf("Couldn't send P2P message: %s", err)

		return nil, ErrP2PMessageSending
	}

	recvEnvelope, err := p2pcommon.ResolveDataEnvelope(recv)
	if err != nil {
		log.Errorf("Couldn't resolve data envelope: %s", err)

		return nil, ErrP2PDataEnvelopeResolving
	}

	// handle client error
	if p2pcommon.MessageTypeError.Equals(recvEnvelope.Header.Type) {
		clientErr := p2pcommon.ConvertClientError(recvEnvelope)

		log.Errorf("P2P client error: %s", clientErr)

		return nil, errors.NewTypedError(ErrP2PClient, clientErr)
	}

	if !p2pcommon.MessageTypeGetDocRep.Equals(recvEnvelope.Header.Type) {
		log.Error("Incorrect response message type")

		return nil, ErrIncorrectResponseMessageType
	}

	r := new(p2ppb.GetDocumentResponse)
	err = proto.Unmarshal(recvEnvelope.Body, r)
	if err != nil {
		log.Errorf("Couldn't decode response: %s", err)

		return nil, ErrResponseDecodeError
	}

	return r, nil
}

// GetSignaturesForDocument requests peer nodes for the signature, verifies them, and returns those signatures.
func (s *p2pPeer) GetSignaturesForDocument(ctx context.Context, model documents.Document) (signatures []*coredocumentpb.Signature, signatureCollectionErrors []error, err error) {
	sender, err := contextutil.Identity(ctx)
	if err != nil {
		log.Errorf("Couldn't get sender identity: %s", err)

		return nil, nil, errors.ErrContextIdentityRetrieval
	}

	signerCollaborators, err := model.GetSignerCollaborators(sender)
	if err != nil {
		log.Errorf("Couldn't get signer collaborators: %s", err)

		return nil, nil, ErrSignerCollaboratorsRetrieval
	}

	peerCtx, cancel := context.WithTimeout(ctx, s.config.GetP2PConnectionTimeout())
	defer cancel()

	var wg sync.WaitGroup

	signatureWrapChan := make(chan signatureResponseWrap, len(signerCollaborators))

	for _, collaborator := range signerCollaborators {
		collaborator := collaborator

		wg.Add(1)

		go func() {
			defer wg.Done()

			resp, err := s.getSignatureForDocument(peerCtx, model, collaborator, sender)

			signatureWrapChan <- signatureResponseWrap{
				resp: resp,
				err:  err,
			}
		}()
	}

	go func() {
		wg.Wait()
		close(signatureWrapChan)
	}()

	for {
		select {
		case <-ctx.Done():
			return nil, nil, fmt.Errorf("context done while collecting signatures: %w", ctx.Err())
		case <-peerCtx.Done():
			return nil, nil, fmt.Errorf("peer context done while collecting signatures: %w", peerCtx.Err())
		case res, ok := <-signatureWrapChan:
			if !ok {
				return signatures, signatureCollectionErrors, nil
			}

			if res.err != nil {
				signatureCollectionErrors = append(signatureCollectionErrors, res.err)
				continue
			}

			signatures = append(signatures, res.resp.Signatures...)
		}
	}
}

// getPeerID returns peerID to contact the remote peer
func (s *p2pPeer) getPeerID(ctx context.Context, accountID *types.AccountID) (libp2pPeer.ID, error) {
	p2pKey, err := s.keystoreAPI.GetLastKeyByPurpose(accountID, keystoreType.KeyPurposeP2PDiscovery)
	if err != nil {
		log.Errorf("Couldn't get P2P key: %s", err)

		return "", ErrP2PKeyRetrievalError
	}

	if err = s.idService.ValidateKey(accountID, p2pKey[:], keystoreType.KeyPurposeP2PDiscovery, time.Now()); err != nil {
		log.Errorf("Invalid P2P key: %s", err)

		return "", ErrInvalidP2PKey
	}

	peerID, err := p2pcommon.ParsePeerID(*p2pKey)
	if err != nil {
		log.Errorf("Couldn't parse peer ID: %s", err)

		return "", ErrPeerIDParsing
	}

	if !s.disablePeerStore {
		ctx, canc := context.WithTimeout(ctx, s.config.GetP2PConnectionTimeout())
		defer canc()

		pinfo, err := s.dht.FindPeer(ctx, peerID)
		if err != nil {
			log.Errorf("Couldn't find peer: %s", err)

			return "", ErrPeerNotFound
		}

		// We have a peer ID and a targetAddr so we add it to the peer store
		// so LibP2P knows how to contact it (this call might be redundant)
		s.host.Peerstore().AddAddrs(peerID, pinfo.Addrs, pstore.PermanentAddrTTL)
	}

	return peerID, nil
}

// getSignatureForDocument requests the target node to sign the document
func (s *p2pPeer) getSignatureForDocument(ctx context.Context, model documents.Document, collaborator, sender *types.AccountID) (*p2ppb.SignatureResponse, error) {
	cd, err := model.PackCoreDocument()
	if err != nil {
		log.Errorf("Couldn't pack core document: %s", err)

		return nil, ErrCoreDocumentPacking
	}

	acc, err := s.cfgService.GetAccount(collaborator.ToBytes())
	if err == nil { // this is a local account
		// create a context with receiving account value
		localPeerCtx := contextutil.WithAccount(ctx, acc)

		resp, err := s.handler.RequestDocumentSignature(localPeerCtx, &p2ppb.SignatureRequest{Document: cd}, sender)
		if err != nil {
			log.Errorf("Couldn't request document signature: %s", err)

			return nil, ErrDocumentSignatureRequest
		}

		header := &p2ppb.Header{NodeVersion: version.GetVersion().String()}

		err = s.validateSignatureResp(model, collaborator, header, resp)
		if err != nil {
			log.Errorf("Couldn't validate signature response: %s", err)

			return nil, ErrInvalidSignatureResponse
		}

		return resp, nil
	}

	// this is a remote account
	err = s.idService.ValidateAccount(collaborator)
	if err != nil {
		log.Errorf("Invalid collaborator account: %s", err)

		return nil, ErrInvalidCollaboratorAccount
	}

	receiverPeer, err := s.getPeerID(ctx, collaborator)
	if err != nil {
		log.Errorf("Couldn't retrieve peer ID: %s", err)

		return nil, ErrPeerIDRetrieval
	}
	envelope, err := p2pcommon.PrepareP2PEnvelope(ctx, s.config.GetNetworkID(), p2pcommon.MessageTypeRequestSignature, &p2ppb.SignatureRequest{Document: cd})
	if err != nil {
		log.Errorf("Couldn't prepare P2P envelope: %s", err)

		return nil, ErrP2PEnvelopePreparation
	}

	log.Infof("Requesting signature from %s\n", receiverPeer)

	recv, err := s.mes.SendMessage(ctx, receiverPeer, envelope, p2pcommon.ProtocolForIdentity(collaborator))

	if err != nil {
		log.Errorf("Couldn't send P2P message: %s", err)

		return nil, ErrP2PMessageSending
	}

	recvEnvelope, err := p2pcommon.ResolveDataEnvelope(recv)
	if err != nil {
		log.Errorf("Couldn't resolve data envelope: %s", err)

		return nil, ErrP2PDataEnvelopeResolving
	}

	// handle client error
	if p2pcommon.MessageTypeError.Equals(recvEnvelope.Header.Type) {
		clientErr := p2pcommon.ConvertClientError(recvEnvelope)

		log.Errorf("P2P client error: %s", clientErr)

		return nil, errors.NewTypedError(ErrP2PClient, clientErr)
	}

	if !p2pcommon.MessageTypeRequestSignatureRep.Equals(recvEnvelope.Header.Type) {
		log.Error("Incorrect response message type")

		return nil, ErrIncorrectResponseMessageType
	}

	resp := new(p2ppb.SignatureResponse)

	if err = proto.Unmarshal(recvEnvelope.Body, resp); err != nil {
		log.Errorf("Couldn't decode response: %s", err)

		return nil, ErrResponseDecodeError
	}

	header := recvEnvelope.Header

	err = s.validateSignatureResp(model, collaborator, header, resp)
	if err != nil {
		log.Errorf("Couldn't validate signature response: %s", err)

		return nil, ErrInvalidSignatureResponse
	}

	log.Infof("Signature successfully received from %s\n", collaborator.ToHexString())

	return resp, nil
}

type signatureResponseWrap struct {
	resp *p2ppb.SignatureResponse
	err  error
}

func (s *p2pPeer) validateSignatureResp(
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

	// TODO(cdamian): Do we need an extra check to ensure that signatures are not empty?
	for _, sig := range resp.Signatures {
		signerAccountID, err := types.NewAccountID(sig.SignerId)
		if err != nil {
			return errors.New("invalid signer account ID")
		}

		if !signerAccountID.Equal(receiver) {
			return errors.New("invalid signature")
		}

		timestamp, err := model.Timestamp()

		if err != nil {
			return errors.New("couldn't retrieve document timestamp: %s", err)
		}

		err = s.idService.ValidateSignature(
			signerAccountID,
			sig.PublicKey,
			documents.ConsensusSignaturePayload(signingRoot, sig.TransitionValidated),
			sig.Signature,
			timestamp,
		)

		if err != nil {
			return errors.New("signature invalid with err: %s", err)
		}
	}

	return nil
}

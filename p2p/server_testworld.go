// +build testworld

package p2p

import (
	"context"

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	p2ppb "github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	protocolpb "github.com/centrifuge/centrifuge-protobufs/gen/go/protocol"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	p2pcommon "github.com/centrifuge/go-centrifuge/p2p/common"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/golang/protobuf/proto"
)

type MessageType string

// AccessPeer allow accessing the peer within a client
func AccessPeer(client documents.Client) *peer {
	if p, ok := client.(*peer); ok {
		return p
	} else {
		return nil
	}
}

//client actions for malicious host

// getSignatureForDocument requests the target node to sign the document
func (s *peer) getSignatureForDocumentIncorrectMessage(ctx context.Context, model documents.Model, collaborator, sender identity.DID, errorType string) (*p2ppb.SignatureResponse, error) {
	nc, err := s.config.GetConfig()
	if err != nil {
		return nil, err
	}

	var resp *p2ppb.SignatureResponse
	var header *p2ppb.Header

	err = s.idService.Exists(ctx, collaborator)
	if err != nil {
		return nil, err
	}
	receiverPeer, err := s.getPeerID(ctx, collaborator)
	if err != nil {
		return nil, err
	}

	cd, err := model.PackCoreDocument()
	if err != nil {
		return nil, errors.New("failed to pack core document: %v", err)
	}

	//select which envelope preparing function to call based on the error type
	var envelope *protocolpb.P2PEnvelope
	var envelopeErr error
	switch errorType {
	case "incorrectNodeVersion":
		envelope, envelopeErr = p2pcommon.PrepareP2PEnvelopeIncorrectNodeVersion(ctx, nc.GetNetworkID(), p2pcommon.MessageTypeRequestSignature, &p2ppb.SignatureRequest{Document: &cd})
		if envelopeErr != nil {
			return nil, envelopeErr
		}
	case "invalidBody":
		envelope, envelopeErr = p2pcommon.PrepareP2PEnvelopeInvalidBody(ctx, nc.GetNetworkID(), p2pcommon.MessageTypeRequestSignature, &p2ppb.SignatureRequest{Document: &cd})
		if envelopeErr != nil {
			return nil, envelopeErr
		}
	default:
		envelope, envelopeErr = p2pcommon.PrepareP2PEnvelopeInvalidHeader(ctx, nc.GetNetworkID(), p2pcommon.MessageTypeRequestSignature, &p2ppb.SignatureRequest{Document: &cd})
		if envelopeErr != nil {
			return nil, envelopeErr
		}
	}
	log.Infof("Requesting signature from %s\n", receiverPeer)
	recv, err := s.mes.SendMessage(ctx, receiverPeer, envelope, p2pcommon.ProtocolForDID(collaborator))
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

	err = s.validateSignatureResp(model, collaborator, header, resp)
	if err != nil {
		return nil, err
	}

	log.Infof("Signature successfully received from %s\n", collaborator)
	return resp, nil
}

func (s *peer) getSignatureAsyncIncorrectMessage(ctx context.Context, model documents.Model, collaborator, sender identity.DID, out chan<- signatureResponseWrap, errorType string) {
	resp, err := s.getSignatureForDocumentIncorrectMessage(ctx, model, collaborator, sender, errorType)
	out <- signatureResponseWrap{
		resp: resp,
		err:  err,
	}
}

// GetSignaturesForDocument requests peer nodes for the signature, verifies them, and returns those signatures. The error type specifies which type of incorrect message should be sent
func (s *peer) GetSignaturesForDocumentIncorrectMessage(ctx context.Context, model documents.Model, errorType string) (signatures []*coredocumentpb.Signature, signatureCollectionErrors []error, err error) {
	in := make(chan signatureResponseWrap)
	defer close(in)

	nc, err := s.config.GetConfig()
	if err != nil {
		return nil, nil, err
	}

	selfDID, err := contextutil.AccountDID(ctx)
	if err != nil {
		return nil, nil, errors.New("failed to get self ID")
	}

	cs, err := model.GetSignerCollaborators(selfDID)
	if err != nil {
		return nil, nil, errors.New("failed to get external collaborators")
	}

	var count int
	peerCtx, cancel := context.WithTimeout(ctx, nc.GetP2PConnectionTimeout())
	defer cancel()
	for _, c := range cs {
		count++
		go s.getSignatureAsyncIncorrectMessage(peerCtx, model, c, selfDID, in, errorType)
	}

	var responses []signatureResponseWrap
	for i := 0; i < count; i++ {
		responses = append(responses, <-in)
	}

	for _, resp := range responses {
		if resp.err != nil {
			log.Warn(resp.err)
			signatureCollectionErrors = append(signatureCollectionErrors, resp.err)
			continue
		}

		signatures = append(signatures, resp.resp.Signatures...)
	}

	return signatures, signatureCollectionErrors, nil
}

//send message over the accepted maximum message size
func (s *peer) SendOverSizedMessage(ctx context.Context, model documents.Model, length int) (envelope *protocolpb.P2PEnvelope, err error) {
	nc, err := s.config.GetConfig()
	if err != nil {
		return nil, err
	}

	selfDID, err := contextutil.AccountDID(ctx)
	if err != nil {
		return nil, err
	}

	cs, err := model.GetSignerCollaborators(selfDID)
	if err != nil {
		return nil, err
	}

	//get the first collaborator only
	collaborator := cs[0]
	err = s.idService.Exists(ctx, collaborator)
	if err != nil {
		return nil, err
	}
	receiverPeer, err := s.getPeerID(ctx, collaborator)
	if err != nil {
		return nil, err
	}

	p2pEnv, err := p2pcommon.PrepareP2PEnvelope(ctx, nc.GetNetworkID(), p2pcommon.MessageTypeRequestSignature, &p2ppb.Envelope{Body: utils.RandomSlice(length)})
	if err != nil {
		return nil, err
	}
	msg, err := s.mes.SendMessage(ctx, receiverPeer, p2pEnv, p2pcommon.ProtocolForDID(collaborator))
	return msg, err
}

package p2pcommon

import (
	"context"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/protocol"
	"github.com/centrifuge/go-centrifuge/version"
	"github.com/golang/protobuf/proto"
)

type MessageType string

const (
	MessageTypeError MessageType = "MessageTypeError"
	MessageTypeInvalid MessageType = "MessageTypeInvalid"
	MessageTypeRequestSignature MessageType = "MessageTypeRequestSignature"
	MessageTypeRequestSignatureRep MessageType = "MessageTypeRequestSignatureRep"
	MessageTypeSendAnchoredDoc MessageType = "MessageTypeSendAnchoredDoc"
	MessageTypeSendAnchoredDocRep MessageType = "MessageTypeSendAnchoredDocRep"
)

func (mt MessageType) Equals(mt2 string) bool {
	return mt.String() == mt2
}

func (mt MessageType) String() string {
	return string(mt)
}

func MessageTypeFromString(mt string) MessageType {
	var found MessageType
	if MessageTypeError.Equals(mt) {
		found = MessageTypeError
	} else if MessageTypeInvalid.Equals(mt) {
		found = MessageTypeInvalid
	} else if MessageTypeRequestSignature.Equals(mt) {
		found = MessageTypeRequestSignature
	} else if MessageTypeRequestSignatureRep.Equals(mt) {
		found = MessageTypeRequestSignatureRep
	} else if MessageTypeSendAnchoredDoc.Equals(mt) {
		found = MessageTypeSendAnchoredDoc
	} else if MessageTypeSendAnchoredDocRep.Equals(mt) {
		found = MessageTypeSendAnchoredDocRep
	}
	return found
}


func ResolveDataEnvelope(mes proto.Message) (*p2ppb.CentrifugeEnvelope, error) {
	recv, ok := mes.(*protocolpb.P2PEnvelope)
	if !ok {
		return nil, errors.New("cannot unmarshall response payload: %v", recv)
	}
	recvEnvelope := new(p2ppb.CentrifugeEnvelope)
	err := proto.Unmarshal(recv.Body, recvEnvelope)
	if err != nil {
		return nil, err
	}

	// Validate at least not-nil fields
	if recvEnvelope.Header == nil {
		return nil, errors.New("Header field is empty")
	}

	return recvEnvelope, nil
}

func PrepareP2PEnvelope(ctx context.Context, networkID uint32, messageType MessageType, mes proto.Message) (*protocolpb.P2PEnvelope, error) {
	self, err := contextutil.Self(ctx)
	if err != nil {
		return nil, err
	}

	centIDBytes := self.ID[:]
	p2pheader := &p2ppb.CentrifugeHeader{
		SenderCentrifugeId: centIDBytes,
		CentNodeVersion:    version.GetVersion().String(),
		NetworkIdentifier:  networkID,
		Type: messageType.String(),
	}

	body, err := proto.Marshal(mes)
	if err != nil {
		return nil, err
	}

	envelope := &p2ppb.CentrifugeEnvelope{
		Header: p2pheader,
		Body: body,
	}

	marshalledRequest, err := proto.Marshal(envelope)
	if err != nil {
		return nil, err
	}

	return &protocolpb.P2PEnvelope{Body: marshalledRequest}, nil
}
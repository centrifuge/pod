package p2pcommon

import (
	"context"
	"fmt"
	"strings"
	"time"

	errorspb "github.com/centrifuge/centrifuge-protobufs/gen/go/errors"
	p2ppb "github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	protocolpb "github.com/centrifuge/centrifuge-protobufs/gen/go/protocol"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/go-centrifuge/version"
	"github.com/golang/protobuf/proto"
	"github.com/libp2p/go-libp2p-core/protocol"
)

// MessageType holds the protocol message type
type MessageType string

const (
	// CentrifugeProtocol is the centrifuge wire protocol
	CentrifugeProtocol protocol.ID = "/centrifuge/0.0.1"

	// MessageTypeError defines any protocol error
	MessageTypeError MessageType = "MessageTypeError"
	// MessageTypeInvalid defines invalid protocol type
	MessageTypeInvalid MessageType = "MessageTypeInvalid"
	// MessageTypeRequestSignature defines RequestSignature type
	MessageTypeRequestSignature MessageType = "MessageTypeRequestSignature"
	// MessageTypeRequestSignatureRep defines RequestSignature response type
	MessageTypeRequestSignatureRep MessageType = "MessageTypeRequestSignatureRep"
	// MessageTypeSendAnchoredDoc defines SendAnchored type
	MessageTypeSendAnchoredDoc MessageType = "MessageTypeSendAnchoredDoc"
	// MessageTypeSendAnchoredDocRep defines SendAnchored response type
	MessageTypeSendAnchoredDocRep MessageType = "MessageTypeSendAnchoredDocRep"
	//MessageTypeGetDoc defines GetAnchoredDoc type
	MessageTypeGetDoc MessageType = "MessageTypeGetDoc"
	//MessageTypeGetDocRep defines GetAnchoredDoc response type
	MessageTypeGetDocRep MessageType = "MessageTypeGetDocRep"
)

//MessageTypes map for MessageTypeFromString function
var messageTypes = map[string]MessageType{
	"MessageTypeError":               "MessageTypeError",
	"MessageTypeInvalid":             "MessageTypeInvalid",
	"MessageTypeRequestSignature":    "MessageTypeRequestSignature",
	"MessageTypeRequestSignatureRep": "MessageTypeRequestSignatureRep",
	"MessageTypeSendAnchoredDoc":     "MessageTypeSendAnchoredDoc",
	"MessageTypeSendAnchoredDocRep":  "MessageTypeSendAnchoredDocRep",
	"MessageTypeGetDoc":              "MessageTypeGetDoc",
	"MessageTypeGetDocRep":           "MessageTypeGetDocRep",
}

// Equals compares if string is of a particular MessageType
func (mt MessageType) Equals(mt2 string) bool {
	return mt.String() == mt2
}

// String representation
func (mt MessageType) String() string {
	return string(mt)
}

// MessageTypeFromString Resolves MessageType out of string
func MessageTypeFromString(ht string) MessageType {
	var messageType MessageType
	if mt, exists := messageTypes[ht]; exists {
		messageType = mt
	}
	return messageType
}

// ProtocolForDID creates the protocol string for the given CID
func ProtocolForDID(did identity.DID) protocol.ID {
	return protocol.ID(fmt.Sprintf("%s/%s", CentrifugeProtocol, did.String()))
}

// ExtractDID extracts DID from a protocol string
func ExtractDID(id protocol.ID) (identity.DID, error) {
	parts := strings.Split(string(id), "/")
	cidHexStr := parts[len(parts)-1]
	return identity.NewDIDFromString(cidHexStr)
}

// ResolveDataEnvelope unwraps Content Envelope out of p2pEnvelope
func ResolveDataEnvelope(mes proto.Message) (*p2ppb.Envelope, error) {
	recv, ok := mes.(*protocolpb.P2PEnvelope)
	if !ok {
		return nil, errors.New("cannot cast proto.Message to protocolpb.P2PEnvelope: %v", recv)
	}
	recvEnvelope := new(p2ppb.Envelope)
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

// PrepareP2PEnvelope wraps content message into p2p envelope
func PrepareP2PEnvelope(ctx context.Context, networkID uint32, messageType MessageType, mes proto.Message) (*protocolpb.P2PEnvelope, error) {
	self, err := contextutil.Account(ctx)
	if err != nil {
		return nil, err
	}

	centIDBytes := self.GetIdentityID()
	tm, err := utils.ToTimestamp(time.Now().UTC())
	if err != nil {
		return nil, err
	}

	p2pheader := &p2ppb.Header{
		SenderId:          centIDBytes,
		NodeVersion:       version.GetVersion().String(),
		NetworkIdentifier: networkID,
		Type:              messageType.String(),
		Timestamp:         tm,
	}

	body, err := proto.Marshal(mes)
	if err != nil {
		return nil, err
	}

	envelope := &p2ppb.Envelope{
		Header: p2pheader,
		Body:   body,
	}

	marshalledRequest, err := proto.Marshal(envelope)
	if err != nil {
		return nil, err
	}

	return &protocolpb.P2PEnvelope{Body: marshalledRequest}, nil
}

// ConvertClientError converts Envelope to error
func ConvertClientError(recv *p2ppb.Envelope) error {
	resp := new(errorspb.Error)
	err := proto.Unmarshal(recv.Body, resp)
	if err != nil {
		return err
	}
	return errors.New(resp.Message)
}

// ConvertP2PEnvelopeToError converts p2pEnvelope containing an error to Error
func ConvertP2PEnvelopeToError(p2pEnvelope *protocolpb.P2PEnvelope) error {
	envelope, err := ResolveDataEnvelope(p2pEnvelope)
	if err != nil {
		return err
	}
	return ConvertClientError(envelope)
}

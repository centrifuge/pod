// +build unit integration testworld

package p2pcommon

import (
	"context"
	"strconv"
	"time"

	p2ppb "github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	protocolpb "github.com/centrifuge/centrifuge-protobufs/gen/go/protocol"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/go-centrifuge/version"
	"github.com/golang/protobuf/proto"
)

// prepare incorrect protobuf messages

// send message with incorrect node version
func PrepareP2PEnvelopeIncorrectNodeVersion(ctx context.Context, networkID uint32, messageType MessageType, mes proto.Message) (*protocolpb.P2PEnvelope, error) {
	self, err := contextutil.Account(ctx)
	if err != nil {
		return nil, err
	}

	centIDBytes := self.GetIdentityID()
	tm, err := utils.ToTimestamp(time.Now().UTC())
	if err != nil {
		return nil, err
	}

	currentNodeVersion := version.GetVersion().String()
	// increment the node version by one
	modifiedMajorVersion := strconv.FormatInt(version.GetVersion().Major()+1, 10)
	modifiedNodeVersion := modifiedMajorVersion + currentNodeVersion[1:]

	// create new header with incorrect node version
	p2pheader := &p2ppb.Header{
		SenderId:          centIDBytes,
		NodeVersion:       modifiedNodeVersion,
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

// PrepareP2PEnvelopeInvalidBody send message with a random byte array as the body, but with a valid header
func PrepareP2PEnvelopeInvalidBody(ctx context.Context, networkID uint32, messageType MessageType, mes proto.Message) (*protocolpb.P2PEnvelope, error) {

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

	// create a random byte array to send as the message body
	invalidMessageBody := utils.RandomSlice(512)

	envelope := &p2ppb.Envelope{
		Header: p2pheader,
		Body:   invalidMessageBody,
	}

	marshalledRequest, err := proto.Marshal(envelope)
	if err != nil {
		return nil, err
	}

	return &protocolpb.P2PEnvelope{Body: marshalledRequest}, nil
}

// PrepareP2PEnvelopeInvalidHeader send message with random values in the header fields
func PrepareP2PEnvelopeInvalidHeader(ctx context.Context, networkID uint32, messageType MessageType, mes proto.Message) (*protocolpb.P2PEnvelope, error) {

	tm, err := utils.ToTimestamp(time.Now().UTC())
	if err != nil {
		return nil, err
	}

	//get random values for header fields
	InvalidSenderId := utils.RandomSlice(32)
	InvalidNodeVersion := " "
	byteArrayforNetworkID := utils.RandomByte32()
	InvalidNetworkID := uint32(utils.ConvertByte32ToInt(byteArrayforNetworkID))
	InvalidmessageType := " "

	p2pheader := &p2ppb.Header{
		SenderId:          InvalidSenderId,
		NodeVersion:       InvalidNodeVersion,
		NetworkIdentifier: InvalidNetworkID,
		Type:              InvalidmessageType,
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

//go:build unit

package p2pcommon

import (
	"context"
	"fmt"
	"testing"

	errorspb "github.com/centrifuge/centrifuge-protobufs/gen/go/errors"
	p2ppb "github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	protocolpb "github.com/centrifuge/centrifuge-protobufs/gen/go/protocol"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/contextutil"
	testingcommons "github.com/centrifuge/go-centrifuge/testingutils/common"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/go-centrifuge/version"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestProtocolForIdentity(t *testing.T) {
	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	protocolID := ProtocolForIdentity(accountID)

	expectedProtocolID := protocol.ID(fmt.Sprintf("%s/%s", CentrifugeProtocol, accountID.ToHexString()))

	assert.Equal(t, expectedProtocolID, protocolID)
}

func TestExtractIdentity(t *testing.T) {
	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	protocolID := ProtocolForIdentity(accountID)

	res, err := ExtractIdentity(protocolID)
	assert.NoError(t, err)
	assert.Equal(t, res, accountID)

	invalidProtocolID := protocol.ID(utils.RandomSlice(32))

	res, err = ExtractIdentity(invalidProtocolID)
	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestResolveDataEnvelope(t *testing.T) {
	p2pEnvelope := &p2ppb.Envelope{
		Header: &p2ppb.Header{
			NetworkIdentifier: 36,
			NodeVersion:       "node_version",
			SenderId:          utils.RandomSlice(32),
			Type:              "message_type",
			Timestamp:         timestamppb.Now(),
		},
		Body: utils.RandomSlice(32),
	}

	encodedBody, err := proto.Marshal(p2pEnvelope)
	assert.NoError(t, err)

	protocolEnvelope := &protocolpb.P2PEnvelope{
		Body: encodedBody,
	}

	res, err := ResolveDataEnvelope(protocolEnvelope)
	assert.NoError(t, err)
	assert.Equal(t, p2pEnvelope.GetHeader().GetNetworkIdentifier(), res.GetHeader().GetNetworkIdentifier())
	assert.Equal(t, p2pEnvelope.GetHeader().GetNodeVersion(), res.GetHeader().GetNodeVersion())
	assert.Equal(t, p2pEnvelope.GetHeader().GetSenderId(), res.GetHeader().GetSenderId())
	assert.Equal(t, p2pEnvelope.GetHeader().GetType(), res.GetHeader().GetType())
	assert.Equal(t, p2pEnvelope.GetHeader().GetTimestamp().AsTime(), res.GetHeader().GetTimestamp().AsTime())
	assert.Equal(t, p2pEnvelope.GetBody(), res.GetBody())
}

func TestResolveDataEnvelope_InvalidProtoMessage(t *testing.T) {
	p2pEnvelope := &p2ppb.Envelope{
		Header: &p2ppb.Header{
			NetworkIdentifier: 36,
			NodeVersion:       "node_version",
			SenderId:          utils.RandomSlice(32),
			Type:              "message_type",
			Timestamp:         timestamppb.Now(),
		},
		Body: utils.RandomSlice(32),
	}

	res, err := ResolveDataEnvelope(p2pEnvelope)
	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestResolveDataEnvelope_InvalidProtocolEnvelopeBody(t *testing.T) {
	protocolEnvelope := &protocolpb.P2PEnvelope{
		Body: utils.RandomSlice(32),
	}

	res, err := ResolveDataEnvelope(protocolEnvelope)
	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestResolveDataEnvelope_InvalidP2PEnvelope(t *testing.T) {
	p2pEnvelope := &p2ppb.Envelope{
		Header: nil,
		Body:   utils.RandomSlice(32),
	}

	encodedBody, err := proto.Marshal(p2pEnvelope)
	assert.NoError(t, err)

	protocolEnvelope := &protocolpb.P2PEnvelope{
		Body: encodedBody,
	}

	res, err := ResolveDataEnvelope(protocolEnvelope)
	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestPrepareP2PEnvelope(t *testing.T) {
	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(identity).
		Once()

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	msg := &p2ppb.GetDocumentRequest{
		DocumentIdentifier: utils.RandomSlice(32),
		AccessType:         p2ppb.AccessType_ACCESS_TYPE_REQUESTER_VERIFICATION,
	}

	messageType := MessageTypeGetDoc

	networkID := uint32(36)

	res, err := PrepareP2PEnvelope(ctx, networkID, messageType, msg)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	var p2pEnvelope p2ppb.Envelope

	err = proto.Unmarshal(res.GetBody(), &p2pEnvelope)
	assert.NoError(t, err)

	assert.Equal(t, p2pEnvelope.GetHeader().GetSenderId(), identity.ToBytes())
	assert.Equal(t, p2pEnvelope.GetHeader().GetNodeVersion(), version.GetVersion().String())
	assert.Equal(t, p2pEnvelope.GetHeader().GetNetworkIdentifier(), networkID)
	assert.Equal(t, p2pEnvelope.GetHeader().GetType(), messageType.String())
	assert.NotNil(t, p2pEnvelope.GetHeader().GetTimestamp())

	var req p2ppb.GetDocumentRequest

	err = proto.Unmarshal(p2pEnvelope.GetBody(), &req)
	assert.NoError(t, err)

	assert.Equal(t, msg.GetDocumentIdentifier(), req.GetDocumentIdentifier())
	assert.Equal(t, msg.GetAccessType(), req.GetAccessType())
}

func TestPrepareP2PEnvelope_NoSenderIdentity(t *testing.T) {
	msg := &p2ppb.GetDocumentRequest{
		DocumentIdentifier: utils.RandomSlice(32),
		AccessType:         p2ppb.AccessType_ACCESS_TYPE_REQUESTER_VERIFICATION,
	}

	messageType := MessageTypeGetDoc

	networkID := uint32(36)

	res, err := PrepareP2PEnvelope(context.Background(), networkID, messageType, msg)
	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestPrepareP2PEnvelope_NilMessage(t *testing.T) {
	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(identity).
		Once()

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	messageType := MessageTypeGetDoc

	networkID := uint32(36)

	res, err := PrepareP2PEnvelope(ctx, networkID, messageType, nil)
	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestConvertClientError(t *testing.T) {
	protoError := &errorspb.Error{
		Code:    11,
		Message: "error_msg",
		Errors: map[string]string{
			"first_error":  "error1",
			"second_error": "error2",
		},
	}

	encodedError, err := proto.Marshal(protoError)
	assert.NoError(t, err)

	envelope := &p2ppb.Envelope{Body: encodedError}

	err = ConvertClientError(envelope)
	assert.Equal(t, protoError.GetMessage(), err.Error())
}

func TestConvertClientError_InvalidP2PEnvelopeBody(t *testing.T) {
	envelope := &p2ppb.Envelope{Body: utils.RandomSlice(32)}

	var res errorspb.Error

	expectedErr := proto.Unmarshal(envelope.GetBody(), &res)

	err := ConvertClientError(envelope)
	assert.Equal(t, expectedErr, err)
}

func TestConvertP2PEnvelopeToError(t *testing.T) {
	protoError := &errorspb.Error{
		Code:    11,
		Message: "error_msg",
		Errors: map[string]string{
			"first_error":  "error1",
			"second_error": "error2",
		},
	}

	encodedError, err := proto.Marshal(protoError)
	assert.NoError(t, err)

	p2pEnvelope := &p2ppb.Envelope{
		Header: &p2ppb.Header{
			NetworkIdentifier: 36,
			NodeVersion:       "node_version",
			SenderId:          utils.RandomSlice(32),
			Type:              "message_type",
			Timestamp:         timestamppb.Now(),
		},
		Body: encodedError,
	}

	encodedBody, err := proto.Marshal(p2pEnvelope)
	assert.NoError(t, err)

	protocolEnvelope := &protocolpb.P2PEnvelope{
		Body: encodedBody,
	}

	err = ConvertP2PEnvelopeToError(protocolEnvelope)
	assert.Equal(t, protoError.GetMessage(), err.Error())
}

func TestConvertP2PEnvelopeToError_ResolveEnvelopeError(t *testing.T) {
	p2pEnvelope := &p2ppb.Envelope{
		Header: nil,
		Body:   utils.RandomSlice(32),
	}

	encodedBody, err := proto.Marshal(p2pEnvelope)
	assert.NoError(t, err)

	protocolEnvelope := &protocolpb.P2PEnvelope{
		Body: encodedBody,
	}

	res, expectedErr := ResolveDataEnvelope(protocolEnvelope)
	assert.NotNil(t, expectedErr)
	assert.Nil(t, res)

	err = ConvertP2PEnvelopeToError(protocolEnvelope)
	assert.Equal(t, expectedErr.Error(), err.Error())
}

func TestConvertP2PEnvelopeToError_ConversionError(t *testing.T) {
	p2pEnvelope := &p2ppb.Envelope{
		Header: &p2ppb.Header{
			NetworkIdentifier: 36,
			NodeVersion:       "node_version",
			SenderId:          utils.RandomSlice(32),
			Type:              "message_type",
			Timestamp:         timestamppb.Now(),
		},
		Body: nil,
	}

	encodedBody, err := proto.Marshal(p2pEnvelope)
	assert.NoError(t, err)

	protocolEnvelope := &protocolpb.P2PEnvelope{
		Body: encodedBody,
	}

	expectedErr := ConvertClientError(p2pEnvelope)

	err = ConvertP2PEnvelopeToError(protocolEnvelope)
	assert.Equal(t, expectedErr.Error(), err.Error())
}

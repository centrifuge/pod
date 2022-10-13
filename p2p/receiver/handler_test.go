//go:build unit

package receiver

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types/codec"

	"github.com/centrifuge/go-centrifuge/errors"

	"github.com/centrifuge/go-centrifuge/contextutil"

	errorspb "github.com/centrifuge/centrifuge-protobufs/gen/go/errors"

	"github.com/centrifuge/go-centrifuge/utils"

	"github.com/stretchr/testify/mock"

	genericUtils "github.com/centrifuge/go-centrifuge/testingutils/generic"

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	protocolpb "github.com/centrifuge/centrifuge-protobufs/gen/go/protocol"
	libp2ppeer "github.com/libp2p/go-libp2p-core/peer"
	"google.golang.org/protobuf/proto"

	p2ppb "github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	p2pcommon "github.com/centrifuge/go-centrifuge/p2p/common"
	testingcommons "github.com/centrifuge/go-centrifuge/testingutils/common"
	"github.com/centrifuge/go-centrifuge/version"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/documents"
	v2 "github.com/centrifuge/go-centrifuge/identity/v2"
	nftv3 "github.com/centrifuge/go-centrifuge/nft/v3"
)

func TestHandler_HandleInterceptor_RequestDocumentSignature(t *testing.T) {
	handler, mocks := getHandlerWithMocks(t)

	ctx := context.Background()

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	senderAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	cd := &coredocumentpb.CoreDocument{}

	req := &p2ppb.SignatureRequest{
		Document: cd,
	}

	encodedReq, err := proto.Marshal(req)
	assert.NoError(t, err)

	env := &p2ppb.Envelope{
		Header: &p2ppb.Header{
			NetworkIdentifier: 36,
			NodeVersion:       version.GetVersion().String(),
			SenderId:          senderAccountID.ToBytes(),
			Type:              p2pcommon.MessageTypeRequestSignature.String(),
			Timestamp:         timestamppb.Now(),
		},
		Body: encodedReq,
	}

	encodedEnv, err := proto.Marshal(env)
	assert.NoError(t, err)

	msg := &protocolpb.P2PEnvelope{
		Body: encodedEnv,
	}

	peerID := libp2ppeer.ID("peer-id")
	protocolID := p2pcommon.ProtocolForIdentity(identity)

	genericUtils.GetMock[*config.ConfigurationMock](mocks).
		On("GetP2PResponseDelay").
		Return(0 * time.Second).Once()

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(identity)

	genericUtils.GetMock[*config.ServiceMock](mocks).
		On("GetAccount", identity.ToBytes()).
		Return(accountMock, nil).Once()

	genericUtils.GetMock[*ValidatorMock](mocks).
		On("Validate", mock.IsType(env.GetHeader()), senderAccountID, &peerID).
		Run(func(args mock.Arguments) {
			header, ok := args.Get(0).(*p2ppb.Header)
			assert.True(t, ok)

			assertExpectedHeaderMatchesActual(t, env.GetHeader(), header)
		}).
		Return(nil).Once()

	// RequestDocumentSignature

	documentMock := documents.NewDocumentMock(t)

	genericUtils.GetMock[*documents.ServiceMock](mocks).
		On("DeriveFromCoreDocument", req.GetDocument()).
		Return(documentMock, nil).Once()

	signatures := []*coredocumentpb.Signature{
		{
			SignatureId:         utils.RandomSlice(32),
			SignerId:            utils.RandomSlice(32),
			PublicKey:           utils.RandomSlice(32),
			Signature:           utils.RandomSlice(32),
			TransitionValidated: true,
		},
	}

	genericUtils.GetMock[*documents.ServiceMock](mocks).
		On(
			"RequestDocumentSignature",
			mock.Anything,
			documentMock,
			senderAccountID,
		).
		Return(signatures, nil).Once()

	genericUtils.GetMock[*config.ConfigurationMock](mocks).
		On("GetNetworkID").
		Return(uint32(36)).Once()

	res, err := handler.HandleInterceptor(ctx, peerID, protocolID, msg)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	var responseEnvelope p2ppb.Envelope

	err = proto.Unmarshal(res.GetBody(), &responseEnvelope)
	assert.NoError(t, err)

	var signatureRes p2ppb.SignatureResponse

	err = proto.Unmarshal(responseEnvelope.GetBody(), &signatureRes)
	assert.NoError(t, err)

	assert.Len(t, signatureRes.GetSignatures(), 1)
	assert.Equal(t, signatures[0].GetSignatureId(), signatureRes.GetSignatures()[0].GetSignatureId())
	assert.Equal(t, signatures[0].GetSignerId(), signatureRes.GetSignatures()[0].GetSignerId())
	assert.Equal(t, signatures[0].GetPublicKey(), signatureRes.GetSignatures()[0].GetPublicKey())
	assert.Equal(t, signatures[0].GetSignature(), signatureRes.GetSignatures()[0].GetSignature())
	assert.Equal(t, signatures[0].GetTransitionValidated(), signatureRes.GetSignatures()[0].GetTransitionValidated())
}

func TestHandler_HandleInterceptor_SendAnchoredDocument(t *testing.T) {
	handler, mocks := getHandlerWithMocks(t)

	ctx := context.Background()

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	senderAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	cd := &coredocumentpb.CoreDocument{}

	req := &p2ppb.AnchorDocumentRequest{
		Document: cd,
	}

	encodedReq, err := proto.Marshal(req)
	assert.NoError(t, err)

	env := &p2ppb.Envelope{
		Header: &p2ppb.Header{
			NetworkIdentifier: 36,
			NodeVersion:       version.GetVersion().String(),
			SenderId:          senderAccountID.ToBytes(),
			Type:              p2pcommon.MessageTypeSendAnchoredDoc.String(),
			Timestamp:         timestamppb.Now(),
		},
		Body: encodedReq,
	}

	encodedEnv, err := proto.Marshal(env)
	assert.NoError(t, err)

	msg := &protocolpb.P2PEnvelope{
		Body: encodedEnv,
	}

	peerID := libp2ppeer.ID("peer-id")
	protocolID := p2pcommon.ProtocolForIdentity(identity)

	genericUtils.GetMock[*config.ConfigurationMock](mocks).
		On("GetP2PResponseDelay").
		Return(0 * time.Second).Once()

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(identity)

	genericUtils.GetMock[*config.ServiceMock](mocks).
		On("GetAccount", identity.ToBytes()).
		Return(accountMock, nil).Once()

	genericUtils.GetMock[*ValidatorMock](mocks).
		On("Validate", mock.IsType(env.GetHeader()), senderAccountID, &peerID).
		Run(func(args mock.Arguments) {
			header, ok := args.Get(0).(*p2ppb.Header)
			assert.True(t, ok)

			assertExpectedHeaderMatchesActual(t, env.GetHeader(), header)
		}).
		Return(nil).Once()

	// SendAnchoredDocument

	documentMock := documents.NewDocumentMock(t)

	genericUtils.GetMock[*documents.ServiceMock](mocks).
		On("DeriveFromCoreDocument", req.GetDocument()).
		Return(documentMock, nil).Once()

	genericUtils.GetMock[*documents.ServiceMock](mocks).
		On(
			"ReceiveAnchoredDocument",
			mock.Anything,
			documentMock,
			senderAccountID,
		).
		Return(nil).Once()

	genericUtils.GetMock[*config.ConfigurationMock](mocks).
		On("GetNetworkID").
		Return(uint32(36)).Once()

	res, err := handler.HandleInterceptor(ctx, peerID, protocolID, msg)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	var responseEnvelope p2ppb.Envelope

	err = proto.Unmarshal(res.GetBody(), &responseEnvelope)
	assert.NoError(t, err)

	var anchorDoc p2ppb.AnchorDocumentResponse

	err = proto.Unmarshal(responseEnvelope.GetBody(), &anchorDoc)
	assert.NoError(t, err)

	assert.True(t, anchorDoc.GetAccepted())
}

func TestHandler_HandleInterceptor_GetDocument(t *testing.T) {
	handler, mocks := getHandlerWithMocks(t)

	ctx := context.Background()

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	senderAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	req := &p2ppb.GetDocumentRequest{
		DocumentIdentifier: utils.RandomSlice(32),
		AccessType:         p2ppb.AccessType_ACCESS_TYPE_REQUESTER_VERIFICATION,
	}

	encodedReq, err := proto.Marshal(req)
	assert.NoError(t, err)

	env := &p2ppb.Envelope{
		Header: &p2ppb.Header{
			NetworkIdentifier: 36,
			NodeVersion:       version.GetVersion().String(),
			SenderId:          senderAccountID.ToBytes(),
			Type:              p2pcommon.MessageTypeGetDoc.String(),
			Timestamp:         timestamppb.Now(),
		},
		Body: encodedReq,
	}

	encodedEnv, err := proto.Marshal(env)
	assert.NoError(t, err)

	msg := &protocolpb.P2PEnvelope{
		Body: encodedEnv,
	}

	peerID := libp2ppeer.ID("peer-id")
	protocolID := p2pcommon.ProtocolForIdentity(identity)

	genericUtils.GetMock[*config.ConfigurationMock](mocks).
		On("GetP2PResponseDelay").
		Return(0 * time.Second).Once()

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(identity)

	genericUtils.GetMock[*config.ServiceMock](mocks).
		On("GetAccount", identity.ToBytes()).
		Return(accountMock, nil).Once()

	genericUtils.GetMock[*ValidatorMock](mocks).
		On("Validate", mock.IsType(env.GetHeader()), senderAccountID, &peerID).
		Run(func(args mock.Arguments) {
			header, ok := args.Get(0).(*p2ppb.Header)
			assert.True(t, ok)

			assertExpectedHeaderMatchesActual(t, env.GetHeader(), header)
		}).
		Return(nil).Once()

	// RequestDocumentSignature

	documentMock := documents.NewDocumentMock(t)

	genericUtils.GetMock[*documents.ServiceMock](mocks).
		On(
			"GetCurrentVersion",
			mock.Anything,
			req.GetDocumentIdentifier(),
		).
		Return(documentMock, nil).Once()

	cd := &coredocumentpb.CoreDocument{}

	documentMock.On("AccountCanRead", senderAccountID).
		Return(true).
		Once()

	documentMock.On("PackCoreDocument").
		Return(cd, nil).
		Once()

	genericUtils.GetMock[*config.ConfigurationMock](mocks).
		On("GetNetworkID").
		Return(uint32(36)).Once()

	res, err := handler.HandleInterceptor(ctx, peerID, protocolID, msg)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	var responseEnvelope p2ppb.Envelope

	err = proto.Unmarshal(res.GetBody(), &responseEnvelope)
	assert.NoError(t, err)

	var getDocumentRes p2ppb.GetDocumentResponse

	err = proto.Unmarshal(responseEnvelope.GetBody(), &getDocumentRes)
	assert.NoError(t, err)

	assert.NotNil(t, getDocumentRes.GetDocument())
}

func TestHandler_HandleInterceptor_InvalidMessageType(t *testing.T) {
	handler, mocks := getHandlerWithMocks(t)

	ctx := context.Background()

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	senderAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	env := &p2ppb.Envelope{
		Header: &p2ppb.Header{
			NetworkIdentifier: 36,
			NodeVersion:       version.GetVersion().String(),
			SenderId:          senderAccountID.ToBytes(),
			Type:              "invalid-message-type",
			Timestamp:         timestamppb.Now(),
		},
		Body: nil,
	}

	encodedEnv, err := proto.Marshal(env)
	assert.NoError(t, err)

	msg := &protocolpb.P2PEnvelope{
		Body: encodedEnv,
	}

	peerID := libp2ppeer.ID("peer-id")
	protocolID := p2pcommon.ProtocolForIdentity(identity)

	genericUtils.GetMock[*config.ConfigurationMock](mocks).
		On("GetP2PResponseDelay").
		Return(0 * time.Second).Once()

	accountMock := config.NewAccountMock(t)

	genericUtils.GetMock[*config.ServiceMock](mocks).
		On("GetAccount", identity.ToBytes()).
		Return(accountMock, nil).Once()

	genericUtils.GetMock[*ValidatorMock](mocks).
		On("Validate", mock.IsType(env.GetHeader()), senderAccountID, &peerID).
		Run(func(args mock.Arguments) {
			header, ok := args.Get(0).(*p2ppb.Header)
			assert.True(t, ok)

			assertExpectedHeaderMatchesActual(t, env.GetHeader(), header)

		}).
		Return(nil).Once()

	res, err := handler.HandleInterceptor(ctx, peerID, protocolID, msg)
	assert.Nil(t, err)
	assert.NotNil(t, res)

	assertErrorEnvelope(t, res)
}

func TestHandler_HandleRequestDocumentSignature(t *testing.T) {
	handler, mocks := getHandlerWithMocks(t)

	ctx := context.Background()

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(identity)

	ctx = contextutil.WithAccount(ctx, accountMock)

	senderAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	cd := &coredocumentpb.CoreDocument{}

	req := &p2ppb.SignatureRequest{
		Document: cd,
	}

	encodedReq, err := proto.Marshal(req)
	assert.NoError(t, err)

	env := &p2ppb.Envelope{
		Header: &p2ppb.Header{
			NetworkIdentifier: 36,
			NodeVersion:       version.GetVersion().String(),
			SenderId:          senderAccountID.ToBytes(),
			Type:              p2pcommon.MessageTypeRequestSignature.String(),
			Timestamp:         timestamppb.Now(),
		},
		Body: encodedReq,
	}

	peerID := libp2ppeer.ID("peer-id")
	protocolID := p2pcommon.ProtocolForIdentity(identity)

	documentMock := documents.NewDocumentMock(t)

	genericUtils.GetMock[*documents.ServiceMock](mocks).
		On("DeriveFromCoreDocument", req.GetDocument()).
		Return(documentMock, nil).Once()

	signatures := []*coredocumentpb.Signature{
		{
			SignatureId:         utils.RandomSlice(32),
			SignerId:            utils.RandomSlice(32),
			PublicKey:           utils.RandomSlice(32),
			Signature:           utils.RandomSlice(32),
			TransitionValidated: true,
		},
	}

	genericUtils.GetMock[*documents.ServiceMock](mocks).
		On(
			"RequestDocumentSignature",
			mock.Anything,
			documentMock,
			senderAccountID,
		).
		Return(signatures, nil).Once()

	genericUtils.GetMock[*config.ConfigurationMock](mocks).
		On("GetNetworkID").
		Return(uint32(36)).Once()

	res, err := handler.HandleRequestDocumentSignature(ctx, peerID, protocolID, env)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	var responseEnvelope p2ppb.Envelope

	err = proto.Unmarshal(res.GetBody(), &responseEnvelope)
	assert.NoError(t, err)

	var signatureRes p2ppb.SignatureResponse

	err = proto.Unmarshal(responseEnvelope.GetBody(), &signatureRes)
	assert.NoError(t, err)

	assert.Len(t, signatureRes.GetSignatures(), 1)
	assert.Equal(t, signatures[0].GetSignatureId(), signatureRes.GetSignatures()[0].GetSignatureId())
	assert.Equal(t, signatures[0].GetSignerId(), signatureRes.GetSignatures()[0].GetSignerId())
	assert.Equal(t, signatures[0].GetPublicKey(), signatureRes.GetSignatures()[0].GetPublicKey())
	assert.Equal(t, signatures[0].GetSignature(), signatureRes.GetSignatures()[0].GetSignature())
	assert.Equal(t, signatures[0].GetTransitionValidated(), signatureRes.GetSignatures()[0].GetTransitionValidated())
}

func TestHandler_HandleRequestDocumentSignature_RequestDecodeError(t *testing.T) {
	handler, _ := getHandlerWithMocks(t)

	ctx := context.Background()

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	senderAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	env := &p2ppb.Envelope{
		Header: &p2ppb.Header{
			NetworkIdentifier: 36,
			NodeVersion:       version.GetVersion().String(),
			SenderId:          senderAccountID.ToBytes(),
			Type:              p2pcommon.MessageTypeRequestSignature.String(),
			Timestamp:         timestamppb.Now(),
		},
		Body: utils.RandomSlice(32),
	}

	peerID := libp2ppeer.ID("peer-id")
	protocolID := p2pcommon.ProtocolForIdentity(identity)

	res, err := handler.HandleRequestDocumentSignature(ctx, peerID, protocolID, env)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	assertErrorEnvelope(t, res)
}

func TestHandler_HandleRequestDocumentSignature_InvalidSenderID(t *testing.T) {
	handler, _ := getHandlerWithMocks(t)

	ctx := context.Background()

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	cd := &coredocumentpb.CoreDocument{}

	req := &p2ppb.SignatureRequest{
		Document: cd,
	}

	encodedReq, err := proto.Marshal(req)
	assert.NoError(t, err)

	env := &p2ppb.Envelope{
		Header: &p2ppb.Header{
			NetworkIdentifier: 36,
			NodeVersion:       version.GetVersion().String(),
			// Invalid sender ID account bytes.
			SenderId:  utils.RandomSlice(31),
			Type:      p2pcommon.MessageTypeRequestSignature.String(),
			Timestamp: timestamppb.Now(),
		},
		Body: encodedReq,
	}

	peerID := libp2ppeer.ID("peer-id")
	protocolID := p2pcommon.ProtocolForIdentity(identity)

	res, err := handler.HandleRequestDocumentSignature(ctx, peerID, protocolID, env)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	assertErrorEnvelope(t, res)
}

func TestHandler_HandleRequestDocumentSignature_NilDocument(t *testing.T) {
	handler, _ := getHandlerWithMocks(t)

	ctx := context.Background()

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	senderAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	req := &p2ppb.SignatureRequest{}

	encodedReq, err := proto.Marshal(req)
	assert.NoError(t, err)

	env := &p2ppb.Envelope{
		Header: &p2ppb.Header{
			NetworkIdentifier: 36,
			NodeVersion:       version.GetVersion().String(),
			SenderId:          senderAccountID.ToBytes(),
			Type:              p2pcommon.MessageTypeRequestSignature.String(),
			Timestamp:         timestamppb.Now(),
		},
		Body: encodedReq,
	}

	peerID := libp2ppeer.ID("peer-id")
	protocolID := p2pcommon.ProtocolForIdentity(identity)

	res, err := handler.HandleRequestDocumentSignature(ctx, peerID, protocolID, env)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	assertErrorEnvelope(t, res)
}

func TestHandler_HandleRequestDocumentSignature_DocDeriveError(t *testing.T) {
	handler, mocks := getHandlerWithMocks(t)

	ctx := context.Background()

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	senderAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	cd := &coredocumentpb.CoreDocument{}

	req := &p2ppb.SignatureRequest{
		Document: cd,
	}

	encodedReq, err := proto.Marshal(req)
	assert.NoError(t, err)

	env := &p2ppb.Envelope{
		Header: &p2ppb.Header{
			NetworkIdentifier: 36,
			NodeVersion:       version.GetVersion().String(),
			SenderId:          senderAccountID.ToBytes(),
			Type:              p2pcommon.MessageTypeRequestSignature.String(),
			Timestamp:         timestamppb.Now(),
		},
		Body: encodedReq,
	}

	peerID := libp2ppeer.ID("peer-id")
	protocolID := p2pcommon.ProtocolForIdentity(identity)

	genericUtils.GetMock[*documents.ServiceMock](mocks).
		On("DeriveFromCoreDocument", req.GetDocument()).
		Return(nil, errors.New("error")).Once()

	res, err := handler.HandleRequestDocumentSignature(ctx, peerID, protocolID, env)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	assertErrorEnvelope(t, res)
}

func TestHandler_HandleRequestDocumentSignature_RequestDocumentSignatureError(t *testing.T) {
	handler, mocks := getHandlerWithMocks(t)

	ctx := context.Background()

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	senderAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	cd := &coredocumentpb.CoreDocument{}

	req := &p2ppb.SignatureRequest{
		Document: cd,
	}

	encodedReq, err := proto.Marshal(req)
	assert.NoError(t, err)

	env := &p2ppb.Envelope{
		Header: &p2ppb.Header{
			NetworkIdentifier: 36,
			NodeVersion:       version.GetVersion().String(),
			SenderId:          senderAccountID.ToBytes(),
			Type:              p2pcommon.MessageTypeRequestSignature.String(),
			Timestamp:         timestamppb.Now(),
		},
		Body: encodedReq,
	}

	peerID := libp2ppeer.ID("peer-id")
	protocolID := p2pcommon.ProtocolForIdentity(identity)

	documentMock := documents.NewDocumentMock(t)

	genericUtils.GetMock[*documents.ServiceMock](mocks).
		On("DeriveFromCoreDocument", req.GetDocument()).
		Return(documentMock, nil).Once()

	genericUtils.GetMock[*documents.ServiceMock](mocks).
		On(
			"RequestDocumentSignature",
			mock.Anything,
			documentMock,
			senderAccountID,
		).
		Return(nil, errors.New("error")).Once()

	res, err := handler.HandleRequestDocumentSignature(ctx, peerID, protocolID, env)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	assertErrorEnvelope(t, res)
}

func TestHandler_HandleSendAnchoredDocument(t *testing.T) {
	handler, mocks := getHandlerWithMocks(t)

	ctx := context.Background()

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(identity)

	ctx = contextutil.WithAccount(ctx, accountMock)

	senderAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	cd := &coredocumentpb.CoreDocument{}

	req := &p2ppb.AnchorDocumentRequest{
		Document: cd,
	}

	encodedReq, err := proto.Marshal(req)
	assert.NoError(t, err)

	env := &p2ppb.Envelope{
		Header: &p2ppb.Header{
			NetworkIdentifier: 36,
			NodeVersion:       version.GetVersion().String(),
			SenderId:          senderAccountID.ToBytes(),
			Type:              p2pcommon.MessageTypeSendAnchoredDoc.String(),
			Timestamp:         timestamppb.Now(),
		},
		Body: encodedReq,
	}

	peerID := libp2ppeer.ID("peer-id")
	protocolID := p2pcommon.ProtocolForIdentity(identity)

	documentMock := documents.NewDocumentMock(t)

	genericUtils.GetMock[*documents.ServiceMock](mocks).
		On("DeriveFromCoreDocument", req.GetDocument()).
		Return(documentMock, nil).Once()

	genericUtils.GetMock[*documents.ServiceMock](mocks).
		On(
			"ReceiveAnchoredDocument",
			mock.Anything,
			documentMock,
			senderAccountID,
		).
		Return(nil).Once()

	genericUtils.GetMock[*config.ConfigurationMock](mocks).
		On("GetNetworkID").
		Return(uint32(36)).Once()

	res, err := handler.HandleSendAnchoredDocument(ctx, peerID, protocolID, env)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	var responseEnvelope p2ppb.Envelope

	err = proto.Unmarshal(res.GetBody(), &responseEnvelope)
	assert.NoError(t, err)

	var anchorDoc p2ppb.AnchorDocumentResponse

	err = proto.Unmarshal(responseEnvelope.GetBody(), &anchorDoc)
	assert.NoError(t, err)

	assert.True(t, anchorDoc.GetAccepted())
}

func TestHandler_HandleSendAnchoredDocument_RequestDecodeError(t *testing.T) {
	handler, _ := getHandlerWithMocks(t)

	ctx := context.Background()

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	senderAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	env := &p2ppb.Envelope{
		Header: &p2ppb.Header{
			NetworkIdentifier: 36,
			NodeVersion:       version.GetVersion().String(),
			SenderId:          senderAccountID.ToBytes(),
			Type:              p2pcommon.MessageTypeSendAnchoredDoc.String(),
			Timestamp:         timestamppb.Now(),
		},
		Body: utils.RandomSlice(32),
	}

	peerID := libp2ppeer.ID("peer-id")
	protocolID := p2pcommon.ProtocolForIdentity(identity)

	res, err := handler.HandleSendAnchoredDocument(ctx, peerID, protocolID, env)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	assertErrorEnvelope(t, res)
}

func TestHandler_HandleSendAnchoredDocument_InvalidSenderID(t *testing.T) {
	handler, _ := getHandlerWithMocks(t)

	ctx := context.Background()

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	cd := &coredocumentpb.CoreDocument{}

	req := &p2ppb.AnchorDocumentRequest{
		Document: cd,
	}

	encodedReq, err := proto.Marshal(req)
	assert.NoError(t, err)

	env := &p2ppb.Envelope{
		Header: &p2ppb.Header{
			NetworkIdentifier: 36,
			NodeVersion:       version.GetVersion().String(),
			// Invalid sender ID account bytes.
			SenderId:  utils.RandomSlice(11),
			Type:      p2pcommon.MessageTypeSendAnchoredDoc.String(),
			Timestamp: timestamppb.Now(),
		},
		Body: encodedReq,
	}

	peerID := libp2ppeer.ID("peer-id")
	protocolID := p2pcommon.ProtocolForIdentity(identity)

	res, err := handler.HandleSendAnchoredDocument(ctx, peerID, protocolID, env)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	assertErrorEnvelope(t, res)
}

func TestHandler_HandleSendAnchoredDocument_NilDocument(t *testing.T) {
	handler, _ := getHandlerWithMocks(t)

	ctx := context.Background()

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	senderAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	req := &p2ppb.AnchorDocumentRequest{}

	encodedReq, err := proto.Marshal(req)
	assert.NoError(t, err)

	env := &p2ppb.Envelope{
		Header: &p2ppb.Header{
			NetworkIdentifier: 36,
			NodeVersion:       version.GetVersion().String(),
			SenderId:          senderAccountID.ToBytes(),
			Type:              p2pcommon.MessageTypeSendAnchoredDoc.String(),
			Timestamp:         timestamppb.Now(),
		},
		Body: encodedReq,
	}

	peerID := libp2ppeer.ID("peer-id")
	protocolID := p2pcommon.ProtocolForIdentity(identity)

	res, err := handler.HandleSendAnchoredDocument(ctx, peerID, protocolID, env)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	assertErrorEnvelope(t, res)
}

func TestHandler_HandleSendAnchoredDocument_DocDeriveError(t *testing.T) {
	handler, mocks := getHandlerWithMocks(t)

	ctx := context.Background()

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	senderAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	cd := &coredocumentpb.CoreDocument{}

	req := &p2ppb.AnchorDocumentRequest{
		Document: cd,
	}

	encodedReq, err := proto.Marshal(req)
	assert.NoError(t, err)

	env := &p2ppb.Envelope{
		Header: &p2ppb.Header{
			NetworkIdentifier: 36,
			NodeVersion:       version.GetVersion().String(),
			SenderId:          senderAccountID.ToBytes(),
			Type:              p2pcommon.MessageTypeSendAnchoredDoc.String(),
			Timestamp:         timestamppb.Now(),
		},
		Body: encodedReq,
	}

	peerID := libp2ppeer.ID("peer-id")
	protocolID := p2pcommon.ProtocolForIdentity(identity)

	genericUtils.GetMock[*documents.ServiceMock](mocks).
		On("DeriveFromCoreDocument", req.GetDocument()).
		Return(nil, errors.New("error")).Once()

	res, err := handler.HandleSendAnchoredDocument(ctx, peerID, protocolID, env)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	assertErrorEnvelope(t, res)
}

func TestHandler_HandleSendAnchoredDocument_ReceiveDocumentError(t *testing.T) {
	handler, mocks := getHandlerWithMocks(t)

	ctx := context.Background()

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	senderAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	cd := &coredocumentpb.CoreDocument{}

	req := &p2ppb.AnchorDocumentRequest{
		Document: cd,
	}

	encodedReq, err := proto.Marshal(req)
	assert.NoError(t, err)

	env := &p2ppb.Envelope{
		Header: &p2ppb.Header{
			NetworkIdentifier: 36,
			NodeVersion:       version.GetVersion().String(),
			SenderId:          senderAccountID.ToBytes(),
			Type:              p2pcommon.MessageTypeSendAnchoredDoc.String(),
			Timestamp:         timestamppb.Now(),
		},
		Body: encodedReq,
	}

	peerID := libp2ppeer.ID("peer-id")
	protocolID := p2pcommon.ProtocolForIdentity(identity)

	documentMock := documents.NewDocumentMock(t)

	genericUtils.GetMock[*documents.ServiceMock](mocks).
		On("DeriveFromCoreDocument", req.GetDocument()).
		Return(documentMock, nil).Once()

	genericUtils.GetMock[*documents.ServiceMock](mocks).
		On(
			"ReceiveAnchoredDocument",
			mock.Anything,
			documentMock,
			senderAccountID,
		).
		Return(errors.New("error")).Once()

	res, err := handler.HandleSendAnchoredDocument(ctx, peerID, protocolID, env)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	assertErrorEnvelope(t, res)
}

func TestHandler_HandleGetDocument_RequesterVerification(t *testing.T) {
	handler, mocks := getHandlerWithMocks(t)

	ctx := context.Background()

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(identity)

	ctx = contextutil.WithAccount(ctx, accountMock)

	senderAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	req := &p2ppb.GetDocumentRequest{
		DocumentIdentifier: utils.RandomSlice(32),
		AccessType:         p2ppb.AccessType_ACCESS_TYPE_REQUESTER_VERIFICATION,
	}

	encodedReq, err := proto.Marshal(req)
	assert.NoError(t, err)

	env := &p2ppb.Envelope{
		Header: &p2ppb.Header{
			NetworkIdentifier: 36,
			NodeVersion:       version.GetVersion().String(),
			SenderId:          senderAccountID.ToBytes(),
			Type:              p2pcommon.MessageTypeGetDoc.String(),
			Timestamp:         timestamppb.Now(),
		},
		Body: encodedReq,
	}

	peerID := libp2ppeer.ID("peer-id")
	protocolID := p2pcommon.ProtocolForIdentity(identity)

	documentMock := documents.NewDocumentMock(t)

	genericUtils.GetMock[*documents.ServiceMock](mocks).
		On(
			"GetCurrentVersion",
			mock.Anything,
			req.GetDocumentIdentifier(),
		).
		Return(documentMock, nil).Once()

	cd := &coredocumentpb.CoreDocument{}

	documentMock.On("AccountCanRead", senderAccountID).
		Return(true).
		Once()

	documentMock.On("PackCoreDocument").
		Return(cd, nil).
		Once()

	genericUtils.GetMock[*config.ConfigurationMock](mocks).
		On("GetNetworkID").
		Return(uint32(36)).Once()

	res, err := handler.HandleGetDocument(ctx, peerID, protocolID, env)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	var responseEnvelope p2ppb.Envelope

	err = proto.Unmarshal(res.GetBody(), &responseEnvelope)
	assert.NoError(t, err)

	var getDocumentRes p2ppb.GetDocumentResponse

	err = proto.Unmarshal(responseEnvelope.GetBody(), &getDocumentRes)
	assert.NoError(t, err)

	assert.NotNil(t, getDocumentRes.GetDocument())
}

func TestHandler_HandleGetDocument_NFTOwnerVerification(t *testing.T) {
	handler, mocks := getHandlerWithMocks(t)

	ctx := context.Background()

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(identity)

	ctx = contextutil.WithAccount(ctx, accountMock)

	senderAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	collectionID := types.U64(1111)
	encodedCollectionID, err := codec.Encode(collectionID)
	assert.NoError(t, err)

	itemID := types.NewU128(*big.NewInt(2222))
	encodedItemID, err := codec.Encode(itemID)
	assert.NoError(t, err)

	req := &p2ppb.GetDocumentRequest{
		DocumentIdentifier: utils.RandomSlice(32),
		AccessType:         p2ppb.AccessType_ACCESS_TYPE_NFT_OWNER_VERIFICATION,
		NftCollectionId:    encodedCollectionID,
		NftItemId:          encodedItemID,
	}

	encodedReq, err := proto.Marshal(req)
	assert.NoError(t, err)

	env := &p2ppb.Envelope{
		Header: &p2ppb.Header{
			NetworkIdentifier: 36,
			NodeVersion:       version.GetVersion().String(),
			SenderId:          senderAccountID.ToBytes(),
			Type:              p2pcommon.MessageTypeGetDoc.String(),
			Timestamp:         timestamppb.Now(),
		},
		Body: encodedReq,
	}

	peerID := libp2ppeer.ID("peer-id")
	protocolID := p2pcommon.ProtocolForIdentity(identity)

	documentMock := documents.NewDocumentMock(t)

	genericUtils.GetMock[*documents.ServiceMock](mocks).
		On(
			"GetCurrentVersion",
			mock.Anything,
			req.GetDocumentIdentifier(),
		).
		Return(documentMock, nil).Once()

	cd := &coredocumentpb.CoreDocument{}

	documentMock.On("AccountCanRead", senderAccountID).
		Return(true).
		Once()

	documentMock.On("NFTCanRead", encodedCollectionID, encodedItemID).
		Return(true).
		Once()

	genericUtils.GetMock[*nftv3.ServiceMock](mocks).
		On("GetNFTOwner", collectionID, itemID).
		Return(senderAccountID, nil).Once()

	documentMock.On("PackCoreDocument").
		Return(cd, nil).
		Once()

	genericUtils.GetMock[*config.ConfigurationMock](mocks).
		On("GetNetworkID").
		Return(uint32(36)).Once()

	res, err := handler.HandleGetDocument(ctx, peerID, protocolID, env)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	var responseEnvelope p2ppb.Envelope

	err = proto.Unmarshal(res.GetBody(), &responseEnvelope)
	assert.NoError(t, err)

	var getDocumentRes p2ppb.GetDocumentResponse

	err = proto.Unmarshal(responseEnvelope.GetBody(), &getDocumentRes)
	assert.NoError(t, err)

	assert.NotNil(t, getDocumentRes.GetDocument())
}

func TestHandler_HandleGetDocument_AccessTokenVerification(t *testing.T) {
	handler, mocks := getHandlerWithMocks(t)

	ctx := context.Background()

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(identity)

	ctx = contextutil.WithAccount(ctx, accountMock)

	senderAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	req := &p2ppb.GetDocumentRequest{
		DocumentIdentifier: utils.RandomSlice(32),
		AccessType:         p2ppb.AccessType_ACCESS_TYPE_ACCESS_TOKEN_VERIFICATION,
		AccessTokenRequest: &p2ppb.AccessTokenRequest{
			DelegatingDocumentIdentifier: utils.RandomSlice(32),
			AccessTokenId:                utils.RandomSlice(32),
		},
	}

	encodedReq, err := proto.Marshal(req)
	assert.NoError(t, err)

	env := &p2ppb.Envelope{
		Header: &p2ppb.Header{
			NetworkIdentifier: 36,
			NodeVersion:       version.GetVersion().String(),
			SenderId:          senderAccountID.ToBytes(),
			Type:              p2pcommon.MessageTypeGetDoc.String(),
			Timestamp:         timestamppb.Now(),
		},
		Body: encodedReq,
	}

	peerID := libp2ppeer.ID("peer-id")
	protocolID := p2pcommon.ProtocolForIdentity(identity)

	documentMock := documents.NewDocumentMock(t)

	genericUtils.GetMock[*documents.ServiceMock](mocks).
		On(
			"GetCurrentVersion",
			mock.Anything,
			req.GetDocumentIdentifier(),
		).
		Return(documentMock, nil).Once()

	entityRelationshipMock := documents.NewDocumentMock(t)

	genericUtils.GetMock[*documents.ServiceMock](mocks).
		On(
			"GetCurrentVersion",
			mock.Anything,
			req.GetAccessTokenRequest().GetDelegatingDocumentIdentifier(),
		).
		Return(entityRelationshipMock, nil).Once()

	entityRelationshipMock.On(
		"ATGranteeCanRead",
		mock.Anything,
		handler.docSrv,
		handler.identityService,
		req.GetAccessTokenRequest().GetAccessTokenId(),
		req.GetDocumentIdentifier(),
		senderAccountID,
	).Return(nil).Once()

	cd := &coredocumentpb.CoreDocument{}

	documentMock.On("PackCoreDocument").
		Return(cd, nil).
		Once()

	genericUtils.GetMock[*config.ConfigurationMock](mocks).
		On("GetNetworkID").
		Return(uint32(36)).Once()

	res, err := handler.HandleGetDocument(ctx, peerID, protocolID, env)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	var responseEnvelope p2ppb.Envelope

	err = proto.Unmarshal(res.GetBody(), &responseEnvelope)
	assert.NoError(t, err)

	var getDocumentRes p2ppb.GetDocumentResponse

	err = proto.Unmarshal(responseEnvelope.GetBody(), &getDocumentRes)
	assert.NoError(t, err)

	assert.NotNil(t, getDocumentRes.GetDocument())
}

func TestHandler_HandleGetDocument_RequestDecodeError(t *testing.T) {
	handler, _ := getHandlerWithMocks(t)

	ctx := context.Background()

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	senderAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	env := &p2ppb.Envelope{
		Header: &p2ppb.Header{
			NetworkIdentifier: 36,
			NodeVersion:       version.GetVersion().String(),
			SenderId:          senderAccountID.ToBytes(),
			Type:              p2pcommon.MessageTypeGetDoc.String(),
			Timestamp:         timestamppb.Now(),
		},
		Body: utils.RandomSlice(32),
	}

	peerID := libp2ppeer.ID("peer-id")
	protocolID := p2pcommon.ProtocolForIdentity(identity)

	res, err := handler.HandleGetDocument(ctx, peerID, protocolID, env)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	assertErrorEnvelope(t, res)
}

func TestHandler_HandleGetDocument_InvalidSenderID(t *testing.T) {
	handler, _ := getHandlerWithMocks(t)

	ctx := context.Background()

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	req := &p2ppb.GetDocumentRequest{
		DocumentIdentifier: utils.RandomSlice(32),
		AccessType:         p2ppb.AccessType_ACCESS_TYPE_REQUESTER_VERIFICATION,
	}

	encodedReq, err := proto.Marshal(req)
	assert.NoError(t, err)

	env := &p2ppb.Envelope{
		Header: &p2ppb.Header{
			NetworkIdentifier: 36,
			NodeVersion:       version.GetVersion().String(),
			// Invalid sender ID account bytes.
			SenderId:  utils.RandomSlice(11),
			Type:      p2pcommon.MessageTypeGetDoc.String(),
			Timestamp: timestamppb.Now(),
		},
		Body: encodedReq,
	}

	peerID := libp2ppeer.ID("peer-id")
	protocolID := p2pcommon.ProtocolForIdentity(identity)

	res, err := handler.HandleGetDocument(ctx, peerID, protocolID, env)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	assertErrorEnvelope(t, res)
}

func TestHandler_HandleGetDocument_GetCurrentVersionError(t *testing.T) {
	handler, mocks := getHandlerWithMocks(t)

	ctx := context.Background()

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	senderAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	req := &p2ppb.GetDocumentRequest{
		DocumentIdentifier: utils.RandomSlice(32),
		AccessType:         p2ppb.AccessType_ACCESS_TYPE_REQUESTER_VERIFICATION,
	}

	encodedReq, err := proto.Marshal(req)
	assert.NoError(t, err)

	env := &p2ppb.Envelope{
		Header: &p2ppb.Header{
			NetworkIdentifier: 36,
			NodeVersion:       version.GetVersion().String(),
			SenderId:          senderAccountID.ToBytes(),
			Type:              p2pcommon.MessageTypeGetDoc.String(),
			Timestamp:         timestamppb.Now(),
		},
		Body: encodedReq,
	}

	peerID := libp2ppeer.ID("peer-id")
	protocolID := p2pcommon.ProtocolForIdentity(identity)

	genericUtils.GetMock[*documents.ServiceMock](mocks).
		On(
			"GetCurrentVersion",
			mock.Anything,
			req.GetDocumentIdentifier(),
		).
		Return(nil, errors.New("error")).Once()

	res, err := handler.HandleGetDocument(ctx, peerID, protocolID, env)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	assertErrorEnvelope(t, res)
}

func TestHandler_HandleGetDocument_PackCoreDocumentError(t *testing.T) {
	handler, mocks := getHandlerWithMocks(t)

	ctx := context.Background()

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	senderAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	req := &p2ppb.GetDocumentRequest{
		DocumentIdentifier: utils.RandomSlice(32),
		AccessType:         p2ppb.AccessType_ACCESS_TYPE_REQUESTER_VERIFICATION,
	}

	encodedReq, err := proto.Marshal(req)
	assert.NoError(t, err)

	env := &p2ppb.Envelope{
		Header: &p2ppb.Header{
			NetworkIdentifier: 36,
			NodeVersion:       version.GetVersion().String(),
			SenderId:          senderAccountID.ToBytes(),
			Type:              p2pcommon.MessageTypeGetDoc.String(),
			Timestamp:         timestamppb.Now(),
		},
		Body: encodedReq,
	}

	peerID := libp2ppeer.ID("peer-id")
	protocolID := p2pcommon.ProtocolForIdentity(identity)

	documentMock := documents.NewDocumentMock(t)

	genericUtils.GetMock[*documents.ServiceMock](mocks).
		On(
			"GetCurrentVersion",
			mock.Anything,
			req.GetDocumentIdentifier(),
		).
		Return(documentMock, nil).Once()

	documentMock.On("AccountCanRead", senderAccountID).
		Return(true).
		Once()

	documentMock.On("PackCoreDocument").
		Return(nil, errors.New("error")).
		Once()

	res, err := handler.HandleGetDocument(ctx, peerID, protocolID, env)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	assertErrorEnvelope(t, res)
}

func TestHandler_HandleGetDocument_RequesterVerification_AccountCannotRead(t *testing.T) {
	handler, mocks := getHandlerWithMocks(t)

	ctx := context.Background()

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	senderAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	req := &p2ppb.GetDocumentRequest{
		DocumentIdentifier: utils.RandomSlice(32),
		AccessType:         p2ppb.AccessType_ACCESS_TYPE_REQUESTER_VERIFICATION,
	}

	encodedReq, err := proto.Marshal(req)
	assert.NoError(t, err)

	env := &p2ppb.Envelope{
		Header: &p2ppb.Header{
			NetworkIdentifier: 36,
			NodeVersion:       version.GetVersion().String(),
			SenderId:          senderAccountID.ToBytes(),
			Type:              p2pcommon.MessageTypeGetDoc.String(),
			Timestamp:         timestamppb.Now(),
		},
		Body: encodedReq,
	}

	peerID := libp2ppeer.ID("peer-id")
	protocolID := p2pcommon.ProtocolForIdentity(identity)

	documentMock := documents.NewDocumentMock(t)

	genericUtils.GetMock[*documents.ServiceMock](mocks).
		On(
			"GetCurrentVersion",
			mock.Anything,
			req.GetDocumentIdentifier(),
		).
		Return(documentMock, nil).Once()

	documentMock.On("AccountCanRead", senderAccountID).
		Return(false).
		Once()

	res, err := handler.HandleGetDocument(ctx, peerID, protocolID, env)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	assertErrorEnvelope(t, res)
}

func TestHandler_HandleGetDocument_NFTOwnerVerification_AccountCannotRead(t *testing.T) {
	handler, mocks := getHandlerWithMocks(t)

	ctx := context.Background()

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	senderAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	collectionID := types.U64(1111)
	encodedCollectionID, err := codec.Encode(collectionID)
	assert.NoError(t, err)

	itemID := types.NewU128(*big.NewInt(2222))
	encodedItemID, err := codec.Encode(itemID)
	assert.NoError(t, err)

	req := &p2ppb.GetDocumentRequest{
		DocumentIdentifier: utils.RandomSlice(32),
		AccessType:         p2ppb.AccessType_ACCESS_TYPE_NFT_OWNER_VERIFICATION,
		NftCollectionId:    encodedCollectionID,
		NftItemId:          encodedItemID,
	}

	encodedReq, err := proto.Marshal(req)
	assert.NoError(t, err)

	env := &p2ppb.Envelope{
		Header: &p2ppb.Header{
			NetworkIdentifier: 36,
			NodeVersion:       version.GetVersion().String(),
			SenderId:          senderAccountID.ToBytes(),
			Type:              p2pcommon.MessageTypeGetDoc.String(),
			Timestamp:         timestamppb.Now(),
		},
		Body: encodedReq,
	}

	peerID := libp2ppeer.ID("peer-id")
	protocolID := p2pcommon.ProtocolForIdentity(identity)

	documentMock := documents.NewDocumentMock(t)

	genericUtils.GetMock[*documents.ServiceMock](mocks).
		On(
			"GetCurrentVersion",
			mock.Anything,
			req.GetDocumentIdentifier(),
		).
		Return(documentMock, nil).Once()

	documentMock.On("AccountCanRead", senderAccountID).
		Return(false).
		Once()

	res, err := handler.HandleGetDocument(ctx, peerID, protocolID, env)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	assertErrorEnvelope(t, res)
}

func TestHandler_HandleGetDocument_NFTOwnerVerification_NFTCannotRead(t *testing.T) {
	handler, mocks := getHandlerWithMocks(t)

	ctx := context.Background()

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	senderAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	collectionID := types.U64(1111)
	encodedCollectionID, err := codec.Encode(collectionID)
	assert.NoError(t, err)

	itemID := types.NewU128(*big.NewInt(2222))
	encodedItemID, err := codec.Encode(itemID)
	assert.NoError(t, err)

	req := &p2ppb.GetDocumentRequest{
		DocumentIdentifier: utils.RandomSlice(32),
		AccessType:         p2ppb.AccessType_ACCESS_TYPE_NFT_OWNER_VERIFICATION,
		NftCollectionId:    encodedCollectionID,
		NftItemId:          encodedItemID,
	}

	encodedReq, err := proto.Marshal(req)
	assert.NoError(t, err)

	env := &p2ppb.Envelope{
		Header: &p2ppb.Header{
			NetworkIdentifier: 36,
			NodeVersion:       version.GetVersion().String(),
			SenderId:          senderAccountID.ToBytes(),
			Type:              p2pcommon.MessageTypeGetDoc.String(),
			Timestamp:         timestamppb.Now(),
		},
		Body: encodedReq,
	}

	peerID := libp2ppeer.ID("peer-id")
	protocolID := p2pcommon.ProtocolForIdentity(identity)

	documentMock := documents.NewDocumentMock(t)

	genericUtils.GetMock[*documents.ServiceMock](mocks).
		On(
			"GetCurrentVersion",
			mock.Anything,
			req.GetDocumentIdentifier(),
		).
		Return(documentMock, nil).Once()

	documentMock.On("AccountCanRead", senderAccountID).
		Return(true).
		Once()

	documentMock.On("NFTCanRead", encodedCollectionID, encodedItemID).
		Return(false).
		Once()

	res, err := handler.HandleGetDocument(ctx, peerID, protocolID, env)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	assertErrorEnvelope(t, res)
}

func TestHandler_HandleGetDocument_NFTOwnerVerification_InvalidNFTCollectionID(t *testing.T) {
	handler, mocks := getHandlerWithMocks(t)

	ctx := context.Background()

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	senderAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	itemID := types.NewU128(*big.NewInt(2222))
	encodedItemID, err := codec.Encode(itemID)
	assert.NoError(t, err)

	req := &p2ppb.GetDocumentRequest{
		DocumentIdentifier: utils.RandomSlice(32),
		AccessType:         p2ppb.AccessType_ACCESS_TYPE_NFT_OWNER_VERIFICATION,
		// Invalid collection ID bytes.
		NftCollectionId: []byte{1},
		NftItemId:       encodedItemID,
	}

	encodedReq, err := proto.Marshal(req)
	assert.NoError(t, err)

	env := &p2ppb.Envelope{
		Header: &p2ppb.Header{
			NetworkIdentifier: 36,
			NodeVersion:       version.GetVersion().String(),
			SenderId:          senderAccountID.ToBytes(),
			Type:              p2pcommon.MessageTypeGetDoc.String(),
			Timestamp:         timestamppb.Now(),
		},
		Body: encodedReq,
	}

	peerID := libp2ppeer.ID("peer-id")
	protocolID := p2pcommon.ProtocolForIdentity(identity)

	documentMock := documents.NewDocumentMock(t)

	genericUtils.GetMock[*documents.ServiceMock](mocks).
		On(
			"GetCurrentVersion",
			mock.Anything,
			req.GetDocumentIdentifier(),
		).
		Return(documentMock, nil).Once()

	documentMock.On("AccountCanRead", senderAccountID).
		Return(true).
		Once()

	documentMock.On("NFTCanRead", req.GetNftCollectionId(), req.GetNftItemId()).
		Return(true).
		Once()

	res, err := handler.HandleGetDocument(ctx, peerID, protocolID, env)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	assertErrorEnvelope(t, res)
}

func TestHandler_HandleGetDocument_NFTOwnerVerification_InvalidNFTItemID(t *testing.T) {
	handler, mocks := getHandlerWithMocks(t)

	ctx := context.Background()

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	senderAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	collectionID := types.U64(1111)
	encodedCollectionID, err := codec.Encode(collectionID)
	assert.NoError(t, err)

	req := &p2ppb.GetDocumentRequest{
		DocumentIdentifier: utils.RandomSlice(32),
		AccessType:         p2ppb.AccessType_ACCESS_TYPE_NFT_OWNER_VERIFICATION,
		NftCollectionId:    encodedCollectionID,
		// Invalid item ID bytes.
		NftItemId: []byte{1},
	}

	encodedReq, err := proto.Marshal(req)
	assert.NoError(t, err)

	env := &p2ppb.Envelope{
		Header: &p2ppb.Header{
			NetworkIdentifier: 36,
			NodeVersion:       version.GetVersion().String(),
			SenderId:          senderAccountID.ToBytes(),
			Type:              p2pcommon.MessageTypeGetDoc.String(),
			Timestamp:         timestamppb.Now(),
		},
		Body: encodedReq,
	}

	peerID := libp2ppeer.ID("peer-id")
	protocolID := p2pcommon.ProtocolForIdentity(identity)

	documentMock := documents.NewDocumentMock(t)

	genericUtils.GetMock[*documents.ServiceMock](mocks).
		On(
			"GetCurrentVersion",
			mock.Anything,
			req.GetDocumentIdentifier(),
		).
		Return(documentMock, nil).Once()

	documentMock.On("AccountCanRead", senderAccountID).
		Return(true).
		Once()

	documentMock.On("NFTCanRead", req.GetNftCollectionId(), req.GetNftItemId()).
		Return(true).
		Once()

	res, err := handler.HandleGetDocument(ctx, peerID, protocolID, env)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	assertErrorEnvelope(t, res)
}

func TestHandler_HandleGetDocument_NFTOwnerVerification_NFTOwnerRetrievalError(t *testing.T) {
	handler, mocks := getHandlerWithMocks(t)

	ctx := context.Background()

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	senderAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	collectionID := types.U64(1111)
	encodedCollectionID, err := codec.Encode(collectionID)
	assert.NoError(t, err)

	itemID := types.NewU128(*big.NewInt(2222))
	encodedItemID, err := codec.Encode(itemID)
	assert.NoError(t, err)

	req := &p2ppb.GetDocumentRequest{
		DocumentIdentifier: utils.RandomSlice(32),
		AccessType:         p2ppb.AccessType_ACCESS_TYPE_NFT_OWNER_VERIFICATION,
		NftCollectionId:    encodedCollectionID,
		NftItemId:          encodedItemID,
	}

	encodedReq, err := proto.Marshal(req)
	assert.NoError(t, err)

	env := &p2ppb.Envelope{
		Header: &p2ppb.Header{
			NetworkIdentifier: 36,
			NodeVersion:       version.GetVersion().String(),
			SenderId:          senderAccountID.ToBytes(),
			Type:              p2pcommon.MessageTypeGetDoc.String(),
			Timestamp:         timestamppb.Now(),
		},
		Body: encodedReq,
	}

	peerID := libp2ppeer.ID("peer-id")
	protocolID := p2pcommon.ProtocolForIdentity(identity)

	documentMock := documents.NewDocumentMock(t)

	genericUtils.GetMock[*documents.ServiceMock](mocks).
		On(
			"GetCurrentVersion",
			mock.Anything,
			req.GetDocumentIdentifier(),
		).
		Return(documentMock, nil).Once()

	documentMock.On("AccountCanRead", senderAccountID).
		Return(true).
		Once()

	documentMock.On("NFTCanRead", encodedCollectionID, encodedItemID).
		Return(true).
		Once()

	genericUtils.GetMock[*nftv3.ServiceMock](mocks).
		On("GetNFTOwner", collectionID, itemID).
		Return(nil, errors.New("error")).Once()

	res, err := handler.HandleGetDocument(ctx, peerID, protocolID, env)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	assertErrorEnvelope(t, res)
}

func TestHandler_HandleGetDocument_NFTOwnerVerification_NFTOwnerDifference(t *testing.T) {
	handler, mocks := getHandlerWithMocks(t)

	ctx := context.Background()

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	senderAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	collectionID := types.U64(1111)
	encodedCollectionID, err := codec.Encode(collectionID)
	assert.NoError(t, err)

	itemID := types.NewU128(*big.NewInt(2222))
	encodedItemID, err := codec.Encode(itemID)
	assert.NoError(t, err)

	req := &p2ppb.GetDocumentRequest{
		DocumentIdentifier: utils.RandomSlice(32),
		AccessType:         p2ppb.AccessType_ACCESS_TYPE_NFT_OWNER_VERIFICATION,
		NftCollectionId:    encodedCollectionID,
		NftItemId:          encodedItemID,
	}

	encodedReq, err := proto.Marshal(req)
	assert.NoError(t, err)

	env := &p2ppb.Envelope{
		Header: &p2ppb.Header{
			NetworkIdentifier: 36,
			NodeVersion:       version.GetVersion().String(),
			SenderId:          senderAccountID.ToBytes(),
			Type:              p2pcommon.MessageTypeGetDoc.String(),
			Timestamp:         timestamppb.Now(),
		},
		Body: encodedReq,
	}

	peerID := libp2ppeer.ID("peer-id")
	protocolID := p2pcommon.ProtocolForIdentity(identity)

	documentMock := documents.NewDocumentMock(t)

	genericUtils.GetMock[*documents.ServiceMock](mocks).
		On(
			"GetCurrentVersion",
			mock.Anything,
			req.GetDocumentIdentifier(),
		).
		Return(documentMock, nil).Once()

	documentMock.On("AccountCanRead", senderAccountID).
		Return(true).
		Once()

	documentMock.On("NFTCanRead", encodedCollectionID, encodedItemID).
		Return(true).
		Once()

	randomAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	genericUtils.GetMock[*nftv3.ServiceMock](mocks).
		On("GetNFTOwner", collectionID, itemID).
		Return(randomAccountID, nil).Once()

	res, err := handler.HandleGetDocument(ctx, peerID, protocolID, env)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	assertErrorEnvelope(t, res)
}

func TestHandler_HandleGetDocument_AccessTokenVerification_NilAccessTokenRequest(t *testing.T) {
	handler, mocks := getHandlerWithMocks(t)

	ctx := context.Background()

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	senderAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	req := &p2ppb.GetDocumentRequest{
		DocumentIdentifier: utils.RandomSlice(32),
		AccessType:         p2ppb.AccessType_ACCESS_TYPE_ACCESS_TOKEN_VERIFICATION,
	}

	encodedReq, err := proto.Marshal(req)
	assert.NoError(t, err)

	env := &p2ppb.Envelope{
		Header: &p2ppb.Header{
			NetworkIdentifier: 36,
			NodeVersion:       version.GetVersion().String(),
			SenderId:          senderAccountID.ToBytes(),
			Type:              p2pcommon.MessageTypeGetDoc.String(),
			Timestamp:         timestamppb.Now(),
		},
		Body: encodedReq,
	}

	peerID := libp2ppeer.ID("peer-id")
	protocolID := p2pcommon.ProtocolForIdentity(identity)

	documentMock := documents.NewDocumentMock(t)

	genericUtils.GetMock[*documents.ServiceMock](mocks).
		On(
			"GetCurrentVersion",
			mock.Anything,
			req.GetDocumentIdentifier(),
		).
		Return(documentMock, nil).Once()

	res, err := handler.HandleGetDocument(ctx, peerID, protocolID, env)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	assertErrorEnvelope(t, res)
}

func TestHandler_HandleGetDocument_AccessTokenVerification_EntityRelationshipRetrievalError(t *testing.T) {
	handler, mocks := getHandlerWithMocks(t)

	ctx := context.Background()

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	senderAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	req := &p2ppb.GetDocumentRequest{
		DocumentIdentifier: utils.RandomSlice(32),
		AccessType:         p2ppb.AccessType_ACCESS_TYPE_ACCESS_TOKEN_VERIFICATION,
		AccessTokenRequest: &p2ppb.AccessTokenRequest{
			DelegatingDocumentIdentifier: utils.RandomSlice(32),
			AccessTokenId:                utils.RandomSlice(32),
		},
	}

	encodedReq, err := proto.Marshal(req)
	assert.NoError(t, err)

	env := &p2ppb.Envelope{
		Header: &p2ppb.Header{
			NetworkIdentifier: 36,
			NodeVersion:       version.GetVersion().String(),
			SenderId:          senderAccountID.ToBytes(),
			Type:              p2pcommon.MessageTypeGetDoc.String(),
			Timestamp:         timestamppb.Now(),
		},
		Body: encodedReq,
	}

	peerID := libp2ppeer.ID("peer-id")
	protocolID := p2pcommon.ProtocolForIdentity(identity)

	documentMock := documents.NewDocumentMock(t)

	genericUtils.GetMock[*documents.ServiceMock](mocks).
		On(
			"GetCurrentVersion",
			mock.Anything,
			req.GetDocumentIdentifier(),
		).
		Return(documentMock, nil).Once()

	genericUtils.GetMock[*documents.ServiceMock](mocks).
		On(
			"GetCurrentVersion",
			mock.Anything,
			req.GetAccessTokenRequest().GetDelegatingDocumentIdentifier(),
		).
		Return(nil, errors.New("error")).Once()

	res, err := handler.HandleGetDocument(ctx, peerID, protocolID, env)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	assertErrorEnvelope(t, res)
}

func TestHandler_HandleGetDocument_AccessTokenVerification_AccessTokenCannotRead(t *testing.T) {
	handler, mocks := getHandlerWithMocks(t)

	ctx := context.Background()

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	senderAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	req := &p2ppb.GetDocumentRequest{
		DocumentIdentifier: utils.RandomSlice(32),
		AccessType:         p2ppb.AccessType_ACCESS_TYPE_ACCESS_TOKEN_VERIFICATION,
		AccessTokenRequest: &p2ppb.AccessTokenRequest{
			DelegatingDocumentIdentifier: utils.RandomSlice(32),
			AccessTokenId:                utils.RandomSlice(32),
		},
	}

	encodedReq, err := proto.Marshal(req)
	assert.NoError(t, err)

	env := &p2ppb.Envelope{
		Header: &p2ppb.Header{
			NetworkIdentifier: 36,
			NodeVersion:       version.GetVersion().String(),
			SenderId:          senderAccountID.ToBytes(),
			Type:              p2pcommon.MessageTypeGetDoc.String(),
			Timestamp:         timestamppb.Now(),
		},
		Body: encodedReq,
	}

	peerID := libp2ppeer.ID("peer-id")
	protocolID := p2pcommon.ProtocolForIdentity(identity)

	documentMock := documents.NewDocumentMock(t)

	genericUtils.GetMock[*documents.ServiceMock](mocks).
		On(
			"GetCurrentVersion",
			mock.Anything,
			req.GetDocumentIdentifier(),
		).
		Return(documentMock, nil).Once()

	entityRelationshipMock := documents.NewDocumentMock(t)

	genericUtils.GetMock[*documents.ServiceMock](mocks).
		On(
			"GetCurrentVersion",
			mock.Anything,
			req.GetAccessTokenRequest().GetDelegatingDocumentIdentifier(),
		).
		Return(entityRelationshipMock, nil).Once()

	entityRelationshipMock.On(
		"ATGranteeCanRead",
		mock.Anything,
		handler.docSrv,
		handler.identityService,
		req.GetAccessTokenRequest().GetAccessTokenId(),
		req.GetDocumentIdentifier(),
		senderAccountID,
	).Return(errors.New("error")).Once()

	res, err := handler.HandleGetDocument(ctx, peerID, protocolID, env)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	assertErrorEnvelope(t, res)
}

func assertErrorEnvelope(t *testing.T, env *protocolpb.P2PEnvelope) {
	var responseEnvelope p2ppb.Envelope

	err := proto.Unmarshal(env.GetBody(), &responseEnvelope)
	assert.NoError(t, err)

	var errorEnv errorspb.Error

	err = proto.Unmarshal(responseEnvelope.GetBody(), &errorEnv)
	assert.NoError(t, err)

	assert.NotEmpty(t, errorEnv.GetMessage())
}

func assertExpectedHeaderMatchesActual(t *testing.T, expected *p2ppb.Header, actual *p2ppb.Header) {
	assert.Equal(t, expected.GetSenderId(), actual.GetSenderId())
	assert.Equal(t, expected.GetNodeVersion(), actual.GetNodeVersion())
	assert.Equal(t, expected.GetNetworkIdentifier(), actual.GetNetworkIdentifier())
	assert.Equal(t, expected.GetType(), actual.GetType())
	assert.Equal(t, expected.GetTimestamp().AsTime(), actual.GetTimestamp().AsTime())
}

func getHandlerWithMocks(t *testing.T) (*handler, []any) {
	cfgMock := config.NewConfigurationMock(t)
	cfgServiceMock := config.NewServiceMock(t)
	validatorMock := NewValidatorMock(t)
	documentServiceMock := documents.NewServiceMock(t)
	identityServiceMock := v2.NewServiceMock(t)
	nftServiceMock := nftv3.NewServiceMock(t)

	h := &handler{
		cfgMock,
		cfgServiceMock,
		validatorMock,
		documentServiceMock,
		identityServiceMock,
		nftServiceMock,
	}

	return h, []any{
		cfgMock,
		cfgServiceMock,
		validatorMock,
		documentServiceMock,
		identityServiceMock,
		nftServiceMock,
	}
}

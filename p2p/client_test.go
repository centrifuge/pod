//go:build unit

package p2p

import (
	"context"
	"testing"
	"time"

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	p2ppb "github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	protocolpb "github.com/centrifuge/centrifuge-protobufs/gen/go/protocol"
	keystoreType "github.com/centrifuge/chain-custom-types/pkg/keystore"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/crypto/ed25519"
	"github.com/centrifuge/go-centrifuge/dispatcher"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	v2 "github.com/centrifuge/go-centrifuge/identity/v2"
	p2pcommon "github.com/centrifuge/go-centrifuge/p2p/common"
	ms "github.com/centrifuge/go-centrifuge/p2p/messenger"
	p2pMocks "github.com/centrifuge/go-centrifuge/p2p/mocks"
	"github.com/centrifuge/go-centrifuge/p2p/receiver"
	"github.com/centrifuge/go-centrifuge/pallets/keystore"
	testingcommons "github.com/centrifuge/go-centrifuge/testingutils/common"
	genericUtils "github.com/centrifuge/go-centrifuge/testingutils/generic"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/go-centrifuge/version"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	libp2ppeer "github.com/libp2p/go-libp2p-core/peer"
	pstore "github.com/libp2p/go-libp2p-core/peerstore"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/proto"
)

func TestPeer_Client_SendAnchoredDocument(t *testing.T) {
	peer, mocks := getPeerMocks(t)

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	senderAccountMock := config.NewAccountMock(t)
	senderAccountMock.On("GetIdentity").
		Return(identity)

	ctx := contextutil.WithAccount(context.Background(), senderAccountMock)

	receiverID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	cd := &coredocumentpb.CoreDocument{}

	req := &p2ppb.AnchorDocumentRequest{
		Document: cd,
	}

	genericUtils.GetMock[*config.ServiceMock](mocks).On("GetAccount", receiverID.ToBytes()).
		Return(nil, errors.New("error")).Once()

	genericUtils.GetMock[*v2.ServiceMock](mocks).On("ValidateAccount", receiverID).
		Return(nil).Once()

	peerID := mockPeerIDRetrievalCalls(t, receiverID, mocks)

	anchorDocRes := &p2ppb.AnchorDocumentResponse{}

	anchorsDocResBytes, err := proto.Marshal(anchorDocRes)
	assert.NoError(t, err)

	envelopeRes := &p2ppb.Envelope{
		Header: &p2ppb.Header{
			NetworkIdentifier: 36,
			NodeVersion:       "test-version",
			SenderId:          utils.RandomSlice(32),
			Type:              p2pcommon.MessageTypeSendAnchoredDocRep.String(),
		},
		Body: anchorsDocResBytes,
	}

	envelopeResBytes, err := proto.Marshal(envelopeRes)
	assert.NoError(t, err)

	protocolEnvelopeRes := &protocolpb.P2PEnvelope{
		Body: envelopeResBytes,
	}

	genericUtils.GetMock[*ms.MessengerMock](mocks).
		On(
			"SendMessage",
			mock.Anything,
			peerID,
			mock.IsType(&protocolpb.P2PEnvelope{}),
			p2pcommon.ProtocolForIdentity(receiverID),
		).
		Run(func(args mock.Arguments) {
			protocolEnv, ok := args.Get(2).(*protocolpb.P2PEnvelope)
			assert.True(t, ok)

			var env p2ppb.Envelope

			err = proto.Unmarshal(protocolEnv.GetBody(), &env)
			assert.NoError(t, err)

			reqBytes, err := proto.Marshal(req)
			assert.NoError(t, err)

			assert.Equal(t, reqBytes, env.GetBody())
			assert.Equal(t, identity.ToBytes(), env.GetHeader().GetSenderId())
			assert.Equal(t, networkID, env.GetHeader().GetNetworkIdentifier())
			assert.Equal(t, p2pcommon.MessageTypeSendAnchoredDoc.String(), env.GetHeader().GetType())
			assert.Equal(t, version.GetVersion().String(), env.GetHeader().GetNodeVersion())
		}).
		Return(protocolEnvelopeRes, nil).Once()

	res, err := peer.SendAnchoredDocument(ctx, receiverID, req)
	assert.NoError(t, err)
	assert.Equal(t, anchorDocRes, res)
}

func TestPeer_Client_SendAnchoredDocument_SenderIdentityError(t *testing.T) {
	peer, _ := getPeerMocks(t)

	ctx := context.Background()

	receiverID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	cd := &coredocumentpb.CoreDocument{}

	req := &p2ppb.AnchorDocumentRequest{
		Document: cd,
	}

	res, err := peer.SendAnchoredDocument(ctx, receiverID, req)
	assert.NotNil(t, err, errors.ErrContextIdentityRetrieval)
	assert.Nil(t, res)
}

func TestPeer_Client_SendAnchoredDocument_ValidateAccountError(t *testing.T) {
	peer, mocks := getPeerMocks(t)

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	senderAccountMock := config.NewAccountMock(t)
	senderAccountMock.On("GetIdentity").
		Return(identity)

	ctx := contextutil.WithAccount(context.Background(), senderAccountMock)

	receiverID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	cd := &coredocumentpb.CoreDocument{}

	req := &p2ppb.AnchorDocumentRequest{
		Document: cd,
	}

	genericUtils.GetMock[*config.ServiceMock](mocks).On("GetAccount", receiverID.ToBytes()).
		Return(nil, errors.New("error")).Once()

	genericUtils.GetMock[*v2.ServiceMock](mocks).On("ValidateAccount", receiverID).
		Return(errors.New("error")).Once()

	res, err := peer.SendAnchoredDocument(ctx, receiverID, req)
	assert.ErrorIs(t, err, ErrInvalidReceiverAccount)
	assert.Nil(t, res)
}

func TestPeer_Client_SendAnchoredDocument_PeerIDRetrievalError(t *testing.T) {
	peer, mocks := getPeerMocks(t)

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	senderAccountMock := config.NewAccountMock(t)
	senderAccountMock.On("GetIdentity").
		Return(identity)

	ctx := contextutil.WithAccount(context.Background(), senderAccountMock)

	receiverID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	cd := &coredocumentpb.CoreDocument{}

	req := &p2ppb.AnchorDocumentRequest{
		Document: cd,
	}

	genericUtils.GetMock[*config.ServiceMock](mocks).On("GetAccount", receiverID.ToBytes()).
		Return(nil, errors.New("error")).Once()

	genericUtils.GetMock[*v2.ServiceMock](mocks).On("ValidateAccount", receiverID).
		Return(nil).Once()

	genericUtils.GetMock[*keystore.APIMock](mocks).
		On(
			"GetLastKeyByPurpose",
			receiverID,
			keystoreType.KeyPurposeP2PDiscovery,
		).
		Return(nil, errors.New("error")).Once()

	res, err := peer.SendAnchoredDocument(ctx, receiverID, req)
	assert.ErrorIs(t, err, ErrPeerIDRetrieval)
	assert.Nil(t, res)
}

func TestPeer_Client_SendAnchoredDocument_MessengerError(t *testing.T) {
	peer, mocks := getPeerMocks(t)

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	senderAccountMock := config.NewAccountMock(t)
	senderAccountMock.On("GetIdentity").
		Return(identity)

	ctx := contextutil.WithAccount(context.Background(), senderAccountMock)

	receiverID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	cd := &coredocumentpb.CoreDocument{}

	req := &p2ppb.AnchorDocumentRequest{
		Document: cd,
	}

	genericUtils.GetMock[*config.ServiceMock](mocks).On("GetAccount", receiverID.ToBytes()).
		Return(nil, errors.New("error")).Once()

	genericUtils.GetMock[*v2.ServiceMock](mocks).On("ValidateAccount", receiverID).
		Return(nil).Once()

	peerID := mockPeerIDRetrievalCalls(t, receiverID, mocks)

	genericUtils.GetMock[*ms.MessengerMock](mocks).
		On(
			"SendMessage",
			mock.Anything,
			peerID,
			mock.IsType(&protocolpb.P2PEnvelope{}),
			p2pcommon.ProtocolForIdentity(receiverID),
		).
		Run(func(args mock.Arguments) {
			protocolEnv, ok := args.Get(2).(*protocolpb.P2PEnvelope)
			assert.True(t, ok)

			var env p2ppb.Envelope

			err = proto.Unmarshal(protocolEnv.GetBody(), &env)
			assert.NoError(t, err)

			reqBytes, err := proto.Marshal(req)
			assert.NoError(t, err)

			assert.Equal(t, reqBytes, env.GetBody())
			assert.Equal(t, identity.ToBytes(), env.GetHeader().GetSenderId())
			assert.Equal(t, networkID, env.GetHeader().GetNetworkIdentifier())
			assert.Equal(t, p2pcommon.MessageTypeSendAnchoredDoc.String(), env.GetHeader().GetType())
			assert.Equal(t, version.GetVersion().String(), env.GetHeader().GetNodeVersion())
		}).
		Return(nil, errors.New("error")).Once()

	res, err := peer.SendAnchoredDocument(ctx, receiverID, req)
	assert.ErrorIs(t, err, ErrP2PMessageSending)
	assert.Nil(t, res)
}

func TestPeer_Client_SendAnchoredDocument_ResolveEnvelopeError(t *testing.T) {
	peer, mocks := getPeerMocks(t)

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	senderAccountMock := config.NewAccountMock(t)
	senderAccountMock.On("GetIdentity").
		Return(identity)

	ctx := contextutil.WithAccount(context.Background(), senderAccountMock)

	receiverID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	cd := &coredocumentpb.CoreDocument{}

	req := &p2ppb.AnchorDocumentRequest{
		Document: cd,
	}

	genericUtils.GetMock[*config.ServiceMock](mocks).On("GetAccount", receiverID.ToBytes()).
		Return(nil, errors.New("error")).Once()

	genericUtils.GetMock[*v2.ServiceMock](mocks).On("ValidateAccount", receiverID).
		Return(nil).Once()

	peerID := mockPeerIDRetrievalCalls(t, receiverID, mocks)

	// Set invalid body to ensure an error when resolving the envelope.
	protocolEnvelopeRes := &protocolpb.P2PEnvelope{
		Body: utils.RandomSlice(32),
	}

	genericUtils.GetMock[*ms.MessengerMock](mocks).
		On(
			"SendMessage",
			mock.Anything,
			peerID,
			mock.IsType(&protocolpb.P2PEnvelope{}),
			p2pcommon.ProtocolForIdentity(receiverID),
		).
		Run(func(args mock.Arguments) {
			protocolEnv, ok := args.Get(2).(*protocolpb.P2PEnvelope)
			assert.True(t, ok)

			var env p2ppb.Envelope

			err = proto.Unmarshal(protocolEnv.GetBody(), &env)
			assert.NoError(t, err)

			reqBytes, err := proto.Marshal(req)
			assert.NoError(t, err)

			assert.Equal(t, reqBytes, env.GetBody())
			assert.Equal(t, identity.ToBytes(), env.GetHeader().GetSenderId())
			assert.Equal(t, networkID, env.GetHeader().GetNetworkIdentifier())
			assert.Equal(t, p2pcommon.MessageTypeSendAnchoredDoc.String(), env.GetHeader().GetType())
			assert.Equal(t, version.GetVersion().String(), env.GetHeader().GetNodeVersion())
		}).
		Return(protocolEnvelopeRes, nil).Once()

	res, err := peer.SendAnchoredDocument(ctx, receiverID, req)
	assert.ErrorIs(t, err, ErrP2PDataEnvelopeResolving)
	assert.Nil(t, res)
}

func TestPeer_Client_SendAnchoredDocument_ReceiveEnvelopeError(t *testing.T) {
	peer, mocks := getPeerMocks(t)

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	senderAccountMock := config.NewAccountMock(t)
	senderAccountMock.On("GetIdentity").
		Return(identity)

	ctx := contextutil.WithAccount(context.Background(), senderAccountMock)

	receiverID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	cd := &coredocumentpb.CoreDocument{}

	req := &p2ppb.AnchorDocumentRequest{
		Document: cd,
	}

	genericUtils.GetMock[*config.ServiceMock](mocks).On("GetAccount", receiverID.ToBytes()).
		Return(nil, errors.New("error")).Once()

	genericUtils.GetMock[*v2.ServiceMock](mocks).On("ValidateAccount", receiverID).
		Return(nil).Once()

	peerID := mockPeerIDRetrievalCalls(t, receiverID, mocks)

	envelopeRes := &p2ppb.Envelope{
		Header: &p2ppb.Header{
			NetworkIdentifier: networkID,
			NodeVersion:       "test-version",
			SenderId:          utils.RandomSlice(32),
			// Error type.
			Type: p2pcommon.MessageTypeError.String(),
		},
		Body: nil,
	}

	envelopeResBytes, err := proto.Marshal(envelopeRes)
	assert.NoError(t, err)

	protocolEnvelopeRes := &protocolpb.P2PEnvelope{
		Body: envelopeResBytes,
	}

	genericUtils.GetMock[*ms.MessengerMock](mocks).
		On(
			"SendMessage",
			mock.Anything,
			peerID,
			mock.IsType(&protocolpb.P2PEnvelope{}),
			p2pcommon.ProtocolForIdentity(receiverID),
		).
		Run(func(args mock.Arguments) {
			protocolEnv, ok := args.Get(2).(*protocolpb.P2PEnvelope)
			assert.True(t, ok)

			var env p2ppb.Envelope

			err = proto.Unmarshal(protocolEnv.GetBody(), &env)
			assert.NoError(t, err)

			reqBytes, err := proto.Marshal(req)
			assert.NoError(t, err)

			assert.Equal(t, reqBytes, env.GetBody())
			assert.Equal(t, identity.ToBytes(), env.GetHeader().GetSenderId())
			assert.Equal(t, networkID, env.GetHeader().GetNetworkIdentifier())
			assert.Equal(t, p2pcommon.MessageTypeSendAnchoredDoc.String(), env.GetHeader().GetType())
			assert.Equal(t, version.GetVersion().String(), env.GetHeader().GetNodeVersion())
		}).
		Return(protocolEnvelopeRes, nil).Once()

	res, err := peer.SendAnchoredDocument(ctx, receiverID, req)
	assert.True(t, errors.IsOfType(ErrP2PClient, err))
	assert.Nil(t, res)
}

func TestPeer_Client_SendAnchoredDocument_ReceiveEnvelopeInvalidType(t *testing.T) {
	peer, mocks := getPeerMocks(t)

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	senderAccountMock := config.NewAccountMock(t)
	senderAccountMock.On("GetIdentity").
		Return(identity)

	ctx := contextutil.WithAccount(context.Background(), senderAccountMock)

	receiverID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	cd := &coredocumentpb.CoreDocument{}

	req := &p2ppb.AnchorDocumentRequest{
		Document: cd,
	}

	genericUtils.GetMock[*config.ServiceMock](mocks).On("GetAccount", receiverID.ToBytes()).
		Return(nil, errors.New("error")).Once()

	genericUtils.GetMock[*v2.ServiceMock](mocks).On("ValidateAccount", receiverID).
		Return(nil).Once()

	peerID := mockPeerIDRetrievalCalls(t, receiverID, mocks)

	anchorDocRes := &p2ppb.AnchorDocumentResponse{}

	anchorsDocResBytes, err := proto.Marshal(anchorDocRes)
	assert.NoError(t, err)

	envelopeRes := &p2ppb.Envelope{
		Header: &p2ppb.Header{
			NetworkIdentifier: networkID,
			NodeVersion:       "test-version",
			SenderId:          utils.RandomSlice(32),
			// Invalid type.
			Type: p2pcommon.MessageTypeSendAnchoredDoc.String(),
		},
		Body: anchorsDocResBytes,
	}

	envelopeResBytes, err := proto.Marshal(envelopeRes)
	assert.NoError(t, err)

	protocolEnvelopeRes := &protocolpb.P2PEnvelope{
		Body: envelopeResBytes,
	}

	genericUtils.GetMock[*ms.MessengerMock](mocks).
		On(
			"SendMessage",
			mock.Anything,
			peerID,
			mock.IsType(&protocolpb.P2PEnvelope{}),
			p2pcommon.ProtocolForIdentity(receiverID),
		).
		Run(func(args mock.Arguments) {
			protocolEnv, ok := args.Get(2).(*protocolpb.P2PEnvelope)
			assert.True(t, ok)

			var env p2ppb.Envelope

			err = proto.Unmarshal(protocolEnv.GetBody(), &env)
			assert.NoError(t, err)

			reqBytes, err := proto.Marshal(req)
			assert.NoError(t, err)

			assert.Equal(t, reqBytes, env.GetBody())
			assert.Equal(t, identity.ToBytes(), env.GetHeader().GetSenderId())
			assert.Equal(t, networkID, env.GetHeader().GetNetworkIdentifier())
			assert.Equal(t, p2pcommon.MessageTypeSendAnchoredDoc.String(), env.GetHeader().GetType())
			assert.Equal(t, version.GetVersion().String(), env.GetHeader().GetNodeVersion())
		}).
		Return(protocolEnvelopeRes, nil).Once()

	res, err := peer.SendAnchoredDocument(ctx, receiverID, req)
	assert.ErrorIs(t, err, ErrIncorrectResponseMessageType)
	assert.Nil(t, res)
}

func TestPeer_Client_SendAnchoredDocument_ResponseDecodeError(t *testing.T) {
	peer, mocks := getPeerMocks(t)

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	senderAccountMock := config.NewAccountMock(t)
	senderAccountMock.On("GetIdentity").
		Return(identity)

	ctx := contextutil.WithAccount(context.Background(), senderAccountMock)

	receiverID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	cd := &coredocumentpb.CoreDocument{}

	req := &p2ppb.AnchorDocumentRequest{
		Document: cd,
	}

	genericUtils.GetMock[*config.ServiceMock](mocks).On("GetAccount", receiverID.ToBytes()).
		Return(nil, errors.New("error")).Once()

	genericUtils.GetMock[*v2.ServiceMock](mocks).On("ValidateAccount", receiverID).
		Return(nil).Once()

	peerID := mockPeerIDRetrievalCalls(t, receiverID, mocks)

	envelopeRes := &p2ppb.Envelope{
		Header: &p2ppb.Header{
			NetworkIdentifier: networkID,
			NodeVersion:       "test-version",
			SenderId:          utils.RandomSlice(32),
			Type:              p2pcommon.MessageTypeSendAnchoredDocRep.String(),
		},
		// Not a valid a p2ppb.AnchorDocumentResponse
		Body: utils.RandomSlice(32),
	}

	envelopeResBytes, err := proto.Marshal(envelopeRes)
	assert.NoError(t, err)

	protocolEnvelopeRes := &protocolpb.P2PEnvelope{
		Body: envelopeResBytes,
	}

	genericUtils.GetMock[*ms.MessengerMock](mocks).
		On(
			"SendMessage",
			mock.Anything,
			peerID,
			mock.IsType(&protocolpb.P2PEnvelope{}),
			p2pcommon.ProtocolForIdentity(receiverID),
		).
		Run(func(args mock.Arguments) {
			protocolEnv, ok := args.Get(2).(*protocolpb.P2PEnvelope)
			assert.True(t, ok)

			var env p2ppb.Envelope

			err = proto.Unmarshal(protocolEnv.GetBody(), &env)
			assert.NoError(t, err)

			reqBytes, err := proto.Marshal(req)
			assert.NoError(t, err)

			assert.Equal(t, reqBytes, env.GetBody())
			assert.Equal(t, identity.ToBytes(), env.GetHeader().GetSenderId())
			assert.Equal(t, networkID, env.GetHeader().GetNetworkIdentifier())
			assert.Equal(t, p2pcommon.MessageTypeSendAnchoredDoc.String(), env.GetHeader().GetType())
			assert.Equal(t, version.GetVersion().String(), env.GetHeader().GetNodeVersion())
		}).
		Return(protocolEnvelopeRes, nil).Once()

	res, err := peer.SendAnchoredDocument(ctx, receiverID, req)
	assert.ErrorIs(t, err, ErrResponseDecodeError)
	assert.Nil(t, res)
}

func TestPeer_Client_SendAnchoredDocument_LocalAccount(t *testing.T) {
	peer, mocks := getPeerMocks(t)

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	senderAccountMock := config.NewAccountMock(t)
	senderAccountMock.On("GetIdentity").
		Return(identity)

	ctx := contextutil.WithAccount(context.Background(), senderAccountMock)

	receiverID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	cd := &coredocumentpb.CoreDocument{}

	req := &p2ppb.AnchorDocumentRequest{
		Document: cd,
	}

	p2pConnTimeout := 1 * time.Second

	genericUtils.GetMock[*config.ConfigurationMock](mocks).On("GetP2PConnectionTimeout").
		Return(p2pConnTimeout).Once()

	receiverAccountMock := config.NewAccountMock(t)

	genericUtils.GetMock[*config.ServiceMock](mocks).On("GetAccount", receiverID.ToBytes()).
		Return(receiverAccountMock, nil).Once()

	anchorDocRes := &p2ppb.AnchorDocumentResponse{}

	genericUtils.GetMock[*receiver.HandlerMock](mocks).
		On(
			"SendAnchoredDocument",
			mock.Anything,
			req,
			identity,
		).
		Run(func(args mock.Arguments) {
			handlerCtx, ok := args.Get(0).(context.Context)
			assert.True(t, ok)

			handlerCtxAccount, err := contextutil.Account(handlerCtx)
			assert.NoError(t, err)

			assert.Equal(t, receiverAccountMock, handlerCtxAccount)
		}).
		Return(anchorDocRes, nil).
		Once()

	res, err := peer.SendAnchoredDocument(ctx, receiverID, req)
	assert.NoError(t, err)
	assert.Equal(t, anchorDocRes, res)
}

func TestPeer_Client_SendAnchoredDocument_LocalAccount_HandlerError(t *testing.T) {
	peer, mocks := getPeerMocks(t)

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	senderAccountMock := config.NewAccountMock(t)
	senderAccountMock.On("GetIdentity").
		Return(identity)

	ctx := contextutil.WithAccount(context.Background(), senderAccountMock)

	receiverID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	cd := &coredocumentpb.CoreDocument{}

	req := &p2ppb.AnchorDocumentRequest{
		Document: cd,
	}

	p2pConnTimeout := 1 * time.Second

	genericUtils.GetMock[*config.ConfigurationMock](mocks).On("GetP2PConnectionTimeout").
		Return(p2pConnTimeout).Once()

	receiverAccountMock := config.NewAccountMock(t)

	genericUtils.GetMock[*config.ServiceMock](mocks).On("GetAccount", receiverID.ToBytes()).
		Return(receiverAccountMock, nil).Once()

	handlerErr := errors.New("error")

	genericUtils.GetMock[*receiver.HandlerMock](mocks).
		On(
			"SendAnchoredDocument",
			mock.Anything,
			req,
			identity,
		).
		Run(func(args mock.Arguments) {
			handlerCtx, ok := args.Get(0).(context.Context)
			assert.True(t, ok)

			handlerCtxAccount, err := contextutil.Account(handlerCtx)
			assert.NoError(t, err)

			assert.Equal(t, receiverAccountMock, handlerCtxAccount)
		}).
		Return(nil, handlerErr).
		Once()

	res, err := peer.SendAnchoredDocument(ctx, receiverID, req)
	assert.ErrorIs(t, err, handlerErr)
	assert.Nil(t, res)
}

func TestPeer_Client_GetDocumentRequest(t *testing.T) {
	peer, mocks := getPeerMocks(t)

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	senderAccountMock := config.NewAccountMock(t)
	senderAccountMock.On("GetIdentity").
		Return(identity)

	ctx := contextutil.WithAccount(context.Background(), senderAccountMock)

	requesterID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	req := &p2ppb.GetDocumentRequest{
		DocumentIdentifier: utils.RandomSlice(32),
		AccessType:         p2ppb.AccessType_ACCESS_TYPE_REQUESTER_VERIFICATION,
	}

	genericUtils.GetMock[*config.ServiceMock](mocks).On("GetAccount", requesterID.ToBytes()).
		Return(nil, errors.New("error")).Once()

	genericUtils.GetMock[*v2.ServiceMock](mocks).On("ValidateAccount", requesterID).
		Return(nil).Once()

	peerID := mockPeerIDRetrievalCalls(t, requesterID, mocks)

	getDocRes := &p2ppb.GetDocumentResponse{}

	getDocResBytes, err := proto.Marshal(getDocRes)
	assert.NoError(t, err)

	envelopeRes := &p2ppb.Envelope{
		Header: &p2ppb.Header{
			NetworkIdentifier: networkID,
			NodeVersion:       "test-version",
			SenderId:          utils.RandomSlice(32),
			Type:              p2pcommon.MessageTypeGetDocRep.String(),
		},
		Body: getDocResBytes,
	}

	envelopeResBytes, err := proto.Marshal(envelopeRes)
	assert.NoError(t, err)

	protocolEnvelopeRes := &protocolpb.P2PEnvelope{
		Body: envelopeResBytes,
	}

	genericUtils.GetMock[*ms.MessengerMock](mocks).
		On(
			"SendMessage",
			mock.Anything,
			peerID,
			mock.IsType(&protocolpb.P2PEnvelope{}),
			p2pcommon.ProtocolForIdentity(requesterID),
		).
		Run(func(args mock.Arguments) {
			protocolEnv, ok := args.Get(2).(*protocolpb.P2PEnvelope)
			assert.True(t, ok)

			var env p2ppb.Envelope

			err = proto.Unmarshal(protocolEnv.GetBody(), &env)
			assert.NoError(t, err)

			reqBytes, err := proto.Marshal(req)
			assert.NoError(t, err)

			assert.Equal(t, reqBytes, env.GetBody())
			assert.Equal(t, identity.ToBytes(), env.GetHeader().GetSenderId())
			assert.Equal(t, networkID, env.GetHeader().GetNetworkIdentifier())
			assert.Equal(t, p2pcommon.MessageTypeGetDoc.String(), env.GetHeader().GetType())
			assert.Equal(t, version.GetVersion().String(), env.GetHeader().GetNodeVersion())
		}).
		Return(protocolEnvelopeRes, nil).Once()

	res, err := peer.GetDocumentRequest(ctx, requesterID, req)
	assert.NoError(t, err)
	assert.Equal(t, getDocRes, res)
}

func TestPeer_Client_GetDocumentRequest_SenderIdentityError(t *testing.T) {
	peer, _ := getPeerMocks(t)

	ctx := context.Background()

	requesterID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	req := &p2ppb.GetDocumentRequest{
		DocumentIdentifier: utils.RandomSlice(32),
		AccessType:         p2ppb.AccessType_ACCESS_TYPE_REQUESTER_VERIFICATION,
	}

	res, err := peer.GetDocumentRequest(ctx, requesterID, req)
	assert.ErrorIs(t, err, errors.ErrContextIdentityRetrieval)
	assert.Nil(t, res)
}

func TestPeer_Client_GetDocumentRequest_ValidateAccountError(t *testing.T) {
	peer, mocks := getPeerMocks(t)

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	senderAccountMock := config.NewAccountMock(t)
	senderAccountMock.On("GetIdentity").
		Return(identity)

	ctx := contextutil.WithAccount(context.Background(), senderAccountMock)

	requesterID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	req := &p2ppb.GetDocumentRequest{
		DocumentIdentifier: utils.RandomSlice(32),
		AccessType:         p2ppb.AccessType_ACCESS_TYPE_REQUESTER_VERIFICATION,
	}

	genericUtils.GetMock[*config.ServiceMock](mocks).On("GetAccount", requesterID.ToBytes()).
		Return(nil, errors.New("error")).Once()

	genericUtils.GetMock[*v2.ServiceMock](mocks).On("ValidateAccount", requesterID).
		Return(errors.New("error")).Once()

	res, err := peer.GetDocumentRequest(ctx, requesterID, req)
	assert.ErrorIs(t, err, ErrInvalidRequesterAccount)
	assert.Nil(t, res)
}

func TestPeer_Client_GetDocumentRequest_PeerIDRetrievalError(t *testing.T) {
	peer, mocks := getPeerMocks(t)

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	senderAccountMock := config.NewAccountMock(t)
	senderAccountMock.On("GetIdentity").
		Return(identity)

	ctx := contextutil.WithAccount(context.Background(), senderAccountMock)

	requesterID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	req := &p2ppb.GetDocumentRequest{
		DocumentIdentifier: utils.RandomSlice(32),
		AccessType:         p2ppb.AccessType_ACCESS_TYPE_REQUESTER_VERIFICATION,
	}

	genericUtils.GetMock[*config.ServiceMock](mocks).On("GetAccount", requesterID.ToBytes()).
		Return(nil, errors.New("error")).Once()

	genericUtils.GetMock[*v2.ServiceMock](mocks).On("ValidateAccount", requesterID).
		Return(nil).Once()

	genericUtils.GetMock[*keystore.APIMock](mocks).
		On(
			"GetLastKeyByPurpose",
			requesterID,
			keystoreType.KeyPurposeP2PDiscovery,
		).
		Return(nil, errors.New("error")).Once()

	res, err := peer.GetDocumentRequest(ctx, requesterID, req)
	assert.ErrorIs(t, err, ErrPeerIDRetrieval)
	assert.Nil(t, res)
}

func TestPeer_Client_GetDocumentRequest_MessengerError(t *testing.T) {
	peer, mocks := getPeerMocks(t)

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	senderAccountMock := config.NewAccountMock(t)
	senderAccountMock.On("GetIdentity").
		Return(identity)

	ctx := contextutil.WithAccount(context.Background(), senderAccountMock)

	requesterID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	req := &p2ppb.GetDocumentRequest{
		DocumentIdentifier: utils.RandomSlice(32),
		AccessType:         p2ppb.AccessType_ACCESS_TYPE_REQUESTER_VERIFICATION,
	}

	genericUtils.GetMock[*config.ServiceMock](mocks).On("GetAccount", requesterID.ToBytes()).
		Return(nil, errors.New("error")).Once()

	genericUtils.GetMock[*v2.ServiceMock](mocks).On("ValidateAccount", requesterID).
		Return(nil).Once()

	peerID := mockPeerIDRetrievalCalls(t, requesterID, mocks)

	genericUtils.GetMock[*ms.MessengerMock](mocks).
		On(
			"SendMessage",
			mock.Anything,
			peerID,
			mock.IsType(&protocolpb.P2PEnvelope{}),
			p2pcommon.ProtocolForIdentity(requesterID),
		).
		Return(nil, errors.New("error")).Once()

	res, err := peer.GetDocumentRequest(ctx, requesterID, req)
	assert.ErrorIs(t, err, ErrP2PMessageSending)
	assert.Nil(t, res)
}

func TestPeer_Client_GetDocumentRequest_ResolveEnvelopeError(t *testing.T) {
	peer, mocks := getPeerMocks(t)

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	senderAccountMock := config.NewAccountMock(t)
	senderAccountMock.On("GetIdentity").
		Return(identity)

	ctx := contextutil.WithAccount(context.Background(), senderAccountMock)

	requesterID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	req := &p2ppb.GetDocumentRequest{
		DocumentIdentifier: utils.RandomSlice(32),
		AccessType:         p2ppb.AccessType_ACCESS_TYPE_REQUESTER_VERIFICATION,
	}

	genericUtils.GetMock[*config.ServiceMock](mocks).On("GetAccount", requesterID.ToBytes()).
		Return(nil, errors.New("error")).Once()

	genericUtils.GetMock[*v2.ServiceMock](mocks).On("ValidateAccount", requesterID).
		Return(nil).Once()

	peerID := mockPeerIDRetrievalCalls(t, requesterID, mocks)

	// Set invalid body to ensure an error when resolving the envelope.
	protocolEnvelopeRes := &protocolpb.P2PEnvelope{
		Body: utils.RandomSlice(32),
	}

	genericUtils.GetMock[*ms.MessengerMock](mocks).
		On(
			"SendMessage",
			mock.Anything,
			peerID,
			mock.IsType(&protocolpb.P2PEnvelope{}),
			p2pcommon.ProtocolForIdentity(requesterID),
		).
		Run(func(args mock.Arguments) {
			protocolEnv, ok := args.Get(2).(*protocolpb.P2PEnvelope)
			assert.True(t, ok)

			var env p2ppb.Envelope

			err = proto.Unmarshal(protocolEnv.GetBody(), &env)
			assert.NoError(t, err)

			reqBytes, err := proto.Marshal(req)
			assert.NoError(t, err)

			assert.Equal(t, reqBytes, env.GetBody())
			assert.Equal(t, identity.ToBytes(), env.GetHeader().GetSenderId())
			assert.Equal(t, networkID, env.GetHeader().GetNetworkIdentifier())
			assert.Equal(t, p2pcommon.MessageTypeGetDoc.String(), env.GetHeader().GetType())
			assert.Equal(t, version.GetVersion().String(), env.GetHeader().GetNodeVersion())
		}).
		Return(protocolEnvelopeRes, nil).Once()

	res, err := peer.GetDocumentRequest(ctx, requesterID, req)
	assert.ErrorIs(t, err, ErrP2PDataEnvelopeResolving)
	assert.Nil(t, res)
}

func TestPeer_Client_GetDocumentRequest_ReceiveEnvelopeError(t *testing.T) {
	peer, mocks := getPeerMocks(t)

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	senderAccountMock := config.NewAccountMock(t)
	senderAccountMock.On("GetIdentity").
		Return(identity)

	ctx := contextutil.WithAccount(context.Background(), senderAccountMock)

	requesterID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	req := &p2ppb.GetDocumentRequest{
		DocumentIdentifier: utils.RandomSlice(32),
		AccessType:         p2ppb.AccessType_ACCESS_TYPE_REQUESTER_VERIFICATION,
	}

	genericUtils.GetMock[*config.ServiceMock](mocks).On("GetAccount", requesterID.ToBytes()).
		Return(nil, errors.New("error")).Once()

	genericUtils.GetMock[*v2.ServiceMock](mocks).On("ValidateAccount", requesterID).
		Return(nil).Once()

	peerID := mockPeerIDRetrievalCalls(t, requesterID, mocks)

	envelopeRes := &p2ppb.Envelope{
		Header: &p2ppb.Header{
			NetworkIdentifier: networkID,
			NodeVersion:       "test-version",
			SenderId:          utils.RandomSlice(32),
			// Error type.
			Type: p2pcommon.MessageTypeError.String(),
		},
		Body: nil,
	}

	envelopeResBytes, err := proto.Marshal(envelopeRes)
	assert.NoError(t, err)

	protocolEnvelopeRes := &protocolpb.P2PEnvelope{
		Body: envelopeResBytes,
	}

	genericUtils.GetMock[*ms.MessengerMock](mocks).
		On(
			"SendMessage",
			mock.Anything,
			peerID,
			mock.IsType(&protocolpb.P2PEnvelope{}),
			p2pcommon.ProtocolForIdentity(requesterID),
		).
		Run(func(args mock.Arguments) {
			protocolEnv, ok := args.Get(2).(*protocolpb.P2PEnvelope)
			assert.True(t, ok)

			var env p2ppb.Envelope

			err = proto.Unmarshal(protocolEnv.GetBody(), &env)
			assert.NoError(t, err)

			reqBytes, err := proto.Marshal(req)
			assert.NoError(t, err)

			assert.Equal(t, reqBytes, env.GetBody())
			assert.Equal(t, identity.ToBytes(), env.GetHeader().GetSenderId())
			assert.Equal(t, networkID, env.GetHeader().GetNetworkIdentifier())
			assert.Equal(t, p2pcommon.MessageTypeGetDoc.String(), env.GetHeader().GetType())
			assert.Equal(t, version.GetVersion().String(), env.GetHeader().GetNodeVersion())
		}).
		Return(protocolEnvelopeRes, nil).Once()

	res, err := peer.GetDocumentRequest(ctx, requesterID, req)
	assert.True(t, errors.IsOfType(ErrP2PClient, err))
	assert.Nil(t, res)
}

func TestPeer_Client_GetDocumentRequest_ReceiveEnvelopeInvalidType(t *testing.T) {
	peer, mocks := getPeerMocks(t)

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	senderAccountMock := config.NewAccountMock(t)
	senderAccountMock.On("GetIdentity").
		Return(identity)

	ctx := contextutil.WithAccount(context.Background(), senderAccountMock)

	requesterID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	req := &p2ppb.GetDocumentRequest{
		DocumentIdentifier: utils.RandomSlice(32),
		AccessType:         p2ppb.AccessType_ACCESS_TYPE_REQUESTER_VERIFICATION,
	}

	genericUtils.GetMock[*config.ServiceMock](mocks).On("GetAccount", requesterID.ToBytes()).
		Return(nil, errors.New("error")).Once()

	genericUtils.GetMock[*v2.ServiceMock](mocks).On("ValidateAccount", requesterID).
		Return(nil).Once()

	peerID := mockPeerIDRetrievalCalls(t, requesterID, mocks)

	getDocRes := &p2ppb.GetDocumentResponse{}

	getDocResBytes, err := proto.Marshal(getDocRes)
	assert.NoError(t, err)

	envelopeRes := &p2ppb.Envelope{
		Header: &p2ppb.Header{
			NetworkIdentifier: networkID,
			NodeVersion:       "test-version",
			SenderId:          utils.RandomSlice(32),
			// Invalid type.
			Type: p2pcommon.MessageTypeSendAnchoredDoc.String(),
		},
		Body: getDocResBytes,
	}

	envelopeResBytes, err := proto.Marshal(envelopeRes)
	assert.NoError(t, err)

	protocolEnvelopeRes := &protocolpb.P2PEnvelope{
		Body: envelopeResBytes,
	}

	genericUtils.GetMock[*ms.MessengerMock](mocks).
		On(
			"SendMessage",
			mock.Anything,
			peerID,
			mock.IsType(&protocolpb.P2PEnvelope{}),
			p2pcommon.ProtocolForIdentity(requesterID),
		).
		Run(func(args mock.Arguments) {
			protocolEnv, ok := args.Get(2).(*protocolpb.P2PEnvelope)
			assert.True(t, ok)

			var env p2ppb.Envelope

			err = proto.Unmarshal(protocolEnv.GetBody(), &env)
			assert.NoError(t, err)

			reqBytes, err := proto.Marshal(req)
			assert.NoError(t, err)

			assert.Equal(t, reqBytes, env.GetBody())
			assert.Equal(t, identity.ToBytes(), env.GetHeader().GetSenderId())
			assert.Equal(t, networkID, env.GetHeader().GetNetworkIdentifier())
			assert.Equal(t, p2pcommon.MessageTypeGetDoc.String(), env.GetHeader().GetType())
			assert.Equal(t, version.GetVersion().String(), env.GetHeader().GetNodeVersion())
		}).
		Return(protocolEnvelopeRes, nil).Once()

	res, err := peer.GetDocumentRequest(ctx, requesterID, req)
	assert.ErrorIs(t, err, ErrIncorrectResponseMessageType)
	assert.Nil(t, res)
}

func TestPeer_Client_GetDocumentRequest_ResponseDecodeError(t *testing.T) {
	peer, mocks := getPeerMocks(t)

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	senderAccountMock := config.NewAccountMock(t)
	senderAccountMock.On("GetIdentity").
		Return(identity)

	ctx := contextutil.WithAccount(context.Background(), senderAccountMock)

	requesterID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	req := &p2ppb.GetDocumentRequest{
		DocumentIdentifier: utils.RandomSlice(32),
		AccessType:         p2ppb.AccessType_ACCESS_TYPE_REQUESTER_VERIFICATION,
	}

	genericUtils.GetMock[*config.ServiceMock](mocks).On("GetAccount", requesterID.ToBytes()).
		Return(nil, errors.New("error")).Once()

	genericUtils.GetMock[*v2.ServiceMock](mocks).On("ValidateAccount", requesterID).
		Return(nil).Once()

	peerID := mockPeerIDRetrievalCalls(t, requesterID, mocks)

	envelopeRes := &p2ppb.Envelope{
		Header: &p2ppb.Header{
			NetworkIdentifier: networkID,
			NodeVersion:       "test-version",
			SenderId:          utils.RandomSlice(32),
			Type:              p2pcommon.MessageTypeGetDocRep.String(),
		},
		// Not a valid a p2ppb.GetDocumentResponse
		Body: utils.RandomSlice(32),
	}

	envelopeResBytes, err := proto.Marshal(envelopeRes)
	assert.NoError(t, err)

	protocolEnvelopeRes := &protocolpb.P2PEnvelope{
		Body: envelopeResBytes,
	}

	genericUtils.GetMock[*ms.MessengerMock](mocks).
		On(
			"SendMessage",
			mock.Anything,
			peerID,
			mock.IsType(&protocolpb.P2PEnvelope{}),
			p2pcommon.ProtocolForIdentity(requesterID),
		).
		Run(func(args mock.Arguments) {
			protocolEnv, ok := args.Get(2).(*protocolpb.P2PEnvelope)
			assert.True(t, ok)

			var env p2ppb.Envelope

			err = proto.Unmarshal(protocolEnv.GetBody(), &env)
			assert.NoError(t, err)

			reqBytes, err := proto.Marshal(req)
			assert.NoError(t, err)

			assert.Equal(t, reqBytes, env.GetBody())
			assert.Equal(t, identity.ToBytes(), env.GetHeader().GetSenderId())
			assert.Equal(t, networkID, env.GetHeader().GetNetworkIdentifier())
			assert.Equal(t, p2pcommon.MessageTypeGetDoc.String(), env.GetHeader().GetType())
			assert.Equal(t, version.GetVersion().String(), env.GetHeader().GetNodeVersion())
		}).
		Return(protocolEnvelopeRes, nil).Once()

	res, err := peer.GetDocumentRequest(ctx, requesterID, req)
	assert.ErrorIs(t, err, ErrResponseDecodeError)
	assert.Nil(t, res)
}

func TestPeer_Client_GetDocumentRequest_LocalAccount(t *testing.T) {
	peer, mocks := getPeerMocks(t)

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	senderAccountMock := config.NewAccountMock(t)
	senderAccountMock.On("GetIdentity").
		Return(identity)

	ctx := contextutil.WithAccount(context.Background(), senderAccountMock)

	requesterID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	req := &p2ppb.GetDocumentRequest{
		DocumentIdentifier: utils.RandomSlice(32),
		AccessType:         p2ppb.AccessType_ACCESS_TYPE_REQUESTER_VERIFICATION,
	}

	requesterAccountMock := config.NewAccountMock(t)

	genericUtils.GetMock[*config.ServiceMock](mocks).On("GetAccount", requesterID.ToBytes()).
		Return(requesterAccountMock, nil).Once()

	p2pConnTimeout := 1 * time.Second

	genericUtils.GetMock[*config.ConfigurationMock](mocks).On("GetP2PConnectionTimeout").
		Return(p2pConnTimeout).Once()

	getDocRes := &p2ppb.GetDocumentResponse{}

	genericUtils.GetMock[*receiver.HandlerMock](mocks).
		On(
			"GetDocument",
			mock.Anything,
			req,
			identity,
		).
		Run(func(args mock.Arguments) {
			handlerCtx, ok := args.Get(0).(context.Context)
			assert.True(t, ok)

			handlerCtxAccount, err := contextutil.Account(handlerCtx)
			assert.NoError(t, err)

			assert.Equal(t, requesterAccountMock, handlerCtxAccount)
		}).
		Return(getDocRes, nil).
		Once()

	res, err := peer.GetDocumentRequest(ctx, requesterID, req)
	assert.NoError(t, err)
	assert.Equal(t, getDocRes, res)
}

func TestPeer_Client_GetDocumentRequest_LocalAccount_HandlerError(t *testing.T) {
	peer, mocks := getPeerMocks(t)

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	senderAccountMock := config.NewAccountMock(t)
	senderAccountMock.On("GetIdentity").
		Return(identity)

	ctx := contextutil.WithAccount(context.Background(), senderAccountMock)

	requesterID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	req := &p2ppb.GetDocumentRequest{
		DocumentIdentifier: utils.RandomSlice(32),
		AccessType:         p2ppb.AccessType_ACCESS_TYPE_REQUESTER_VERIFICATION,
	}

	requesterAccountMock := config.NewAccountMock(t)

	genericUtils.GetMock[*config.ServiceMock](mocks).On("GetAccount", requesterID.ToBytes()).
		Return(requesterAccountMock, nil).Once()

	p2pConnTimeout := 1 * time.Second

	genericUtils.GetMock[*config.ConfigurationMock](mocks).On("GetP2PConnectionTimeout").
		Return(p2pConnTimeout).Once()

	handlerErr := errors.New("error")

	genericUtils.GetMock[*receiver.HandlerMock](mocks).
		On(
			"GetDocument",
			mock.Anything,
			req,
			identity,
		).
		Run(func(args mock.Arguments) {
			handlerCtx, ok := args.Get(0).(context.Context)
			assert.True(t, ok)

			handlerCtxAccount, err := contextutil.Account(handlerCtx)
			assert.NoError(t, err)

			assert.Equal(t, requesterAccountMock, handlerCtxAccount)
		}).
		Return(nil, handlerErr).
		Once()

	res, err := peer.GetDocumentRequest(ctx, requesterID, req)
	assert.Equal(t, err, handlerErr)
	assert.Nil(t, res)
}

func TestPeer_Client_GetSignaturesForDocument(t *testing.T) {
	peer, mocks := getPeerMocks(t)

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	senderAccountMock := config.NewAccountMock(t)
	senderAccountMock.On("GetIdentity").
		Return(identity)

	ctx := contextutil.WithAccount(context.Background(), senderAccountMock)

	documentMock := documents.NewDocumentMock(t)

	localCollaborator, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)
	externalCollaborator, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	signerCollaborators := []*types.AccountID{localCollaborator, externalCollaborator}

	documentMock.On("GetSignerCollaborators", identity).
		Return(signerCollaborators, nil).
		Once()

	coreDocument := &coredocumentpb.CoreDocument{}

	documentMock.On("PackCoreDocument").
		Return(coreDocument, nil).
		Times(len(signerCollaborators))

	// Local collaborator
	localAccountMock := config.NewAccountMock(t)

	genericUtils.GetMock[*config.ServiceMock](mocks).On("GetAccount", localCollaborator.ToBytes()).
		Return(localAccountMock, nil).Once()

	localSignature := &coredocumentpb.Signature{
		SignatureId:         utils.RandomSlice(32),
		SignerId:            localCollaborator.ToBytes(),
		PublicKey:           utils.RandomSlice(32),
		Signature:           utils.RandomSlice(32),
		TransitionValidated: false,
	}

	signatureRes := &p2ppb.SignatureResponse{
		Signatures: []*coredocumentpb.Signature{localSignature},
	}

	genericUtils.GetMock[*receiver.HandlerMock](mocks).
		On(
			"RequestDocumentSignature",
			mock.Anything,
			mock.IsType(&p2ppb.SignatureRequest{}),
			identity,
		).
		Run(func(args mock.Arguments) {
			handlerCtx, ok := args.Get(0).(context.Context)
			assert.True(t, ok)

			handlerCtxAccount, err := contextutil.Account(handlerCtx)
			assert.NoError(t, err)

			assert.Equal(t, localAccountMock, handlerCtxAccount)

			sigReq, ok := args.Get(1).(*p2ppb.SignatureRequest)
			assert.True(t, ok)

			assert.Equal(t, coreDocument, sigReq.Document)
		}).
		Return(signatureRes, nil).
		Once()

	signingRoot := utils.RandomSlice(32)

	documentMock.On("CalculateSigningRoot").
		Return(signingRoot, nil)

	documentTimestamp := time.Now()

	documentMock.On("Timestamp").
		Return(documentTimestamp, nil)

	genericUtils.GetMock[*v2.ServiceMock](mocks).On(
		"ValidateSignature",
		localCollaborator,
		localSignature.GetPublicKey(),
		documents.ConsensusSignaturePayload(signingRoot, localSignature.GetTransitionValidated()),
		localSignature.GetSignature(),
		documentTimestamp,
	).Return(nil).Once()

	// External collaborator
	genericUtils.GetMock[*config.ServiceMock](mocks).On("GetAccount", externalCollaborator.ToBytes()).
		Return(nil, errors.New("error")).Once()

	genericUtils.GetMock[*v2.ServiceMock](mocks).On("ValidateAccount", externalCollaborator).
		Return(nil).Once()

	peerID := mockPeerIDRetrievalCalls(t, externalCollaborator, mocks)

	externalSignature := &coredocumentpb.Signature{
		SignatureId:         utils.RandomSlice(32),
		SignerId:            externalCollaborator.ToBytes(),
		PublicKey:           utils.RandomSlice(32),
		Signature:           utils.RandomSlice(32),
		TransitionValidated: true,
	}

	signatureRes = &p2ppb.SignatureResponse{
		Signatures: []*coredocumentpb.Signature{externalSignature},
	}

	signatureResBytes, err := proto.Marshal(signatureRes)
	assert.NoError(t, err)

	envelopeRes := &p2ppb.Envelope{
		Header: &p2ppb.Header{
			NetworkIdentifier: networkID,
			NodeVersion:       version.GetVersion().String(),
			SenderId:          utils.RandomSlice(32),
			Type:              p2pcommon.MessageTypeRequestSignatureRep.String(),
		},
		Body: signatureResBytes,
	}

	envelopeResBytes, err := proto.Marshal(envelopeRes)
	assert.NoError(t, err)

	protocolEnvelopeRes := &protocolpb.P2PEnvelope{
		Body: envelopeResBytes,
	}

	genericUtils.GetMock[*ms.MessengerMock](mocks).
		On(
			"SendMessage",
			mock.Anything,
			peerID,
			mock.IsType(&protocolpb.P2PEnvelope{}),
			p2pcommon.ProtocolForIdentity(externalCollaborator),
		).
		Run(func(args mock.Arguments) {
			protocolEnv, ok := args.Get(2).(*protocolpb.P2PEnvelope)
			assert.True(t, ok)

			var env p2ppb.Envelope

			err = proto.Unmarshal(protocolEnv.GetBody(), &env)
			assert.NoError(t, err)

			req := &p2ppb.SignatureRequest{Document: coreDocument}

			reqBytes, err := proto.Marshal(req)
			assert.NoError(t, err)

			assert.Equal(t, reqBytes, env.GetBody())
			assert.Equal(t, identity.ToBytes(), env.GetHeader().GetSenderId())
			assert.Equal(t, networkID, env.GetHeader().GetNetworkIdentifier())
			assert.Equal(t, p2pcommon.MessageTypeRequestSignature.String(), env.GetHeader().GetType())
			assert.Equal(t, version.GetVersion().String(), env.GetHeader().GetNodeVersion())
		}).
		Return(protocolEnvelopeRes, nil).Once()

	genericUtils.GetMock[*v2.ServiceMock](mocks).On(
		"ValidateSignature",
		externalCollaborator,
		externalSignature.GetPublicKey(),
		documents.ConsensusSignaturePayload(signingRoot, externalSignature.GetTransitionValidated()),
		externalSignature.GetSignature(),
		documentTimestamp,
	).Return(nil).Once()

	signatures, signatureErrors, err := peer.GetSignaturesForDocument(ctx, documentMock)
	assert.NoError(t, err)
	assert.Nil(t, signatureErrors)
	assert.Len(t, signatures, 2)
}

func TestPeer_Client_GetSignaturesForDocument_SenderIdentityError(t *testing.T) {
	peer, _ := getPeerMocks(t)

	ctx := context.Background()

	documentMock := documents.NewDocumentMock(t)

	signatures, signatureErrors, err := peer.GetSignaturesForDocument(ctx, documentMock)
	assert.ErrorIs(t, err, errors.ErrContextIdentityRetrieval)
	assert.Nil(t, signatureErrors)
	assert.Nil(t, signatures)
}

func TestPeer_Client_GetSignaturesForDocument_SignerCollaboratorsError(t *testing.T) {
	peer, _ := getPeerMocks(t)

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	senderAccountMock := config.NewAccountMock(t)
	senderAccountMock.On("GetIdentity").
		Return(identity)

	ctx := contextutil.WithAccount(context.Background(), senderAccountMock)

	documentMock := documents.NewDocumentMock(t)

	documentMock.On("GetSignerCollaborators", identity).
		Return(nil, errors.New("error")).
		Once()

	signatures, signatureErrors, err := peer.GetSignaturesForDocument(ctx, documentMock)
	assert.ErrorIs(t, err, ErrSignerCollaboratorsRetrieval)
	assert.Nil(t, signatureErrors)
	assert.Nil(t, signatures)
}

func TestPeer_Client_GetSignaturesForDocument_LocalCollaboratorError(t *testing.T) {
	peer, mocks := getPeerMocks(t)

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	senderAccountMock := config.NewAccountMock(t)
	senderAccountMock.On("GetIdentity").
		Return(identity)

	ctx := contextutil.WithAccount(context.Background(), senderAccountMock)

	documentMock := documents.NewDocumentMock(t)

	localCollaborator, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)
	externalCollaborator, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	signerCollaborators := []*types.AccountID{localCollaborator, externalCollaborator}

	documentMock.On("GetSignerCollaborators", identity).
		Return(signerCollaborators, nil).
		Once()

	coreDocument := &coredocumentpb.CoreDocument{}

	documentMock.On("PackCoreDocument").
		Return(coreDocument, nil).
		Times(len(signerCollaborators))

	signingRoot := utils.RandomSlice(32)

	documentMock.On("CalculateSigningRoot").
		Return(signingRoot, nil)

	documentTimestamp := time.Now()

	documentMock.On("Timestamp").
		Return(documentTimestamp, nil)

	// Local collaborator
	localAccountMock := config.NewAccountMock(t)

	genericUtils.GetMock[*config.ServiceMock](mocks).On("GetAccount", localCollaborator.ToBytes()).
		Return(localAccountMock, nil).Once()

	genericUtils.GetMock[*receiver.HandlerMock](mocks).
		On(
			"RequestDocumentSignature",
			mock.Anything,
			mock.IsType(&p2ppb.SignatureRequest{}),
			identity,
		).
		Return(nil, errors.New("error")).
		Once()

	// External collaborator
	genericUtils.GetMock[*config.ServiceMock](mocks).On("GetAccount", externalCollaborator.ToBytes()).
		Return(nil, errors.New("error")).Once()

	genericUtils.GetMock[*v2.ServiceMock](mocks).On("ValidateAccount", externalCollaborator).
		Return(nil).Once()

	peerID := mockPeerIDRetrievalCalls(t, externalCollaborator, mocks)

	externalSignature := &coredocumentpb.Signature{
		SignatureId:         utils.RandomSlice(32),
		SignerId:            externalCollaborator.ToBytes(),
		PublicKey:           utils.RandomSlice(32),
		Signature:           utils.RandomSlice(32),
		TransitionValidated: true,
	}

	signatureRes := &p2ppb.SignatureResponse{
		Signatures: []*coredocumentpb.Signature{externalSignature},
	}

	signatureResBytes, err := proto.Marshal(signatureRes)
	assert.NoError(t, err)

	envelopeRes := &p2ppb.Envelope{
		Header: &p2ppb.Header{
			NetworkIdentifier: networkID,
			NodeVersion:       version.GetVersion().String(),
			SenderId:          utils.RandomSlice(32),
			Type:              p2pcommon.MessageTypeRequestSignatureRep.String(),
		},
		Body: signatureResBytes,
	}

	envelopeResBytes, err := proto.Marshal(envelopeRes)
	assert.NoError(t, err)

	protocolEnvelopeRes := &protocolpb.P2PEnvelope{
		Body: envelopeResBytes,
	}

	genericUtils.GetMock[*ms.MessengerMock](mocks).
		On(
			"SendMessage",
			mock.Anything,
			peerID,
			mock.IsType(&protocolpb.P2PEnvelope{}),
			p2pcommon.ProtocolForIdentity(externalCollaborator),
		).
		Run(func(args mock.Arguments) {
			protocolEnv, ok := args.Get(2).(*protocolpb.P2PEnvelope)
			assert.True(t, ok)

			var env p2ppb.Envelope

			err = proto.Unmarshal(protocolEnv.GetBody(), &env)
			assert.NoError(t, err)

			req := &p2ppb.SignatureRequest{Document: coreDocument}

			reqBytes, err := proto.Marshal(req)
			assert.NoError(t, err)

			assert.Equal(t, reqBytes, env.GetBody())
			assert.Equal(t, identity.ToBytes(), env.GetHeader().GetSenderId())
			assert.Equal(t, networkID, env.GetHeader().GetNetworkIdentifier())
			assert.Equal(t, p2pcommon.MessageTypeRequestSignature.String(), env.GetHeader().GetType())
			assert.Equal(t, version.GetVersion().String(), env.GetHeader().GetNodeVersion())
		}).
		Return(protocolEnvelopeRes, nil).Once()

	genericUtils.GetMock[*v2.ServiceMock](mocks).On(
		"ValidateSignature",
		externalCollaborator,
		externalSignature.GetPublicKey(),
		documents.ConsensusSignaturePayload(signingRoot, externalSignature.GetTransitionValidated()),
		externalSignature.GetSignature(),
		documentTimestamp,
	).Return(nil).Once()

	signatures, signatureErrors, err := peer.GetSignaturesForDocument(ctx, documentMock)
	assert.NoError(t, err)
	assert.Contains(t, signatureErrors, ErrDocumentSignatureRequest)
	assert.Len(t, signatures, 1)
}

func TestPeer_Client_GetSignaturesForDocument_ExternalCollaboratorError(t *testing.T) {
	peer, mocks := getPeerMocks(t)

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	senderAccountMock := config.NewAccountMock(t)
	senderAccountMock.On("GetIdentity").
		Return(identity)

	ctx := contextutil.WithAccount(context.Background(), senderAccountMock)

	documentMock := documents.NewDocumentMock(t)

	localCollaborator, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)
	externalCollaborator, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	signerCollaborators := []*types.AccountID{localCollaborator, externalCollaborator}

	documentMock.On("GetSignerCollaborators", identity).
		Return(signerCollaborators, nil).
		Once()

	p2pConnTimeout := 1 * time.Minute

	genericUtils.GetMock[*config.ConfigurationMock](mocks).On("GetP2PConnectionTimeout").
		Return(p2pConnTimeout)

	coreDocument := &coredocumentpb.CoreDocument{}

	documentMock.On("PackCoreDocument").
		Return(coreDocument, nil).
		Times(len(signerCollaborators))

	// Local collaborator
	localAccountMock := config.NewAccountMock(t)

	genericUtils.GetMock[*config.ServiceMock](mocks).On("GetAccount", localCollaborator.ToBytes()).
		Return(localAccountMock, nil).Once()

	localSignature := &coredocumentpb.Signature{
		SignatureId:         utils.RandomSlice(32),
		SignerId:            localCollaborator.ToBytes(),
		PublicKey:           utils.RandomSlice(32),
		Signature:           utils.RandomSlice(32),
		TransitionValidated: false,
	}

	signatureRes := &p2ppb.SignatureResponse{
		Signatures: []*coredocumentpb.Signature{localSignature},
	}

	genericUtils.GetMock[*receiver.HandlerMock](mocks).
		On(
			"RequestDocumentSignature",
			mock.Anything,
			mock.IsType(&p2ppb.SignatureRequest{}),
			identity,
		).
		Run(func(args mock.Arguments) {
			handlerCtx, ok := args.Get(0).(context.Context)
			assert.True(t, ok)

			handlerCtxAccount, err := contextutil.Account(handlerCtx)
			assert.NoError(t, err)

			assert.Equal(t, localAccountMock, handlerCtxAccount)

			sigReq, ok := args.Get(1).(*p2ppb.SignatureRequest)
			assert.True(t, ok)

			assert.Equal(t, coreDocument, sigReq.Document)
		}).
		Return(signatureRes, nil).
		Once()

	signingRoot := utils.RandomSlice(32)

	documentMock.On("CalculateSigningRoot").
		Return(signingRoot, nil)

	documentTimestamp := time.Now()

	documentMock.On("Timestamp").
		Return(documentTimestamp, nil)

	genericUtils.GetMock[*v2.ServiceMock](mocks).On(
		"ValidateSignature",
		localCollaborator,
		localSignature.GetPublicKey(),
		documents.ConsensusSignaturePayload(signingRoot, localSignature.GetTransitionValidated()),
		localSignature.GetSignature(),
		documentTimestamp,
	).Return(nil).Once()

	// External collaborator
	genericUtils.GetMock[*config.ServiceMock](mocks).On("GetAccount", externalCollaborator.ToBytes()).
		Return(nil, errors.New("error")).Once()

	genericUtils.GetMock[*v2.ServiceMock](mocks).On("ValidateAccount", externalCollaborator).
		Return(errors.New("error")).Once()

	signatures, signatureErrors, err := peer.GetSignaturesForDocument(ctx, documentMock)
	assert.NoError(t, err)
	assert.Contains(t, signatureErrors, ErrInvalidCollaboratorAccount)
	assert.Len(t, signatures, 1)
}

func TestPeer_Client_getPeerID(t *testing.T) {
	peer, mocks := getPeerMocks(t)
	peer.disablePeerStore = true

	ctx := context.Background()

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	pubKey, _, err := ed25519.GenerateSigningKeyPair()
	assert.NoError(t, err)

	p2pKey := types.NewHash(pubKey)

	genericUtils.GetMock[*keystore.APIMock](mocks).On(
		"GetLastKeyByPurpose",
		accountID,
		keystoreType.KeyPurposeP2PDiscovery,
	).Return(&p2pKey, nil).Once()

	genericUtils.GetMock[*v2.ServiceMock](mocks).On(
		"ValidateKey",
		accountID,
		p2pKey[:],
		keystoreType.KeyPurposeP2PDiscovery,
		mock.IsType(time.Now()),
	).Return(nil).Once()

	peerID, err := p2pcommon.ParsePeerID(p2pKey)
	assert.NoError(t, err)

	res, err := peer.getPeerID(ctx, accountID)
	assert.NoError(t, err)
	assert.Equal(t, peerID, res)
}

func TestPeer_Client_getPeerID_KeyRetrievalError(t *testing.T) {
	peer, mocks := getPeerMocks(t)
	peer.disablePeerStore = true

	ctx := context.Background()

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	genericUtils.GetMock[*keystore.APIMock](mocks).On(
		"GetLastKeyByPurpose",
		accountID,
		keystoreType.KeyPurposeP2PDiscovery,
	).Return(nil, errors.New("error")).Once()

	res, err := peer.getPeerID(ctx, accountID)
	assert.ErrorIs(t, err, ErrP2PKeyRetrievalError)
	assert.Empty(t, res)
}

func TestPeer_Client_getPeerID_KeyValidationError(t *testing.T) {
	peer, mocks := getPeerMocks(t)
	peer.disablePeerStore = true

	ctx := context.Background()

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	pubKey, _, err := ed25519.GenerateSigningKeyPair()
	assert.NoError(t, err)

	p2pKey := types.NewHash(pubKey)

	genericUtils.GetMock[*keystore.APIMock](mocks).On(
		"GetLastKeyByPurpose",
		accountID,
		keystoreType.KeyPurposeP2PDiscovery,
	).Return(&p2pKey, nil).Once()

	genericUtils.GetMock[*v2.ServiceMock](mocks).On(
		"ValidateKey",
		accountID,
		p2pKey[:],
		keystoreType.KeyPurposeP2PDiscovery,
		mock.IsType(time.Now()),
	).Return(errors.New("error")).Once()

	res, err := peer.getPeerID(ctx, accountID)
	assert.ErrorIs(t, err, ErrInvalidP2PKey)
	assert.Empty(t, res)
}

func TestPeer_Client_getPeerID_WithPeerstore(t *testing.T) {
	peer, mocks := getPeerMocks(t)

	ctx := context.Background()

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	pubKey, _, err := ed25519.GenerateSigningKeyPair()
	assert.NoError(t, err)

	p2pKey := types.NewHash(pubKey)

	genericUtils.GetMock[*keystore.APIMock](mocks).On(
		"GetLastKeyByPurpose",
		accountID,
		keystoreType.KeyPurposeP2PDiscovery,
	).Return(&p2pKey, nil).Once()

	genericUtils.GetMock[*v2.ServiceMock](mocks).On(
		"ValidateKey",
		accountID,
		p2pKey[:],
		keystoreType.KeyPurposeP2PDiscovery,
		mock.IsType(time.Now()),
	).Return(nil).Once()

	p2pConnTimeout := 1 * time.Second

	genericUtils.GetMock[*config.ConfigurationMock](mocks).On("GetP2PConnectionTimeout").
		Return(p2pConnTimeout).Once()

	peerID, err := p2pcommon.ParsePeerID(p2pKey)
	assert.NoError(t, err)

	addrInfo := libp2ppeer.AddrInfo{}

	genericUtils.GetMock[*p2pMocks.IpfsDHTMock](mocks).
		On(
			"FindPeer",
			mock.Anything,
			peerID,
		).
		Return(addrInfo, nil).Once()

	peerstoreMock := p2pMocks.NewPeerstoreMock(t)

	genericUtils.GetMock[*p2pMocks.HostMock](mocks).On("Peerstore").
		Return(peerstoreMock, nil).Once()

	peerstoreMock.On("AddAddrs", peerID, addrInfo.Addrs, time.Duration(pstore.PermanentAddrTTL)).
		Once()

	res, err := peer.getPeerID(ctx, accountID)
	assert.NoError(t, err)
	assert.Equal(t, peerID, res)
}

func TestPeer_Client_getPeerID_WithPeerstore_DHTError(t *testing.T) {
	peer, mocks := getPeerMocks(t)

	ctx := context.Background()

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	pubKey, _, err := ed25519.GenerateSigningKeyPair()
	assert.NoError(t, err)

	p2pKey := types.NewHash(pubKey)

	genericUtils.GetMock[*keystore.APIMock](mocks).On(
		"GetLastKeyByPurpose",
		accountID,
		keystoreType.KeyPurposeP2PDiscovery,
	).Return(&p2pKey, nil).Once()

	genericUtils.GetMock[*v2.ServiceMock](mocks).On(
		"ValidateKey",
		accountID,
		p2pKey[:],
		keystoreType.KeyPurposeP2PDiscovery,
		mock.IsType(time.Now()),
	).Return(nil).Once()

	p2pConnTimeout := 1 * time.Second

	genericUtils.GetMock[*config.ConfigurationMock](mocks).On("GetP2PConnectionTimeout").
		Return(p2pConnTimeout).Once()

	peerID, err := p2pcommon.ParsePeerID(p2pKey)
	assert.NoError(t, err)

	addrInfo := libp2ppeer.AddrInfo{}

	genericUtils.GetMock[*p2pMocks.IpfsDHTMock](mocks).
		On(
			"FindPeer",
			mock.Anything,
			peerID,
		).
		Return(addrInfo, errors.New("error")).Once()

	res, err := peer.getPeerID(ctx, accountID)
	assert.ErrorIs(t, err, ErrPeerNotFound)
	assert.Empty(t, res)
}

func TestPeer_Client_getSignatureForDocument_LocalAccount(t *testing.T) {
	peer, mocks := getPeerMocks(t)

	ctx := context.Background()

	sender, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	collaborator, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	documentMock := documents.NewDocumentMock(t)

	cd := &coredocumentpb.CoreDocument{}

	documentMock.On("PackCoreDocument").
		Return(cd, nil).
		Once()

	localAccountMock := config.NewAccountMock(t)

	genericUtils.GetMock[*config.ServiceMock](mocks).On("GetAccount", collaborator.ToBytes()).
		Return(localAccountMock, nil).Once()

	localSignature := &coredocumentpb.Signature{
		SignatureId:         utils.RandomSlice(32),
		SignerId:            collaborator.ToBytes(),
		PublicKey:           utils.RandomSlice(32),
		Signature:           utils.RandomSlice(32),
		TransitionValidated: false,
	}

	signatureRes := &p2ppb.SignatureResponse{
		Signatures: []*coredocumentpb.Signature{localSignature},
	}

	genericUtils.GetMock[*receiver.HandlerMock](mocks).
		On(
			"RequestDocumentSignature",
			mock.Anything,
			mock.IsType(&p2ppb.SignatureRequest{}),
			sender,
		).
		Run(func(args mock.Arguments) {
			handlerCtx, ok := args.Get(0).(context.Context)
			assert.True(t, ok)

			handlerCtxAccount, err := contextutil.Account(handlerCtx)
			assert.NoError(t, err)

			assert.Equal(t, localAccountMock, handlerCtxAccount)

			sigReq, ok := args.Get(1).(*p2ppb.SignatureRequest)
			assert.True(t, ok)

			assert.Equal(t, cd, sigReq.Document)
		}).
		Return(signatureRes, nil).
		Once()

	signingRoot := utils.RandomSlice(32)

	documentMock.On("CalculateSigningRoot").
		Return(signingRoot, nil)

	documentTimestamp := time.Now()

	documentMock.On("Timestamp").
		Return(documentTimestamp, nil)

	genericUtils.GetMock[*v2.ServiceMock](mocks).On(
		"ValidateSignature",
		collaborator,
		localSignature.GetPublicKey(),
		documents.ConsensusSignaturePayload(signingRoot, localSignature.GetTransitionValidated()),
		localSignature.GetSignature(),
		documentTimestamp,
	).Return(nil).Once()

	res, err := peer.getSignatureForDocument(ctx, documentMock, collaborator, sender)
	assert.NoError(t, err)
	assert.Equal(t, signatureRes, res)
}

func TestPeer_Client_getSignatureForDocument_ExternalAccount(t *testing.T) {
	peer, mocks := getPeerMocks(t)

	sender, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	senderAccountMock := config.NewAccountMock(t)
	senderAccountMock.On("GetIdentity").
		Return(sender)

	ctx := contextutil.WithAccount(context.Background(), senderAccountMock)

	collaborator, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	documentMock := documents.NewDocumentMock(t)

	cd := &coredocumentpb.CoreDocument{}

	documentMock.On("PackCoreDocument").
		Return(cd, nil).
		Once()

	genericUtils.GetMock[*config.ServiceMock](mocks).On("GetAccount", collaborator.ToBytes()).
		Return(nil, errors.New("error")).Once()

	genericUtils.GetMock[*v2.ServiceMock](mocks).On("ValidateAccount", collaborator).
		Return(nil).Once()

	peerID := mockPeerIDRetrievalCalls(t, collaborator, mocks)

	externalSignature := &coredocumentpb.Signature{
		SignatureId:         utils.RandomSlice(32),
		SignerId:            collaborator.ToBytes(),
		PublicKey:           utils.RandomSlice(32),
		Signature:           utils.RandomSlice(32),
		TransitionValidated: true,
	}

	signatureRes := &p2ppb.SignatureResponse{
		Signatures: []*coredocumentpb.Signature{externalSignature},
	}

	signatureResBytes, err := proto.Marshal(signatureRes)
	assert.NoError(t, err)

	envelopeRes := &p2ppb.Envelope{
		Header: &p2ppb.Header{
			NetworkIdentifier: networkID,
			NodeVersion:       version.GetVersion().String(),
			SenderId:          utils.RandomSlice(32),
			Type:              p2pcommon.MessageTypeRequestSignatureRep.String(),
		},
		Body: signatureResBytes,
	}

	envelopeResBytes, err := proto.Marshal(envelopeRes)
	assert.NoError(t, err)

	protocolEnvelopeRes := &protocolpb.P2PEnvelope{
		Body: envelopeResBytes,
	}

	genericUtils.GetMock[*ms.MessengerMock](mocks).
		On(
			"SendMessage",
			mock.Anything,
			peerID,
			mock.IsType(&protocolpb.P2PEnvelope{}),
			p2pcommon.ProtocolForIdentity(collaborator),
		).
		Run(func(args mock.Arguments) {
			protocolEnv, ok := args.Get(2).(*protocolpb.P2PEnvelope)
			assert.True(t, ok)

			var env p2ppb.Envelope

			err = proto.Unmarshal(protocolEnv.GetBody(), &env)
			assert.NoError(t, err)

			req := &p2ppb.SignatureRequest{Document: cd}

			reqBytes, err := proto.Marshal(req)
			assert.NoError(t, err)

			assert.Equal(t, reqBytes, env.GetBody())
			assert.Equal(t, sender.ToBytes(), env.GetHeader().GetSenderId())
			assert.Equal(t, networkID, env.GetHeader().GetNetworkIdentifier())
			assert.Equal(t, p2pcommon.MessageTypeRequestSignature.String(), env.GetHeader().GetType())
			assert.Equal(t, version.GetVersion().String(), env.GetHeader().GetNodeVersion())
		}).
		Return(protocolEnvelopeRes, nil).Once()

	signingRoot := utils.RandomSlice(32)

	documentMock.On("CalculateSigningRoot").
		Return(signingRoot, nil)

	documentTimestamp := time.Now()

	documentMock.On("Timestamp").
		Return(documentTimestamp, nil)

	genericUtils.GetMock[*v2.ServiceMock](mocks).On(
		"ValidateSignature",
		collaborator,
		externalSignature.GetPublicKey(),
		documents.ConsensusSignaturePayload(signingRoot, externalSignature.GetTransitionValidated()),
		externalSignature.GetSignature(),
		documentTimestamp,
	).Return(nil).Once()

	res, err := peer.getSignatureForDocument(ctx, documentMock, collaborator, sender)
	assert.NoError(t, err)
	assert.Len(t, res.GetSignatures(), 1)
	assert.Equal(t, externalSignature.GetSignature(), res.GetSignatures()[0].GetSignature())
	assert.Equal(t, externalSignature.GetSignatureId(), res.GetSignatures()[0].GetSignatureId())
	assert.Equal(t, externalSignature.GetPublicKey(), res.GetSignatures()[0].GetPublicKey())
	assert.Equal(t, externalSignature.GetTransitionValidated(), res.GetSignatures()[0].GetTransitionValidated())
	assert.Equal(t, externalSignature.GetSignerId(), res.GetSignatures()[0].GetSignerId())
}

func TestPeer_Client_getSignatureForDocument_PackCoreDocumentError(t *testing.T) {
	peer, _ := getPeerMocks(t)

	ctx := context.Background()

	sender, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	collaborator, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	documentMock := documents.NewDocumentMock(t)

	documentMock.On("PackCoreDocument").
		Return(nil, errors.New("error")).
		Once()

	res, err := peer.getSignatureForDocument(ctx, documentMock, collaborator, sender)
	assert.ErrorIs(t, err, ErrCoreDocumentPacking)
	assert.Nil(t, res)
}

func TestPeer_Client_getSignatureForDocument_LocalAccount_HandlerError(t *testing.T) {
	peer, mocks := getPeerMocks(t)

	ctx := context.Background()

	sender, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	collaborator, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	documentMock := documents.NewDocumentMock(t)

	cd := &coredocumentpb.CoreDocument{}

	documentMock.On("PackCoreDocument").
		Return(cd, nil).
		Once()

	localAccountMock := config.NewAccountMock(t)

	genericUtils.GetMock[*config.ServiceMock](mocks).On("GetAccount", collaborator.ToBytes()).
		Return(localAccountMock, nil).Once()

	genericUtils.GetMock[*receiver.HandlerMock](mocks).
		On(
			"RequestDocumentSignature",
			mock.Anything,
			mock.IsType(&p2ppb.SignatureRequest{}),
			sender,
		).
		Run(func(args mock.Arguments) {
			handlerCtx, ok := args.Get(0).(context.Context)
			assert.True(t, ok)

			handlerCtxAccount, err := contextutil.Account(handlerCtx)
			assert.NoError(t, err)

			assert.Equal(t, localAccountMock, handlerCtxAccount)

			sigReq, ok := args.Get(1).(*p2ppb.SignatureRequest)
			assert.True(t, ok)

			assert.Equal(t, cd, sigReq.Document)
		}).
		Return(nil, errors.New("error")).
		Once()

	res, err := peer.getSignatureForDocument(ctx, documentMock, collaborator, sender)
	assert.ErrorIs(t, err, ErrDocumentSignatureRequest)
	assert.Nil(t, res)
}

func TestPeer_Client_getSignatureForDocument_LocalAccount_InvalidSignatureResponse(t *testing.T) {
	peer, mocks := getPeerMocks(t)

	ctx := context.Background()

	sender, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	collaborator, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	documentMock := documents.NewDocumentMock(t)

	cd := &coredocumentpb.CoreDocument{}

	documentMock.On("PackCoreDocument").
		Return(cd, nil).
		Once()

	localAccountMock := config.NewAccountMock(t)

	genericUtils.GetMock[*config.ServiceMock](mocks).On("GetAccount", collaborator.ToBytes()).
		Return(localAccountMock, nil).Once()

	localSignature := &coredocumentpb.Signature{
		SignatureId:         utils.RandomSlice(32),
		SignerId:            collaborator.ToBytes(),
		PublicKey:           utils.RandomSlice(32),
		Signature:           utils.RandomSlice(32),
		TransitionValidated: false,
	}

	signatureRes := &p2ppb.SignatureResponse{
		Signatures: []*coredocumentpb.Signature{localSignature},
	}

	genericUtils.GetMock[*receiver.HandlerMock](mocks).
		On(
			"RequestDocumentSignature",
			mock.Anything,
			mock.IsType(&p2ppb.SignatureRequest{}),
			sender,
		).
		Run(func(args mock.Arguments) {
			handlerCtx, ok := args.Get(0).(context.Context)
			assert.True(t, ok)

			handlerCtxAccount, err := contextutil.Account(handlerCtx)
			assert.NoError(t, err)

			assert.Equal(t, localAccountMock, handlerCtxAccount)

			sigReq, ok := args.Get(1).(*p2ppb.SignatureRequest)
			assert.True(t, ok)

			assert.Equal(t, cd, sigReq.Document)
		}).
		Return(signatureRes, nil).
		Once()

	signingRoot := utils.RandomSlice(32)

	documentMock.On("CalculateSigningRoot").
		Return(signingRoot, nil)

	documentTimestamp := time.Now()

	documentMock.On("Timestamp").
		Return(documentTimestamp, nil)

	genericUtils.GetMock[*v2.ServiceMock](mocks).On(
		"ValidateSignature",
		collaborator,
		localSignature.GetPublicKey(),
		documents.ConsensusSignaturePayload(signingRoot, localSignature.GetTransitionValidated()),
		localSignature.GetSignature(),
		documentTimestamp,
	).Return(errors.New("error")).Once()

	res, err := peer.getSignatureForDocument(ctx, documentMock, collaborator, sender)
	assert.ErrorIs(t, err, ErrInvalidSignatureResponse)
	assert.Nil(t, res)
}

func TestPeer_Client_getSignatureForDocument_ExternalAccount_InvalidAccount(t *testing.T) {
	peer, mocks := getPeerMocks(t)

	sender, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	ctx := context.Background()

	collaborator, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	documentMock := documents.NewDocumentMock(t)

	cd := &coredocumentpb.CoreDocument{}

	documentMock.On("PackCoreDocument").
		Return(cd, nil).
		Once()

	genericUtils.GetMock[*config.ServiceMock](mocks).On("GetAccount", collaborator.ToBytes()).
		Return(nil, errors.New("error")).Once()

	genericUtils.GetMock[*v2.ServiceMock](mocks).On("ValidateAccount", collaborator).
		Return(errors.New("error")).Once()

	res, err := peer.getSignatureForDocument(ctx, documentMock, collaborator, sender)
	assert.ErrorIs(t, err, ErrInvalidCollaboratorAccount)
	assert.Nil(t, res)
}

func TestPeer_Client_getSignatureForDocument_ExternalAccount_PeerIDRetrievalError(t *testing.T) {
	peer, mocks := getPeerMocks(t)

	sender, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	ctx := context.Background()

	collaborator, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	documentMock := documents.NewDocumentMock(t)

	cd := &coredocumentpb.CoreDocument{}

	documentMock.On("PackCoreDocument").
		Return(cd, nil).
		Once()

	genericUtils.GetMock[*config.ServiceMock](mocks).On("GetAccount", collaborator.ToBytes()).
		Return(nil, errors.New("error")).Once()

	genericUtils.GetMock[*v2.ServiceMock](mocks).On("ValidateAccount", collaborator).
		Return(nil).Once()

	genericUtils.GetMock[*keystore.APIMock](mocks).
		On(
			"GetLastKeyByPurpose",
			collaborator,
			keystoreType.KeyPurposeP2PDiscovery,
		).
		Return(nil, errors.New("error")).Once()

	res, err := peer.getSignatureForDocument(ctx, documentMock, collaborator, sender)
	assert.ErrorIs(t, err, ErrPeerIDRetrieval)
	assert.Nil(t, res)
}

func TestPeer_Client_getSignatureForDocument_ExternalAccount_MessengerError(t *testing.T) {
	peer, mocks := getPeerMocks(t)

	sender, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	senderAccountMock := config.NewAccountMock(t)
	senderAccountMock.On("GetIdentity").
		Return(sender)

	ctx := contextutil.WithAccount(context.Background(), senderAccountMock)

	collaborator, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	documentMock := documents.NewDocumentMock(t)

	cd := &coredocumentpb.CoreDocument{}

	documentMock.On("PackCoreDocument").
		Return(cd, nil).
		Once()

	genericUtils.GetMock[*config.ServiceMock](mocks).On("GetAccount", collaborator.ToBytes()).
		Return(nil, errors.New("error")).Once()

	genericUtils.GetMock[*v2.ServiceMock](mocks).On("ValidateAccount", collaborator).
		Return(nil).Once()

	peerID := mockPeerIDRetrievalCalls(t, collaborator, mocks)

	genericUtils.GetMock[*ms.MessengerMock](mocks).
		On(
			"SendMessage",
			mock.Anything,
			peerID,
			mock.IsType(&protocolpb.P2PEnvelope{}),
			p2pcommon.ProtocolForIdentity(collaborator),
		).
		Run(func(args mock.Arguments) {
			protocolEnv, ok := args.Get(2).(*protocolpb.P2PEnvelope)
			assert.True(t, ok)

			var env p2ppb.Envelope

			err = proto.Unmarshal(protocolEnv.GetBody(), &env)
			assert.NoError(t, err)

			req := &p2ppb.SignatureRequest{Document: cd}

			reqBytes, err := proto.Marshal(req)
			assert.NoError(t, err)

			assert.Equal(t, reqBytes, env.GetBody())
			assert.Equal(t, sender.ToBytes(), env.GetHeader().GetSenderId())
			assert.Equal(t, networkID, env.GetHeader().GetNetworkIdentifier())
			assert.Equal(t, p2pcommon.MessageTypeRequestSignature.String(), env.GetHeader().GetType())
			assert.Equal(t, version.GetVersion().String(), env.GetHeader().GetNodeVersion())
		}).
		Return(nil, errors.New("error")).Once()

	res, err := peer.getSignatureForDocument(ctx, documentMock, collaborator, sender)
	assert.ErrorIs(t, err, ErrP2PMessageSending)
	assert.Nil(t, res)
}

func TestPeer_Client_getSignatureForDocument_ExternalAccount_ResolveEnvelopeError(t *testing.T) {
	peer, mocks := getPeerMocks(t)

	sender, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	senderAccountMock := config.NewAccountMock(t)
	senderAccountMock.On("GetIdentity").
		Return(sender)

	ctx := contextutil.WithAccount(context.Background(), senderAccountMock)

	collaborator, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	documentMock := documents.NewDocumentMock(t)

	cd := &coredocumentpb.CoreDocument{}

	documentMock.On("PackCoreDocument").
		Return(cd, nil).
		Once()

	genericUtils.GetMock[*config.ServiceMock](mocks).On("GetAccount", collaborator.ToBytes()).
		Return(nil, errors.New("error")).Once()

	genericUtils.GetMock[*v2.ServiceMock](mocks).On("ValidateAccount", collaborator).
		Return(nil).Once()

	peerID := mockPeerIDRetrievalCalls(t, collaborator, mocks)

	// Set invalid body to ensure an error when resolving the envelope.
	protocolEnvelopeRes := &protocolpb.P2PEnvelope{
		Body: utils.RandomSlice(32),
	}

	genericUtils.GetMock[*ms.MessengerMock](mocks).
		On(
			"SendMessage",
			mock.Anything,
			peerID,
			mock.IsType(&protocolpb.P2PEnvelope{}),
			p2pcommon.ProtocolForIdentity(collaborator),
		).
		Run(func(args mock.Arguments) {
			protocolEnv, ok := args.Get(2).(*protocolpb.P2PEnvelope)
			assert.True(t, ok)

			var env p2ppb.Envelope

			err = proto.Unmarshal(protocolEnv.GetBody(), &env)
			assert.NoError(t, err)

			req := &p2ppb.SignatureRequest{Document: cd}

			reqBytes, err := proto.Marshal(req)
			assert.NoError(t, err)

			assert.Equal(t, reqBytes, env.GetBody())
			assert.Equal(t, sender.ToBytes(), env.GetHeader().GetSenderId())
			assert.Equal(t, networkID, env.GetHeader().GetNetworkIdentifier())
			assert.Equal(t, p2pcommon.MessageTypeRequestSignature.String(), env.GetHeader().GetType())
			assert.Equal(t, version.GetVersion().String(), env.GetHeader().GetNodeVersion())
		}).
		Return(protocolEnvelopeRes, nil).Once()

	res, err := peer.getSignatureForDocument(ctx, documentMock, collaborator, sender)
	assert.ErrorIs(t, err, ErrP2PDataEnvelopeResolving)
	assert.Nil(t, res)
}

func TestPeer_Client_getSignatureForDocument_ExternalAccount_ReceiveEnvelopeError(t *testing.T) {
	peer, mocks := getPeerMocks(t)

	sender, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	senderAccountMock := config.NewAccountMock(t)
	senderAccountMock.On("GetIdentity").
		Return(sender)

	ctx := contextutil.WithAccount(context.Background(), senderAccountMock)

	collaborator, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	documentMock := documents.NewDocumentMock(t)

	cd := &coredocumentpb.CoreDocument{}

	documentMock.On("PackCoreDocument").
		Return(cd, nil).
		Once()

	genericUtils.GetMock[*config.ServiceMock](mocks).On("GetAccount", collaborator.ToBytes()).
		Return(nil, errors.New("error")).Once()

	genericUtils.GetMock[*v2.ServiceMock](mocks).On("ValidateAccount", collaborator).
		Return(nil).Once()

	peerID := mockPeerIDRetrievalCalls(t, collaborator, mocks)

	envelopeRes := &p2ppb.Envelope{
		Header: &p2ppb.Header{
			NetworkIdentifier: networkID,
			NodeVersion:       version.GetVersion().String(),
			SenderId:          utils.RandomSlice(32),
			// Error type.
			Type: p2pcommon.MessageTypeError.String(),
		},
		Body: nil,
	}

	envelopeResBytes, err := proto.Marshal(envelopeRes)
	assert.NoError(t, err)

	protocolEnvelopeRes := &protocolpb.P2PEnvelope{
		Body: envelopeResBytes,
	}

	genericUtils.GetMock[*ms.MessengerMock](mocks).
		On(
			"SendMessage",
			mock.Anything,
			peerID,
			mock.IsType(&protocolpb.P2PEnvelope{}),
			p2pcommon.ProtocolForIdentity(collaborator),
		).
		Run(func(args mock.Arguments) {
			protocolEnv, ok := args.Get(2).(*protocolpb.P2PEnvelope)
			assert.True(t, ok)

			var env p2ppb.Envelope

			err = proto.Unmarshal(protocolEnv.GetBody(), &env)
			assert.NoError(t, err)

			req := &p2ppb.SignatureRequest{Document: cd}

			reqBytes, err := proto.Marshal(req)
			assert.NoError(t, err)

			assert.Equal(t, reqBytes, env.GetBody())
			assert.Equal(t, sender.ToBytes(), env.GetHeader().GetSenderId())
			assert.Equal(t, networkID, env.GetHeader().GetNetworkIdentifier())
			assert.Equal(t, p2pcommon.MessageTypeRequestSignature.String(), env.GetHeader().GetType())
			assert.Equal(t, version.GetVersion().String(), env.GetHeader().GetNodeVersion())
		}).
		Return(protocolEnvelopeRes, nil).Once()

	res, err := peer.getSignatureForDocument(ctx, documentMock, collaborator, sender)
	assert.True(t, errors.IsOfType(ErrP2PClient, err))
	assert.Nil(t, res)
}

func TestPeer_Client_getSignatureForDocument_ExternalAccount_ReceiveEnvelopeInvalidType(t *testing.T) {
	peer, mocks := getPeerMocks(t)

	sender, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	senderAccountMock := config.NewAccountMock(t)
	senderAccountMock.On("GetIdentity").
		Return(sender)

	ctx := contextutil.WithAccount(context.Background(), senderAccountMock)

	collaborator, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	documentMock := documents.NewDocumentMock(t)

	cd := &coredocumentpb.CoreDocument{}

	documentMock.On("PackCoreDocument").
		Return(cd, nil).
		Once()

	genericUtils.GetMock[*config.ServiceMock](mocks).On("GetAccount", collaborator.ToBytes()).
		Return(nil, errors.New("error")).Once()

	genericUtils.GetMock[*v2.ServiceMock](mocks).On("ValidateAccount", collaborator).
		Return(nil).Once()

	peerID := mockPeerIDRetrievalCalls(t, collaborator, mocks)

	envelopeRes := &p2ppb.Envelope{
		Header: &p2ppb.Header{
			NetworkIdentifier: networkID,
			NodeVersion:       version.GetVersion().String(),
			SenderId:          utils.RandomSlice(32),
			Type:              p2pcommon.MessageTypeSendAnchoredDoc.String(),
		},
		Body: nil,
	}

	envelopeResBytes, err := proto.Marshal(envelopeRes)
	assert.NoError(t, err)

	protocolEnvelopeRes := &protocolpb.P2PEnvelope{
		Body: envelopeResBytes,
	}

	genericUtils.GetMock[*ms.MessengerMock](mocks).
		On(
			"SendMessage",
			mock.Anything,
			peerID,
			mock.IsType(&protocolpb.P2PEnvelope{}),
			p2pcommon.ProtocolForIdentity(collaborator),
		).
		Run(func(args mock.Arguments) {
			protocolEnv, ok := args.Get(2).(*protocolpb.P2PEnvelope)
			assert.True(t, ok)

			var env p2ppb.Envelope

			err = proto.Unmarshal(protocolEnv.GetBody(), &env)
			assert.NoError(t, err)

			req := &p2ppb.SignatureRequest{Document: cd}

			reqBytes, err := proto.Marshal(req)
			assert.NoError(t, err)

			assert.Equal(t, reqBytes, env.GetBody())
			assert.Equal(t, sender.ToBytes(), env.GetHeader().GetSenderId())
			assert.Equal(t, networkID, env.GetHeader().GetNetworkIdentifier())
			assert.Equal(t, p2pcommon.MessageTypeRequestSignature.String(), env.GetHeader().GetType())
			assert.Equal(t, version.GetVersion().String(), env.GetHeader().GetNodeVersion())
		}).
		Return(protocolEnvelopeRes, nil).Once()

	res, err := peer.getSignatureForDocument(ctx, documentMock, collaborator, sender)
	assert.ErrorIs(t, err, ErrIncorrectResponseMessageType)
	assert.Nil(t, res)
}

func TestPeer_Client_getSignatureForDocument_ExternalAccount_ResponseDecodeError(t *testing.T) {
	peer, mocks := getPeerMocks(t)

	sender, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	senderAccountMock := config.NewAccountMock(t)
	senderAccountMock.On("GetIdentity").
		Return(sender)

	ctx := contextutil.WithAccount(context.Background(), senderAccountMock)

	collaborator, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	documentMock := documents.NewDocumentMock(t)

	cd := &coredocumentpb.CoreDocument{}

	documentMock.On("PackCoreDocument").
		Return(cd, nil).
		Once()

	genericUtils.GetMock[*config.ServiceMock](mocks).On("GetAccount", collaborator.ToBytes()).
		Return(nil, errors.New("error")).Once()

	genericUtils.GetMock[*v2.ServiceMock](mocks).On("ValidateAccount", collaborator).
		Return(nil).Once()

	peerID := mockPeerIDRetrievalCalls(t, collaborator, mocks)

	envelopeRes := &p2ppb.Envelope{
		Header: &p2ppb.Header{
			NetworkIdentifier: networkID,
			NodeVersion:       version.GetVersion().String(),
			SenderId:          utils.RandomSlice(32),
			Type:              p2pcommon.MessageTypeRequestSignatureRep.String(),
		},
		// Invalid p2ppb.SignatureResponse bytes
		Body: utils.RandomSlice(32),
	}

	envelopeResBytes, err := proto.Marshal(envelopeRes)
	assert.NoError(t, err)

	protocolEnvelopeRes := &protocolpb.P2PEnvelope{
		Body: envelopeResBytes,
	}

	genericUtils.GetMock[*ms.MessengerMock](mocks).
		On(
			"SendMessage",
			mock.Anything,
			peerID,
			mock.IsType(&protocolpb.P2PEnvelope{}),
			p2pcommon.ProtocolForIdentity(collaborator),
		).
		Run(func(args mock.Arguments) {
			protocolEnv, ok := args.Get(2).(*protocolpb.P2PEnvelope)
			assert.True(t, ok)

			var env p2ppb.Envelope

			err = proto.Unmarshal(protocolEnv.GetBody(), &env)
			assert.NoError(t, err)

			req := &p2ppb.SignatureRequest{Document: cd}

			reqBytes, err := proto.Marshal(req)
			assert.NoError(t, err)

			assert.Equal(t, reqBytes, env.GetBody())
			assert.Equal(t, sender.ToBytes(), env.GetHeader().GetSenderId())
			assert.Equal(t, networkID, env.GetHeader().GetNetworkIdentifier())
			assert.Equal(t, p2pcommon.MessageTypeRequestSignature.String(), env.GetHeader().GetType())
			assert.Equal(t, version.GetVersion().String(), env.GetHeader().GetNodeVersion())
		}).
		Return(protocolEnvelopeRes, nil).Once()

	res, err := peer.getSignatureForDocument(ctx, documentMock, collaborator, sender)
	assert.ErrorIs(t, err, ErrResponseDecodeError)
	assert.Nil(t, res)
}

func TestPeer_Client_getSignatureForDocument_ExternalAccount_InvalidSignatureResponse(t *testing.T) {
	peer, mocks := getPeerMocks(t)

	sender, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	senderAccountMock := config.NewAccountMock(t)
	senderAccountMock.On("GetIdentity").
		Return(sender)

	ctx := contextutil.WithAccount(context.Background(), senderAccountMock)

	collaborator, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	documentMock := documents.NewDocumentMock(t)

	cd := &coredocumentpb.CoreDocument{}

	documentMock.On("PackCoreDocument").
		Return(cd, nil).
		Once()

	genericUtils.GetMock[*config.ServiceMock](mocks).On("GetAccount", collaborator.ToBytes()).
		Return(nil, errors.New("error")).Once()

	genericUtils.GetMock[*v2.ServiceMock](mocks).On("ValidateAccount", collaborator).
		Return(nil).Once()

	peerID := mockPeerIDRetrievalCalls(t, collaborator, mocks)

	externalSignature := &coredocumentpb.Signature{
		SignatureId:         utils.RandomSlice(32),
		SignerId:            collaborator.ToBytes(),
		PublicKey:           utils.RandomSlice(32),
		Signature:           utils.RandomSlice(32),
		TransitionValidated: true,
	}

	signatureRes := &p2ppb.SignatureResponse{
		Signatures: []*coredocumentpb.Signature{externalSignature},
	}

	signatureResBytes, err := proto.Marshal(signatureRes)
	assert.NoError(t, err)

	envelopeRes := &p2ppb.Envelope{
		Header: &p2ppb.Header{
			NetworkIdentifier: networkID,
			NodeVersion:       version.GetVersion().String(),
			SenderId:          utils.RandomSlice(32),
			Type:              p2pcommon.MessageTypeRequestSignatureRep.String(),
		},
		Body: signatureResBytes,
	}

	envelopeResBytes, err := proto.Marshal(envelopeRes)
	assert.NoError(t, err)

	protocolEnvelopeRes := &protocolpb.P2PEnvelope{
		Body: envelopeResBytes,
	}

	genericUtils.GetMock[*ms.MessengerMock](mocks).
		On(
			"SendMessage",
			mock.Anything,
			peerID,
			mock.IsType(&protocolpb.P2PEnvelope{}),
			p2pcommon.ProtocolForIdentity(collaborator),
		).
		Run(func(args mock.Arguments) {
			protocolEnv, ok := args.Get(2).(*protocolpb.P2PEnvelope)
			assert.True(t, ok)

			var env p2ppb.Envelope

			err = proto.Unmarshal(protocolEnv.GetBody(), &env)
			assert.NoError(t, err)

			req := &p2ppb.SignatureRequest{Document: cd}

			reqBytes, err := proto.Marshal(req)
			assert.NoError(t, err)

			assert.Equal(t, reqBytes, env.GetBody())
			assert.Equal(t, sender.ToBytes(), env.GetHeader().GetSenderId())
			assert.Equal(t, networkID, env.GetHeader().GetNetworkIdentifier())
			assert.Equal(t, p2pcommon.MessageTypeRequestSignature.String(), env.GetHeader().GetType())
			assert.Equal(t, version.GetVersion().String(), env.GetHeader().GetNodeVersion())
		}).
		Return(protocolEnvelopeRes, nil).Once()

	signingRoot := utils.RandomSlice(32)

	documentMock.On("CalculateSigningRoot").
		Return(signingRoot, nil)

	documentTimestamp := time.Now()

	documentMock.On("Timestamp").
		Return(documentTimestamp, nil)

	genericUtils.GetMock[*v2.ServiceMock](mocks).On(
		"ValidateSignature",
		collaborator,
		externalSignature.GetPublicKey(),
		documents.ConsensusSignaturePayload(signingRoot, externalSignature.GetTransitionValidated()),
		externalSignature.GetSignature(),
		documentTimestamp,
	).Return(errors.New("error")).Once()

	res, err := peer.getSignatureForDocument(ctx, documentMock, collaborator, sender)
	assert.ErrorIs(t, err, ErrInvalidSignatureResponse)
	assert.Nil(t, res)
}

func TestPeer_Client_validateSignatureResp(t *testing.T) {
	peer, mocks := getPeerMocks(t)

	documentMock := documents.NewDocumentMock(t)

	receiver, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	header := &p2ppb.Header{NodeVersion: version.GetVersion().String()}

	signature := &coredocumentpb.Signature{
		SignatureId:         utils.RandomSlice(32),
		SignerId:            receiver.ToBytes(),
		PublicKey:           utils.RandomSlice(32),
		Signature:           utils.RandomSlice(32),
		TransitionValidated: false,
	}

	resp := &p2ppb.SignatureResponse{
		Signatures: []*coredocumentpb.Signature{signature},
	}

	signingRoot := utils.RandomSlice(32)

	documentMock.On("CalculateSigningRoot").
		Return(signingRoot, nil).
		Once()

	timestamp := time.Now()

	documentMock.On("Timestamp").
		Return(timestamp, nil).
		Once()

	genericUtils.GetMock[*v2.ServiceMock](mocks).On(
		"ValidateSignature",
		receiver,
		signature.PublicKey,
		documents.ConsensusSignaturePayload(signingRoot, signature.TransitionValidated),
		signature.Signature,
		timestamp,
	).Return(nil).Once()

	err = peer.validateSignatureResp(documentMock, receiver, header, resp)
	assert.NoError(t, err)
}

func TestPeer_Client_validateSignatureResp_InvalidVersion(t *testing.T) {
	peer, _ := getPeerMocks(t)

	documentMock := documents.NewDocumentMock(t)

	receiver, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	header := &p2ppb.Header{NodeVersion: "invalid-version"}

	resp := &p2ppb.SignatureResponse{}

	err = peer.validateSignatureResp(documentMock, receiver, header, resp)
	assert.True(t, errors.IsOfType(version.ErrIncompatibleVersion, err))
}

func TestPeer_Client_validateSignatureResp_DocumentSigningRootError(t *testing.T) {
	peer, _ := getPeerMocks(t)

	documentMock := documents.NewDocumentMock(t)

	receiver, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	header := &p2ppb.Header{NodeVersion: version.GetVersion().String()}

	resp := &p2ppb.SignatureResponse{}

	documentMock.On("CalculateSigningRoot").
		Return(nil, errors.New("error")).
		Once()

	err = peer.validateSignatureResp(documentMock, receiver, header, resp)
	assert.NotNil(t, err)
}

func TestPeer_Client_validateSignatureResp_InvalidSignerAccount(t *testing.T) {
	peer, _ := getPeerMocks(t)

	documentMock := documents.NewDocumentMock(t)

	receiver, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	header := &p2ppb.Header{NodeVersion: version.GetVersion().String()}

	signature := &coredocumentpb.Signature{
		SignatureId: utils.RandomSlice(32),
		// Invalid account ID bytes.
		SignerId:            utils.RandomSlice(11),
		PublicKey:           utils.RandomSlice(32),
		Signature:           utils.RandomSlice(32),
		TransitionValidated: false,
	}

	resp := &p2ppb.SignatureResponse{
		Signatures: []*coredocumentpb.Signature{signature},
	}

	signingRoot := utils.RandomSlice(32)

	documentMock.On("CalculateSigningRoot").
		Return(signingRoot, nil).
		Once()

	err = peer.validateSignatureResp(documentMock, receiver, header, resp)
	assert.NotNil(t, err)
}

func TestPeer_Client_validateSignatureResp_DifferentSignerAccount(t *testing.T) {
	peer, _ := getPeerMocks(t)

	documentMock := documents.NewDocumentMock(t)

	receiver, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	header := &p2ppb.Header{NodeVersion: version.GetVersion().String()}

	signature := &coredocumentpb.Signature{
		SignatureId: utils.RandomSlice(32),
		// This is not the receiver.
		SignerId:            utils.RandomSlice(32),
		PublicKey:           utils.RandomSlice(32),
		Signature:           utils.RandomSlice(32),
		TransitionValidated: false,
	}

	resp := &p2ppb.SignatureResponse{
		Signatures: []*coredocumentpb.Signature{signature},
	}

	signingRoot := utils.RandomSlice(32)

	documentMock.On("CalculateSigningRoot").
		Return(signingRoot, nil).
		Once()

	err = peer.validateSignatureResp(documentMock, receiver, header, resp)
	assert.NotNil(t, err)
}

func TestPeer_Client_validateSignatureResp_DocumentTimestampError(t *testing.T) {
	peer, _ := getPeerMocks(t)

	documentMock := documents.NewDocumentMock(t)

	receiver, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	header := &p2ppb.Header{NodeVersion: version.GetVersion().String()}

	signature := &coredocumentpb.Signature{
		SignatureId:         utils.RandomSlice(32),
		SignerId:            receiver.ToBytes(),
		PublicKey:           utils.RandomSlice(32),
		Signature:           utils.RandomSlice(32),
		TransitionValidated: false,
	}

	resp := &p2ppb.SignatureResponse{
		Signatures: []*coredocumentpb.Signature{signature},
	}

	signingRoot := utils.RandomSlice(32)

	documentMock.On("CalculateSigningRoot").
		Return(signingRoot, nil).
		Once()

	documentMock.On("Timestamp").
		Return(time.Now(), errors.New("error")).
		Once()

	err = peer.validateSignatureResp(documentMock, receiver, header, resp)
	assert.NotNil(t, err)
}

func TestPeer_Client_validateSignatureResp_SignatureValidationError(t *testing.T) {
	peer, mocks := getPeerMocks(t)

	documentMock := documents.NewDocumentMock(t)

	receiver, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	header := &p2ppb.Header{NodeVersion: version.GetVersion().String()}

	signature := &coredocumentpb.Signature{
		SignatureId:         utils.RandomSlice(32),
		SignerId:            receiver.ToBytes(),
		PublicKey:           utils.RandomSlice(32),
		Signature:           utils.RandomSlice(32),
		TransitionValidated: false,
	}

	resp := &p2ppb.SignatureResponse{
		Signatures: []*coredocumentpb.Signature{signature},
	}

	signingRoot := utils.RandomSlice(32)

	documentMock.On("CalculateSigningRoot").
		Return(signingRoot, nil).
		Once()

	timestamp := time.Now()

	documentMock.On("Timestamp").
		Return(timestamp, nil).
		Once()

	genericUtils.GetMock[*v2.ServiceMock](mocks).On(
		"ValidateSignature",
		receiver,
		signature.PublicKey,
		documents.ConsensusSignaturePayload(signingRoot, signature.TransitionValidated),
		signature.Signature,
		timestamp,
	).Return(errors.New("error")).Once()

	err = peer.validateSignatureResp(documentMock, receiver, header, resp)
	assert.NotNil(t, err)
}

const (
	networkID = uint32(36)
)

func mockPeerIDRetrievalCalls(t *testing.T, accountID *types.AccountID, mocks []any) libp2ppeer.ID {
	pubKey, _, err := ed25519.GenerateSigningKeyPair()
	assert.NoError(t, err)

	p2pKey := types.NewHash(pubKey)

	genericUtils.GetMock[*keystore.APIMock](mocks).
		On(
			"GetLastKeyByPurpose",
			accountID,
			keystoreType.KeyPurposeP2PDiscovery,
		).
		Return(&p2pKey, nil).Once()

	genericUtils.GetMock[*v2.ServiceMock](mocks).
		On(
			"ValidateKey",
			accountID,
			p2pKey[:],
			keystoreType.KeyPurposeP2PDiscovery,
			mock.Anything,
		).
		Return(nil).Once()

	peerID, err := p2pcommon.ParsePeerID(p2pKey)
	assert.NoError(t, err)

	p2pConnTimeout := 1 * time.Second

	genericUtils.GetMock[*config.ConfigurationMock](mocks).On("GetP2PConnectionTimeout").
		Return(p2pConnTimeout)

	addrInfo := libp2ppeer.AddrInfo{}

	genericUtils.GetMock[*p2pMocks.IpfsDHTMock](mocks).
		On(
			"FindPeer",
			mock.Anything,
			peerID,
		).
		Return(addrInfo, nil).Once()

	peerstoreMock := p2pMocks.NewPeerstoreMock(t)

	genericUtils.GetMock[*p2pMocks.HostMock](mocks).On("Peerstore").
		Return(peerstoreMock, nil).Once()

	peerstoreMock.On("AddAddrs", peerID, addrInfo.Addrs, time.Duration(pstore.PermanentAddrTTL)).
		Once()

	genericUtils.GetMock[*config.ConfigurationMock](mocks).On("GetNetworkID").
		Return(networkID).Once()

	return peerID
}

func getPeerMocks(t *testing.T) (*p2pPeer, []any) {
	cfgServiceMock := config.NewServiceMock(t)
	identityServiceMock := v2.NewServiceMock(t)
	keystoreAPIMock := keystore.NewAPIMock(t)
	protocolIDDispMock := dispatcher.NewDispatcherMock[protocol.ID](t)
	handlerMock := receiver.NewHandlerMock(t)
	configMock := config.NewConfigurationMock(t)
	p2pHostMock := p2pMocks.NewHostMock(t)
	ipfsDHTMock := p2pMocks.NewIpfsDHTMock(t)
	messengerMock := ms.NewMessengerMock(t)

	peer := newPeer(
		configMock,
		cfgServiceMock,
		identityServiceMock,
		keystoreAPIMock,
		protocolIDDispMock,
		handlerMock,
	)

	peer.host = p2pHostMock
	peer.dht = ipfsDHTMock
	peer.mes = messengerMock

	return peer, []any{
		configMock,
		cfgServiceMock,
		identityServiceMock,
		keystoreAPIMock,
		protocolIDDispMock,
		handlerMock,
		p2pHostMock,
		ipfsDHTMock,
		messengerMock,
	}
}

//go:build unit

package p2p

//func TestGetSignatureForDocument_fail_connect(t *testing.T) {
//	c, err := cfg.GetConfig()
//	assert.NoError(t, err)
//	c = updateKeys(c)
//	ctx := testingconfig.CreateAccountContext(t, c)
//	idService := getIDMocks(ctx, did)
//	m := &MockMessenger{}
//	testClient := &peer{config: cfg, idService: idService, mes: m, disablePeerStore: true}
//	model, cd := generic.CreateGenericWithEmbedCD(t, ctx, did, nil)
//	_, err = p2pcommon.PrepareP2PEnvelope(ctx, c.GetNetworkID(), p2pcommon.MessageTypeRequestSignature, &p2ppb.SignatureRequest{Document: &cd})
//	assert.NoError(t, err, "signature request could not be created")
//
//	m.On("SendMessage", ctx, mock.Anything, mock.Anything, p2pcommon.ProtocolForIdentity(did)).Return(nil, errors.New("some error"))
//	resp, err := testClient.getSignatureForDocument(ctx, model, did, did)
//	m.AssertExpectations(t)
//	assert.Error(t, err, "must fail")
//	assert.Nil(t, resp, "must be nil")
//}
//
//func TestGetSignatureForDocument_fail_version_check(t *testing.T) {
//	c, err := cfg.GetConfig()
//	assert.NoError(t, err)
//	c = updateKeys(c)
//	ctx := testingconfig.CreateAccountContext(t, c)
//	idService := getIDMocks(ctx, did)
//	m := &MockMessenger{}
//	testClient := &peer{config: cfg, idService: idService, mes: m, disablePeerStore: true}
//	model, cd := generic.CreateGenericWithEmbedCD(t, ctx, did, nil)
//	_, err = p2pcommon.PrepareP2PEnvelope(ctx, c.GetNetworkID(), p2pcommon.MessageTypeRequestSignature, &p2ppb.SignatureRequest{Document: &cd})
//	assert.NoError(t, err, "signature request could not be created")
//
//	m.On("SendMessage", ctx, mock.Anything, mock.Anything, p2pcommon.ProtocolForIdentity(did)).Return(testClient.createSignatureResp("", nil), nil)
//	resp, err := testClient.getSignatureForDocument(ctx, model, did, did)
//	m.AssertExpectations(t)
//	assert.Error(t, err, "must fail")
//	assert.Contains(t, err.Error(), "Incompatible version")
//	assert.Nil(t, resp, "must be nil")
//}
//
//func TestGetSignatureForDocument_fail_did(t *testing.T) {
//	c, err := cfg.GetConfig()
//	assert.NoError(t, err)
//	c = updateKeys(c)
//	ctx := testingconfig.CreateAccountContext(t, c)
//	idService := getIDMocks(ctx, did)
//	m := &MockMessenger{}
//	testClient := &peer{config: cfg, idService: idService, mes: m, disablePeerStore: true}
//	model, cd := generic.CreateGenericWithEmbedCD(t, ctx, did, nil)
//	err = model.AddUpdateLog(did)
//	assert.NoError(t, err)
//	_, err = p2pcommon.PrepareP2PEnvelope(ctx, c.GetNetworkID(), p2pcommon.MessageTypeRequestSignature, &p2ppb.SignatureRequest{Document: &cd})
//	assert.NoError(t, err, "signature request could not be created")
//
//	randomBytes := utils.RandomSlice(identity.DIDLength)
//	signatures := []*coredocumentpb.Signature{{SignatureId: utils.RandomSlice(52), SignerId: randomBytes, PublicKey: utils.RandomSlice(32)}}
//	m.On("SendMessage", ctx, mock.Anything, mock.Anything, p2pcommon.ProtocolForIdentity(did)).Return(testClient.createSignatureResp(version.GetVersion().String(), signatures), nil)
//
//	resp, err := testClient.getSignatureForDocument(ctx, model, did, did)
//
//	m.AssertExpectations(t)
//	assert.Nil(t, resp, "must be nil")
//	assert.Error(t, err, "must not be nil")
//	assert.Contains(t, err.Error(), "signature invalid with err: provided bytes doesn't match centID")
//
//}
//
//func getIDMocks(ctx context.Context, did identity.DID) *testingcommons.MockIdentityService {
//	idService := &testingcommons.MockIdentityService{}
//	idService.On("CurrentP2PKey", did).Return("QmVf6EN6mkqWejWKW2qPu16XpdG3kJo1T3mhahPB5Se5n1", nil)
//	idService.On("Exists", ctx, did).Return(nil)
//	return idService
//}
//
//func (s *peer) createSignatureResp(centNodeVer string, signatures []*coredocumentpb.Signature) *protocolpb.P2PEnvelope {
//	req, err := proto.Marshal(&p2ppb.SignatureResponse{Signatures: signatures})
//	if err != nil {
//		return nil
//	}
//
//	dataReq := &p2ppb.Envelope{
//		Header: &p2ppb.Header{
//			NodeVersion: centNodeVer,
//			Type:        p2pcommon.MessageTypeRequestSignatureRep.String(),
//		},
//		Body: req,
//	}
//
//	reqB, err := proto.Marshal(dataReq)
//	if err != nil {
//		return nil
//	}
//
//	return &protocolpb.P2PEnvelope{Body: reqB}
//}

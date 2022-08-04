//go:build unit
// +build unit

package documents

//func TestDefaultProcessor_PrepareForSignatureRequests(t *testing.T) {
//	srv := &testingcommons.MockIdentityService{}
//	dp := DefaultProcessor(srv, nil, nil, cfg).(defaultProcessor)
//
//	ctxh := testingconfig.CreateAccountContext(t, cfg)
//
//	// failed to get self
//	err := dp.PrepareForSignatureRequests(context.Background(), nil)
//	assert.Error(t, err)
//	assert.True(t, errors.IsOfType(contextutil.ErrSelfNotFound, err))
//
//	// failed compute field execution
//	model := new(mockModel)
//	model.On("AddUpdateLog").Return(nil)
//	model.On("SetUsedAnchorRepoAddress", cfg.GetContractAddress(config.AnchorRepo)).Return()
//	model.On("ExecuteComputeFields", computeFieldsTimeout).Return(errors.New("failed to execute compute fields")).Once()
//	err = dp.PrepareForSignatureRequests(ctxh, model)
//	model.AssertExpectations(t)
//	assert.Error(t, err)
//
//	// failed signing root
//	model.On("ExecuteComputeFields", computeFieldsTimeout).Return(nil)
//	model.On("CalculateSigningRoot").Return(nil, errors.New("failed signing root")).Once()
//	err = dp.PrepareForSignatureRequests(ctxh, model)
//	model.AssertExpectations(t)
//	assert.Error(t, err)
//	assert.Contains(t, err.Error(), "failed signing root")
//
//	// success
//	sr := utils.RandomSlice(32)
//	model.On("CalculateSigningRoot").Return(sr, nil)
//	model.On("AppendSignatures", mock.Anything).Return().Once()
//	err = dp.PrepareForSignatureRequests(ctxh, model)
//	model.AssertExpectations(t)
//	assert.Nil(t, err)
//	assert.NotNil(t, model.sigs)
//	assert.Len(t, model.sigs, 1)
//	sig := model.sigs[0]
//	self, err := contextutil.Account(ctxh)
//	assert.NoError(t, err)
//	keys, err := self.GetKeys()
//	assert.NoError(t, err)
//	assert.True(t, crypto.VerifyMessage(keys[identity.KeyPurposeSigning.Name].PublicKey, ConsensusSignaturePayload(sr, false), sig.Signature, crypto.CurveEd25519))
//}
//
//type p2pClient struct {
//	mock.Mock
//	Client
//}
//
//func (p *p2pClient) GetSignaturesForDocument(ctx context.Context, model Document) ([]*coredocumentpb.Signature, []error, error) {
//	args := p.Called(ctx, model)
//	sigs, _ := args.Get(0).([]*coredocumentpb.Signature)
//	return sigs, nil, args.Error(1)
//}
//
//func (p *p2pClient) SendAnchoredDocument(ctx context.Context, receiverID identity.DID, in *p2ppb.AnchorDocumentRequest) (*p2ppb.AnchorDocumentResponse, error) {
//	args := p.Called(ctx, receiverID, in)
//	resp, _ := args.Get(0).(*p2ppb.AnchorDocumentResponse)
//	return resp, args.Error(1)
//}
//
//func TestDefaultProcessor_RequestSignatures(t *testing.T) {
//	srv := &testingcommons.MockIdentityService{}
//	dp := DefaultProcessor(srv, nil, nil, cfg).(defaultProcessor)
//	ctxh := testingconfig.CreateAccountContext(t, cfg)
//
//	self, err := contextutil.Account(ctxh)
//	assert.NoError(t, err)
//	did := self.GetIdentityID()
//	sr := utils.RandomSlice(32)
//	sig, err := self.SignMsg(sr)
//	assert.NoError(t, err)
//
//	did1, err := identity.NewDIDFromBytes(did)
//	assert.NoError(t, err)
//
//	// data validations failed
//	model := new(mockModel)
//	model.On("ID").Return([]byte{})
//	model.On("CurrentVersion").Return([]byte{})
//	model.On("NextVersion").Return([]byte{})
//	model.On("CalculateSigningRoot").Return(nil, errors.New("error"))
//	model.On("Timestamp").Return(time.Now().UTC(), nil)
//	model.On("GetAttributes").Return(nil)
//	model.On("GetComputeFieldsRules").Return(nil)
//	err = dp.RequestSignatures(ctxh, model)
//	model.AssertExpectations(t)
//	assert.Error(t, err)
//
//	// key validation failed
//	model = new(mockModel)
//	id := utils.RandomSlice(32)
//	next := utils.RandomSlice(32)
//	model.On("ID").Return(id)
//	model.On("CurrentVersion").Return(id)
//	model.On("NextVersion").Return(next)
//	model.On("CalculateSigningRoot").Return(sr, nil)
//	model.On("Signatures").Return()
//	model.On("Author").Return(did1, nil)
//	model.On("Timestamp").Return(time.Now(), nil)
//	model.On("GetAttributes").Return(nil)
//	model.On("GetComputeFieldsRules").Return(nil)
//	model.On("GetSignerCollaborators", mock.Anything).Return([]identity.DID{did1, testingidentity.GenerateRandomDID()}, nil)
//	model.sigs = append(model.sigs, sig)
//	c := new(p2pClient)
//	srv.On("ValidateSignature", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("cannot validate key")).Once()
//	err = dp.RequestSignatures(ctxh, model)
//	model.AssertExpectations(t)
//	c.AssertExpectations(t)
//	assert.Error(t, err)
//	assert.Contains(t, err.Error(), "cannot validate key")
//
//	// failed signature collection
//	model = new(mockModel)
//	model.On("ID").Return(id)
//	model.On("CurrentVersion").Return(id)
//	model.On("NextVersion").Return(next)
//	model.On("CalculateSigningRoot").Return(sr, nil)
//	model.On("Signatures").Return()
//	model.On("Author").Return(did1, nil)
//	model.On("Timestamp").Return(time.Now(), nil)
//	model.On("GetAttributes").Return(nil)
//	model.On("GetComputeFieldsRules").Return(nil)
//	model.On("GetSignerCollaborators", mock.Anything).Return([]identity.DID{did1, testingidentity.GenerateRandomDID()}, nil)
//	model.sigs = append(model.sigs, sig)
//	c = new(p2pClient)
//	srv.On("ValidateSignature", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
//	c.On("GetSignaturesForDocument", ctxh, model).Return(nil, errors.New("failed to get signatures")).Once()
//	dp.p2pClient = c
//	err = dp.RequestSignatures(ctxh, model)
//	model.AssertExpectations(t)
//	c.AssertExpectations(t)
//	assert.Error(t, err)
//	assert.Contains(t, err.Error(), "failed to get signatures")
//
//	// success
//	model = new(mockModel)
//	model.On("ID").Return(id)
//	model.On("CurrentVersion").Return(id)
//	model.On("NextVersion").Return(next)
//	model.On("CalculateSigningRoot").Return(sr, nil)
//	model.On("Signatures").Return()
//	model.On("AppendSignatures", []*coredocumentpb.Signature{sig}).Return().Once()
//	model.On("Author").Return(did1, nil)
//	model.On("GetSignerCollaborators", mock.Anything).Return([]identity.DID{did1, testingidentity.GenerateRandomDID()}, nil)
//	model.On("Timestamp").Return(time.Now(), nil)
//	model.On("GetAttributes").Return(nil)
//	model.On("GetComputeFieldsRules").Return(nil)
//	model.sigs = append(model.sigs, sig)
//	c = new(p2pClient)
//	c.On("GetSignaturesForDocument", ctxh, model).Return([]*coredocumentpb.Signature{sig}, nil).Once()
//	dp.p2pClient = c
//	err = dp.RequestSignatures(ctxh, model)
//	model.AssertExpectations(t)
//	c.AssertExpectations(t)
//	assert.Nil(t, err)
//}
//
//func TestDefaultProcessor_PrepareForAnchoring(t *testing.T) {
//	srv := &testingcommons.MockIdentityService{}
//	dp := DefaultProcessor(srv, nil, nil, cfg).(defaultProcessor)
//
//	ctxh := testingconfig.CreateAccountContext(t, cfg)
//	self, err := contextutil.Account(ctxh)
//	assert.NoError(t, err)
//	did := self.GetIdentityID()
//	sr := utils.RandomSlice(32)
//	payload := ConsensusSignaturePayload(sr, false)
//	sig, err := self.SignMsg(payload)
//	assert.NoError(t, err)
//	did1, err := identity.NewDIDFromBytes(did)
//	assert.NoError(t, err)
//
//	// validation failed
//	model := new(mockModel)
//	id := utils.RandomSlice(32)
//	next := utils.RandomSlice(32)
//	model.On("ID").Return(id)
//	model.On("CurrentVersion").Return(id)
//	model.On("NextVersion").Return(next)
//	model.On("CalculateSigningRoot").Return(sr, nil)
//	model.On("Signatures").Return()
//	model.On("Author").Return(did1, nil)
//	model.On("GetSignerCollaborators", mock.Anything).Return([]identity.DID{did1, testingidentity.GenerateRandomDID()}, nil)
//	tm := time.Now()
//	model.On("Timestamp").Return(tm, nil)
//	model.On("GetAttributes").Return(nil)
//	model.On("GetComputeFieldsRules").Return(nil)
//	model.sigs = append(model.sigs, sig)
//	srv = &testingcommons.MockIdentityService{}
//	cid, _ := identity.NewDIDFromBytes(did)
//	srv.On("ValidateSignature", cid, sig.PublicKey, sig.Signature, payload, tm).Return(errors.New("validation failed")).Once()
//	dp.identityService = srv
//	err = dp.PrepareForAnchoring(ctxh, model)
//	model.AssertExpectations(t)
//	srv.AssertExpectations(t)
//	assert.Error(t, err)
//
//	// success
//	model = new(mockModel)
//	model.On("ID").Return(id)
//	model.On("CurrentVersion").Return(id)
//	model.On("NextVersion").Return(next)
//	model.On("CalculateSigningRoot").Return(sr, nil)
//	model.On("Signatures").Return()
//	model.On("Author").Return(did1, nil)
//	model.On("GetSignerCollaborators", mock.Anything).Return([]identity.DID{did1, testingidentity.GenerateRandomDID()}, nil)
//	model.On("Timestamp").Return(tm, nil)
//	model.On("GetAttributes").Return(nil)
//	model.On("GetComputeFieldsRules").Return(nil)
//	model.sigs = append(model.sigs, sig)
//	srv = &testingcommons.MockIdentityService{}
//	srv.On("ValidateSignature", cid, sig.PublicKey, sig.Signature, payload, tm).Return(nil).Once()
//	dp.identityService = srv
//	err = dp.PrepareForAnchoring(ctxh, model)
//	model.AssertExpectations(t)
//	srv.AssertExpectations(t)
//	assert.NoError(t, err)
//}
//
//func TestDefaultProcessor_AnchorDocument(t *testing.T) {
//	srv := &testingcommons.MockIdentityService{}
//	dp := DefaultProcessor(srv, nil, nil, cfg).(defaultProcessor)
//	ctxh := testingconfig.CreateAccountContext(t, cfg)
//	self, err := contextutil.Account(ctxh)
//	assert.NoError(t, err)
//	did := self.GetIdentityID()
//	assert.NoError(t, err)
//	sr := utils.RandomSlice(32)
//	payload := ConsensusSignaturePayload(sr, false)
//	sig, err := self.SignMsg(payload)
//	assert.NoError(t, err)
//	did1, err := identity.NewDIDFromBytes(did)
//	assert.NoError(t, err)
//
//	// validations failed
//	id := utils.RandomSlice(32)
//	next := utils.RandomSlice(32)
//	model := new(mockModel)
//	model.On("ID").Return(id)
//	model.On("CurrentVersion").Return(id)
//	model.On("NextVersion").Return(next)
//	model.On("CalculateSigningRoot").Return(sr, nil)
//	model.On("Signatures").Return()
//	model.On("CalculateDocumentRoot").Return(nil, errors.New("error"))
//	model.On("Author").Return(did1, nil)
//	model.On("GetSignerCollaborators", mock.Anything).Return([]identity.DID{did1, testingidentity.GenerateRandomDID()}, nil)
//	tm := time.Now()
//	model.On("Timestamp").Return(tm, nil)
//	model.On("GetAttributes").Return(nil)
//	model.On("GetComputeFieldsRules").Return(nil)
//	model.sigs = append(model.sigs, sig)
//	srv = &testingcommons.MockIdentityService{}
//	cid, err := identity.NewDIDFromBytes(did)
//	assert.NoError(t, err)
//	srv.On("ValidateSignature", cid, sig.PublicKey, sig.Signature, payload, tm).Return(nil).Once()
//	dp.identityService = srv
//	err = dp.AnchorDocument(ctxh, model)
//	model.AssertExpectations(t)
//	srv.AssertExpectations(t)
//	assert.Error(t, err)
//	assert.Contains(t, err.Error(), "pre anchor validation failed")
//
//	id = utils.RandomSlice(32)
//	next = utils.RandomSlice(32)
//	model = new(mockModel)
//	model.On("ID").Return(id)
//	model.On("CurrentVersion").Return(id)
//	model.On("CurrentVersionPreimage").Return(id)
//	model.On("NextVersion").Return(next)
//	model.On("CalculateSigningRoot").Return(sr, nil)
//	model.On("Signatures").Return()
//	model.On("CalculateDocumentRoot").Return(utils.RandomSlice(32), nil)
//	model.On("Author").Return(did1, nil)
//	model.On("GetSignerCollaborators", mock.Anything).Return([]identity.DID{did1, testingidentity.GenerateRandomDID()}, nil)
//	model.On("Timestamp").Return(tm, nil)
//	model.On("CalculateSignaturesRoot").Return(nil, errors.New("error"))
//	model.On("GetAttributes").Return(nil)
//	model.On("GetComputeFieldsRules").Return(nil)
//	model.sigs = append(model.sigs, sig)
//	srv = &testingcommons.MockIdentityService{}
//	srv.On("ValidateSignature", did1, sig.PublicKey, sig.Signature, payload, tm).Return(nil).Once()
//	dp.identityService = srv
//	err = dp.AnchorDocument(ctxh, model)
//	model.AssertExpectations(t)
//	srv.AssertExpectations(t)
//	assert.Error(t, err)
//	assert.Contains(t, err.Error(), "failed to get signature root")
//
//	// success
//	model = new(mockModel)
//	model.On("ID").Return(id)
//	model.On("CurrentVersion").Return(id)
//	model.On("CurrentVersionPreimage").Return(id)
//	model.On("NextVersion").Return(next)
//	model.On("CalculateSigningRoot").Return(sr, nil)
//	model.On("CalculateSignaturesRoot").Return(utils.RandomSlice(32), nil)
//	model.On("Signatures").Return()
//	model.On("CalculateDocumentRoot").Return(utils.RandomSlice(32), nil)
//	model.On("Author").Return(did1, nil)
//	model.On("GetSignerCollaborators", mock.Anything).Return([]identity.DID{did1, testingidentity.GenerateRandomDID()}, nil)
//	model.On("Timestamp").Return(tm, nil)
//	model.On("GetAttributes").Return(nil)
//	model.On("GetComputeFieldsRules").Return(nil)
//	model.sigs = append(model.sigs, sig)
//	srv = &testingcommons.MockIdentityService{}
//	srv.On("ValidateSignature", cid, sig.PublicKey, sig.Signature, payload, tm).Return(nil).Once()
//	dp.identityService = srv
//	anchorSrv := new(anchors.MockAnchorService)
//	ch := make(chan error, 1)
//	ch <- nil
//	anchorSrv.On("CommitAnchor", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
//	dp.anchorSrv = anchorSrv
//	err = dp.AnchorDocument(ctxh, model)
//	model.AssertExpectations(t)
//	srv.AssertExpectations(t)
//	anchorSrv.AssertExpectations(t)
//	assert.Nil(t, err)
//}
//
//func TestDefaultProcessor_SendDocument(t *testing.T) {
//	srv := &testingcommons.MockIdentityService{}
//	srv.On("ValidateSignature", mock.Anything, mock.Anything).Return(nil).Once()
//	dp := DefaultProcessor(srv, nil, nil, cfg).(defaultProcessor)
//	ctxh := testingconfig.CreateAccountContext(t, cfg)
//	self, err := contextutil.Account(ctxh)
//	assert.NoError(t, err)
//	didb := self.GetIdentityID()
//	assert.NoError(t, err)
//	did1, err := identity.NewDIDFromBytes(didb)
//	assert.NoError(t, err)
//	sr := utils.RandomSlice(32)
//	payload := ConsensusSignaturePayload(sr, false)
//	sig, err := self.SignMsg(payload)
//	assert.NoError(t, err)
//	zeros := [32]byte{}
//	zeroRoot, err := anchors.ToDocumentRoot(zeros[:])
//
//	// validations failed
//	id := utils.RandomSlice(32)
//	aid, err := anchors.ToAnchorID(id)
//	assert.NoError(t, err)
//	next := utils.RandomSlice(32)
//	nextAid, err := anchors.ToAnchorID(next)
//	assert.NoError(t, err)
//	model := new(mockModel)
//	model.On("ID").Return(id)
//	model.On("CurrentVersion").Return(id)
//	model.On("NextVersion").Return(next)
//	model.On("CalculateSigningRoot").Return(sr, nil)
//	model.On("Signatures").Return()
//	model.On("CalculateDocumentRoot").Return(utils.RandomSlice(32), nil)
//	model.On("Author").Return(did1, nil)
//	model.On("GetSignerCollaborators", mock.Anything).Return([]identity.DID{did1, testingidentity.GenerateRandomDID()}, nil)
//	tm := time.Now()
//	model.On("Timestamp").Return(tm, nil)
//	model.On("GetAttributes").Return(nil)
//	model.On("GetComputeFieldsRules").Return(nil)
//	model.sigs = append(model.sigs, sig)
//	srv = &testingcommons.MockIdentityService{}
//	cid, err := identity.NewDIDFromBytes(didb)
//	assert.NoError(t, err)
//	srv.On("ValidateSignature", cid, sig.PublicKey, sig.Signature, payload, tm).Return(nil).Once()
//	dp.identityService = srv
//	anchorSrv := new(anchors.MockAnchorService)
//
//	anchorSrv.On("GetAnchorData", nextAid).Return(zeroRoot, nil)
//	anchorSrv.On("GetAnchorData", aid).Return(nil, errors.New("error"))
//	dp.anchorSrv = anchorSrv
//	err = dp.SendDocument(ctxh, model)
//	model.AssertExpectations(t)
//	srv.AssertExpectations(t)
//	anchorSrv.AssertExpectations(t)
//	assert.Error(t, err)
//	assert.Contains(t, err.Error(), "post anchor validations failed")
//
//	// get collaborators failed
//	dr, err := anchors.ToDocumentRoot(utils.RandomSlice(32))
//	assert.NoError(t, err)
//
//	assert.NoError(t, err)
//	model = new(mockModel)
//	model.On("ID").Return(id)
//	model.On("CurrentVersion").Return(id)
//	model.On("NextVersion").Return(next)
//	model.On("CalculateSigningRoot").Return(sr, nil)
//	model.On("Signatures").Return()
//	model.On("CalculateDocumentRoot").Return(dr[:], nil)
//	model.On("GetSignerCollaborators", mock.Anything).Return(nil, errors.New("error")).Once()
//	model.On("Author").Return(did1, nil)
//	model.On("Timestamp").Return(tm, nil)
//	model.On("GetAttributes").Return(nil)
//	model.On("GetComputeFieldsRules").Return(nil)
//	model.sigs = append(model.sigs, sig)
//	srv = &testingcommons.MockIdentityService{}
//	dp.identityService = srv
//	anchorSrv = new(anchors.MockAnchorService)
//	anchorSrv.On("GetAnchorData", nextAid).Return(zeroRoot, nil)
//	anchorSrv.On("GetAnchorData", aid).Return(dr, nil)
//	dp.anchorSrv = anchorSrv
//	err = dp.SendDocument(ctxh, model)
//	model.AssertExpectations(t)
//	srv.AssertExpectations(t)
//	anchorSrv.AssertExpectations(t)
//	assert.Error(t, err)
//
//	// pack core document failed
//	model = new(mockModel)
//	model.On("ID").Return(id)
//	model.On("CurrentVersion").Return(id)
//	model.On("NextVersion").Return(next)
//	model.On("CalculateSigningRoot").Return(sr, nil)
//	model.On("Signatures").Return()
//	model.On("CalculateDocumentRoot").Return(dr[:], nil)
//	model.On("GetSignerCollaborators", mock.Anything).Return([]identity.DID{testingidentity.GenerateRandomDID()}, nil)
//	model.On("PackCoreDocument").Return(nil, errors.New("error")).Once()
//	model.On("Author").Return(did1, nil)
//	model.On("Timestamp").Return(tm, nil)
//	model.On("GetAttributes").Return(nil)
//	model.On("GetComputeFieldsRules").Return(nil)
//	model.sigs = append(model.sigs, sig)
//	srv = &testingcommons.MockIdentityService{}
//	srv.On("ValidateSignature", cid, sig.PublicKey, sig.Signature, payload, tm).Return(nil).Once()
//	dp.identityService = srv
//	anchorSrv = new(anchors.MockAnchorService)
//	anchorSrv.On("GetAnchorData", nextAid).Return(zeroRoot, errors.New("missing"))
//	anchorSrv.On("GetAnchorData", aid).Return(dr, nil)
//	dp.anchorSrv = anchorSrv
//	err = dp.SendDocument(ctxh, model)
//	model.AssertExpectations(t)
//	srv.AssertExpectations(t)
//	anchorSrv.AssertExpectations(t)
//	assert.Error(t, err)
//
//	// successful
//	cd := coredocumentpb.CoreDocument{}
//	did := testingidentity.GenerateRandomDID()
//	model = new(mockModel)
//	model.On("ID").Return(id)
//	model.On("CurrentVersion").Return(id)
//	model.On("NextVersion").Return(next)
//	model.On("CalculateSigningRoot").Return(sr, nil)
//	model.On("Signatures").Return()
//	model.On("CalculateDocumentRoot").Return(dr[:], nil)
//	model.On("GetSignerCollaborators", mock.Anything).Return([]identity.DID{did}, nil)
//	model.On("PackCoreDocument").Return(cd, nil).Once()
//	model.On("Author").Return(did1, nil)
//	model.On("Timestamp").Return(tm, nil)
//	model.On("GetAttributes").Return(nil)
//	model.On("GetComputeFieldsRules").Return(nil)
//	model.sigs = append(model.sigs, sig)
//	srv = &testingcommons.MockIdentityService{}
//	srv.On("ValidateSignature", cid, sig.PublicKey, sig.Signature, payload, tm).Return(nil).Once()
//	dp.identityService = srv
//	anchorSrv = new(anchors.MockAnchorService)
//	anchorSrv.On("GetAnchorData", aid).Return(dr, nil)
//	anchorSrv.On("GetAnchorData", nextAid).Return([32]byte{}, errors.New("missing"))
//	client := new(p2pClient)
//	client.On("SendAnchoredDocument", mock.Anything, did, mock.Anything).Return(&p2ppb.AnchorDocumentResponse{Accepted: true}, nil).Once()
//	dp.anchorSrv = anchorSrv
//	dp.p2pClient = client
//	err = dp.SendDocument(ctxh, model)
//	model.AssertExpectations(t)
//	srv.AssertExpectations(t)
//	anchorSrv.AssertExpectations(t)
//	client.AssertExpectations(t)
//	assert.NoError(t, err)
//}

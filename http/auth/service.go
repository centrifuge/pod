package auth

//type AccountHeader struct {
//	Identity *types.AccountID
//	IsAdmin  bool
//}
//
//func NewAccountHeader(payload *token.JW3TPayload) (*AccountHeader, error) {
//	delegatorAccountID, err := decodeSS58Address(payload.OnBehalfOf)
//
//	if err != nil {
//		return nil, fmt.Errorf("couldn't decode delegator address: %w", err)
//	}
//
//	accountHeader := &AccountHeader{
//		Identity: delegatorAccountID,
//	}
//
//	switch payload.ProxyType {
//	case PodAdminProxyType:
//		accountHeader.IsAdmin = true
//	default:
//		if _, ok := proxyTypes.ProxyTypeValue[payload.ProxyType]; !ok {
//			return nil, fmt.Errorf("invalid proxy type - %s", payload.ProxyType)
//		}
//	}
//
//	return accountHeader, nil
//}
//
////go:generate mockery --name Service --structname ServiceMock --filename service_mock.go --inpackage
//
//type Service interface {
//	Validate(ctx context.Context, token *token.JW3Token) (*AccountHeader, error)
//}
//
//type service struct {
//	authenticationEnabled bool
//	log                   *logging.ZapEventLogger
//	proxyAPI              proxy.API
//	configSrv             config.Service
//}
//
//func NewService(
//	authenticationEnabled bool,
//	proxyAPI proxy.API,
//	configSrv config.Service,
//) Service {
//	log := logging.Logger("http-auth")
//
//	return &service{
//		authenticationEnabled: authenticationEnabled,
//		log:                   log,
//		proxyAPI:              proxyAPI,
//		configSrv:             configSrv,
//	}
//}
//
//func (s *service) Validate(_ context.Context, token *token.JW3Token) (*AccountHeader, error) {
//	if !s.authenticationEnabled {
//		return NewAccountHeader(jw3tPayload)
//	}
//
//	if token.Payload.ProxyType == PodAdminProxyType {
//		if err := s.validateAdminAccount(delegateAccountID); err != nil {
//			s.log.Errorf("Invalid admin account: %s", err)
//
//			return nil, err
//		}
//
//		return NewAccountHeader(jw3tPayload)
//	}
//
//	delegatorAccountID, err := decodeSS58Address(jw3tPayload.OnBehalfOf)
//
//	if err != nil {
//		s.log.Errorf("Couldn't decode delegator address: %s", err)
//
//		return nil, ErrSS58AddressDecode
//	}
//
//	// Verify OnBehalfOf is a valid Identity on the pod
//	_, err = s.configSrv.GetAccount(delegatorAccountID.ToBytes())
//	if err != nil {
//		s.log.Errorf("Invalid identity: %s", err)
//
//		return nil, ErrInvalidIdentity
//	}
//
//	// Verify that Address is a valid proxy of OnBehalfOf against the Proxy Pallet with the desired level ProxyType
//	proxyStorageEntry, err := s.proxyAPI.GetProxies(delegatorAccountID)
//
//	if err != nil {
//		s.log.Errorf("Couldn't retrieve account proxies: %s", err)
//
//		return nil, ErrAccountProxiesRetrieval
//	}
//
//	pt := proxyTypes.ProxyTypeValue[jw3tPayload.ProxyType]
//
//	valid := false
//	for _, proxyDefinition := range proxyStorageEntry.ProxyDefinitions {
//		if proxyDefinition.Delegate.Equal(delegateAccountID) {
//			if uint8(proxyDefinition.ProxyType) == uint8(pt) {
//				valid = true
//				break
//			}
//		}
//	}
//
//	if !valid {
//		s.log.Errorf("Invalid delegate")
//
//		return nil, ErrInvalidDelegate
//	}
//
//	return NewAccountHeader(jw3tPayload)
//}
//
//func (s *service) validateAdminAccount(accountID *types.AccountID) error {
//	podAdmin, err := s.configSrv.GetPodAdmin()
//
//	if err != nil {
//		return ErrPodAdminRetrieval
//	}
//
//	if !podAdmin.GetAccountID().Equal(accountID) {
//		return ErrNotAdminAccount
//	}
//
//	return nil
//}

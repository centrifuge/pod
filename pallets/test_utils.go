//go:build integration || testworld

package pallets

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	keystoreTypes "github.com/centrifuge/chain-custom-types/pkg/keystore"
	loansTypes "github.com/centrifuge/chain-custom-types/pkg/loans"
	proxyType "github.com/centrifuge/chain-custom-types/pkg/proxy"
	"github.com/centrifuge/go-substrate-rpc-client/v4/registry"
	"github.com/centrifuge/go-substrate-rpc-client/v4/registry/parser"
	"github.com/centrifuge/go-substrate-rpc-client/v4/scale"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/pod/centchain"
	"github.com/centrifuge/pod/config"
	"github.com/centrifuge/pod/crypto"
	"github.com/centrifuge/pod/pallets/keystore"
	"github.com/centrifuge/pod/pallets/proxy"
	"github.com/centrifuge/pod/pallets/uniques"
	genericUtils "github.com/centrifuge/pod/testingutils/generic"
	logging "github.com/ipfs/go-log"
)

var (
	log = logging.Logger("pallet-test-utils")
)

const (
	createAnonymousProxyTimeout = 10 * time.Minute
)

func CreateAnonymousProxy(
	serviceCtx map[string]any,
	originKrp signature.KeyringPair,
) (*types.AccountID, error) {
	cfg := genericUtils.GetService[config.Configuration](serviceCtx)

	testClient, err := centchain.NewTestClient(cfg.GetCentChainNodeURL())

	if err != nil {
		return nil, fmt.Errorf("couldn't create funds client: %w", err)
	}

	defer testClient.Close()

	ctx, cancel := context.WithTimeout(context.Background(), createAnonymousProxyTimeout)
	defer cancel()

	// TODO(cdamian): The following retry logic is required because, in some cases, there are no "AnonymousCreated"
	// events in the block where the extrinsic was created.
	anonymousProxyCreateFn := func() (*types.AccountID, error) {
		fn := getCreateAnonymousProxyCallCreationFn(proxyType.Any, 0, 0)

		blockHash, err := testClient.SubmitAndWait(ctx, originKrp, fn)

		if err != nil {
			return nil, fmt.Errorf("couldn't create anonymous proxy: %w", err)
		}

		events, err := testClient.GetEvents(*blockHash)

		if err != nil {
			return nil, fmt.Errorf("couldn't get events: %w", err)
		}

		return getAnonymousProxyCreatedByAccount(originKrp, events)
	}

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("context expired while creating proxy: %w", ctx.Err())
		default:
			anonymousProxyAccountID, err := anonymousProxyCreateFn()

			if err == nil {
				return anonymousProxyAccountID, nil
			}

			log.Errorf("Couldn't create anonymous proxy, retrying. Error: %s", err)
		}
	}
}

const (
	ProxyPureCreatedEventName    = "Proxy.PureCreated"
	ProxyPureCreateWhoFieldName  = "sp_core.crypto.AccountId32.who"
	ProxyPureCreatePureFieldName = "sp_core.crypto.AccountId32.pure"
)

func getAnonymousProxyCreatedByAccount(
	originKrp signature.KeyringPair,
	events []*parser.Event,
) (*types.AccountID, error) {
	for _, event := range events {
		if event.Name != ProxyPureCreatedEventName {
			continue
		}

		who, err := registry.GetDecodedFieldAsType[types.AccountID](
			event.Fields,
			func(fieldIndex int, field *registry.DecodedField) bool {
				return field.Name == ProxyPureCreateWhoFieldName
			},
		)

		if err != nil {
			return nil, fmt.Errorf("cannot find proxy event field: %w", err)
		}

		if bytes.Equal(who.ToBytes(), originKrp.PublicKey) {
			proxyAccountID, err := registry.GetDecodedFieldAsType[types.AccountID](
				event.Fields,
				func(fieldIndex int, field *registry.DecodedField) bool {
					return field.Name == ProxyPureCreatePureFieldName
				},
			)

			if err != nil {
				return nil, fmt.Errorf("couldn't get pure proxy: %w", err)
			}

			return &proxyAccountID, nil
		}
	}

	return nil, errors.New("pure proxy not found")
}

func ExecuteWithTestClient(
	ctx context.Context,
	serviceCtx map[string]any,
	originKrp signature.KeyringPair,
	callProviderFn centchain.CallProviderFn,
) error {
	cfg := genericUtils.GetService[config.Configuration](serviceCtx)

	testClient, err := centchain.NewTestClient(cfg.GetCentChainNodeURL())

	if err != nil {
		return fmt.Errorf("couldn't create test client: %w", err)
	}

	defer testClient.Close()

	if _, err = testClient.SubmitAndWait(ctx, originKrp, callProviderFn); err != nil {
		return fmt.Errorf("couldn't submit batch call: %w", err)
	}

	return nil
}

type ProxyPair struct {
	Delegate  *types.AccountID
	ProxyType proxyType.CentrifugeProxyType
}

type ProxyPairs []ProxyPair

func (p ProxyPairs) GetDelegateAccountIDs() []*types.AccountID {
	accountIDMap := make(map[string]struct{})

	var accountIDs []*types.AccountID

	for _, proxyPair := range p {
		if _, ok := accountIDMap[proxyPair.Delegate.ToHexString()]; ok {
			continue
		}

		accountIDMap[proxyPair.Delegate.ToHexString()] = struct{}{}

		accountIDs = append(accountIDs, proxyPair.Delegate)
	}

	return accountIDs
}

func getUnstoredAccountKeys(
	serviceCtx map[string]any,
	acc config.Account,
) ([]*keystoreTypes.KeyID, error) {
	cfgService := genericUtils.GetService[config.Service](serviceCtx)
	cfg, err := cfgService.GetConfig()

	if err != nil {
		return nil, fmt.Errorf("couldn't get config: %w", err)
	}

	_, p2pPublicKey, err := crypto.ObtainP2PKeypair(cfg.GetP2PKeyPair())

	if err != nil {
		return nil, fmt.Errorf("couldn't obtain P2P key pair: %w", err)
	}

	p2pPublicKeyRaw, err := p2pPublicKey.Raw()

	if err != nil {
		return nil, fmt.Errorf("couldn't get raw P2P public key: %w", err)
	}

	keys := []*keystoreTypes.KeyID{
		{
			Hash:       types.NewHash(p2pPublicKeyRaw),
			KeyPurpose: keystoreTypes.KeyPurposeP2PDiscovery,
		},
		{
			Hash:       types.NewHash(acc.GetSigningPublicKey()),
			KeyPurpose: keystoreTypes.KeyPurposeP2PDocumentSigning,
		},
	}

	return filterUnstoredAccountKeys(serviceCtx, acc.GetIdentity(), keys)
}

func filterUnstoredAccountKeys(serviceCtx map[string]any, accountID *types.AccountID, keys []*keystoreTypes.KeyID) ([]*keystoreTypes.KeyID, error) {
	keystoreAPI := genericUtils.GetService[keystore.API](serviceCtx)

	return genericUtils.FilterSlice(keys, func(key *keystoreTypes.KeyID) (bool, error) {
		_, err := keystoreAPI.GetKey(accountID, key)

		if err != nil {
			if errors.Is(err, keystore.ErrKeyNotFound) {
				return true, nil
			}

			return false, err
		}

		return false, nil
	})
}

func GetAddProxyCallCreationFns(anonymousProxyAccountID *types.AccountID, proxyPairs ProxyPairs) []centchain.CallProviderFn {
	var callCreationFns []centchain.CallProviderFn

	for _, proxyPair := range proxyPairs {
		callCreationFn := getAddProxyToAnonymousProxyCall(anonymousProxyAccountID, proxyPair.Delegate, proxyPair.ProxyType)

		callCreationFns = append(callCreationFns, callCreationFn)
	}

	return callCreationFns
}

func getCreateAnonymousProxyCallCreationFn(
	pt proxyType.CentrifugeProxyType,
	delay types.U32,
	index types.U16,
) centchain.CallProviderFn {
	return func(meta *types.Metadata) (*types.Call, error) {
		call, err := types.NewCall(
			meta,
			proxy.ProxyCreatePure,
			pt,
			delay,
			index,
		)

		if err != nil {
			return nil, err
		}

		return &call, nil
	}
}

func getAddProxyToAnonymousProxyCall(
	anonymousProxyID *types.AccountID,
	delegate *types.AccountID,
	pt proxyType.CentrifugeProxyType,
) centchain.CallProviderFn {
	return func(meta *types.Metadata) (*types.Call, error) {
		delegateMultiAddress, err := types.NewMultiAddressFromAccountID(delegate.ToBytes())

		if err != nil {
			return nil, err
		}

		addProxyCall, err := types.NewCall(
			meta,
			proxy.ProxyAdd,
			delegateMultiAddress,
			pt,
			types.U32(0),
		)

		if err != nil {
			return nil, err
		}

		delegatorMultiAddress, err := types.NewMultiAddressFromAccountID(anonymousProxyID.ToBytes())

		if err != nil {
			return nil, err
		}

		proxyCall, err := types.NewCall(
			meta,
			proxy.ProxyCall,
			delegatorMultiAddress,
			types.NewOption(proxyType.Any),
			addProxyCall,
		)

		if err != nil {
			return nil, err
		}

		return &proxyCall, nil
	}
}

func GetBalanceTransferCallCreationFn(balance string, receiverAccountID []byte) centchain.CallProviderFn {
	return func(meta *types.Metadata) (*types.Call, error) {
		dest, err := types.NewMultiAddressFromAccountID(receiverAccountID)

		if err != nil {
			return nil, err
		}

		b, ok := big.NewInt(0).SetString(balance, 10)

		if !ok {
			return nil, errors.New("couldn't create balance int")
		}

		call, err := types.NewCall(meta, "Balances.transfer", dest, types.NewUCompact(b))

		if err != nil {
			return nil, err
		}

		return &call, nil
	}
}

func getAddKeysArgsForAccount(
	serviceCtx map[string]any,
	acc config.Account,
) ([]*keystoreTypes.AddKey, error) {
	unstoredAccountKeys, err := getUnstoredAccountKeys(serviceCtx, acc)
	if err != nil {
		return nil, fmt.Errorf("couldn't get account keys: %w", err)
	}

	var keys []*keystoreTypes.AddKey

	for _, unstoredAccountKey := range unstoredAccountKeys {
		keys = append(keys, &keystoreTypes.AddKey{
			Key:     unstoredAccountKey.Hash,
			Purpose: unstoredAccountKey.KeyPurpose,
			KeyType: keystoreTypes.KeyTypeECDSA,
		})
	}

	return keys, nil
}

func GetAddKeysCall(
	serviceCtx map[string]any,
	acc config.Account,
) (centchain.CallProviderFn, error) {
	keys, err := getAddKeysArgsForAccount(serviceCtx, acc)

	if err != nil {
		return nil, err
	}

	return func(meta *types.Metadata) (*types.Call, error) {
		addKeysCall, err := types.NewCall(meta, keystore.AddKeysCall, keys)

		if err != nil {
			return nil, err
		}

		delegatorMultiAddress, err := types.NewMultiAddressFromAccountID(acc.GetIdentity().ToBytes())

		if err != nil {
			return nil, err
		}

		proxyCall, err := types.NewCall(
			meta,
			proxy.ProxyCall,
			delegatorMultiAddress,
			types.NewOption(proxyType.Any),
			addKeysCall,
		)

		if err != nil {
			return nil, err
		}

		return &proxyCall, nil
	}, nil
}

// Pools

type TrancheInput struct {
	TrancheType     TrancheType
	Seniority       types.Option[types.U32]
	TrancheMetadata TrancheMetadata
}

func (t TrancheInput) Encode(encoder scale.Encoder) error {
	if err := encoder.Encode(t.TrancheType); err != nil {
		return err
	}

	if err := encoder.Encode(t.Seniority); err != nil {
		return err
	}

	return encoder.Encode(t.TrancheMetadata)
}

type TrancheType struct {
	IsResidual bool

	IsNonResidual bool
	AsNonResidual NonResidual
}

func (t TrancheType) Encode(encoder scale.Encoder) error {
	switch {
	case t.IsResidual:
		return encoder.PushByte(0)
	case t.IsNonResidual:
		if err := encoder.PushByte(1); err != nil {
			return err
		}

		return encoder.Encode(t.AsNonResidual)
	default:
		return fmt.Errorf("unsupported tranche type")
	}
}

type NonResidual struct {
	InterestRatePerSec types.U128
	MinRiskBuffer      types.U64
}

func (n NonResidual) Encode(encoder scale.Encoder) error {
	if err := encoder.Encode(n.InterestRatePerSec); err != nil {
		return err
	}

	return encoder.Encode(n.MinRiskBuffer)
}

type TrancheMetadata struct {
	TokenName   []byte
	TokenSymbol []byte
}

func (t TrancheMetadata) Encode(encoder scale.Encoder) error {
	if err := encoder.Encode(t.TokenName); err != nil {
		return err
	}

	return encoder.Encode(t.TokenSymbol)
}

type CurrencyID struct {
	IsNative bool

	IsTranche bool
	AsTranche Tranche

	IsKSM bool

	IsAUSD bool

	IsForeignAsset bool
	AsForeignAsset types.U32

	IsStaking bool
	AsStaking StakingCurrency
}

func (c CurrencyID) Encode(encoder scale.Encoder) error {
	switch {
	case c.IsNative:
		return encoder.PushByte(0)
	case c.IsTranche:
		if err := encoder.PushByte(1); err != nil {
			return err
		}

		return encoder.Encode(c.AsTranche)
	case c.IsKSM:
		return encoder.PushByte(2)
	case c.IsAUSD:
		return encoder.PushByte(3)
	case c.IsForeignAsset:
		if err := encoder.PushByte(4); err != nil {
			return err
		}

		return encoder.Encode(c.AsForeignAsset)
	case c.IsStaking:
		if err := encoder.PushByte(5); err != nil {
			return err
		}

		return encoder.Encode(c.AsStaking)
	default:
		return errors.New("unsupported currency ID")
	}
}

type Tranche struct {
	PoolID    types.U64
	TrancheID [16]byte
}

type StakingCurrency struct {
	IsBlockRewards bool
}

const (
	registerPoolCall = "PoolRegistry.register"
)

func GetRegisterPoolCallCreationFn(
	poolAdmin *types.AccountID,
	poolID types.U64,
	trancheInputs []TrancheInput,
	currency CurrencyID,
	maxReserve types.U128,
	metadata []byte,
) centchain.CallProviderFn {
	return func(meta *types.Metadata) (*types.Call, error) {
		call, err := types.NewCall(
			meta,
			registerPoolCall,
			poolAdmin,
			poolID,
			trancheInputs,
			currency,
			maxReserve,
			types.NewOption(metadata),
		)

		if err != nil {
			return nil, err
		}

		return &call, nil
	}
}

// Permissions

type Role struct {
	IsPoolRole bool
	AsPoolRole PoolRole

	IsPermissionedCurrencyRole bool
	PermissionedCurrencyRole   PermissionedCurrencyRole
}

func (r Role) Encode(encoder scale.Encoder) error {
	switch {
	case r.IsPoolRole:
		if err := encoder.PushByte(0); err != nil {
			return err
		}

		return encoder.Encode(r.AsPoolRole)
	case r.IsPermissionedCurrencyRole:
		if err := encoder.PushByte(1); err != nil {
			return err
		}

		return encoder.Encode(r.AsPoolRole)
	default:
		return errors.New("unsupported role")
	}
}

type PoolRole struct {
	IsPoolAdmin bool

	IsBorrower bool

	IsPricingAdmin bool

	IsLiquidityAdmin bool

	IsMemberListAdmin bool

	IsLoanAdmin bool

	IsTrancheInvestor bool
	AsTrancheInvestor TrancheInvestor

	IsPODReadAccess bool
}

func (p PoolRole) Encode(encoder scale.Encoder) error {
	switch {
	case p.IsPoolAdmin:
		return encoder.PushByte(0)
	case p.IsBorrower:
		return encoder.PushByte(1)
	case p.IsPricingAdmin:
		return encoder.PushByte(2)
	case p.IsLiquidityAdmin:
		return encoder.PushByte(3)
	case p.IsMemberListAdmin:
		return encoder.PushByte(4)
	case p.IsLoanAdmin:
		return encoder.PushByte(5)
	case p.IsTrancheInvestor:
		if err := encoder.PushByte(6); err != nil {
			return err
		}

		return encoder.Encode(p.AsTrancheInvestor)
	case p.IsPODReadAccess:
		return encoder.PushByte(7)
	default:
		return errors.New("unsupported pool role")
	}
}

type PermissionedCurrencyRole struct {
	IsHolder bool
	AsHolder types.U64

	IsManager bool

	IsIssuer bool
}

func (p PermissionedCurrencyRole) Encode(encoder scale.Encoder) error {
	switch {
	case p.IsHolder:
		if err := encoder.PushByte(0); err != nil {
			return err
		}

		return encoder.Encode(p.AsHolder)
	case p.IsManager:
		return encoder.PushByte(1)
	case p.IsIssuer:
		return encoder.PushByte(2)
	default:
		return errors.New("unsupported permissioned currency role")
	}
}

type TrancheInvestor struct {
	TrancheID [16]byte
	Moment    types.U64
}

func (t TrancheInvestor) Encode(encoder scale.Encoder) error {
	if err := encoder.Encode(t.TrancheID); err != nil {
		return err
	}

	return encoder.Encode(t.Moment)
}

type PermissionScope struct {
	IsPool bool
	AsPool types.U64

	IsCurrency bool
	AsCurrency CurrencyID
}

func (p PermissionScope) Encode(encoder scale.Encoder) error {
	switch {
	case p.IsPool:
		if err := encoder.PushByte(0); err != nil {
			return err
		}

		return encoder.Encode(p.AsPool)
	case p.IsCurrency:
		if err := encoder.PushByte(1); err != nil {
			return err
		}

		return encoder.Encode(p.AsCurrency)
	default:
		return errors.New("unsupported permission scope")
	}
}

const (
	addPermissionsCall = "Permissions.add"
)

func GetPermissionsCallCreationFn(
	withRole Role,
	to *types.AccountID,
	scope PermissionScope,
	role Role,
) centchain.CallProviderFn {
	return func(meta *types.Metadata) (*types.Call, error) {
		call, err := types.NewCall(meta, addPermissionsCall, withRole, to, scope, role)

		if err != nil {
			return nil, err
		}

		return &call, nil
	}
}

// Loans

const (
	createLoanCall = "Loans.create"
)

func GetCreateLoanCallCreationFn(
	poolID types.U64,
	loanInfo loansTypes.LoanInfo,
) centchain.CallProviderFn {
	return func(meta *types.Metadata) (*types.Call, error) {
		call, err := types.NewCall(meta, createLoanCall, poolID, loanInfo)

		if err != nil {
			return nil, err
		}

		return &call, nil
	}
}

func GetCreateNFTCollectionCallCreationFn(
	collectionID types.U64,
	owner *types.AccountID,
) centchain.CallProviderFn {
	return func(meta *types.Metadata) (*types.Call, error) {
		ownerMultiAddress, err := types.NewMultiAddressFromAccountID(owner.ToBytes())

		if err != nil {
			return nil, err
		}

		call, err := types.NewCall(meta, uniques.CreateCollectionCall, collectionID, ownerMultiAddress)

		if err != nil {
			return nil, err
		}

		return &call, nil
	}
}

func GetNFTMintCallCreationFn(
	collectionID types.U64,
	itemID types.U128,
	owner *types.AccountID,
) centchain.CallProviderFn {
	return func(meta *types.Metadata) (*types.Call, error) {
		ownerMultiAddress, err := types.NewMultiAddressFromAccountID(owner.ToBytes())

		if err != nil {
			return nil, err
		}

		call, err := types.NewCall(meta, uniques.MintCall, collectionID, itemID, ownerMultiAddress)

		if err != nil {
			return nil, err
		}

		return &call, nil
	}
}

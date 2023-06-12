package permissions

import (
	"errors"

	"github.com/centrifuge/go-substrate-rpc-client/v4/scale"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

type PermissionScope struct {
	IsPool bool
	AsPool types.U64

	IsCurrency bool
	AsCurrency types.CurrencyID
}

func (p *PermissionScope) Decode(decoder scale.Decoder) error {
	b, err := decoder.ReadOneByte()
	if err != nil {
		return err
	}

	switch b {
	case 0:
		p.IsPool = true

		return decoder.Decode(p.AsPool)
	case 1:
		p.IsCurrency = true

		return decoder.Decode(p.AsCurrency)
	default:
		return errors.New("unsupported permission scope")
	}
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

type PermissionRoles struct {
	PoolAdmin               PoolAdminRole
	CurrencyAdmin           CurrencyAdminRole
	PermissionedAssetHolder PermissionedCurrencyHolders
	TranceInvestor          TrancheInvestors
}

type TrancheInvestors struct {
	Info []TrancheInvestorInfo
}

type TrancheInvestorInfo struct {
	TrancheID        [16]types.U8
	PermissionedTill types.U64
}

type PermissionedCurrencyHolders struct {
	Info types.Option[PermissionedCurrencyHolderInfo]
}

type PermissionedCurrencyHolderInfo struct {
	PermissionedTill types.U64
}

type PoolAdminRole uint32

const (
	PoolAdmin       PoolAdminRole = 0b00000001
	Borrower        PoolAdminRole = 0b00000010
	PricingAdmin    PoolAdminRole = 0b00000100
	LiquidityAdmin  PoolAdminRole = 0b00001000
	MemberListAdmin PoolAdminRole = 0b00010000
	RiskAdmin       PoolAdminRole = 0b00100000
	PodReadAccess   PoolAdminRole = 0b01000000
)

type CurrencyAdminRole uint32

const (
	PermissionedAssetManager CurrencyAdminRole = 0b00000001
	PermissionedAssetIssuer  CurrencyAdminRole = 0b00000010
)

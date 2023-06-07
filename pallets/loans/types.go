package loans

import (
	"github.com/centrifuge/go-substrate-rpc-client/v4/scale"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/pod/errors"
)

type CreatedLoanStorageEntry struct {
	Info     LoanInfo
	Borrower types.AccountID
}

type LoanInfo struct {
	Schedule     RepaymentSchedule
	Collateral   Asset
	Pricing      Pricing
	Restrictions LoanRestrictions
}

type Pricing struct {
	IsInternal bool
	AsInternal InternalPricing

	IsExternal bool
	AsExternal ExternalPricing
}

func (p *Pricing) Decode(decoder scale.Decoder) error {
	b, err := decoder.ReadOneByte()

	if err != nil {
		return err
	}

	switch b {
	case 0:
		p.IsInternal = true

		return decoder.Decode(&p.AsInternal)
	case 1:
		p.IsExternal = true

		return decoder.Decode(&p.AsExternal)
	default:
		return errors.New("unsupported pricing")
	}
}

func (p Pricing) Encode(encoder scale.Encoder) error {
	switch {
	case p.IsInternal:
		if err := encoder.PushByte(0); err != nil {
			return err
		}

		return encoder.Encode(p.AsInternal)
	case p.IsExternal:
		if err := encoder.PushByte(1); err != nil {
			return err
		}

		return encoder.Encode(p.AsExternal)
	default:
		return errors.New("unsupported pricing")
	}
}

type InternalPricing struct {
	CollateralValue types.U128
	ValuationMethod ValuationMethod
	InterestRate    types.U128
	MaxBorrowAmount MaxBorrowAmount
}

type ExternalPricing struct {
	PriceID           PriceID
	MaxBorrowQuantity types.U128
}

type PriceID struct {
	IsIsin bool
	AsIsin [12]types.U8
}

func (p *PriceID) Decode(decoder scale.Decoder) error {
	b, err := decoder.ReadOneByte()

	if err != nil {
		return err
	}

	switch b {
	case 0:
		p.IsIsin = true

		return decoder.Decode(&p.AsIsin)
	default:
		return errors.New("unsupported price ID")
	}
}

func (p PriceID) Encode(encoder scale.Encoder) error {
	switch {
	case p.IsIsin:
		if err := encoder.PushByte(0); err != nil {
			return err
		}

		return encoder.Encode(p.AsIsin)
	default:
		return errors.New("unsupported price ID")
	}
}

type LoanRestrictions struct {
	Borrows    BorrowRestrictions
	Repayments RepayRestrictions
}

type BorrowRestrictions struct {
	IsNotWrittenOff bool
	IsFullOnce      bool
}

func (r *BorrowRestrictions) Decode(decoder scale.Decoder) error {
	b, err := decoder.ReadOneByte()

	if err != nil {
		return err
	}

	switch b {
	case 0:
		r.IsNotWrittenOff = true

		return nil
	case 1:
		r.IsFullOnce = true

		return nil
	default:
		return errors.New("unsupported borrowed restrictions")
	}
}

func (r BorrowRestrictions) Encode(encoder scale.Encoder) error {
	switch {
	case r.IsNotWrittenOff:
		return encoder.PushByte(0)
	case r.IsFullOnce:
		return encoder.PushByte(1)
	default:
		return errors.New("unsupported borrowed restrictions")
	}
}

type RepayRestrictions struct {
	IsNone     bool
	IsFullOnce bool
}

func (r *RepayRestrictions) Decode(decoder scale.Decoder) error {
	b, err := decoder.ReadOneByte()

	if err != nil {
		return err
	}

	switch b {
	case 0:
		r.IsNone = true

		return nil
	case 1:
		r.IsFullOnce = true

		return nil
	default:
		return errors.New("unsupported repay restrictions")
	}
}

func (r RepayRestrictions) Encode(encoder scale.Encoder) error {
	switch {
	case r.IsNone:
		return encoder.PushByte(0)
	case r.IsFullOnce:
		return encoder.PushByte(1)
	default:
		return errors.New("unsupported repay restrictions")
	}
}

type MaxBorrowAmount struct {
	IsUpToTotalBorrowed bool
	AsUpToTotalBorrowed AdvanceRate

	IsUpToOutstandingDebt bool
	AsUpToOutstandingDebt AdvanceRate
}

type AdvanceRate struct {
	AdvanceRate types.U128
}

func (m *MaxBorrowAmount) Decode(decoder scale.Decoder) error {
	b, err := decoder.ReadOneByte()

	if err != nil {
		return err
	}

	switch b {
	case 0:
		m.IsUpToTotalBorrowed = true

		return decoder.Decode(&m.AsUpToTotalBorrowed)
	case 1:
		m.IsUpToOutstandingDebt = true

		return decoder.Decode(&m.AsUpToOutstandingDebt)
	default:
		return errors.New("unsupported max borrow amount")
	}
}

func (m MaxBorrowAmount) Encode(encoder scale.Encoder) error {
	switch {
	case m.IsUpToTotalBorrowed:
		if err := encoder.PushByte(0); err != nil {
			return err
		}

		return encoder.Encode(m.AsUpToTotalBorrowed)
	case m.IsUpToOutstandingDebt:
		if err := encoder.PushByte(1); err != nil {
			return err
		}

		return encoder.Encode(m.AsUpToOutstandingDebt)
	default:
		return errors.New("unsupported max borrow amount")
	}
}

type ValuationMethod struct {
	IsDiscountedCashFlow bool
	AsDiscountedCashFlow DiscountedCashFlow

	IsOutstandingDebt bool
}

func (v *ValuationMethod) Decode(decoder scale.Decoder) error {
	b, err := decoder.ReadOneByte()

	if err != nil {
		return err
	}

	switch b {
	case 0:
		v.IsDiscountedCashFlow = true

		return decoder.Decode(&v.AsDiscountedCashFlow)
	case 1:
		v.IsOutstandingDebt = true

		return nil
	default:
		return errors.New("unsupported valuation method")
	}
}

func (v ValuationMethod) Encode(encoder scale.Encoder) error {
	switch {
	case v.IsDiscountedCashFlow:
		if err := encoder.PushByte(0); err != nil {
			return nil
		}

		return encoder.Encode(v.AsDiscountedCashFlow)
	case v.IsOutstandingDebt:
		return encoder.PushByte(1)
	default:
		return errors.New("unsupported valuation method")
	}
}

type DiscountedCashFlow struct {
	ProbabilityOfDefault types.U128
	LossGivenDefault     types.U128
	DiscountRate         types.U128
}

type Asset struct {
	CollectionID types.U64
	ItemID       types.U128
}

type RepaymentSchedule struct {
	Maturity         Maturity
	InterestPayments InterestPayments
	PayDownSchedule  PayDownSchedule
}

type PayDownSchedule struct {
	IsNone bool
}

func (p *PayDownSchedule) Decode(decoder scale.Decoder) error {
	b, err := decoder.ReadOneByte()

	if err != nil {
		return err
	}

	switch b {
	case 0:
		p.IsNone = true

		return nil
	default:
		return errors.New("unsupported pay down schedule")
	}
}

func (p PayDownSchedule) Encode(encoder scale.Encoder) error {
	if !p.IsNone {
		return errors.New("invalid pay down schedule")
	}

	return encoder.PushByte(0)
}

type InterestPayments struct {
	IsNone bool
}

func (i *InterestPayments) Decode(decoder scale.Decoder) error {
	b, err := decoder.ReadOneByte()

	if err != nil {
		return err
	}

	switch b {
	case 0:
		i.IsNone = true

		return nil
	default:
		return errors.New("unsupported interest payments")
	}
}

func (i InterestPayments) Encode(encoder scale.Encoder) error {
	if !i.IsNone {
		return errors.New("invalid interest payments")
	}

	return encoder.PushByte(0)
}

type Maturity struct {
	IsFixed bool
	AsFixed types.U64
}

func (m *Maturity) Decode(decoder scale.Decoder) error {
	b, err := decoder.ReadOneByte()

	if err != nil {
		return err
	}

	switch b {
	case 0:
		m.IsFixed = true

		return decoder.Decode(&m.AsFixed)
	default:
		return errors.New("unsupported maturity")
	}
}

func (m Maturity) Encode(encoder scale.Encoder) error {
	if !m.IsFixed {
		return errors.New("invalid maturity")
	}

	if err := encoder.PushByte(0); err != nil {
		return nil
	}

	return encoder.Encode(m.AsFixed)
}

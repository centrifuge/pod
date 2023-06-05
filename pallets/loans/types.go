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

type ActiveLoanStorageEntry struct {
	ActiveLoan ActiveLoan
	Moment     types.U64
}

type ActiveLoan struct {
	LoanID          types.U64
	Info            LoanInfo
	Borrower        types.AccountID
	WriteOffStatus  WriteOffStatus
	OriginationDate types.U64
	NormalizedDebt  types.U128
	TotalBorrowed   types.U128
	TotalRepaid     types.U128
}

type WriteOffStatus struct {
	Percentage types.U128
	Penalty    types.U128
}

type LoanInfo struct {
	Schedule        RepaymentSchedule
	Collateral      Asset
	CollateralValue types.U128
	ValuationMethod ValuationMethod
	Restrictions    LoanRestrictions
	InterestRate    types.U128
}

type LoanRestrictions struct {
	MaxBorrowAmount MaxBorrowAmount
	Borrows         BorrowRestrictions
	Repayments      RepayRestrictions
}

type BorrowRestrictions struct {
	IsWrittenOff bool
}

func (r *BorrowRestrictions) Decode(decoder scale.Decoder) error {
	b, err := decoder.ReadOneByte()

	if err != nil {
		return err
	}

	switch b {
	case 0:
		r.IsWrittenOff = true

		return nil
	default:
		return errors.New("unsupported borrowed restrictions")
	}
}

func (r BorrowRestrictions) Encode(encoder scale.Encoder) error {
	switch {
	case r.IsWrittenOff:
		return encoder.PushByte(0)
	default:
		return errors.New("unsupported borrowed restrictions")
	}
}

type RepayRestrictions struct {
	IsNone bool
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
	default:
		return errors.New("unsupported repay restrictions")
	}
}

func (r RepayRestrictions) Encode(encoder scale.Encoder) error {
	switch {
	case r.IsNone:
		return encoder.PushByte(0)
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

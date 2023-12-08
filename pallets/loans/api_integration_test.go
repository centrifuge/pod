//go:build integration

package loans_test

import (
	"os"
	"testing"

	"github.com/centrifuge/pod/bootstrap"
	"github.com/centrifuge/pod/bootstrap/bootstrappers/integration_test"
	"github.com/centrifuge/pod/bootstrap/bootstrappers/testlogging"
	"github.com/centrifuge/pod/centchain"
	"github.com/centrifuge/pod/config"
	"github.com/centrifuge/pod/config/configstore"
	"github.com/centrifuge/pod/dispatcher"
	v2 "github.com/centrifuge/pod/identity/v2"
	"github.com/centrifuge/pod/jobs"
	"github.com/centrifuge/pod/pallets"
	"github.com/centrifuge/pod/pallets/loans"
	"github.com/centrifuge/pod/storage/leveldb"
	genericUtils "github.com/centrifuge/pod/testingutils/generic"
)

var integrationTestBootstrappers = []bootstrap.TestBootstrapper{
	&integration_test.Bootstrapper{},
	&testlogging.TestLoggingBootstrapper{},
	&config.Bootstrapper{},
	&leveldb.Bootstrapper{},
	&configstore.Bootstrapper{},
	&jobs.Bootstrapper{},
	centchain.Bootstrapper{},
	&pallets.Bootstrapper{},
	&dispatcher.Bootstrapper{},
	&v2.Bootstrapper{},
}

var (
	serviceCtx map[string]any
	loansAPI   loans.API
)

func TestMain(m *testing.M) {
	serviceCtx = bootstrap.RunTestBootstrappers(integrationTestBootstrappers, nil)
	loansAPI = genericUtils.GetService[loans.API](serviceCtx)

	result := m.Run()

	bootstrap.RunTestTeardown(integrationTestBootstrappers)

	os.Exit(result)
}

// NOTE - The following test is disabled for now since we have to update the loan-related entities.
//
//func TestIntegration_CreatedLoanRetrieval(t *testing.T) {
//	testKeyring := keyrings.AliceKeyRingPair
//
//	testAccountID, err := types.NewAccountID(testKeyring.PublicKey)
//	assert.NoError(t, err)
//
//	nftCollectionID := types.U64(rand.Uint32())
//	nftItemID := types.NewU128(*big.NewInt(rand.Int63()))
//
//	poolID := types.U64(rand.Int63())
//
//	// Create NFT.
//
//	nftCollectionCall := pallets.GetCreateNFTCollectionCallCreationFn(nftCollectionID, testAccountID)
//	nftMintCall := pallets.GetNFTMintCallCreationFn(nftCollectionID, nftItemID, testAccountID)
//
//	// Create a pool using Alice's account as the owner.
//
//	registerPoolCall := pallets.GetRegisterPoolCallCreationFn(
//		testAccountID,
//		poolID,
//		[]pallets.TrancheInput{
//			{
//				TrancheType: pallets.TrancheType{
//					IsResidual: true,
//				},
//				Seniority: types.NewOption[types.U32](0),
//				TrancheMetadata: pallets.TrancheMetadata{
//					TokenName:   []byte("CFG-TEST-1"),
//					TokenSymbol: []byte("CFGT1"),
//				},
//			},
//			{
//				TrancheType: pallets.TrancheType{
//					IsNonResidual: true,
//					AsNonResidual: pallets.NonResidual{
//						InterestRatePerSec: types.NewU128(*big.NewInt(1)),
//						MinRiskBuffer:      5,
//					},
//				},
//				Seniority: types.NewOption[types.U32](1),
//				TrancheMetadata: pallets.TrancheMetadata{
//					TokenName:   []byte("CFG-TEST-2"),
//					TokenSymbol: []byte("CFGT2"),
//				},
//			},
//		},
//		pallets.CurrencyID{
//			IsForeignAsset: true,
//			AsForeignAsset: types.U32(1),
//		},
//		types.NewU128(*big.NewInt(rand.Int63())),
//		[]byte("test"),
//	)
//
//	// Assign the Borrower permission to Alice's account.
//
//	addBorrowerPermissionsCall := pallets.GetPermissionsCallCreationFn(
//		pallets.Role{
//			IsPoolRole: true,
//			AsPoolRole: pallets.PoolRole{IsPoolAdmin: true},
//		},
//		testAccountID,
//		pallets.PermissionScope{
//			IsPool: true,
//			AsPool: poolID,
//		},
//		pallets.Role{
//			IsPoolRole: true,
//			AsPoolRole: pallets.PoolRole{IsBorrower: true},
//		},
//	)
//
//	// Create a test loan using some random info.
//
//	loanInfo := loansTypes.LoanInfo{
//		Schedule: loansTypes.RepaymentSchedule{
//			Maturity: loansTypes.Maturity{
//				IsFixed: true,
//				// 1 Year maturity date.
//				AsFixed: loansTypes.FixedMaturity{
//					Date:      types.U64(time.Now().Add(356 * 24 * time.Hour).Unix()),
//					Extension: 0,
//				},
//			},
//			InterestPayments: loansTypes.InterestPayments{
//				IsNone: true,
//			},
//			PayDownSchedule: loansTypes.PayDownSchedule{
//				IsNone: true,
//			},
//		},
//		Collateral: loansTypes.Asset{
//			CollectionID: nftCollectionID,
//			ItemID:       nftItemID,
//		},
//		InterestRate: loansTypes.InterestRate{
//			IsFixed: true,
//			AsFixed: loansTypes.FixedInterestRate{
//				RatePerYear: types.NewU128(*big.NewInt(0)),
//				Compounding: loansTypes.CompoundingSchedule{
//					IsSecondly: true,
//				},
//			},
//		},
//		Pricing: loansTypes.Pricing{
//			IsInternal: true,
//			AsInternal: loansTypes.InternalPricing{
//				CollateralValue: types.NewU128(*big.NewInt(rand.Int63())),
//				ValuationMethod: loansTypes.ValuationMethod{
//					IsOutstandingDebt: true,
//				},
//				MaxBorrowAmount: loansTypes.InternalPricingMaxBorrowAmount{
//					IsUpToTotalBorrowed: true,
//					AsUpToTotalBorrowed: loansTypes.AdvanceRate{
//						AdvanceRate: types.NewU128(*big.NewInt(11)),
//					},
//				},
//			},
//		},
//		Restrictions: loansTypes.LoanRestrictions{
//			Borrows: loansTypes.BorrowRestrictions{
//				IsNotWrittenOff: true,
//			},
//			Repayments: loansTypes.RepayRestrictions{
//				IsNone: true,
//			},
//		},
//	}
//
//	loanCreateCall := pallets.GetCreateLoanCallCreationFn(poolID, loanInfo)
//
//	// Execute the batch call using the test keyring.
//
//	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
//	defer cancel()
//
//	err = pallets.ExecuteWithTestClient(
//		ctx,
//		serviceCtx,
//		testKeyring,
//		utility.BatchCalls(
//			nftCollectionCall,
//			nftMintCall,
//			registerPoolCall,
//			addBorrowerPermissionsCall,
//			loanCreateCall,
//		),
//	)
//	assert.NoError(t, err)
//
//	// The first loan created for a pool will have index/ID 0.
//
//	loanID := types.U64(1)
//
//	res, err := loansAPI.GetCreatedLoan(poolID, loanID)
//	assert.NoError(t, err)
//	assert.Equal(t, loanInfo.Collateral.CollectionID, res.Info.Collateral.CollectionID)
//	assert.Equal(t, loanInfo.Collateral.ItemID, res.Info.Collateral.ItemID)
//}

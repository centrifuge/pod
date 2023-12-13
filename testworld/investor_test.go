//go:build testworld

package testworld

import (
	"context"
	"fmt"
	"math/big"
	"math/rand"
	"net/http"
	"testing"
	"time"

	loansTypes "github.com/centrifuge/chain-custom-types/pkg/loans"
	proxyType "github.com/centrifuge/chain-custom-types/pkg/proxy"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/pod/http/auth/token"
	nftv3 "github.com/centrifuge/pod/nft/v3"
	"github.com/centrifuge/pod/pallets"
	"github.com/centrifuge/pod/pallets/proxy"
	"github.com/centrifuge/pod/pallets/utility"
	"github.com/centrifuge/pod/testworld/park/behavior/client"
	"github.com/centrifuge/pod/testworld/park/host"
	"github.com/stretchr/testify/assert"
)

func TestInvestorAPI_GetAsset(t *testing.T) {
	t.Parallel()

	alice, err := controller.GetHost(host.Alice)
	assert.NoError(t, err)

	// Create document

	aliceClient, err := controller.GetClientForHost(t, host.Alice)
	assert.NoError(t, err)

	payload := genericCoreAPICreate(nil)

	res := aliceClient.CreateDocument("documents", http.StatusCreated, payload)
	assert.Equal(t, client.GetDocumentStatus(res), "pending")
	docID := client.GetDocumentIdentifier(res)

	// Create NFT collection

	collectionID := types.U64(rand.Int63())

	payload = map[string]interface{}{
		"collection_id": collectionID,
	}

	createClassRes := aliceClient.CreateNFTCollection(http.StatusAccepted, payload)

	jobID, err := client.GetJobID(createClassRes)
	assert.NoError(t, err)

	err = aliceClient.WaitForJobCompletion(jobID)
	assert.NoError(t, err)

	// Commit and mint NFT

	ipfsName := "ipfs_name"
	ipfsDescription := "ipfs_description"
	ipfsImage := "ipfs_image"

	ipfsMetadata := nftv3.IPFSMetadata{
		Name:                  ipfsName,
		Description:           ipfsDescription,
		Image:                 ipfsImage,
		DocumentAttributeKeys: nil,
	}

	payload = map[string]interface{}{
		"collection_id":   collectionID,
		"document_id":     docID,
		"owner":           alice.GetMainAccount().GetAccountID().ToHexString(),
		"ipfs_metadata":   ipfsMetadata,
		"freeze_metadata": false,
	}

	mintRes := aliceClient.CommitAndMintNFT(http.StatusAccepted, payload)

	jobID, err = client.GetJobID(mintRes)
	assert.NoError(t, err)

	err = aliceClient.WaitForJobCompletion(jobID)
	assert.NoError(t, err)

	// Get NFT item ID

	docVal := aliceClient.GetDocumentAndVerify(docID, nil, nil)
	itemIDRaw := docVal.Path("$.header.nfts[0].item_id").String().Raw()

	i := new(big.Int)
	bi, ok := i.SetString(itemIDRaw, 10)
	assert.True(t, ok)

	itemID := types.NewU128(*bi)

	// Create a pool with the main account on the Alice host as admin.

	poolID := types.U64(rand.Int63())

	registerPoolCall := pallets.GetRegisterPoolCallCreationFn(
		alice.GetMainAccount().GetAccountID(),
		poolID,
		[]pallets.TrancheInput{
			{
				TrancheType: pallets.TrancheType{
					IsResidual: true,
				},
				Seniority: types.NewOption[types.U32](0),
				TrancheMetadata: pallets.TrancheMetadata{
					TokenName:   []byte("CFG-TEST-1"),
					TokenSymbol: []byte("CFGT1"),
				},
			},
			{
				TrancheType: pallets.TrancheType{
					IsNonResidual: true,
					AsNonResidual: pallets.NonResidual{
						InterestRatePerSec: types.NewU128(*big.NewInt(1)),
						MinRiskBuffer:      5,
					},
				},
				Seniority: types.NewOption[types.U32](1),
				TrancheMetadata: pallets.TrancheMetadata{
					TokenName:   []byte("CFG-TEST-2"),
					TokenSymbol: []byte("CFGT2"),
				},
			},
		},
		pallets.CurrencyID{
			IsForeignAsset: true,
			AsForeignAsset: types.U32(1),
		},
		types.NewU128(*big.NewInt(rand.Int63())),
		[]byte("test"),
		[]pallets.WriteOffRule{},
	)

	// Assign the Borrower permission to the main account on the Alice host.

	addBorrowerPermissionsCall := pallets.GetPermissionsCallCreationFn(
		pallets.Role{
			IsPoolRole: true,
			AsPoolRole: pallets.PoolRole{IsPoolAdmin: true},
		},
		alice.GetMainAccount().GetAccountID(),
		pallets.PermissionScope{
			IsPool: true,
			AsPool: poolID,
		},
		pallets.Role{
			IsPoolRole: true,
			AsPoolRole: pallets.PoolRole{IsBorrower: true},
		},
	)

	// Assign the PODReadAccess permission to the POD auth proxy account on the Alice host.

	addPODReadPermissionsCall := pallets.GetPermissionsCallCreationFn(
		pallets.Role{
			IsPoolRole: true,
			AsPoolRole: pallets.PoolRole{IsPoolAdmin: true},
		},
		alice.GetMainAccount().GetPodAuthProxyAccountID(),
		pallets.PermissionScope{
			IsPool: true,
			AsPool: poolID,
		},
		pallets.Role{
			IsPoolRole: true,
			AsPoolRole: pallets.PoolRole{IsPODReadAccess: true},
		},
	)

	// Create a loan using the NFT that was minted as collateral.

	loanCreateCall := pallets.GetCreateLoanCallCreationFn(poolID, loansTypes.LoanInfo{
		Schedule: loansTypes.RepaymentSchedule{
			Maturity: loansTypes.Maturity{
				IsFixed: true,
				// 1 Year maturity date.
				AsFixed: loansTypes.FixedMaturity{
					Date:      types.U64(time.Now().Add(356 * 24 * time.Hour).Unix()),
					Extension: 0,
				},
			},
			InterestPayments: loansTypes.InterestPayments{
				IsNone: true,
			},
			PayDownSchedule: loansTypes.PayDownSchedule{
				IsNone: true,
			},
		},
		Collateral: loansTypes.Asset{
			CollectionID: collectionID,
			ItemID:       itemID,
		},
		InterestRate: loansTypes.InterestRate{
			IsFixed: true,
			AsFixed: loansTypes.FixedInterestRate{
				RatePerYear: types.NewU128(*big.NewInt(0)),
				Compounding: loansTypes.CompoundingSchedule{
					IsSecondly: true,
				},
			},
		},
		Pricing: loansTypes.Pricing{
			IsInternal: true,
			AsInternal: loansTypes.InternalPricing{
				CollateralValue: types.NewU128(*big.NewInt(rand.Int63())),
				ValuationMethod: loansTypes.ValuationMethod{
					IsOutstandingDebt: true,
				},
				MaxBorrowAmount: loansTypes.InternalPricingMaxBorrowAmount{
					IsUpToTotalBorrowed: true,
					AsUpToTotalBorrowed: loansTypes.AdvanceRate{
						AdvanceRate: types.NewU128(*big.NewInt(11)),
					},
				},
			},
		},
		Restrictions: loansTypes.LoanRestrictions{
			Borrows: loansTypes.BorrowRestrictions{
				IsNotWrittenOff: true,
			},
			Repayments: loansTypes.RepayRestrictions{
				IsNone: true,
			},
		},
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	// Alice will execute a proxied batch call on behalf of the main account.
	err = pallets.ExecuteWithTestClient(
		ctx,
		alice.GetServiceCtx(),
		alice.GetOriginKeyringPair(),
		proxy.WrapWithProxyCall(
			alice.GetMainAccount().GetAccountID(),
			types.NewOption(proxyType.Any),
			utility.BatchCalls(
				registerPoolCall,
				addPODReadPermissionsCall,
				addBorrowerPermissionsCall,
				loanCreateCall,
			),
		),
	)
	assert.NoError(t, err)

	// Use a PodAuth token and client since the POD auth proxy account ID is the one with the PODReadAccess permission.
	podAuthToken, err := alice.GetMainAccount().GetJW3Token(proxyType.ProxyTypeName[proxyType.PodAuth])
	assert.NoError(t, err)

	testClient := client.New(t, controller.GetWebhookReceiver(), alice.GetAPIURL(), podAuthToken)

	// This should be the index of a newly created loan.
	loanID := 1

	assetRequest := &client.AssetRequest{
		PoolID:  fmt.Sprintf("%d", poolID),
		LoanID:  fmt.Sprintf("%d", loanID),
		AssetID: docID,
	}

	assetRes := testClient.GetAsset(assetRequest, http.StatusOK)
	assetID := client.GetDocumentIdentifier(assetRes)

	assert.Equal(t, docID, assetID)

	podOperationAuthToken, err := alice.GetMainAccount().GetJW3Token(proxyType.ProxyTypeName[proxyType.PodOperation])
	assert.NoError(t, err)

	testClient.SetAuthToken(podOperationAuthToken)

	_ = testClient.GetAsset(assetRequest, http.StatusForbidden)

	podAdminAuthToken, err := alice.GetMainAccount().GetJW3Token(token.PodAdminProxyType)
	assert.NoError(t, err)

	testClient.SetAuthToken(podAdminAuthToken)

	_ = testClient.GetAsset(assetRequest, http.StatusForbidden)
}

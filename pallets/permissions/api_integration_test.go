//go:build integration

package permissions_test

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
	"github.com/centrifuge/pod/pallets/permissions"
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
	serviceCtx     map[string]any
	cfgService     config.Service
	permissionsAPI permissions.API
)

func TestMain(m *testing.M) {
	serviceCtx = bootstrap.RunTestBootstrappers(integrationTestBootstrappers, nil)
	cfgService = genericUtils.GetService[config.Service](serviceCtx)
	permissionsAPI = genericUtils.GetService[permissions.API](serviceCtx)

	result := m.Run()

	bootstrap.RunTestTeardown(integrationTestBootstrappers)

	os.Exit(result)
}

// NOTE - The following test is disabled for now since we have to update the pool-related entities.
//
//func TestIntegration_PermissionRolesRetrieval(t *testing.T) {
//	testKeyring := keyrings.AliceKeyRingPair
//
//	testAccountID, err := types.NewAccountID(testKeyring.PublicKey)
//	assert.NoError(t, err)
//
//	poolID := types.U64(rand.Uint32())
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
//			registerPoolCall,
//			addBorrowerPermissionsCall,
//		),
//	)
//	assert.NoError(t, err)
//
//	res, err := permissionsAPI.GetPermissionRoles(testAccountID, poolID)
//	assert.NoError(t, err)
//	assert.True(t, res.PoolAdmin&permissions.Borrower == permissions.Borrower)
//}

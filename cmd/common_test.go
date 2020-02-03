// +build integration

package cmd

import (
	"context"
	"os"
	"os/exec"
	"path"
	"testing"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testlogging"
	"github.com/centrifuge/go-centrifuge/centchain"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/crypto/ed25519"
	"github.com/centrifuge/go-centrifuge/crypto/secp256k1"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/identity/ideth"
	"github.com/centrifuge/go-centrifuge/jobs/jobsv1"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/storage/leveldb"
	"github.com/centrifuge/go-centrifuge/testingutils"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

var ctx = map[string]interface{}{}

func TestMain(m *testing.M) {
	var bootstrappers = []bootstrap.TestBootstrapper{
		&testlogging.TestLoggingBootstrapper{},
		&config.Bootstrapper{},
		&leveldb.Bootstrapper{},
		jobsv1.Bootstrapper{},
		&queue.Bootstrapper{},
		centchain.Bootstrapper{},
		ethereum.Bootstrapper{},
		&ideth.Bootstrapper{},
		&configstore.Bootstrapper{},
		&queue.Starter{},
	}

	bootstrap.RunTestBootstrappers(bootstrappers, ctx)
	result := m.Run()
	bootstrap.RunTestTeardown(bootstrappers)
	os.Exit(result)
}

func TestCreateConfig(t *testing.T) {
	// create config
	dataDir := "testconfig"
	keyPath := path.Join(testingutils.GetProjectDir(), "build/scripts/test-dependencies/test-ethereum/migrateAccount.json")
	scAddrs := testingutils.GetSmartContractAddresses()
	err := CreateConfig(
		dataDir,
		"http://127.0.0.1:9545",
		keyPath,
		"", "russianhill",
		"127.0.0.1", 8028, 38202,
		nil, true, false, "", scAddrs, "",
		"ws://127.0.0.1:9944",
		"0xc81ebbec0559a6acf184535eb19da51ed3ed8c4ac65323999482aaf9b6696e27",
		"0xc166b100911b1e9f780bb66d13badf2c1edbe94a1220f1a0584c09490158be31",
		"5Gb6Zfe8K8NSKrkFLCgqs8LUdk7wKweXM5pN296jVqDpdziR")
	assert.Nil(t, err, "Create Config should be successful")

	// config exists
	cfg := config.LoadConfiguration(path.Join(dataDir, "config.yaml"))
	client := ctx[ethereum.BootstrappedEthereumClient].(ethereum.Client)

	// contract exists
	id, err := cfg.GetIdentityID()
	accountId := identity.NewDID(common.BytesToAddress(id))

	assert.Nil(t, err, "did should exists")
	contractCode, err := client.GetEthClient().CodeAt(context.Background(), common.BytesToAddress(id), nil)
	assert.Nil(t, err, "should be successful to get the contract code")
	assert.Equal(t, true, len(contractCode) > 3000, "current contract code should be around 3378 bytes")

	// Keys exists
	// type KeyPurposeP2P
	idSrv := ctx[identity.BootstrappedDIDService].(identity.Service)
	pk, _, err := ed25519.GetSigningKeyPair(cfg.GetP2PKeyPair())
	assert.Nil(t, err)
	pk32, err := utils.SliceToByte32(pk)
	assert.Nil(t, err)
	response, _ := idSrv.GetKey(accountId, pk32)
	assert.NotNil(t, response)
	assert.Equal(t, &(identity.KeyPurposeP2PDiscovery.Value), response.Purposes[0], "purpose should be P2P")

	// type KeyPurposeSigning
	pk, _, err = secp256k1.GetSigningKeyPair(cfg.GetSigningKeyPair())
	assert.Nil(t, err)
	address32Bytes := utils.AddressTo32Bytes(common.HexToAddress(secp256k1.GetAddress(pk)))
	assert.Nil(t, err)
	response, _ = idSrv.GetKey(accountId, address32Bytes)
	assert.NotNil(t, response)
	assert.Equal(t, &(identity.KeyPurposeSigning.Value), response.Purposes[0], "purpose should be Signing")

	err = exec.Command("rm", "-rf", dataDir).Run()
	assert.Nil(t, err, "removing testconfig folder should be successful")

}

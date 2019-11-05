// +build cmd

package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"testing"

	"github.com/centrifuge/go-centrifuge/testingutils"
	"github.com/centrifuge/go-centrifuge/version"
	"github.com/stretchr/testify/assert"
)

func TestVersion(t *testing.T) {
	o, err := exec.Command(testingutils.GetBinaryPath(), "version").Output()
	assert.NoError(t, err)

	assert.Contains(t, string(o), version.CentrifugeNodeVersion)
}

func TestCreateConfigCmd(t *testing.T) {
	dataDir := path.Join(os.Getenv("HOME"), "datadir_test")
	scAddrs := testingutils.GetSmartContractAddresses()
	keyPath := path.Join(testingutils.GetProjectDir(), "build/scripts/test-dependencies/test-ethereum/migrateAccount.json")
	cmd := exec.Command(
		testingutils.GetBinaryPath(),
		"createconfig", "-n", "testing", "-t", dataDir, "-z", keyPath,
		"--centchainaddr", "5Gb6Zfe8K8NSKrkFLCgqs8LUdk7wKweXM5pN296jVqDpdziR",
		"--centchainid", "0xc81ebbec0559a6acf184535eb19da51ed3ed8c4ac65323999482aaf9b6696e27",
		"--centchainsecret", "0xc166b100911b1e9f780bb66d13badf2c1edbe94a1220f1a0584c09490158be31")
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, fmt.Sprintf("CENT_NETWORKS_TESTING_CONTRACTADDRESSES_IDENTITYFACTORY=%s", scAddrs.IdentityFactoryAddr))
	cmd.Env = append(cmd.Env, fmt.Sprintf("CENT_NETWORKS_TESTING_CONTRACTADDRESSES_ANCHORREPOSITORY=%s", scAddrs.AnchorRepositoryAddr))
	cmd.Env = append(cmd.Env, fmt.Sprintf("CENT_NETWORKS_TESTING_CONTRACTADDRESSES_INVOICEUNPAID=%s", scAddrs.InvoiceUnpaidAddr))
	o, err := cmd.Output()
	assert.NoError(t, err)

	fmt.Printf("Output: %s\n", o)
}

//go:build cmd

package main

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

func pathToCmd() string {
	root := testingutils.GetProjectDir()
	return fmt.Sprintf("%s/cmd/centrifuge", root)
}

func TestVersion(t *testing.T) {
	cmd := exec.Command("go", "run", pathToCmd(), "version")
	o, err := cmd.Output()
	assert.NoError(t, err)
	assert.Contains(t, string(o), version.CentrifugeNodeVersion)
}

func TestCreateConfigCmd(t *testing.T) {
	dataDir := path.Join(os.Getenv("HOME"), "datadir_test")
	scAddrs := testingutils.GetSmartContractAddresses()
	keyPath := path.Join(testingutils.GetProjectDir(), "build/scripts/test-dependencies/test-ethereum/migrateAccount.json")
	cmd := exec.Command(
		"go", "run", pathToCmd(),
		"createconfig", "-n", "testing", "-t", dataDir, "-z", keyPath,
		"--centchainaddr", "5GrwvaEF5zXb26Fz9rcQpDWS57CtERHpNehXCPcNoHGKutQY",
		"--centchainid", "0xd43593c715fdd31c61141abd04a99fd6822c8558854ccde39a5684e7a56da27d",
		"--centchainsecret", "//Alice")
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, fmt.Sprintf("CENT_NETWORKS_TESTING_CONTRACTADDRESSES_IDENTITYFACTORY=%s", scAddrs.IdentityFactoryAddr))
	fmt.Println(cmd.String())
	o, err := cmd.Output()
	assert.NoError(t, err)

	fmt.Printf("Output: %s\n", o)
}

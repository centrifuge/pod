// +build integration

package cmd

import (
	"github.com/centrifuge/go-centrifuge/identity/did"
	"github.com/centrifuge/go-centrifuge/testingutils"
	"github.com/stretchr/testify/assert"
	"os"
	"path"
	"testing"
)

func TestCreateConfig(t *testing.T) {
	dataDir := path.Join(os.Getenv("HOME"), "datadir")
	keyPath := path.Join(testingutils.GetProjectDir(), "build/scripts/test-dependencies/test-ethereum/migrateAccount.json")
	scAddrs := did.GetSmartContractAddresses()
	err := CreateConfig(dataDir,"http://127.0.0.1:9545",keyPath,"","russianhill",8028,38202,nil,true,"",scAddrs)
	assert.Nil(t,err,"Create Config should be successful")

}

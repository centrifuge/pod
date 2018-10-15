// +build unit

package purchaseorder

import (
	"os"
	"testing"

	cc "github.com/centrifuge/go-centrifuge/centrifuge/context/testingbootstrap"
	"github.com/centrifuge/go-centrifuge/centrifuge/coredocument/repository"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	cc.TestIntegrationBootstrap()
	levelDB := cc.GetLevelDBStorage()
	coredocumentrepository.InitLevelDBRepository(levelDB)
	InitLevelDBRepository(levelDB)
	result := m.Run()
	cc.TestIntegrationTearDown()
	os.Exit(result)
}

func TestBootstrapper_Bootstrap(t *testing.T) {
	err := (&Bootstrapper{}).Bootstrap(map[string]interface{}{})
	assert.Error(t, err, "Should throw an error because of empty context")
}

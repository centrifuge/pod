// +build integration

package invoice_test

import (
	"os"
	"testing"

	cc "github.com/centrifuge/go-centrifuge/centrifuge/context/testingbootstrap"
	"github.com/centrifuge/go-centrifuge/centrifuge/testingutils"
)

func TestMain(m *testing.M) {
	cc.DONT_USE_FOR_UNIT_TESTS_TestFunctionalEthereumBootstrap()
	testingutils.CreateIdentityWithKeys()
	result := m.Run()
	cc.TestFunctionalEthereumTearDown()
	os.Exit(result)
}

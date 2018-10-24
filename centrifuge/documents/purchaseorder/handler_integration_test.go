// +build integration

package purchaseorder_test

import (
	"os"
	"testing"

	"github.com/centrifuge/go-centrifuge/centrifuge/config"
	cc "github.com/centrifuge/go-centrifuge/centrifuge/context/testingbootstrap"
	"github.com/centrifuge/go-centrifuge/centrifuge/testingutils"
)

func TestMain(m *testing.M) {
	cc.DONT_USE_FOR_UNIT_TESTS_TestFunctionalEthereumBootstrap()
	config.Config.V.Set("keys.signing.publicKey", "../../../example/resources/signature1.pub.pem")
	config.Config.V.Set("keys.signing.privateKey", "../../../example/resources/signature1.key.pem")
	config.Config.V.Set("keys.ethauth.publicKey", "../../../example/resources/ethauth.pub.pem")
	config.Config.V.Set("keys.ethauth.privateKey", "../../../example/resources/ethauth.key.pem")
	testingutils.CreateIdentityWithKeys()
	result := m.Run()
	cc.TestFunctionalEthereumTearDown()
	os.Exit(result)
}

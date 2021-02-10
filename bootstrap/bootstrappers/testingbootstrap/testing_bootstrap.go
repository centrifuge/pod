// +build integration

package testingbootstrap

import (
	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testlogging"
	"github.com/centrifuge/go-centrifuge/centchain"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/documents/entity"
	"github.com/centrifuge/go-centrifuge/documents/entityrelationship"
	"github.com/centrifuge/go-centrifuge/documents/generic"
	"github.com/centrifuge/go-centrifuge/ethereum"
	v2 "github.com/centrifuge/go-centrifuge/http/v2"
	"github.com/centrifuge/go-centrifuge/identity/ideth"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/nft"
	"github.com/centrifuge/go-centrifuge/oracle"
	"github.com/centrifuge/go-centrifuge/p2p"
	"github.com/centrifuge/go-centrifuge/pending"
	"github.com/centrifuge/go-centrifuge/storage/leveldb"
	"github.com/centrifuge/go-centrifuge/testingutils"
	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("context")

var bootstrappers = []bootstrap.TestBootstrapper{
	&testlogging.TestLoggingBootstrapper{},
	&config.Bootstrapper{},
	&leveldb.Bootstrapper{},
	jobs.Bootstrapper{},
	centchain.Bootstrapper{},
	ethereum.Bootstrapper{},
	&ideth.Bootstrapper{},
	&configstore.Bootstrapper{},
	anchors.Bootstrapper{},
	documents.Bootstrapper{},
	&entityrelationship.Bootstrapper{},
	generic.Bootstrapper{},
	&nft.Bootstrapper{},
	p2p.Bootstrapper{},
	documents.PostBootstrapper{},
	pending.Bootstrapper{},
	&entity.Bootstrapper{},
	oracle.Bootstrapper{},
	v2.Bootstrapper{},
}

func TestFunctionalEthereumBootstrap() map[string]interface{} {
	cm := testingutils.BuildIntegrationTestingContext()
	for _, b := range bootstrappers {
		err := b.TestBootstrap(cm)
		if err != nil {
			log.Error("Error encountered while bootstrapping", err)
			panic(err)
		}
	}

	return cm
}
func TestFunctionalEthereumTearDown() {
	for _, b := range bootstrappers {
		err := b.TestTearDown()
		if err != nil {
			log.Error("Error encountered while bootstrapping", err)
			panic(err)
		}
	}
}

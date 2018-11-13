// +build unit integration

package p2p

import (
	"fmt"
	"strings"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
)

func (b Bootstrapper) TestBootstrap(ctx map[string]interface{}) error {
	cfg := ctx[bootstrap.BootstrappedConfig].(*config.Configuration)
	m, ok := ctx[bootstrap.TestBootstrappedPathMatch]
	if !ok {
		return fmt.Errorf("match path missing")
	}

	path := m.(string)
	oldpub, oldPri := cfg.GetSigningKeyPair()
	newPub := strings.Replace(oldpub, "../../", path+"/", 1)
	newPri := strings.Replace(oldPri, "../../", path+"/", 1)
	cfg.Set("keys.signing.publicKey", newPub)
	cfg.Set("keys.signing.privateKey", newPri)
	return b.Bootstrap(ctx)
}

func (b Bootstrapper) TestTearDown() error {
	return nil
}

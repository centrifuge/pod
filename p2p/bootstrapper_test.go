// +build unit

package p2p

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/nft"
	"github.com/centrifuge/go-centrifuge/node"
	"github.com/centrifuge/go-centrifuge/testingutils/commons"
	"github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/centrifuge/go-centrifuge/testingutils/documents"
	"github.com/stretchr/testify/assert"
)

func TestBootstrapper_Bootstrap(t *testing.T) {
	b := Bootstrapper{}

	// no config
	m := make(map[string]interface{})
	err := b.Bootstrap(m)
	assert.Error(t, err)

	// config
	cs := new(configstore.MockService)
	m[bootstrap.BootstrappedConfig] = new(testingconfig.MockConfig)
	m[config.BootstrappedConfigStorage] = cs
	cs.On("GetConfig").Return(&configstore.NodeConfig{}, nil)
	ids := new(testingcommons.MockIdentityService)
	m[identity.BootstrappedDIDService] = ids
	m[documents.BootstrappedDocumentService] = documents.DefaultService(nil, nil, documents.NewServiceRegistry(), ids)
	m[nft.BootstrappedPayObService] = new(testingdocuments.MockRegistry)

	err = b.Bootstrap(m)
	assert.Nil(t, err)

	assert.NotNil(t, m[bootstrap.BootstrappedPeer])
	_, ok := m[bootstrap.BootstrappedPeer].(node.Server)
	assert.True(t, ok)

	assert.NotNil(t, m[bootstrap.BootstrappedPeer])
	_, ok = m[bootstrap.BootstrappedPeer].(documents.Client)
	assert.True(t, ok)
}

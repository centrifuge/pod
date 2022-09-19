//go:build unit

package v2

import (
	"testing"

	testingcommons "github.com/centrifuge/go-centrifuge/testingutils/common"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/documents/entity"
	"github.com/centrifuge/go-centrifuge/documents/entityrelationship"
	"github.com/centrifuge/go-centrifuge/errors"
	v2 "github.com/centrifuge/go-centrifuge/identity/v2"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/pending"
	"github.com/stretchr/testify/assert"
)

func TestNewService_Errors(t *testing.T) {
	pendingDocSrvMock := pending.NewServiceMock(t)
	dispatcherMock := jobs.NewDispatcherMock(t)
	cfgServiceMock := config.NewServiceMock(t)
	entityServiceMock := entity.NewServiceMock(t)
	identityServiceMock := v2.NewServiceMock(t)
	entityRelationshipServiceMock := entityrelationship.NewServiceMock(t)
	documentServiceMock := documents.NewServiceMock(t)

	cfgServiceMock.On("GetConfig").
		Return(nil, errors.New("error")).
		Once()

	_, err := NewService(
		pendingDocSrvMock,
		dispatcherMock,
		cfgServiceMock,
		entityServiceMock,
		identityServiceMock,
		entityRelationshipServiceMock,
		documentServiceMock,
	)
	assert.NotNil(t, err)

	configMock := config.NewConfigurationMock(t)

	cfgServiceMock.On("GetConfig").
		Return(configMock, nil).
		Once()

	configMock.On("GetP2PKeyPair").
		Return("invalidPath", "invalidPath").
		Once()

	_, err = NewService(
		pendingDocSrvMock,
		dispatcherMock,
		cfgServiceMock,
		entityServiceMock,
		identityServiceMock,
		entityRelationshipServiceMock,
		documentServiceMock,
	)
	assert.NotNil(t, err)

	cfgServiceMock.On("GetConfig").
		Return(configMock, nil)

	configMock.On("GetP2PKeyPair").
		Return(testingcommons.TestPublicSigningKeyPath, testingcommons.TestPrivateSigningKeyPath)

	cfgServiceMock.On("GetPodOperator").
		Return(nil, errors.New("error"))

	_, err = NewService(
		pendingDocSrvMock,
		dispatcherMock,
		cfgServiceMock,
		entityServiceMock,
		identityServiceMock,
		entityRelationshipServiceMock,
		documentServiceMock,
	)
	assert.NotNil(t, err)
}

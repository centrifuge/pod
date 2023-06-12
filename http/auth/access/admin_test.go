//go:build unit

package access

import (
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	configMocks "github.com/centrifuge/pod/config"
	"github.com/centrifuge/pod/errors"
	authToken "github.com/centrifuge/pod/http/auth/token"
	"github.com/centrifuge/pod/utils"
	"github.com/stretchr/testify/assert"
	"github.com/vedhavyas/go-subkey/v2"

	"testing"
)

func TestAdminAccessValidator_Validate(t *testing.T) {
	configSrvMock := configMocks.NewServiceMock(t)

	adminAccessValidator := NewAdminAccessValidator(configSrvMock)

	adminAccountID, err := types.NewAccountID(utils.RandomSlice(32))
	assert.NoError(t, err)

	adminSS58Address := subkey.SS58Encode(adminAccountID.ToBytes(), authToken.CentrifugeNetworkID)

	token := &authToken.JW3Token{
		Payload: &authToken.JW3TPayload{
			Address: adminSS58Address,
		},
	}

	podAdmin := configMocks.NewPodAdminMock(t)
	podAdmin.On("GetAccountID").
		Return(adminAccountID).
		Once()

	configSrvMock.On("GetPodAdmin").
		Return(podAdmin, nil).
		Once()

	res, err := adminAccessValidator.Validate(nil, token)
	assert.NoError(t, err)
	assert.True(t, adminAccountID.Equal(res))
}

func TestAdminAccessValidator_Validate_DelegateDecodeError(t *testing.T) {
	configSrvMock := configMocks.NewServiceMock(t)

	adminAccessValidator := NewAdminAccessValidator(configSrvMock)

	token := &authToken.JW3Token{
		Payload: &authToken.JW3TPayload{
			Address: "",
		},
	}

	res, err := adminAccessValidator.Validate(nil, token)
	assert.Nil(t, res)
	assert.ErrorIs(t, err, ErrSS58AddressDecode)
}

func TestAdminAccessValidator_Validate_PodAdminRetrievalError(t *testing.T) {
	configSrvMock := configMocks.NewServiceMock(t)

	adminAccessValidator := NewAdminAccessValidator(configSrvMock)

	adminAccountID, err := types.NewAccountID(utils.RandomSlice(32))
	assert.NoError(t, err)

	adminSS58Address := subkey.SS58Encode(adminAccountID.ToBytes(), authToken.CentrifugeNetworkID)

	token := &authToken.JW3Token{
		Payload: &authToken.JW3TPayload{
			Address: adminSS58Address,
		},
	}

	configSrvMock.On("GetPodAdmin").
		Return(nil, errors.New("error")).
		Once()

	res, err := adminAccessValidator.Validate(nil, token)
	assert.Nil(t, res)
	assert.ErrorIs(t, err, ErrPodAdminRetrieval)
}

func TestAdminAccessValidator_Validate_ValidationError(t *testing.T) {
	configSrvMock := configMocks.NewServiceMock(t)

	adminAccessValidator := NewAdminAccessValidator(configSrvMock)

	randomAccountID, err := types.NewAccountID(utils.RandomSlice(32))

	adminAccountID, err := types.NewAccountID(utils.RandomSlice(32))
	assert.NoError(t, err)

	randomAccountSS58Address := subkey.SS58Encode(randomAccountID.ToBytes(), authToken.CentrifugeNetworkID)

	token := &authToken.JW3Token{
		Payload: &authToken.JW3TPayload{
			Address: randomAccountSS58Address,
		},
	}

	podAdmin := configMocks.NewPodAdminMock(t)
	podAdmin.On("GetAccountID").
		Return(adminAccountID).
		Once()

	configSrvMock.On("GetPodAdmin").
		Return(podAdmin, nil).
		Once()

	res, err := adminAccessValidator.Validate(nil, token)
	assert.Nil(t, res)
	assert.ErrorIs(t, err, ErrNotAdminAccount)
}

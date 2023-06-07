//go:build unit

package access

import (
	"testing"

	"github.com/centrifuge/pod/errors"

	proxyType "github.com/centrifuge/chain-custom-types/pkg/proxy"
	authToken "github.com/centrifuge/pod/http/auth/token"

	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/pod/pallets/proxy"
	"github.com/centrifuge/pod/utils"
	"github.com/stretchr/testify/assert"
	"github.com/vedhavyas/go-subkey"
)

func TestProxyAccessValidator_Validate(t *testing.T) {
	proxyAPIMock := proxy.NewAPIMock(t)

	proxyAccessValidator := NewProxyAccessValidator(proxyAPIMock)

	pt := proxyType.Any

	delegateAccountID, err := types.NewAccountID(utils.RandomSlice(32))
	assert.NoError(t, err)

	delegatorAccountID, err := types.NewAccountID(utils.RandomSlice(32))
	assert.NoError(t, err)

	delegateAddress := subkey.SS58Encode(delegateAccountID.ToBytes(), authToken.CentrifugeNetworkID)

	delegatorAddress := subkey.SS58Encode(delegatorAccountID.ToBytes(), authToken.CentrifugeNetworkID)

	token := &authToken.JW3Token{
		Payload: &authToken.JW3TPayload{
			Address:    delegateAddress,
			OnBehalfOf: delegatorAddress,
			ProxyType:  proxyType.ProxyTypeName[pt],
		},
	}

	proxies := &types.ProxyStorageEntry{
		ProxyDefinitions: []types.ProxyDefinition{
			{
				Delegate:  *delegateAccountID,
				ProxyType: types.U8(pt),
			},
		},
	}

	proxyAPIMock.On("GetProxies", delegatorAccountID).
		Return(proxies, nil).
		Once()

	res, err := proxyAccessValidator.Validate(nil, token)
	assert.NoError(t, err)
	assert.True(t, res.Equal(delegatorAccountID))
}

func TestProxyAccessValidator_Validate_DelegateDecodeError(t *testing.T) {
	proxyAPIMock := proxy.NewAPIMock(t)

	proxyAccessValidator := NewProxyAccessValidator(proxyAPIMock)

	pt := proxyType.Any

	delegatorAccountID, err := types.NewAccountID(utils.RandomSlice(32))
	assert.NoError(t, err)

	delegatorAddress := subkey.SS58Encode(delegatorAccountID.ToBytes(), authToken.CentrifugeNetworkID)
	assert.NoError(t, err)

	token := &authToken.JW3Token{
		Payload: &authToken.JW3TPayload{
			Address:    "",
			OnBehalfOf: delegatorAddress,
			ProxyType:  proxyType.ProxyTypeName[pt],
		},
	}

	res, err := proxyAccessValidator.Validate(nil, token)
	assert.Nil(t, res)
	assert.ErrorIs(t, err, ErrSS58AddressDecode)
}

func TestProxyAccessValidator_Validate_DelegatorDecodeError(t *testing.T) {
	proxyAPIMock := proxy.NewAPIMock(t)

	proxyAccessValidator := NewProxyAccessValidator(proxyAPIMock)

	pt := proxyType.Any

	delegateAccountID, err := types.NewAccountID(utils.RandomSlice(32))
	assert.NoError(t, err)

	delegateAddress := subkey.SS58Encode(delegateAccountID.ToBytes(), authToken.CentrifugeNetworkID)
	assert.NoError(t, err)

	token := &authToken.JW3Token{
		Payload: &authToken.JW3TPayload{
			Address:    delegateAddress,
			OnBehalfOf: "",
			ProxyType:  proxyType.ProxyTypeName[pt],
		},
	}

	res, err := proxyAccessValidator.Validate(nil, token)
	assert.Nil(t, res)
	assert.ErrorIs(t, err, ErrSS58AddressDecode)
}

func TestProxyAccessValidator_Validate_SkipProxyCheck(t *testing.T) {
	proxyAPIMock := proxy.NewAPIMock(t)

	proxyAccessValidator := NewProxyAccessValidator(proxyAPIMock)

	pt := proxyType.Any

	delegateAccountID, err := types.NewAccountID(utils.RandomSlice(32))
	assert.NoError(t, err)

	delegateAddress := subkey.SS58Encode(delegateAccountID.ToBytes(), authToken.CentrifugeNetworkID)
	assert.NoError(t, err)

	token := &authToken.JW3Token{
		Payload: &authToken.JW3TPayload{
			Address:    delegateAddress,
			OnBehalfOf: delegateAddress,
			ProxyType:  proxyType.ProxyTypeName[pt],
		},
	}

	res, err := proxyAccessValidator.Validate(nil, token)
	assert.NoError(t, err)
	assert.True(t, res.Equal(delegateAccountID))
}

func TestProxyAccessValidator_Validate_ProxyRetrievalError(t *testing.T) {
	proxyAPIMock := proxy.NewAPIMock(t)

	proxyAccessValidator := NewProxyAccessValidator(proxyAPIMock)

	pt := proxyType.Any

	delegateAccountID, err := types.NewAccountID(utils.RandomSlice(32))
	assert.NoError(t, err)

	delegatorAccountID, err := types.NewAccountID(utils.RandomSlice(32))
	assert.NoError(t, err)

	delegateAddress := subkey.SS58Encode(delegateAccountID.ToBytes(), authToken.CentrifugeNetworkID)
	assert.NoError(t, err)

	delegatorAddress := subkey.SS58Encode(delegatorAccountID.ToBytes(), authToken.CentrifugeNetworkID)
	assert.NoError(t, err)

	token := &authToken.JW3Token{
		Payload: &authToken.JW3TPayload{
			Address:    delegateAddress,
			OnBehalfOf: delegatorAddress,
			ProxyType:  proxyType.ProxyTypeName[pt],
		},
	}

	proxyAPIMock.On("GetProxies", delegatorAccountID).
		Return(nil, errors.New("error")).
		Once()

	res, err := proxyAccessValidator.Validate(nil, token)
	assert.Nil(t, res)
	assert.ErrorIs(t, err, ErrAccountProxiesRetrieval)
}

func TestProxyAccessValidator_Validate_InvalidDelegate(t *testing.T) {
	proxyAPIMock := proxy.NewAPIMock(t)

	proxyAccessValidator := NewProxyAccessValidator(proxyAPIMock)

	pt := proxyType.Any

	delegateAccountID, err := types.NewAccountID(utils.RandomSlice(32))
	assert.NoError(t, err)

	delegatorAccountID, err := types.NewAccountID(utils.RandomSlice(32))
	assert.NoError(t, err)

	delegateAddress := subkey.SS58Encode(delegateAccountID.ToBytes(), authToken.CentrifugeNetworkID)
	assert.NoError(t, err)

	delegatorAddress := subkey.SS58Encode(delegatorAccountID.ToBytes(), authToken.CentrifugeNetworkID)
	assert.NoError(t, err)

	token := &authToken.JW3Token{
		Payload: &authToken.JW3TPayload{
			Address:    delegateAddress,
			OnBehalfOf: delegatorAddress,
			ProxyType:  proxyType.ProxyTypeName[pt],
		},
	}

	proxies := &types.ProxyStorageEntry{
		ProxyDefinitions: []types.ProxyDefinition{
			{
				Delegate: *delegateAccountID,
				// Different from the proxy in the token.
				ProxyType: types.U8(proxyType.PodOperation),
			},
		},
	}

	proxyAPIMock.On("GetProxies", delegatorAccountID).
		Return(proxies, nil).
		Once()

	res, err := proxyAccessValidator.Validate(nil, token)
	assert.Nil(t, res)
	assert.ErrorIs(t, err, ErrInvalidDelegate)
}

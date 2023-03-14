//go:build unit

package receiver

import (
	"testing"

	p2ppb "github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	keystoreType "github.com/centrifuge/chain-custom-types/pkg/keystore"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/pod/crypto/ed25519"
	"github.com/centrifuge/pod/errors"
	v2 "github.com/centrifuge/pod/identity/v2"
	p2pcommon "github.com/centrifuge/pod/p2p/common"
	testingcommons "github.com/centrifuge/pod/testingutils/common"
	"github.com/centrifuge/pod/version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_versionValidator(t *testing.T) {
	tests := []struct {
		Name          string
		Header        *p2ppb.Header
		ExpectedError bool
	}{
		{
			Name: "valid header",
			Header: &p2ppb.Header{
				NodeVersion: version.GetVersion().String(),
			},
			ExpectedError: false,
		},
		{
			Name:          "nil header",
			Header:        nil,
			ExpectedError: true,
		},
		{
			Name: "invalid version",
			Header: &p2ppb.Header{
				NodeVersion: "invalid-version",
			},
			ExpectedError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			v := versionValidator()

			err := v.Validate(test.Header, nil, nil)

			if test.ExpectedError {
				assert.NotNil(t, err)
				return
			}

			assert.NoError(t, err)
		})
	}
}

func Test_networkValidator(t *testing.T) {
	networkID := uint32(36)

	tests := []struct {
		Name          string
		Header        *p2ppb.Header
		ExpectedError bool
	}{
		{
			Name: "valid header",
			Header: &p2ppb.Header{
				NetworkIdentifier: networkID,
			},
			ExpectedError: false,
		},
		{
			Name:          "nil header",
			Header:        nil,
			ExpectedError: true,
		},
		{
			Name: "invalid version",
			Header: &p2ppb.Header{
				NetworkIdentifier: networkID + 1,
			},
			ExpectedError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			v := networkValidator(networkID)

			err := v.Validate(test.Header, nil, nil)

			if test.ExpectedError {
				assert.NotNil(t, err)
				return
			}

			assert.NoError(t, err)
		})
	}
}

func Test_peerValidator(t *testing.T) {
	identityServiceMock := v2.NewServiceMock(t)

	v := peerValidator(identityServiceMock)

	err := v.Validate(nil, nil, nil)
	assert.NotNil(t, err)

	err = v.Validate(&p2ppb.Header{}, nil, nil)
	assert.NotNil(t, err)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	err = v.Validate(&p2ppb.Header{}, accountID, nil)
	assert.NotNil(t, err)

	publicKey, _, err := ed25519.GenerateSigningKeyPair()
	assert.NoError(t, err)

	p2pKey := types.NewHash(publicKey)

	peerID, err := p2pcommon.ParsePeerID(p2pKey)
	assert.NoError(t, err)

	identityServiceMock.On(
		"ValidateKey",
		accountID,
		p2pKey[:],
		keystoreType.KeyPurposeP2PDiscovery,
		mock.Anything,
	).Return(nil).Once()

	err = v.Validate(&p2ppb.Header{}, accountID, &peerID)
	assert.NoError(t, err)

	validateErr := errors.New("error")

	identityServiceMock.On(
		"ValidateKey",
		accountID,
		p2pKey[:],
		keystoreType.KeyPurposeP2PDiscovery,
		mock.Anything,
	).Return(validateErr).Once()

	err = v.Validate(&p2ppb.Header{}, accountID, &peerID)
	assert.ErrorIs(t, err, validateErr)
}

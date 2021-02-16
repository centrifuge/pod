// +build unit

package receiver

import (
	"testing"
	"time"

	p2ppb "github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	testingcommons "github.com/centrifuge/go-centrifuge/testingutils/commons"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/go-centrifuge/version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var id1 = []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}

func TestValidate_versionValidator(t *testing.T) {
	vv := versionValidator()

	// Nil header
	err := vv.Validate(nil, nil, nil)
	assert.NotNil(t, err)

	// Empty header
	header := &p2ppb.Header{}
	err = vv.Validate(header, nil, nil)
	assert.NotNil(t, err)

	// Incompatible Major
	header.NodeVersion = "2.1.1"
	err = vv.Validate(header, nil, nil)
	assert.NotNil(t, err)

	// Compatible Minor
	header.NodeVersion = "1.1.1"
	err = vv.Validate(header, nil, nil)
	assert.Nil(t, err)

	// Same version
	header.NodeVersion = version.GetVersion().String()
	err = vv.Validate(header, nil, nil)
	assert.Nil(t, err)
}

func TestValidate_networkValidator(t *testing.T) {
	nv := networkValidator(cfg.GetNetworkID())

	// Nil header
	err := nv.Validate(nil, nil, nil)
	assert.NotNil(t, err)

	header := &p2ppb.Header{}
	err = nv.Validate(header, nil, nil)
	assert.NotNil(t, err)

	// Incompatible network
	header.NetworkIdentifier = 12
	err = nv.Validate(header, nil, nil)
	assert.NotNil(t, err)

	// Compatible network
	header.NetworkIdentifier = cfg.GetNetworkID()
	err = nv.Validate(header, nil, nil)
	assert.Nil(t, err)
}

func TestValidate_peerValidator(t *testing.T) {
	cID, err := identity.NewDIDFromBytes(id1)
	assert.NoError(t, err)

	idService := &testingcommons.MockIdentityService{}
	sv := peerValidator(idService)

	// Nil headers
	err = sv.Validate(nil, nil, nil)
	assert.Error(t, err)

	tm, err := utils.ToTimestamp(time.Now())
	assert.NoError(t, err)

	// Nil centID
	header := &p2ppb.Header{
		Timestamp: tm,
	}
	err = sv.Validate(header, nil, nil)
	assert.Error(t, err)

	// Nil peerID
	err = sv.Validate(header, &cID, nil)
	assert.Error(t, err)

	// Identity validation failure
	idService.On("ValidateKey", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("key not linked to identity")).Once()
	err = sv.Validate(header, &cID, &defaultPID)
	assert.Error(t, err)

	// Success
	idService.On("ValidateKey", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	err = sv.Validate(header, &cID, &defaultPID)
	assert.NoError(t, err)
}

func TestValidate_handshakeValidator(t *testing.T) {
	cID, err := identity.NewDIDFromBytes(id1)
	assert.NoError(t, err)

	idService := &testingcommons.MockIdentityService{}
	hv := HandshakeValidator(cfg.GetNetworkID(), idService)
	tm, err := utils.ToTimestamp(time.Now())
	assert.NoError(t, err)

	// Incompatible version network and wrong signature
	header := &p2ppb.Header{
		NodeVersion:       "version",
		NetworkIdentifier: 52,
		Timestamp:         tm,
	}
	err = hv.Validate(header, nil, nil)
	assert.NotNil(t, err)

	// Incompatible version, correct network
	header.NetworkIdentifier = cfg.GetNetworkID()
	err = hv.Validate(header, nil, nil)
	assert.NotNil(t, err)

	// Compatible version, incorrect network
	header.NetworkIdentifier = 52
	header.NodeVersion = version.GetVersion().String()
	err = hv.Validate(header, nil, nil)
	assert.NotNil(t, err)

	// Compatible version, network and wrong eth key
	idService.On("ValidateKey", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("key not linked to identity")).Once()
	header.NetworkIdentifier = cfg.GetNetworkID()
	err = hv.Validate(header, &cID, &defaultPID)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "key not linked to identity")

	// Compatible version, network and signature
	idService.On("ValidateKey", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	err = hv.Validate(header, &cID, &defaultPID)
	assert.NoError(t, err)
}

func TestDocumentAccessValidator_Collaborator(t *testing.T) {
	//account1, err := identity.CentIDFromString("0x010203040506")
	//assert.NoError(t, err)
	//cd, err := coredocument.NewWithCollaborators([]string{account1.String()})
	//assert.NotNil(t, cd.Collaborators)
	//
	//docId := cd.DocumentIdentifier
	//req := &p2ppb.GetDocumentRequest{DocumentIdentifier: docId, AccessType: p2ppb.AccessType_ACCESS_TYPE_REQUESTER_VERIFICATION}
	//err = DocumentAccessValidator(cd, req, account1)
	//assert.NoError(t, err)
	//
	//account2, err := identity.CentIDFromString("0x012345678910")
	//err = DocumentAccessValidator(cd, req, account2)
	//assert.Error(t, err, "requester does not have access")
}

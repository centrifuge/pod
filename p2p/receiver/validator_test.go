// +build unit

package receiver

import (
	"github.com/centrifuge/go-centrifuge/coredocument"
	"testing"

	"github.com/centrifuge/go-centrifuge/identity"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/version"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	testingcommons "github.com/centrifuge/go-centrifuge/testingutils/commons"
)

var (
	key1Pub = []byte{230, 49, 10, 12, 200, 149, 43, 184, 145, 87, 163, 252, 114, 31, 91, 163, 24, 237, 36, 51, 165, 8, 34, 104, 97, 49, 114, 85, 255, 15, 195, 199}
	key1    = []byte{102, 109, 71, 239, 130, 229, 128, 189, 37, 96, 223, 5, 189, 91, 210, 47, 89, 4, 165, 6, 188, 53, 49, 250, 109, 151, 234, 139, 57, 205, 231, 253, 230, 49, 10, 12, 200, 149, 43, 184, 145, 87, 163, 252, 114, 31, 91, 163, 24, 237, 36, 51, 165, 8, 34, 104, 97, 49, 114, 85, 255, 15, 195, 199}
	id1     = []byte{1, 1, 1, 1, 1, 1}
)

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
	header.NodeVersion = "1.1.1"
	err = vv.Validate(header, nil, nil)
	assert.NotNil(t, err)

	// Compatible Minor
	header.NodeVersion = "0.1.1"
	err = vv.Validate(header, nil, nil)
	assert.Nil(t, err)

	//Same version
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
	cID, _ := identity.ToCentID(id1)

	idService := &testingcommons.MockIDService{}
	sv := peerValidator(idService)

	// Nil headers
	err := sv.Validate(nil, nil, nil)
	assert.Error(t, err)

	// Nil centID
	header := &p2ppb.Header{}
	err = sv.Validate(header, nil, nil)
	assert.Error(t, err)

	// Nil peerID
	err = sv.Validate(header, &cID, nil)
	assert.Error(t, err)

	// Identity validation failure
	idService.On("ValidateKey", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("key not linked to identity")).Once()
	err = sv.Validate(header, &cID, &defaultPID)
	assert.Error(t, err)

	// Success
	idService.On("ValidateKey", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	err = sv.Validate(header, &cID, &defaultPID)
	assert.NoError(t, err)
}

func TestValidate_handshakeValidator(t *testing.T) {
	cID, _ := identity.ToCentID(id1)
	idService := &testingcommons.MockIDService{}
	hv := HandshakeValidator(cfg.GetNetworkID(), idService)

	// Incompatible version network and wrong signature
	header := &p2ppb.Header{
		NodeVersion:       "version",
		NetworkIdentifier: 52,
	}
	err := hv.Validate(header, nil, nil)
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
	idService.On("ValidateKey", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("key not linked to identity")).Once()
	header.NetworkIdentifier = cfg.GetNetworkID()
	err = hv.Validate(header, &cID, &defaultPID)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "key not linked to identity")

	// Compatible version, network and signature
	idService.On("ValidateKey", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	err = hv.Validate(header, &cID, &defaultPID)
	assert.NoError(t, err)
}

func TestDocumentAccessValidator_Collaborator(t *testing.T) {
	account1, err := identity.CentIDFromString("0x010203040506")
	assert.NoError(t, err)
	cd, err := coredocument.NewWithCollaborators([]string{account1.String()})
	assert.NotNil(t, cd.Collaborators)

	docId := cd.DocumentIdentifier
	req := &p2ppb.GetDocumentRequest{DocumentIdentifier: docId, AccessType: p2ppb.AccessType_ACCESS_TYPE_REQUESTER_VERIFICATION}
	err = DocumentAccessValidator(cd, req, account1)
	assert.NoError(t, err)

	account2, err := identity.CentIDFromString("0x012345678910")
	err = DocumentAccessValidator(cd, req, account2)
	assert.Error(t, err, "requester does not have access")
}

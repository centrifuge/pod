//go:build testworld
// +build testworld

package testworld

import (
	"strings"
	"testing"

	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/p2p"
	"github.com/centrifuge/go-centrifuge/p2p/messenger"
	testingconfig "github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/stretchr/testify/assert"
)

// send a valid signature request message
func TestIncorrectProto_ValidMessage(t *testing.T) {
	// Hosts
	bob := doctorFord.getHostTestSuite(t, "Bob")
	eve := doctorFord.getHostTestSuite(t, "Eve")

	ctxh := testingconfig.CreateAccountContext(t, eve.host.config)

	// Get PublicKey and PrivateKey
	publicKey, privateKey := GetSigningKeyPair(t, eve.host.idService, eve.id, ctxh)

	collaborators := [][]byte{bob.id[:]}
	dm := createCDWithEmbeddedDocument(t, collaborators, eve.id, publicKey, privateKey)

	p := p2p.AccessPeer(eve.host.p2pClient)

	// send a signature request message with incorect protocol version
	signatures, signatureErrors, err := p.GetSignaturesForDocument(ctxh, dm)
	assert.NoError(t, err)
	assert.Nil(t, signatureErrors)
	assert.Equal(t, 1, len(signatures))
}

// send a signature request message with an incorrect node version
func TestIncorrectProto_DifferentVersion(t *testing.T) {
	errors.MaskErrs = false
	// Hosts
	bob := doctorFord.getHostTestSuite(t, "Bob")
	eve := doctorFord.getHostTestSuite(t, "Eve")

	ctxh := testingconfig.CreateAccountContext(t, eve.host.config)

	// Get PublicKey and PrivateKey
	publicKey, privateKey := GetSigningKeyPair(t, eve.host.idService, eve.id, ctxh)

	collaborators := [][]byte{bob.id[:]}
	dm := createCDWithEmbeddedDocument(t, collaborators, eve.id, publicKey, privateKey)

	p := p2p.AccessPeer(eve.host.p2pClient)

	// send a signature request message with incorect protocol version
	signatures, signatureErrors, err := p.GetSignaturesForDocumentIncorrectMessage(ctxh, dm, "incorrectNodeVersion")
	assert.NoError(t, err)
	assert.Error(t, signatureErrors[0], "Message failed error")
	assert.Equal(t, true, strings.Contains(signatureErrors[0].Error(), "Incompatible version"))
	assert.Equal(t, 0, len(signatures))
}

// send a signature request message with an invalid body
func TestIncorrectProto_InvalidBody(t *testing.T) {
	errors.MaskErrs = false
	// Hosts
	bob := doctorFord.getHostTestSuite(t, "Bob")
	eve := doctorFord.getHostTestSuite(t, "Eve")

	ctxh := testingconfig.CreateAccountContext(t, eve.host.config)

	// Get PublicKey and PrivateKey
	publicKey, privateKey := GetSigningKeyPair(t, eve.host.idService, eve.id, ctxh)

	collaborators := [][]byte{bob.id[:]}
	dm := createCDWithEmbeddedDocument(t, collaborators, eve.id, publicKey, privateKey)

	p := p2p.AccessPeer(eve.host.p2pClient)

	// send a signature request message with incorect protocol version
	signatures, signatureErrors, err := p.GetSignaturesForDocumentIncorrectMessage(ctxh, dm, "invalidBody")
	assert.NoError(t, err)
	assert.Error(t, signatureErrors[0], "Message failed error")
	// assert.Equal(t, true, strings.Contains(signatureErrors[0].Error(), "unknown wire type") || strings.Contains(signatureErrors[0].Error(), "illegal tag"))
	assert.Equal(t, 0, len(signatures))
}

// send a signature request message with an invalid header
func TestIncorrectProto_InvalidHeader(t *testing.T) {
	errors.MaskErrs = false
	// Hosts
	bob := doctorFord.getHostTestSuite(t, "Bob")
	eve := doctorFord.getHostTestSuite(t, "Eve")

	ctxh := testingconfig.CreateAccountContext(t, eve.host.config)

	// Get PublicKey and PrivateKey
	publicKey, privateKey := GetSigningKeyPair(t, eve.host.idService, eve.id, ctxh)

	collaborators := [][]byte{bob.id[:]}
	dm := createCDWithEmbeddedDocument(t, collaborators, eve.id, publicKey, privateKey)

	p := p2p.AccessPeer(eve.host.p2pClient)

	// send a signature request message with incorect protocol version
	signatures, signatureErrors, err := p.GetSignaturesForDocumentIncorrectMessage(ctxh, dm, "invalidHeader")
	assert.NoError(t, err)
	assert.Error(t, signatureErrors[0], "Message failed error")
	assert.Equal(t, true, strings.Contains(signatureErrors[0].Error(), "invalid DID length"))
	assert.Equal(t, 0, len(signatures))
}

// send a signature request message with a message which is larger than the max allowed size
func TestIncorrectProto_AboveMaxSize(t *testing.T) {
	// Hosts
	bob := doctorFord.getHostTestSuite(t, "Bob")
	eve := doctorFord.getHostTestSuite(t, "Eve")

	ctxh := testingconfig.CreateAccountContext(t, eve.host.config)

	// Get PublicKey and PrivateKey
	publicKey, privateKey := GetSigningKeyPair(t, eve.host.idService, eve.id, ctxh)

	collaborators := [][]byte{bob.id[:]}
	dm := createCDWithEmbeddedDocument(t, collaborators, eve.id, publicKey, privateKey)
	p := p2p.AccessPeer(eve.host.p2pClient)
	_, err := p.SendOverSizedMessage(ctxh, dm, messenger.MessageSizeMax+1)
	assert.Error(t, err)
	assert.Equal(t, "stream reset", err.Error())
}

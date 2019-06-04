// +build testworld

package testworld

import (
	"github.com/centrifuge/go-centrifuge/config"
	testingconfig "github.com/centrifuge/go-centrifuge/testingutils/config"

	"github.com/centrifuge/go-centrifuge/p2p"
	"github.com/stretchr/testify/assert"
	"testing"
)

//send a valid signature request message
func TestIncorrectProto_ValidMessage(t *testing.T){
	t.Parallel()
	// Hosts
	bob := doctorFord.getHostTestSuite(t, "Bob")
	oskar := doctorFord.getHostTestSuite(t, "Oskar")

	ctxh := testingconfig.CreateAccountContext(t, oskar.host.config)

	// Get PublicKey and PrivateKey
	publicKey, privateKey := GetSigningKeyPair(t, oskar.host.idService, oskar.id, ctxh)

	collaborators := [][]byte{bob.id[:]}
	dm := createCDWithEmbeddedPO(t, collaborators, oskar.id, publicKey, privateKey, oskar.host.config.GetContractAddress(config.AnchorRepo))

	p := p2p.AccessPeer(oskar.host.p2pClient)

	//send a signature request message with incorect protocol version
	signatures, signatureErrors, err := p.GetSignaturesForDocument(ctxh, dm)
	assert.NoError(t, err)
	assert.Nil(t, signatureErrors)
	assert.Equal(t, 1, len(signatures))

}

//send a signature request message with an incorrect node version
func TestIncorrectProto_DifferentVersion(t *testing.T){
	t.Parallel()
	// Hosts
	bob := doctorFord.getHostTestSuite(t, "Bob")
	oskar := doctorFord.getHostTestSuite(t, "Oskar")

	ctxh := testingconfig.CreateAccountContext(t, oskar.host.config)

	// Get PublicKey and PrivateKey
	publicKey, privateKey := GetSigningKeyPair(t, oskar.host.idService, oskar.id, ctxh)

	collaborators := [][]byte{bob.id[:]}
	dm := createCDWithEmbeddedPO(t, collaborators, oskar.id, publicKey, privateKey, oskar.host.config.GetContractAddress(config.AnchorRepo))

	p := p2p.AccessPeer(oskar.host.p2pClient)

	//send a signature request message with incorect protocol version
	signatures, signatureErrors, err := p.GetSignaturesForDocumentIncorrectMessage(ctxh, dm, "incorrectNodeVersion")
	assert.NoError(t, err)
	assert.Error(t, signatureErrors[0], "Message failed error")
	assert.Equal(t, 0, len(signatures))

}

//send a signature request message with an invalid body
func TestIncorrectProto_InvalidBody(t *testing.T){
	t.Parallel()
	// Hosts
	bob := doctorFord.getHostTestSuite(t, "Bob")
	oskar := doctorFord.getHostTestSuite(t, "Oskar")

	ctxh := testingconfig.CreateAccountContext(t, oskar.host.config)

	// Get PublicKey and PrivateKey
	publicKey, privateKey := GetSigningKeyPair(t, oskar.host.idService, oskar.id, ctxh)

	collaborators := [][]byte{bob.id[:]}
	dm := createCDWithEmbeddedPO(t, collaborators, oskar.id, publicKey, privateKey, oskar.host.config.GetContractAddress(config.AnchorRepo))

	p := p2p.AccessPeer(oskar.host.p2pClient)

	//send a signature request message with incorect protocol version
	signatures, signatureErrors, err := p.GetSignaturesForDocumentIncorrectMessage(ctxh, dm, "invalidBody")
	assert.NoError(t, err)
	assert.Error(t, signatureErrors[0], "Message failed error")
	assert.Equal(t, 0, len(signatures))

}

//send a signature request message with an invalid header
func TestIncorrectProto_InvalidHeader(t *testing.T){
	t.Parallel()
	// Hosts
	bob := doctorFord.getHostTestSuite(t, "Bob")
	oskar := doctorFord.getHostTestSuite(t, "Oskar")

	ctxh := testingconfig.CreateAccountContext(t, oskar.host.config)

	// Get PublicKey and PrivateKey
	publicKey, privateKey := GetSigningKeyPair(t, oskar.host.idService, oskar.id, ctxh)

	collaborators := [][]byte{bob.id[:]}
	dm := createCDWithEmbeddedPO(t, collaborators, oskar.id, publicKey, privateKey, oskar.host.config.GetContractAddress(config.AnchorRepo))

	p := p2p.AccessPeer(oskar.host.p2pClient)

	//send a signature request message with incorect protocol version
	signatures, signatureErrors, err := p.GetSignaturesForDocumentIncorrectMessage(ctxh, dm, "invalidHeader")
	assert.NoError(t, err)
	assert.Error(t, signatureErrors[0], "Message failed error")
	assert.Equal(t, 0, len(signatures))

}

/*
func TestHost_dodgySignature(t *testing.T) {
	// Hosts
	bob := doctorFord.getHostTestSuite(t, "Bob")
	oskar := doctorFord.getHostTestSuite(t, "Oskar")

	ctxh := testingconfig.CreateAccountContext(t, oskar.host.config)

	// Get PublicKey and PrivateKey
	publicKey, privateKey := GetSigningKeyPair(t, oskar.host.idService, oskar.id, ctxh)


	//p, ok := oskar.host.p2pClient.(p2p.peer); ok{
	//	mes := p.getMessenger()
	//}

	//collaborators := [][]byte{bob.id[:]}
	//dm := createCDWithEmbeddedPO(t, collaborators, oskar.id, publicKey, privateKey, oskar.host.config.GetContractAddress(config.AnchorRepo))


}

*/
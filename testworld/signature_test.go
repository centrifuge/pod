// +build testworld

package testworld

import (
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/centrifuge/go-centrifuge/testingutils/documents"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

func TestHost_FakedSignature(t *testing.T) {
	t.Parallel()

	// Hosts
	alice := doctorFord.getHostTestSuite(t, "Alice")
	bob := doctorFord.getHostTestSuite(t, "Bob")
	eve := doctorFord.getHostTestSuite(t, "Eve")

	ectxh := testingconfig.CreateAccountContext(t, eve.host.config)

	collaborators := [][]byte{bob.id[:]}
	coredoc := prepareCoreDocument(t, collaborators, alice)

	err := eve.host.p2pClient.GetSignaturesForDocument(ectxh, coredoc)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(coredoc.Document.Signatures))
}

func TestHost_RevokedSigningKey(t *testing.T) {
	t.Parallel()
	bob := doctorFord.getHostTestSuite(t, "Bob")
	eve := doctorFord.getHostTestSuite(t, "Eve")

	ectxh := testingconfig.CreateAccountContext(t, eve.host.config)

	keys, err := eve.host.idService.GetKeysByPurpose(eve.id, big.NewInt(identity.KeyPurposeSigning))
	assert.NoError(t, err)

	// Revoke Key
	eve.host.idService.RevokeKey(ectxh, keys[0])
	response, err := eve.host.idService.GetKey(eve.id, keys[0])
	assert.NotEqual(t, utils.ByteSliceToBigInt([]byte{0}), response.RevokedAt, "key should be revoked")

	collaborators := [][]byte{bob.id[:]}
	coredoc := prepareCoreDocument(t, collaborators, eve)

	err = eve.host.p2pClient.GetSignaturesForDocument(ectxh, coredoc)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(coredoc.Document.Signatures))

	// TODO: Validate signatures before and after
}

func prepareCoreDocument(t *testing.T, collaborators [][]byte, hts hostTestSuite) *documents.CoreDocumentModel {
	dm := testingdocuments.GenerateCoreDocumentModelWithCollaborators(collaborators)
	m, err := hts.host.docSrv.DeriveFromCoreDocumentModel(dm)
	assert.Nil(t, err)

	droot, err := m.CalculateDataRoot()
	assert.Nil(t, err)

	dm, err = m.PackCoreDocument()
	assert.NoError(t, err)

	tree, err := dm.GetDocumentSigningTree(droot)
	assert.NoError(t, err)

	dm.Document.SigningRoot = tree.RootHash()

	idConfig, err := identity.GetIdentityConfig(hts.host.config)
	assert.Nil(t, err)

	sig := identity.Sign(idConfig, identity.KeyPurposeSigning, dm.Document.SigningRoot)

	dm.Document.Signatures = append(dm.Document.Signatures, sig)

	tree, err = dm.GetDocumentRootTree()
	assert.NoError(t, err)

	dm.Document.DocumentRoot = tree.RootHash()
	return dm
}

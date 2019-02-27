// +build testworld

package testworld

import (
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/centrifuge/go-centrifuge/testingutils/documents"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHost_Signature(t *testing.T) {
	t.Parallel()
	eve := doctorFord.getHostTestSuite(t, "Eve")
	bob := doctorFord.getHostTestSuite(t, "Bob")

	edp := eve.host.anchorProcessor

	ectxh := testingconfig.CreateAccountContext(t, eve.host.config)

	collaborators := [][]byte{bob.id[:]}
	coredoc := prepareCoreDocument(t, collaborators, eve)

	err := edp.Send(ectxh, coredoc, bob.id)
	assert.Nil(t, err)
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

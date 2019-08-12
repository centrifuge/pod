// +build integration unit testworld

package invoice

import (
	"context"
	"encoding/json"
	"testing"

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/identity"
	testingidentity "github.com/centrifuge/go-centrifuge/testingutils/identity"
	"github.com/stretchr/testify/assert"
)

func (b *Bootstrapper) TestBootstrap(context map[string]interface{}) error {
	return b.Bootstrap(context)
}

func (*Bootstrapper) TestTearDown() error {
	return nil
}

func invoiceData(t *testing.T) []byte {
	did, err := identity.NewDIDFromString("0xEA939D5C0494b072c51565b191eE59B5D34fbf79")
	assert.NoError(t, err)
	dec, err := documents.NewDecimal("42")
	assert.NoError(t, err)
	data := Data{
		Recipient: &did,
		Currency:  "EUR",
		LineItems: []*LineItem{
			{
				ItemNumber:  "12312431",
				Description: "just",
				Quantity:    dec,
			},
		},
	}

	d, err := json.Marshal(data)
	assert.NoError(t, err)
	return d
}

func CreateInvoicePayload(t *testing.T, collaborators []identity.DID) documents.CreatePayload {
	if collaborators == nil {
		collaborators = []identity.DID{testingidentity.GenerateRandomDID()}
	}
	return documents.CreatePayload{
		Scheme: Scheme,
		Collaborators: documents.CollaboratorsAccess{
			ReadWriteCollaborators: collaborators,
		},
		Data: invoiceData(t),
	}
}

func InitInvoice(t *testing.T, did identity.DID, payload documents.CreatePayload) *Invoice {
	inv := new(Invoice)
	payload.Collaborators.ReadWriteCollaborators = append(payload.Collaborators.ReadWriteCollaborators, did)
	assert.NoError(t, inv.DeriveFromCreatePayload(payload))
	return inv
}

func CreateInvoiceWithEmbedCDWithPayload(t *testing.T, ctx context.Context, did identity.DID, payload documents.CreatePayload) (*Invoice, coredocumentpb.CoreDocument) {
	inv := new(Invoice)
	payload.Collaborators.ReadWriteCollaborators = append(payload.Collaborators.ReadWriteCollaborators, did)
	err := inv.DeriveFromCreatePayload(payload)
	assert.NoError(t, err)
	inv.GetTestCoreDocWithReset()
	_, err = inv.CalculateDataRoot()
	assert.NoError(t, err)
	sr, err := inv.CalculateSigningRoot()
	assert.NoError(t, err)
	// if acc errors out, just skip it
	if ctx == nil {
		ctx = context.Background()
	}
	acc, err := contextutil.Account(ctx)
	if err == nil {
		sig, err := acc.SignMsg(sr)
		assert.NoError(t, err)
		inv.AppendSignatures(sig)
	}
	_, err = inv.CalculateDocumentRoot()
	assert.NoError(t, err)
	cd, err := inv.PackCoreDocument()
	assert.NoError(t, err)

	return inv, cd
}

func CreateInvoiceWithEmbedCD(t *testing.T, ctx context.Context, did identity.DID, collaborators []identity.DID) (*Invoice, coredocumentpb.CoreDocument) {
	payload := CreateInvoicePayload(t, collaborators)
	return CreateInvoiceWithEmbedCDWithPayload(t, ctx, did, payload)
}

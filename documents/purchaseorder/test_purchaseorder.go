// +build integration unit

package purchaseorder

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

func poData(t *testing.T) []byte {
	did, err := identity.NewDIDFromString("0xEA939D5C0494b072c51565b191eE59B5D34fbf79")
	assert.NoError(t, err)
	dec, err := documents.NewDecimal("42")
	assert.NoError(t, err)
	data := Data{
		Recipient:   &did,
		Currency:    "EUR",
		TotalAmount: dec,
		LineItems: []*LineItem{
			{
				Status:      "pending",
				AmountTotal: dec,
				Activities: []*LineItemActivity{
					{
						Status:     "pending",
						Amount:     dec,
						ItemNumber: "12345",
					},
				},
				TaxItems: []*TaxItem{
					{
						ItemNumber: "12345",
						TaxAmount:  dec,
					},
				},
			},
		},
	}

	d, err := json.Marshal(data)
	assert.NoError(t, err)
	return d
}

func CreatePOPayload(t *testing.T, collaborators []identity.DID) documents.CreatePayload {
	if collaborators == nil {
		collaborators = []identity.DID{testingidentity.GenerateRandomDID()}
	}
	return documents.CreatePayload{
		Scheme: Scheme,
		Collaborators: documents.CollaboratorsAccess{
			ReadWriteCollaborators: collaborators,
		},
		Data: poData(t),
	}
}

func InitPurchaseOrder(t *testing.T, did identity.DID, payload documents.CreatePayload) *PurchaseOrder {
	po := new(PurchaseOrder)
	assert.NoError(t, po.unpackFromCreatePayload(did, payload))
	return po
}

func CreatePOWithEmbedCDWithPayload(t *testing.T, ctx context.Context, did identity.DID, payload documents.CreatePayload) (*PurchaseOrder, coredocumentpb.CoreDocument) {
	po := new(PurchaseOrder)
	err := po.unpackFromCreatePayload(did, payload)
	assert.NoError(t, err)
	po.GetTestCoreDocWithReset()
	_, err = po.CalculateDataRoot()
	assert.NoError(t, err)
	sr, err := po.CalculateSigningRoot()
	assert.NoError(t, err)
	// if acc errors out, just skip it
	if ctx == nil {
		ctx = context.Background()
	}
	acc, err := contextutil.Account(ctx)
	if err == nil {
		sig, err := acc.SignMsg(sr)
		assert.NoError(t, err)
		po.AppendSignatures(sig)
	}
	_, err = po.CalculateDocumentRoot()
	assert.NoError(t, err)
	cd, err := po.PackCoreDocument()
	assert.NoError(t, err)

	return po, cd
}

func CreatePOWithEmbedCD(t *testing.T, ctx context.Context, did identity.DID, collaborators []identity.DID) (*PurchaseOrder, coredocumentpb.CoreDocument) {
	payload := CreatePOPayload(t, collaborators)
	return CreatePOWithEmbedCDWithPayload(t, ctx, did, payload)
}

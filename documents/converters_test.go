// +build unit

package documents

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/document"
	"github.com/centrifuge/go-centrifuge/testingutils/identity"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/stretchr/testify/assert"
)

func TestBinaryAttachments(t *testing.T) {
	atts := []*BinaryAttachment{
		{
			Name:     "some name",
			FileType: "pdf",
			Size:     1024,
			Data:     utils.RandomSlice(32),
			Checksum: utils.RandomSlice(32),
		},

		{
			Name:     "some name 1",
			FileType: "jpeg",
			Size:     4096,
			Data:     utils.RandomSlice(32),
			Checksum: utils.RandomSlice(32),
		},
	}

	catts := ToClientAttachments(atts)
	patts := ToP2PAttachments(atts)

	fcatts, err := FromClientAttachments(catts)
	assert.NoError(t, err)
	assert.Equal(t, atts, fcatts)

	fpatts := FromP2PAttachments(patts)
	assert.Equal(t, atts, fpatts)

	catts[0].Checksum = "some checksum"
	_, err = FromClientAttachments(catts)
	assert.Error(t, err)

	catts[0].Data = "some data"
	_, err = FromClientAttachments(catts)
	assert.Error(t, err)
}

func TestPaymentDetails(t *testing.T) {
	did := testingidentity.GenerateRandomDID()
	dec := new(Decimal)
	err := dec.SetString("0.99")
	assert.NoError(t, err)
	details := []*PaymentDetails{
		{
			ID:     "some id",
			Payee:  &did,
			Amount: dec,
		},
	}

	cdetails := ToClientPaymentDetails(details)
	pdetails, err := ToP2PPaymentDetails(details)
	assert.NoError(t, err)

	fcdetails, err := FromClientPaymentDetails(cdetails)
	assert.NoError(t, err)
	fpdetails, err := FromP2PPaymentDetails(pdetails)
	assert.NoError(t, err)

	assert.Equal(t, details, fcdetails)
	assert.Equal(t, details, fpdetails)

	cdetails[0].Payee = "some did"
	_, err = FromClientPaymentDetails(cdetails)
	assert.Error(t, err)

	cdetails[0].Amount = "0.1.1"
	_, err = FromClientPaymentDetails(cdetails)
	assert.Error(t, err)

	pdetails[0].Amount = utils.RandomSlice(40)
	_, err = FromP2PPaymentDetails(pdetails)
	assert.Error(t, err)
}

func TestCollaboratorAccess(t *testing.T) {
	id1 := testingidentity.GenerateRandomDID()
	id2 := testingidentity.GenerateRandomDID()
	id3 := testingidentity.GenerateRandomDID()
	rcs := &documentpb.ReadAccess{
		Collaborators: []string{id1.String(), id2.String()},
	}

	wcs := &documentpb.WriteAccess{
		Collaborators: []string{id2.String(), id3.String()},
	}

	ca, err := FromClientCollaboratorAccess(rcs, wcs)
	assert.NoError(t, err)
	assert.Len(t, ca.ReadCollaborators, 1)
	assert.Len(t, ca.ReadWriteCollaborators, 2)

	grcs, gwcs := ToClientCollaboratorAccess(ca)
	assert.Len(t, grcs.Collaborators, 1)
	assert.Equal(t, grcs.Collaborators[0], id1.String())
	assert.Len(t, gwcs.Collaborators, 2)
	assert.Contains(t, gwcs.Collaborators, id2.String(), id3.String())

	wcs.Collaborators[0] = "wrg id"
	_, err = FromClientCollaboratorAccess(rcs, wcs)
	assert.Error(t, err)

	rcs.Collaborators[0] = "wrg id"
	_, err = FromClientCollaboratorAccess(rcs, wcs)
	assert.Error(t, err)
}

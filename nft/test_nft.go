// +build integration unit testworld

package nft

import (
	"fmt"
	"testing"
	"time"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
)

func (b *Bootstrapper) TestBootstrap(context map[string]interface{}) error {
	return b.Bootstrap(context)
}

func (b *Bootstrapper) TestTearDown() error {
	return nil
}

func GetAttributes(t *testing.T, did identity.DID) (map[documents.AttrKey]documents.Attribute, []string) {
	attrs := map[documents.AttrKey]documents.Attribute{}
	attr1, err := documents.NewStringAttribute("Originator", documents.AttrBytes, did.ToAddress().Hex())
	assert.NoError(t, err)
	attrs[attr1.Key] = attr1
	attr2, err := documents.NewStringAttribute("AssetValue", documents.AttrDecimal, "100")
	assert.NoError(t, err)
	attrs[attr2.Key] = attr2
	attr3, err := documents.NewStringAttribute("AssetIdentifier", documents.AttrBytes, hexutil.Encode(utils.RandomSlice(32)))
	assert.NoError(t, err)
	attrs[attr3.Key] = attr3
	attr4, err := documents.NewStringAttribute("MaturityDate", documents.AttrTimestamp, time.Now().Format(time.RFC3339Nano))
	assert.NoError(t, err)
	attrs[attr4.Key] = attr4

	var proofFields []string
	for _, a := range []documents.Attribute{attr1, attr2, attr3, attr4} {
		proofFields = append(proofFields, fmt.Sprintf("%s.attributes[%s].byte_val", documents.CDTreePrefix, a.Key.String()))
	}
	return attrs, proofFields
}

func GetSignatureProofField(t *testing.T, tcr config.Account) string {
	did := tcr.GetIdentityID()
	keys, err := tcr.GetKeys()
	assert.NoError(t, err)
	pub := keys[identity.KeyPurposeSigning.Name].PublicKey
	id := append(did, pub...)
	return fmt.Sprintf("%s.signatures[%s]", documents.SignaturesTreePrefix, hexutil.Encode(id))
}

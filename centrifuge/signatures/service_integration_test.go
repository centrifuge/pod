// +build ethereum

package signatures_test

import (
	"encoding/base64"
	"os"
	"testing"
	"time"

	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/coredocument"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/config"
	cc "github.com/CentrifugeInc/go-centrifuge/centrifuge/context/testing"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/identity"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/signatures"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/testingutils"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/tools"
	"github.com/stretchr/testify/assert"
)

var identityService identity.IdentityService

func TestMain(m *testing.M) {
	// Adding delay to startup (concurrency hack)
	// TODO: look for other sleep statements in tests and fix the underlying issues
	time.Sleep(time.Second + 2)

	cc.TestFunctionalEthereumBootstrap()
	signatures.NewSigningService(signatures.SigningService{IdentityService: &identity.EthereumIdentityService{}})
	config.Config.V.Set("keys.signing.publicKey", "../../example/resources/signingKey.pub.pem")
	config.Config.V.Set("keys.signing.privateKey", "../../example/resources/signingKey.key.pem")

	identityService = &identity.EthereumIdentityService{}
	result := m.Run()
	cc.TestFunctionalEthereumTearDown()
	os.Exit(result)
}

func TestValidateDocumentSignature(t *testing.T) {
	centrifugeId := tools.RandomSlice(identity.CentIdByteLength)
	config.Config.V.Set("identityId", base64.StdEncoding.EncodeToString(centrifugeId))

	identityConfig, err := identity.NewIdentityConfig()
	assert.Nil(t, err, "should not error out wrapping identity config")

	id, confirmations, err := identityService.CreateIdentity(centrifugeId)
	assert.Nil(t, err, "should not error out when creating identity")

	watchRegisteredIdentity := <-confirmations
	assert.Nil(t, watchRegisteredIdentity.Error, "No error thrown by context")
	assert.Equal(t, centrifugeId, watchRegisteredIdentity.Identity.GetCentrifugeId(), "Resulting Identity should have the same ID as the input")

	confirmations, err = id.AddKeyToIdentity(2, identityConfig.PublicKey)
	assert.Nil(t, err, "should not error out when adding key to identity")
	watchReceivedIdentity := <-confirmations
	assert.Equal(t, centrifugeId, watchReceivedIdentity.Identity.GetCentrifugeId(), "Resulting Identity should have the same ID as the input")

	dataRoot := testingutils.Rand32Bytes()
	documentIdentifier := testingutils.Rand32Bytes()
	nextIdentifier := testingutils.Rand32Bytes()

	doc := &coredocumentpb.CoreDocument{
		DataRoot:           dataRoot,
		DocumentIdentifier: documentIdentifier,
		NextIdentifier:     nextIdentifier,
		SigningRoot:        identityConfig.PublicKey,
	}

	signingService := signatures.GetSigningService()
	signingService.Sign(doc)

	valid, err := signingService.ValidateSignaturesOnDocument(doc)
	assert.Nil(t, err)
	assert.True(t, valid)

}

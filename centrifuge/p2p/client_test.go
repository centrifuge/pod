// +build unit

package p2p

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	cc "github.com/centrifuge/go-centrifuge/centrifuge/context/testingbootstrap"
	"github.com/centrifuge/go-centrifuge/centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/centrifuge/testingutils"
	"github.com/centrifuge/go-centrifuge/centrifuge/testingutils/commons"
	"github.com/centrifuge/go-centrifuge/centrifuge/tools"
	"github.com/centrifuge/go-centrifuge/centrifuge/version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMain(m *testing.M) {
	cc.TestIntegrationBootstrap()
	result := m.Run()
	cc.TestIntegrationTearDown()
	os.Exit(result)
}

func TestGetSignatureForDocument_fail_connect(t *testing.T) {
	client := &testingcommons.P2PMockClient{}
	coreDoc := testingutils.GenerateCoreDocument()
	ctx := context.Background()

	centrifugeId, err := identity.ToCentID(tools.RandomSlice(identity.CentIDLength))
	assert.Nil(t, err, "centrifugeId not initialized correctly ")
	client.On("RequestDocumentSignature", ctx, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("signature failed")).Once()
	resp, err := getSignatureForDocument(ctx, *coreDoc, client, centrifugeId)
	client.AssertExpectations(t)
	assert.Error(t, err, "must fail")
	assert.Nil(t, resp, "must be nil")
}

func TestGetSignatureForDocument_fail_version_check(t *testing.T) {
	client := &testingcommons.P2PMockClient{}
	coreDoc := testingutils.GenerateCoreDocument()
	ctx := context.Background()
	resp := &p2ppb.SignatureResponse{CentNodeVersion: "1.0.0"}

	centrifugeId, err := identity.ToCentID(tools.RandomSlice(identity.CentIDLength))
	assert.Nil(t, err, "centrifugeId not initialized correctly ")

	client.On("RequestDocumentSignature", ctx, mock.Anything, mock.Anything).Return(resp, nil).Once()
	resp, err = getSignatureForDocument(ctx, *coreDoc, client, centrifugeId)
	client.AssertExpectations(t)
	assert.Error(t, err, "must fail")
	assert.Contains(t, err.Error(), "Incompatible version")
	assert.Nil(t, resp, "must be nil")
}

// TODO(ved): once keyinfo is done, add a successful test
func TestGetSignatureForDocument_fail_centrifugeId(t *testing.T) {
	client := &testingcommons.P2PMockClient{}
	coreDoc := testingutils.GenerateCoreDocument()
	ctx := context.Background()

	centrifugeId, err := identity.ToCentID(tools.RandomSlice(identity.CentIDLength))
	assert.Nil(t, err, "centrifugeId not initialized correctly ")

	randomBytes := tools.RandomSlice(identity.CentIDLength)

	signature := &coredocumentpb.Signature{EntityId: randomBytes, PublicKey: tools.RandomSlice(32)}
	sigResp := &p2ppb.SignatureResponse{
		CentNodeVersion: version.GetVersion().String(),
		Signature:       signature,
	}

	client.On("RequestDocumentSignature", ctx, mock.Anything, mock.Anything).Return(sigResp, nil).Once()
	resp, err := getSignatureForDocument(ctx, *coreDoc, client, centrifugeId)

	client.AssertExpectations(t)
	assert.Nil(t, resp, "must be nil")
	assert.Error(t, err, "must not be nil")
	assert.Contains(t, err.Error(), "[5]signature entity doesn't match provided centID")

}

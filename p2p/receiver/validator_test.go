// +build unit

package receiver

import (
	"testing"

	"github.com/golang/protobuf/proto"

	"github.com/centrifuge/go-centrifuge/crypto"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	"github.com/centrifuge/go-centrifuge/version"
	"github.com/stretchr/testify/assert"
)

var (
	key1Pub = []byte{230, 49, 10, 12, 200, 149, 43, 184, 145, 87, 163, 252, 114, 31, 91, 163, 24, 237, 36, 51, 165, 8, 34, 104, 97, 49, 114, 85, 255, 15, 195, 199}
	key1    = []byte{102, 109, 71, 239, 130, 229, 128, 189, 37, 96, 223, 5, 189, 91, 210, 47, 89, 4, 165, 6, 188, 53, 49, 250, 109, 151, 234, 139, 57, 205, 231, 253, 230, 49, 10, 12, 200, 149, 43, 184, 145, 87, 163, 252, 114, 31, 91, 163, 24, 237, 36, 51, 165, 8, 34, 104, 97, 49, 114, 85, 255, 15, 195, 199}
	id1     = []byte{1, 1, 1, 1, 1, 1}
)

func TestValidate_versionValidator(t *testing.T) {
	vv := versionValidator()

	// Nil header
	err := vv.Validate(nil)
	assert.NotNil(t, err)

	// Empty header
	envelope := &p2ppb.Envelope{Header: &p2ppb.Header{}}
	err = vv.Validate(envelope)
	assert.NotNil(t, err)

	// Incompatible Major
	envelope.Header.NodeVersion = "1.1.1"
	err = vv.Validate(envelope)
	assert.NotNil(t, err)

	// Compatible Minor
	envelope.Header.NodeVersion = "0.1.1"
	err = vv.Validate(envelope)
	assert.Nil(t, err)

	//Same version
	envelope.Header.NodeVersion = version.GetVersion().String()
	err = vv.Validate(envelope)
	assert.Nil(t, err)
}

func TestValidate_networkValidator(t *testing.T) {
	nv := networkValidator(cfg.GetNetworkID())

	// Nil header
	err := nv.Validate(nil)
	assert.NotNil(t, err)

	envelope := &p2ppb.Envelope{Header: &p2ppb.Header{}}
	err = nv.Validate(envelope)
	assert.NotNil(t, err)

	// Incompatible network
	envelope.Header.NetworkIdentifier = 12
	err = nv.Validate(envelope)
	assert.NotNil(t, err)

	// Compatible network
	envelope.Header.NetworkIdentifier = cfg.GetNetworkID()
	err = nv.Validate(envelope)
	assert.Nil(t, err)
}

func TestValidate_signatureValidator(t *testing.T) {
	sv := signatureValidator()

	// Nil envelope
	err := sv.Validate(nil)
	assert.Error(t, err)

	// Nil Header
	envelope := &p2ppb.Envelope{}
	err = sv.Validate(envelope)
	assert.Error(t, err)

	// Nil Signature
	envelope.Header = &p2ppb.Header{}
	err = sv.Validate(envelope)
	assert.Error(t, err)

	// Signature validation failure
	envelope.Header.Signature = crypto.Sign(id1, key1, key1Pub, key1Pub)
	err = sv.Validate(envelope)
	assert.Error(t, err)

	// Success
	envelope.Header.Signature = nil
	data, err := proto.Marshal(envelope)
	assert.NoError(t, err)
	envelope.Header.Signature = crypto.Sign(id1, key1, key1Pub, data)
	err = sv.Validate(envelope)
	assert.NoError(t, err)
}

func TestValidate_handshakeValidator(t *testing.T) {
	hv := HandshakeValidator(cfg.GetNetworkID())

	// Incompatible version network and wrong signature
	envelope := &p2ppb.Envelope{
		Header: &p2ppb.Header{
			NodeVersion:       "version",
			NetworkIdentifier: 52,
			Signature:         crypto.Sign(id1, key1, key1Pub, key1Pub),
		},
		Body: key1Pub,
	}
	err := hv.Validate(envelope)
	assert.NotNil(t, err)

	// Incompatible version, correct network
	envelope.Header.NetworkIdentifier = cfg.GetNetworkID()
	err = hv.Validate(envelope)
	assert.NotNil(t, err)

	// Compatible version, incorrect network
	envelope.Header.NetworkIdentifier = 52
	envelope.Header.NodeVersion = version.GetVersion().String()
	err = hv.Validate(envelope)
	assert.NotNil(t, err)

	// Compatible version, network and wrong signature
	envelope.Header.NetworkIdentifier = cfg.GetNetworkID()
	err = hv.Validate(envelope)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "signature validation failure")

	// Compatible version, network and signature
	envelope.Header.Signature = nil
	data, err := proto.Marshal(envelope)
	assert.NoError(t, err)
	envelope.Header.Signature = crypto.Sign(id1, key1, key1Pub, data)
	err = hv.Validate(envelope)
	assert.Nil(t, err)
}

package configstore

import (
	"encoding/json"
	"reflect"

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/crypto"
	"github.com/centrifuge/go-centrifuge/identity"
)

type account struct {
	Identity identity.DID `json:"identity" swaggertype:"string"`

	P2PPublicKey  []byte
	P2PPrivateKey []byte

	SigningPublicKey  []byte
	SigningPrivateKey []byte

	WebhookURL       string `json:"webhook_url"`
	PrecommitEnabled bool   `json:"precommit_enabled"`

	AccountProxies config.AccountProxies `json:"account_proxies"`
}

func (acc *account) GetIdentity() identity.DID {
	return acc.Identity
}

func (acc *account) GetP2PPublicKey() []byte {
	return acc.P2PPublicKey
}

func (acc *account) GetSigningPublicKey() []byte {
	return acc.P2PPublicKey
}

func (acc *account) GetWebhookURL() string {
	return acc.WebhookURL
}

// GetPrecommitEnabled gets the enable pre commit value
func (acc *account) GetPrecommitEnabled() bool {
	return acc.PrecommitEnabled
}

func (acc *account) GetAccountProxies() config.AccountProxies {
	return acc.AccountProxies
}

// SignMsg signs a message with the signing key
func (acc *account) SignMsg(msg []byte) (*coredocumentpb.Signature, error) {
	signature, err := crypto.SignMessage(acc.SigningPrivateKey, msg, crypto.CurveEd25519)
	if err != nil {
		return nil, err
	}

	did := acc.GetIdentity()

	return &coredocumentpb.Signature{
		SignatureId: append(did[:], acc.SigningPublicKey...),
		SignerId:    did[:],
		PublicKey:   acc.SigningPublicKey,
		Signature:   signature,
	}, nil
}

// Type Returns the underlying type of the Account
func (acc *account) Type() reflect.Type {
	return reflect.TypeOf(acc)
}

// JSON return the json representation of the model
func (acc *account) JSON() ([]byte, error) {
	return json.Marshal(acc)
}

// FromJSON initialize the model with a json
func (acc *account) FromJSON(data []byte) error {
	return json.Unmarshal(data, acc)
}

package configstore

import (
	"encoding/json"
	"reflect"

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/crypto"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	libp2pcrypto "github.com/libp2p/go-libp2p-core/crypto"
)

type NodeAdmin struct {
	AccountID *types.AccountID `json:"account_id"`
}

func NewNodeAdmin(accountID *types.AccountID) config.NodeAdmin {
	return &NodeAdmin{accountID}
}

func (n *NodeAdmin) GetAccountID() *types.AccountID {
	return n.AccountID
}

// Type Returns the underlying type of the Account
func (n *NodeAdmin) Type() reflect.Type {
	return reflect.TypeOf(n)
}

// JSON return the json representation of the model
func (n *NodeAdmin) JSON() ([]byte, error) {
	return json.Marshal(n)
}

// FromJSON initialize the model with a json
func (n *NodeAdmin) FromJSON(data []byte) error {
	return json.Unmarshal(data, n)
}

type Account struct {
	Identity *types.AccountID `json:"identity" swaggertype:"string"`

	SigningPublicKey  []byte
	SigningPrivateKey []byte

	WebhookURL       string `json:"webhook_url"`
	PrecommitEnabled bool   `json:"precommit_enabled"`
}

func NewAccount(
	identity *types.AccountID,
	signingPublicKey libp2pcrypto.PubKey,
	signingPrivateKey libp2pcrypto.PrivKey,
	webhookURL string,
	precommitEnabled bool,
) (config.Account, error) {
	signingPublicKeyRaw, err := signingPublicKey.Raw()

	if err != nil {
		return nil, err
	}

	signingPrivateKeyRaw, err := signingPrivateKey.Raw()

	if err != nil {
		return nil, err
	}

	return &Account{
		Identity:          identity,
		SigningPublicKey:  signingPublicKeyRaw,
		SigningPrivateKey: signingPrivateKeyRaw,
		WebhookURL:        webhookURL,
		PrecommitEnabled:  precommitEnabled,
	}, nil
}

func (acc *Account) GetIdentity() *types.AccountID {
	return acc.Identity
}

func (acc *Account) GetSigningPublicKey() []byte {
	return acc.SigningPublicKey
}

func (acc *Account) GetWebhookURL() string {
	return acc.WebhookURL
}

// GetPrecommitEnabled gets the enable pre commit value
func (acc *Account) GetPrecommitEnabled() bool {
	return acc.PrecommitEnabled
}

// SignMsg signs a message with the signing key
func (acc *Account) SignMsg(msg []byte) (*coredocumentpb.Signature, error) {
	sign, err := crypto.SignMessage(acc.SigningPrivateKey, msg, crypto.CurveEd25519)
	if err != nil {
		return nil, err
	}

	did := acc.GetIdentity()

	return &coredocumentpb.Signature{
		SignatureId: append(did.ToBytes(), acc.SigningPublicKey...),
		SignerId:    did[:],
		PublicKey:   acc.SigningPublicKey,
		Signature:   sign,
	}, nil
}

// Type Returns the underlying type of the Account
func (acc *Account) Type() reflect.Type {
	return reflect.TypeOf(acc)
}

// JSON return the json representation of the model
func (acc *Account) JSON() ([]byte, error) {
	return json.Marshal(acc)
}

// FromJSON initialize the model with a json
func (acc *Account) FromJSON(data []byte) error {
	return json.Unmarshal(data, acc)
}

type PodOperator struct {
	URI       string           `json:"uri"`
	AccountID *types.AccountID `json:"account_id"`
}

func NewPodOperator(URI string, accountID *types.AccountID) config.PodOperator {
	return &PodOperator{
		URI:       URI,
		AccountID: accountID,
	}
}

func (p *PodOperator) GetURI() string {
	return p.URI
}

func (p *PodOperator) GetAccountID() *types.AccountID {
	return p.AccountID
}

func (p *PodOperator) ToKeyringPair() signature.KeyringPair {
	return signature.KeyringPair{
		URI:       p.URI,
		PublicKey: p.AccountID.ToBytes(),
	}
}

// Type Returns the underlying type of the Account
func (p *PodOperator) Type() reflect.Type {
	return reflect.TypeOf(p)
}

// JSON return the json representation of the model
func (p *PodOperator) JSON() ([]byte, error) {
	return json.Marshal(p)
}

// FromJSON initialize the model with a json
func (p *PodOperator) FromJSON(data []byte) error {
	return json.Unmarshal(data, p)
}

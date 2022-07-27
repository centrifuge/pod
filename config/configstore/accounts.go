package configstore

import (
	"encoding/json"
	"reflect"

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/crypto"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	libp2pcrypto "github.com/libp2p/go-libp2p-core/crypto"
)

type NodeAdmin struct {
	accountID *types.AccountID
}

func NewNodeAdmin(accountID *types.AccountID) config.NodeAdmin {
	return &NodeAdmin{accountID}
}

func (n *NodeAdmin) AccountID() *types.AccountID {
	return n.accountID
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

	P2PPublicKey  []byte
	P2PPrivateKey []byte

	SigningPublicKey  []byte
	SigningPrivateKey []byte

	WebhookURL       string `json:"webhook_url"`
	PrecommitEnabled bool   `json:"precommit_enabled"`

	AccountProxies config.AccountProxies `json:"account_proxies"`
}

func NewAccount(
	identity *types.AccountID,
	p2pPublicKey libp2pcrypto.PubKey,
	p2pPrivateKey libp2pcrypto.PrivKey,
	signingPublicKey libp2pcrypto.PubKey,
	signingPrivateKey libp2pcrypto.PrivKey,
	webhookURL string,
	precommitEnabled bool,
	accountProxies config.AccountProxies,
) (config.Account, error) {
	p2pPublicKeyRaw, err := p2pPublicKey.Raw()

	if err != nil {
		return nil, err
	}

	p2pPrivateKeyRaw, err := p2pPrivateKey.Raw()

	if err != nil {
		return nil, err
	}

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
		P2PPublicKey:      p2pPublicKeyRaw,
		P2PPrivateKey:     p2pPrivateKeyRaw,
		SigningPublicKey:  signingPublicKeyRaw,
		SigningPrivateKey: signingPrivateKeyRaw,
		WebhookURL:        webhookURL,
		PrecommitEnabled:  precommitEnabled,
		AccountProxies:    accountProxies,
	}, nil
}

func (acc *Account) GetIdentity() *types.AccountID {
	return acc.Identity
}

func (acc *Account) GetP2PPublicKey() []byte {
	return acc.P2PPublicKey
}

func (acc *Account) GetSigningPublicKey() []byte {
	return acc.P2PPublicKey
}

func (acc *Account) GetWebhookURL() string {
	return acc.WebhookURL
}

// GetPrecommitEnabled gets the enable pre commit value
func (acc *Account) GetPrecommitEnabled() bool {
	return acc.PrecommitEnabled
}

func (acc *Account) GetAccountProxies() config.AccountProxies {
	return acc.AccountProxies
}

// SignMsg signs a message with the signing key
func (acc *Account) SignMsg(msg []byte) (*coredocumentpb.Signature, error) {
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

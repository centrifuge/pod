//go:build unit || integration || testworld

package keyrings

import (
	"fmt"

	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/vedhavyas/go-subkey/v2/sr25519"
)

const (
	AlicePubKeyHex   = "0xd43593c715fdd31c61141abd04a99fd6822c8558854ccde39a5684e7a56da27d"
	BobPubKeyHex     = "0x8eaf04151687736326c9fea17e25fc5287613693c912909cb226aa4794f26a48"
	CharliePubKeyHex = "0x90b5ab205c6974c9ea841be688864633dc9ca8a357843eeacf2314649965fe22"
	DavePubKeyHex    = "0x306721211d5404bd9da88e0204360a1a9ab8b87c66c1bc2fcdd37f3c2222cc20"
	EvePubKeyHex     = "0xe659a7a1628cdd93febc04a4e0646ea20e9f5f0ce097d9a05290d4a9e054df4e"
	FerdiePubKeyHex  = "0x1cbd2d43530a44705ad088af313e18f80b53ef16b36177cd4b77b846f2a5f07c"
)

var (
	AliceKeyRingPair = signature.KeyringPair{
		URI:       "//Alice",
		Address:   "4bzRqvQ75UxEpP5R8u5eTyDCymgy2jzCezbFALvqydsPLTqF",
		PublicKey: nil,
	}

	BobKeyRingPair = signature.KeyringPair{
		URI:       "//Bob",
		Address:   "kALNreUp6oBmtfG87fe7MakWR8BnmQ4SmKjjfG27iVd3nuTue",
		PublicKey: nil,
	}

	CharlieKeyRingPair = signature.KeyringPair{
		URI:       "//Charlie",
		Address:   "kALRWiex98BfXHsXnCRhgaQxhxKU9PNKckWzzo2Qpn2EKCarm",
		PublicKey: nil,
	}

	DaveKeyRingPair = signature.KeyringPair{
		URI:       "//Dave",
		Address:   "kAJFEnqV7LCiCaxNoSu7esnt96V7diRvYBZ9xeUHZg5k6Dqo4",
		PublicKey: nil,
	}

	EveKeyRingPair = signature.KeyringPair{
		URI:       "//Eve",
		Address:   "kANMoW9QF4aVQJaENC86NzuGq5udJWYKBNbz5MZFDVQqTEovC",
		PublicKey: nil,
	}

	FerdieKeyRingPair = signature.KeyringPair{
		URI:       "//Ferdie",
		Address:   "kAHoTPrQXE9SoYurqWopgFFii1c9Vz2KcEZ6VySRfX691nc51",
		PublicKey: nil,
	}
)

func init() {
	var err error

	AliceKeyRingPair.PublicKey, err = hexutil.Decode(AlicePubKeyHex)

	if err != nil {
		panic("couldn't decode Alice's public key")
	}

	BobKeyRingPair.PublicKey, err = hexutil.Decode(BobPubKeyHex)

	if err != nil {
		panic("couldn't decode Bob's public key")
	}

	CharlieKeyRingPair.PublicKey, err = hexutil.Decode(CharliePubKeyHex)

	if err != nil {
		panic("couldn't decode Charlie's public key")
	}

	DaveKeyRingPair.PublicKey, err = hexutil.Decode(DavePubKeyHex)

	if err != nil {
		panic("couldn't decode Dave's public key")
	}

	EveKeyRingPair.PublicKey, err = hexutil.Decode(EvePubKeyHex)

	if err != nil {
		panic("couldn't decode Dave's public key")
	}

	FerdieKeyRingPair.PublicKey, err = hexutil.Decode(FerdiePubKeyHex)

	if err != nil {
		panic("couldn't decode Ferdie's public key")
	}
}

func GenerateKeyringPair() (*signature.KeyringPair, error) {
	var scheme sr25519.Scheme

	kp, err := scheme.Generate()

	if err != nil {
		return nil, fmt.Errorf("couldn't generate seed: %w", err)
	}

	return &signature.KeyringPair{
		URI:       hexutil.Encode(kp.Seed()),
		Address:   kp.SS58Address(36),
		PublicKey: kp.AccountID(),
	}, nil
}

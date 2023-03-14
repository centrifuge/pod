//go:build testworld

package host

import (
	"fmt"

	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/pod/crypto"
	"github.com/centrifuge/pod/http/auth"
	"github.com/vedhavyas/go-subkey/v2"
	"github.com/vedhavyas/go-subkey/v2/sr25519"
)

type SignerAccount struct {
	AccountID  *types.AccountID
	Address    string
	SecretSeed string
}

func (s *SignerAccount) ToKeyringPair() signature.KeyringPair {
	return signature.KeyringPair{
		URI:       s.SecretSeed,
		Address:   s.Address,
		PublicKey: s.AccountID.ToBytes(),
	}
}

func GetSignerAccount(secretSeed string) (*SignerAccount, error) {
	kp, err := subkey.DeriveKeyPair(sr25519.Scheme{}, secretSeed)

	if err != nil {
		return nil, fmt.Errorf("couldn't derive signer account key pair: %w", err)
	}

	accountID, err := types.NewAccountID(kp.AccountID())

	if err != nil {
		return nil, fmt.Errorf("couldn't create signer account ID: %w", err)
	}

	return &SignerAccount{
		AccountID:  accountID,
		Address:    kp.SS58Address(auth.CentrifugeNetworkID),
		SecretSeed: secretSeed,
	}, nil
}

func GenerateSignerAccount() (*SignerAccount, error) {
	secretSeed, err := crypto.GenerateSR25519SecretSeed()

	if err != nil {
		return nil, fmt.Errorf("couldn't generate secret seed: %w", err)
	}

	signerAccount, err := GetSignerAccount(secretSeed)

	if err != nil {
		return nil, fmt.Errorf("couldn't get signer account: %w", err)
	}

	return signerAccount, nil
}

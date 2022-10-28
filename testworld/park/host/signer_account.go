//go:build testworld

package host

import (
	"fmt"

	"github.com/centrifuge/go-centrifuge/http/auth"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/vedhavyas/go-subkey/v2"
	"github.com/vedhavyas/go-subkey/v2/sr25519"
)

type SignerAccount struct {
	AccountID  *types.AccountID
	Address    string
	SecretSeed string
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

//go:build unit || integration || testworld

package p2p

import (
	"fmt"

	"github.com/centrifuge/pod/config"
	"github.com/centrifuge/pod/crypto"
	p2pcommon "github.com/centrifuge/pod/p2p/common"
	"github.com/centrifuge/pod/utils"
)

func GetLocalP2PAddress(cfg config.Configuration) (string, error) {
	_, pubKey, err := crypto.ObtainP2PKeypair(cfg.GetP2PKeyPair())

	if err != nil {
		return "", fmt.Errorf("couldn't obtain key pair: %w", err)
	}

	rawPubKey, err := pubKey.Raw()

	if err != nil {
		return "", fmt.Errorf("couldn't obtain raw public key: %w", err)
	}

	p, err := utils.SliceToByte32(rawPubKey)

	if err != nil {
		return "", fmt.Errorf("couldn't convert public key: %w", err)
	}

	peerID, err := p2pcommon.ParsePeerID(p)

	if err != nil {
		return "", fmt.Errorf("couldn't parse peer ID: %w", err)
	}

	addr := fmt.Sprintf("/ip4/127.0.0.1/tcp/%d/ipfs/%s", cfg.GetP2PPort(), peerID.String())

	return addr, nil
}

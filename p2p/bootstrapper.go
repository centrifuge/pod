package p2p

import (
	"fmt"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/keytools/ed25519"
)

// Bootstrapper implements Bootstrapper with p2p details
type Bootstrapper struct{}

// Bootstrap initiates p2p server and client into context
func (b Bootstrapper) Bootstrap(ctx map[string]interface{}) error {
	if _, ok := ctx[bootstrap.BootstrappedConfig]; !ok {
		return fmt.Errorf("config not initialised")
	}

	cfg := ctx[bootstrap.BootstrappedConfig].(*config.Configuration)
	publicKey, privateKey, err := ed25519.GetSigningKeyPairFromConfig()
	if err != nil {
		return fmt.Errorf("failed to get p2p keys: %v", err)
	}

	srv := &p2pServer{
		port:           cfg.GetP2PPort(),
		bootstrapPeers: cfg.GetBootstrapPeers(),
		publicKey:      publicKey,
		privateKey:     privateKey,
	}
	ctx[bootstrap.BootstrappedP2PServer] = srv
	ctx[bootstrap.BootstrappedP2PClient] = srv
	fmt.Println(ctx)
	return nil
}

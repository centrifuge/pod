package dispatcher

import (
	"context"

	"github.com/libp2p/go-libp2p-core/protocol"
)

const BootstrappedProtocolIDDispatcher = "BootstrappedProtocolIDDispatcher"

type Bootstrapper struct{}

func (b *Bootstrapper) Bootstrap(ctx map[string]any) error {
	protocolIDDispatcher := NewDispatcher[protocol.ID](context.Background())

	ctx[BootstrappedProtocolIDDispatcher] = protocolIDDispatcher

	return nil
}

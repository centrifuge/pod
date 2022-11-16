//go:build unit || integration || testworld

package proxy

import (
	"context"
	"fmt"
	"time"

	"github.com/centrifuge/go-centrifuge/pallets/proxy"
	genericUtils "github.com/centrifuge/go-centrifuge/testingutils/generic"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

const (
	proxiesCheckInterval = 10 * time.Second
	proxiesCheckTimeout  = 3 * time.Minute
)

func WaitForProxiesToBeAdded(
	ctx context.Context,
	serviceCtx map[string]any,
	delegatorAccountID *types.AccountID,
	expectedProxies ...*types.AccountID,
) error {
	proxyAPI := genericUtils.GetService[proxy.API](serviceCtx)
	expectedProxiesMap := make(map[string]struct{})

	for _, expectedProxy := range expectedProxies {
		expectedProxiesMap[expectedProxy.ToHexString()] = struct{}{}
	}

	ctx, cancel := context.WithTimeout(ctx, proxiesCheckTimeout)
	defer cancel()

	t := time.NewTicker(proxiesCheckInterval)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context expired while checking for proxies: %w", ctx.Err())
		case <-t.C:
			res, err := proxyAPI.GetProxies(delegatorAccountID)

			if err != nil {
				continue
			}

			for _, proxyDefinition := range res.ProxyDefinitions {
				delete(expectedProxiesMap, proxyDefinition.Delegate.ToHexString())
			}

			if len(expectedProxiesMap) == 0 {
				return nil
			}
		}
	}
}

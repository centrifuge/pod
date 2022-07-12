package proxy

import (
	"fmt"
	"github.com/centrifuge/go-centrifuge/centchain"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/vedhavyas/go-subkey/v2"
	"strconv"
)

import "github.com/centrifuge/go-substrate-rpc-client/v4/types"

const (
	Pallet        = "Proxy"
	ProxiesMethod = "Proxies"
	NodeAdminRole = "NodeAdmin"
)

type ProxyDelegate struct {
	Delegate  types.AccountID
	ProxyType types.ProxyType
	Delay     types.U32
}

type ProxyDefinition struct {
	Delegates []ProxyDelegate
	Amount    types.U128
}

type Service interface {
	GetProxy(address string) (*ProxyDefinition, error)
	ProxyHasProxyType(proxyDef *ProxyDefinition, proxied []byte, proxyType string) bool
}

type service struct {
	api centchain.API
}

func newService(api centchain.API) Service {
	return &service{api: api}
}

func (s service) GetProxy(address string) (*ProxyDefinition, error) {
	_, proxyPublicKey, err := subkey.SS58Decode(address)
	if err != nil {
		return nil, err
	}

	meta, err := s.api.GetMetadataLatest()
	if err != nil {
		return nil, err
	}

	key, err := types.CreateStorageKey(meta, Pallet, ProxiesMethod, proxyPublicKey)
	if err != nil {
		return nil, err
	}

	var proxyDef ProxyDefinition
	err = s.api.GetStorageLatest(key, &proxyDef)
	if err != nil {
		return nil, fmt.Errorf("failed to get the proxy definition: %w", err)
	}

	return &proxyDef, nil
}

func (s service) ProxyHasProxyType(proxyDef *ProxyDefinition, proxied []byte, proxyType string) bool {
	valid := false
	for _, d := range proxyDef.Delegates {
		if utils.IsSameByteSlice(utils.Byte32ToSlice(d.Delegate), proxied) {
			pxInt, err := strconv.Atoi(proxyType)
			if err != nil {
				return false
			}

			if uint8(d.ProxyType) == uint8(pxInt) {
				valid = true
				break
			}
		}

	}
	return valid
}

package proxy

import (
	"fmt"

	"github.com/centrifuge/go-centrifuge/centchain"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/vedhavyas/go-subkey/v2"
)

const (
	Pallet        = "Proxy"
	ProxiesMethod = "Proxies"
	NodeAdminRole = "NodeAdmin"
)

type Delegate struct {
	Delegate  types.AccountID
	ProxyType types.ProxyType
	Delay     types.U32
}

type Definition struct {
	Delegates []Delegate
	Amount    types.U128
}

type Service interface {
	GetProxy(address string) (*Definition, error)
	ProxyHasProxyType(proxyDef *Definition, proxied []byte, proxyType string) bool
}

type service struct {
	api centchain.API
}

func newService(api centchain.API) Service {
	return &service{api: api}
}

func (s service) GetProxy(address string) (*Definition, error) {
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

	var proxyDef Definition
	err = s.api.GetStorageLatest(key, &proxyDef)
	if err != nil {
		return nil, fmt.Errorf("failed to get the proxy definition: %w", err)
	}

	return &proxyDef, nil
}

// TODO: Annoying way of converting types, we can have a better approach
func convertToProxy(proxyType string) (types.ProxyType, error) {
	switch proxyType {
	case "Any":
		return types.Any, nil
	default:
		return types.ProxyType(255), errors.New("unsupported proxy type")
	}
}

func (s service) ProxyHasProxyType(proxyDef *Definition, proxied []byte, proxyType string) bool {
	valid := false
	for _, d := range proxyDef.Delegates {
		if utils.IsSameByteSlice(utils.Byte32ToSlice(d.Delegate), proxied) {
			pt, err := convertToProxy(proxyType)
			if err != nil {
				fmt.Println(err.Error())
				return false
			}
			if pt == d.ProxyType {
				valid = true
				break
			}
		}
	}
	return valid
}

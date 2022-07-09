package identity

import "github.com/centrifuge/go-substrate-rpc-client/v4/types"

const (
	ProxyPallet   = "Proxy"
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

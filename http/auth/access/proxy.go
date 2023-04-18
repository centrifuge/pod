package access

import (
	"net/http"

	proxyTypes "github.com/centrifuge/chain-custom-types/pkg/proxy"

	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/pod/config"
	authToken "github.com/centrifuge/pod/http/auth/token"
	"github.com/centrifuge/pod/pallets/proxy"
	logging "github.com/ipfs/go-log"
)

type proxyAccessValidator struct {
	log       *logging.ZapEventLogger
	configSrv config.Service
	proxyAPI  proxy.API
}

func NewProxyAccessValidator(
	configSrv config.Service,
	proxyAPI proxy.API,
) Validator {
	log := logging.Logger("http-proxy-access-validator")

	return &proxyAccessValidator{
		log:       log,
		configSrv: configSrv,
		proxyAPI:  proxyAPI,
	}
}

func (p *proxyAccessValidator) Validate(_ *http.Request, token *authToken.JW3Token) (*types.AccountID, error) {
	delegateAccountID, err := authToken.DecodeSS58Address(token.Payload.Address)

	if err != nil {
		p.log.Error("Couldn't decode delegate account ID: %s", err)

		return nil, ErrSS58AddressDecode
	}

	delegatorAccountID, err := authToken.DecodeSS58Address(token.Payload.OnBehalfOf)

	if err != nil {
		p.log.Error("Couldn't decode delegator account ID: %s", err)

		return nil, ErrSS58AddressDecode
	}

	// Verify that the delegator is a valid Identity on the pod
	if _, err = p.configSrv.GetAccount(delegatorAccountID.ToBytes()); err != nil {
		p.log.Errorf("Delegator account not found: %s", err)

		return nil, ErrDelegatorNotFound
	}

	// Verify that the delegate is a valid proxy of the delegator.
	proxyStorageEntry, err := p.proxyAPI.GetProxies(delegatorAccountID)

	if err != nil {
		p.log.Errorf("Couldn't retrieve account proxies: %s", err)

		return nil, ErrAccountProxiesRetrieval
	}

	pt := proxyTypes.ProxyTypeValue[token.Payload.ProxyType]

	valid := false
	for _, proxyDefinition := range proxyStorageEntry.ProxyDefinitions {
		if proxyDefinition.Delegate.Equal(delegateAccountID) {
			if uint8(proxyDefinition.ProxyType) == uint8(pt) {
				valid = true
				break
			}
		}
	}

	if !valid {
		p.log.Errorf("Invalid delegate")

		return nil, ErrInvalidDelegate
	}

	return delegatorAccountID, nil
}

package access

import (
	"net/http"

	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/pod/config"
	authToken "github.com/centrifuge/pod/http/auth/token"
	logging "github.com/ipfs/go-log"
)

type adminAccessValidator struct {
	log       *logging.ZapEventLogger
	configSrv config.Service
}

func NewAdminAccessValidator(configSrv config.Service) Validator {
	log := logging.Logger("http-admin-access-validator")

	return &adminAccessValidator{
		log:       log,
		configSrv: configSrv,
	}
}

func (a *adminAccessValidator) Validate(_ *http.Request, token *authToken.JW3Token) (*types.AccountID, error) {
	delegateAccountID, err := authToken.DecodeSS58Address(token.Payload.Address)

	if err != nil {
		a.log.Error("Couldn't decode admin account ID: %s", err)

		return nil, ErrSS58AddressDecode
	}

	podAdmin, err := a.configSrv.GetPodAdmin()

	if err != nil {
		a.log.Error("Couldn't retrieve POD admin: %s", err)

		return nil, ErrPodAdminRetrieval
	}

	if !podAdmin.GetAccountID().Equal(delegateAccountID) {
		return nil, ErrNotAdminAccount
	}

	return delegateAccountID, nil
}

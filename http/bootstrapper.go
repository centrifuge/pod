package http

import (
	"context"
	"fmt"
	"sync"

	"github.com/centrifuge/pod/pallets/permissions"
	"github.com/centrifuge/pod/pallets/uniques"

	"github.com/centrifuge/pod/pallets/loans"

	"github.com/centrifuge/pod/http/auth/access"

	"github.com/centrifuge/pod/bootstrap"
	"github.com/centrifuge/pod/config"
	"github.com/centrifuge/pod/errors"
	"github.com/centrifuge/pod/pallets"
	"github.com/centrifuge/pod/pallets/proxy"
)

const (
	BootstrappedValidationWrapperFactory = "BootstrappedValidationWrapperFactory"
)

// Bootstrapper implements bootstrapper.Bootstrapper
type Bootstrapper struct {
	testServerWg        sync.WaitGroup
	testServerCtx       context.Context
	testServerCtxCancel context.CancelFunc
}

// Bootstrap initiates api server
func (b *Bootstrapper) Bootstrap(ctx map[string]interface{}) error {
	cfgService, ok := ctx[config.BootstrappedConfigStorage].(config.Service)
	if !ok {
		return errors.New("config storage not initialised")
	}

	proxyAPI, ok := ctx[pallets.BootstrappedProxyAPI].(proxy.API)
	if !ok {
		return errors.New("proxy API not initialised")
	}

	loansAPI, ok := ctx[pallets.BootstrappedLoansAPI].(loans.API)
	if !ok {
		return errors.New("loans API not initialised")
	}

	permissionsAPI, ok := ctx[pallets.BootstrappedPermissionsAPI].(permissions.API)
	if !ok {
		return errors.New("permissions API not initialised")
	}

	uniquesAPI, ok := ctx[pallets.BootstrappedUniquesAPI].(uniques.API)
	if !ok {
		return errors.New("uniques API not initialised")
	}

	cfg, err := cfgService.GetConfig()

	if err != nil {
		return fmt.Errorf("couldn't retrieve config: %w", err)
	}

	proxyAccessValidator := access.NewProxyAccessValidator(cfgService, proxyAPI)
	adminAccessValidator := access.NewAdminAccessValidator(cfgService)
	investorAccessValidator := access.NewInvestorAccessValidator(loansAPI, permissionsAPI, uniquesAPI)

	validationWrapperFactory := access.NewValidationWrapperFactory(cfgService, proxyAccessValidator, adminAccessValidator, investorAccessValidator)

	ctx[BootstrappedValidationWrapperFactory] = validationWrapperFactory

	srv := apiServer{config: cfg}
	ctx[bootstrap.BootstrappedAPIServer] = srv
	return nil
}

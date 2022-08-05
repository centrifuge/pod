package v2

import (
	"context"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity/v2/keystore"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/gocelery/v2"
	logging "github.com/ipfs/go-log"
)

const (
	addKeysJob = "Add keys to keystore"
)

type AddKeysJob struct {
	jobs.Base

	keystoreAPI keystore.API
	log         *logging.ZapEventLogger
}

func (a *AddKeysJob) New() gocelery.Runner {
	log := logging.Logger("add_keys_to_keystore_dispatcher")

	aj := &AddKeysJob{
		keystoreAPI: a.keystoreAPI,
		log:         log,
	}

	aj.Base = jobs.NewBase(aj.loadTasks())
	return aj
}

func (a *AddKeysJob) convertArgs(args []interface{}) (
	config.Account,
	[]*types.AddKey,
	error,
) {
	account, ok := args[0].(config.Account)

	if !ok {
		return nil, nil, errors.New("account not provided in args")
	}

	keys, ok := args[1].([]*types.AddKey)

	if !ok {
		return nil, nil, errors.New("keys not provided in args")
	}

	return account, keys, nil
}

func (a *AddKeysJob) loadTasks() map[string]jobs.Task {
	return map[string]jobs.Task{
		"add_keys_to_keystore": {
			RunnerFunc: func(args []interface{}, overrides map[string]interface{}) (result interface{}, err error) {
				account, keys, err := a.convertArgs(args)

				if err != nil {
					a.log.Errorf("Couldn't convert args: %s", err)

					return nil, err
				}

				ctx := contextutil.WithAccount(context.Background(), account)

				extInfo, err := a.keystoreAPI.AddKeys(ctx, keys)

				if err != nil {
					a.log.Errorf("Couldn't add keys to keystore: %s", err)

					return nil, err
				}

				a.log.Infof("Successfully added keys to keystore, ext hash - %s", extInfo.Hash.Hex())

				overrides["ext_info"] = extInfo

				return nil, nil
			},
		},
	}
}

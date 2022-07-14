package v2

import (
	"context"

	"github.com/centrifuge/go-substrate-rpc-client/v4/types"

	"github.com/centrifuge/go-centrifuge/identity/v2/keystore"
	"github.com/centrifuge/go-centrifuge/jobs"
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
	ctx context.Context,
	keys []*types.AddKey,
	err error,
) {
	ctx = args[0].(context.Context)
	keys = args[1].([]*types.AddKey)

	return ctx, keys, nil
}

func (a *AddKeysJob) loadTasks() map[string]jobs.Task {
	return map[string]jobs.Task{
		"add_keys_to_keystore": {
			RunnerFunc: func(args []interface{}, overrides map[string]interface{}) (result interface{}, err error) {
				ctx, keys, err := a.convertArgs(args)

				if err != nil {
					a.log.Errorf("Couldn't convert args: %s", err)

					return nil, err
				}

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

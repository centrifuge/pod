package documents

import (
	"context"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/coredocument"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/transactions"
	"github.com/centrifuge/gocelery"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/satori/go.uuid"
)

const (
	// nftCreatedTaskName is the name of the NFT Task
	nftCreatedTaskName string = "NFT created task"

	// TokenRegistryParam maps to token registry address
	TokenRegistryParam string = "Token Registry"

	// TokenIDParam maps to NFT token ID
	TokenIDParam string = "Token ID"
)

type nftCreatedTask struct {
	transactions.BaseTask
	accountID     identity.CentID
	documentID    []byte
	tokenRegistry common.Address
	tokenID       []byte

	// state
	docSrv Service
	cfgSrv config.Service
}

func (t *nftCreatedTask) ParseKwargs(kwargs map[string]interface{}) (err error) {
	defer func() {
		log.Error(err)
	}()

	err = t.ParseTransactionID(kwargs)
	if err != nil {
		return err
	}

	acc, ok := kwargs[AccountIDParam].(string)
	if !ok {
		return errors.New("account identifier not set")
	}

	t.accountID, err = identity.CentIDFromString(acc)
	if err != nil {
		return errors.New("invalid accountID")
	}

	vr, ok := kwargs[DocumentIDParam].(string)
	if !ok {
		return errors.New("model identifier not set")
	}

	t.documentID, err = hexutil.Decode(vr)
	if err != nil {
		return err
	}

	tr, ok := kwargs[TokenRegistryParam].(string)
	if !ok {
		return errors.New("token registry not set")
	}

	t.tokenRegistry = common.HexToAddress(tr)

	tidr, ok := kwargs[TokenIDParam].(string)
	if !ok {
		return errors.New("token ID not set")
	}

	t.tokenID, err = hexutil.Decode(tidr)
	if err != nil {
		return err
	}

	return nil
}

func (t *nftCreatedTask) Copy() (gocelery.CeleryTask, error) {
	return &nftCreatedTask{
		BaseTask: t.BaseTask,
		docSrv:   t.docSrv,
		cfgSrv:   t.cfgSrv,
	}, nil
}

func (t *nftCreatedTask) RunTask() (result interface{}, err error) {
	defer func() {
		log.Error(err)
	}()

	defer func() {
		err = t.UpdateTransaction(t.accountID, t.TaskTypeName(), err, false)
	}()

	ctx, err := contextutil.Context(context.Background(), t.cfgSrv)
	if err != nil {
		return nil, err
	}

	model, err := t.docSrv.GetCurrentVersion(ctx, t.documentID)
	if err != nil {
		return nil, err
	}

	cd, err := model.PackCoreDocument()
	if err != nil {
		return nil, err
	}

	cd, err = coredocument.PrepareNewVersion(*cd, nil)
	if err != nil {
		return nil, err
	}

	err = coredocument.AddNFTToReadRules(cd, t.tokenRegistry, t.tokenID)
	if err != nil {
		return nil, err
	}

	model, err = t.docSrv.DeriveFromCoreDocument(cd)
	if err != nil {
		return nil, err
	}

	_, _, err = t.docSrv.Update(ctx, model, t.TxID)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (t *nftCreatedTask) TaskTypeName() string {
	return nftCreatedTaskName
}

// InitNFTCreatedTask initiates a new nft created task
func InitNFTCreatedTask(
	queuer queue.TaskQueuer,
	txID uuid.UUID,
	cid identity.CentID,
	documentID []byte,
	registry common.Address,
	tokenID []byte,
) (queue.TaskResult, error) {
	log.Infof("Starting NFT created task: %v\n", txID.String())
	return queuer.EnqueueJob(nftCreatedTaskName, map[string]interface{}{
		transactions.TxIDParam: txID.String(),
		AccountIDParam:         cid.String(),
		DocumentIDParam:        hexutil.Encode(documentID),
		TokenRegistryParam:     registry.String(),
		TokenIDParam:           hexutil.Encode(tokenID),
	})
}

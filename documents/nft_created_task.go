package documents

import (
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/transactions"
	"github.com/centrifuge/gocelery"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

const (
	// NFTCreatedTask is the name of the NFT Task
	NFTCreatedTask string = "NFT created task"

	// ModelIDParam maps the model identifier
	ModelIDParam string = "Model ID"

	// TokenRegistryParam maps to token registry address
	TokenRegistryParam string = "Token Registry"

	// TokenIDParam maps to NFT token ID
	TokenIDParam string = "Token ID"
)

type nftCreatedTask struct {
	transactions.BaseTask
	modelID       []byte
	tokenRegistry common.Address
	tokenID       []byte

	// state
	docSrv Service
}

func (t *nftCreatedTask) ParseKwargs(kwargs map[string]interface{}) error {
	err := t.ParseTransactionID(kwargs)
	if err != nil {
		return err
	}

	vr, ok := kwargs[ModelIDParam].(string)
	if !ok {
		return errors.New("model identifier not set")
	}

	t.modelID, err = hexutil.Decode(vr)
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
	}, nil
}

func (t *nftCreatedTask) RunTask() (interface{}, error) {

	return nil, nil
}

func (t *nftCreatedTask) TaskTypeName() string {
	return NFTCreatedTask
}

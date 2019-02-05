package did

import (
	"context"
	"fmt"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/transactions"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/satori/go.uuid"

	"github.com/centrifuge/go-centrifuge/ethereum"
	id "github.com/centrifuge/go-centrifuge/identity"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("identity")

// Service is the interface for identity related interactions
type Service interface {
	CreateIdentity(ctx context.Context) (id *DID, confirmations chan *id.WatchIdentity, err error)
}

type service struct {
	config          id.Config
	factoryContract *FactoryContract
	client          ethereum.Client
	txManager transactions.Manager
	queue            *queue.Server
}

// NewService returns a new identity service
func NewService(config id.Config, factoryContract *FactoryContract, client ethereum.Client, txManager transactions.Manager, queue *queue.Server) Service {

	return &service{config: config, factoryContract: factoryContract, client: client, txManager:txManager,queue:queue}
}

func (s *service) getNonceAt(ctx context.Context, address common.Address) (uint64, error) {
	// TODO: add blockNumber of the transaction which created the contract
	return s.client.GetEthClient().NonceAt(ctx, getFactoryAddress(), nil)
}

// CalculateCreatedAddress calculates the Ethereum address based on address and nonce
func CalculateCreatedAddress(address common.Address, nonce uint64) common.Address {
	// How is a Ethereum address calculated:
	// See https://ethereum.stackexchange.com/questions/760/how-is-the-address-of-an-ethereum-contract-computed
	return crypto.CreateAddress(address, nonce)
}


func (s *service) createIdentityTX(opts *bind.TransactOpts) func(accountID id.CentID, txID uuid.UUID, txMan transactions.Manager, errOut chan<- error)  {
	return func(accountID id.CentID, txID uuid.UUID, txMan transactions.Manager, errOut chan<- error) {

		ethTX, err := s.client.SubmitTransactionWithRetries(s.factoryContract.CreateIdentity, opts)
		if err != nil {
			log.Infof("Failed to send identity for creation [txHash: %s] : %v", ethTX.Hash(), err)
			return
		}

		log.Infof("Sent off identity creation Ethereum transaction hash [%x] and Nonce [%v] and Check [%v]", ethTX.Hash(), ethTX.Nonce(), ethTX.CheckNonce())
		log.Infof("Transfer pending: 0x%x\n", ethTX.Hash())


		res, err := ethereum.QueueEthTXStatusTask(accountID, txID, ethTX.Hash(), s.queue)
		if err != nil {
			errOut <- err
			return
		}

		_, err = res.Get(txMan.GetDefaultTaskTimeout())
		if err != nil {
			errOut <- err
			return
		}
		errOut <- nil
	}

}


func (s *service) CreateIdentity(ctx context.Context) (did *DID, confirmations chan *id.WatchIdentity, err error) {
	tc, err := contextutil.Account(ctx)
	if err != nil {
		return nil, confirmations, err
	}

	opts, err := s.client.GetTxOpts(tc.GetEthereumDefaultAccountName())
	if err != nil {
		log.Infof("Failed to get txOpts from Ethereum client: %v", err)
		return nil, nil, err
	}

	idConfig, err := contextutil.Self(ctx)
	if err != nil {
		return nil, nil, err
	}

	txID, done, err := s.txManager.ExecuteWithinTX(context.Background(), idConfig.ID, uuid.Nil, "Check TX status", s.createIdentityTX(opts))

	<- done

	fmt.Println("is not done")
	err = s.txManager.WaitForTransaction(idConfig.ID, txID)

	if err != nil {
			return nil, nil, err
	}


	factoryAddress := getFactoryAddress()
	nonce, err := s.getNonceAt(ctx, factoryAddress)
	if err != nil {
		return nil, nil, err
	}

	identityAddress := CalculateCreatedAddress(factoryAddress, nonce)
	log.Infof("Address of created identity contract: 0x%x\n", identityAddress)

	createdDID := NewDID(identityAddress)

	return &createdDID, nil, nil
}

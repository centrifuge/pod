package did

import (
	"context"

	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("identity")

// Service is the interface for identity related interactions
type Service interface {
	CreateIdentity(ctx context.Context) (id *DID, confirmations chan *identity.WatchIdentity, err error)
}

type service struct {
	config          identity.Config
	factoryContract *FactoryContract
	client          ethereum.Client
}

// NewService returns a new identity service
func NewService(config identity.Config, factoryContract *FactoryContract, client ethereum.Client) Service {

	return &service{config: config, factoryContract: factoryContract, client: client}
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

func (s *service) CreateIdentity(ctx context.Context) (id *DID, confirmations chan *identity.WatchIdentity, err error) {
	opts, err := s.client.GetTxOpts(s.config.GetEthereumDefaultAccountName())
	if err != nil {
		log.Infof("Failed to get txOpts from Ethereum client: %v", err)
		return nil, nil, err
	}

	tx, err := s.client.SubmitTransactionWithRetries(s.factoryContract.CreateIdentity, opts)
	if err != nil {
		log.Infof("Failed to send identity for creation [txHash: %s] : %v", tx.Hash(), err)
		return nil, nil, err
	}

	log.Infof("Sent off identity creation Ethereum transaction hash [%x] and Nonce [%v] and Check [%v]", tx.Hash(), tx.Nonce(), tx.CheckNonce())
	log.Infof("Transfer pending: 0x%x\n", tx.Hash())

	// TODO use transactionStatusTask and following code as a statusHandler

	factoryAddress := getFactoryAddress()
	nonce, err := s.getNonceAt(ctx, factoryAddress)
	if err != nil {
		return nil, nil, err
	}

	identityAddress := CalculateCreatedAddress(factoryAddress, nonce)
	log.Infof("Address of created identity contract: 0x%x\n", identityAddress)

	did := NewDID(identityAddress)

	return &did, nil, nil
}

package ideth

import (
	"context"
	"fmt"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"

	"github.com/centrifuge/go-centrifuge/errors"

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


type factory struct {
	factoryAddress  common.Address
	factoryContract *FactoryContract
	client          ethereum.Client
	txManager       transactions.Manager
	queue           *queue.Server
}

// NewFactory returns a new identity factory service
func NewFactory(factoryContract *FactoryContract, client ethereum.Client, txManager transactions.Manager, queue *queue.Server, factoryAddress common.Address) id.Factory {
	return &factory{factoryAddress: factoryAddress, factoryContract: factoryContract, client: client, txManager: txManager, queue: queue}
}

func (s *factory) getNonceAt(ctx context.Context, address common.Address) (uint64, error) {
	// TODO: add blockNumber of the transaction which created the contract
	return s.client.GetEthClient().NonceAt(ctx, s.factoryAddress, nil)
}

// CalculateCreatedAddress calculates the Ethereum address based on address and nonce
func CalculateCreatedAddress(address common.Address, nonce uint64) common.Address {
	// How is a Ethereum address calculated:
	// See https://ethereum.stackexchange.com/questions/760/how-is-the-address-of-an-ethereum-contract-computed
	return crypto.CreateAddress(address, nonce)
}

func (s *factory) createIdentityTX(opts *bind.TransactOpts) func(accountID id.DID, txID uuid.UUID, txMan transactions.Manager, errOut chan<- error) {
	return func(accountID id.DID, txID uuid.UUID, txMan transactions.Manager, errOut chan<- error) {
		ethTX, err := s.client.SubmitTransactionWithRetries(s.factoryContract.CreateIdentity, opts)
		if err != nil {
			errOut <- err
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

func (s *factory) CalculateIdentityAddress(ctx context.Context) (*common.Address, error) {
	nonce, err := s.getNonceAt(ctx, s.factoryAddress)
	if err != nil {
		return nil, err
	}

	identityAddress := CalculateCreatedAddress(s.factoryAddress, nonce)
	log.Infof("Calculated Address of the identity contract: 0x%x\n", identityAddress)
	return &identityAddress, nil
}

func isIdentityContract(identityAddress common.Address, client ethereum.Client) error {
	contractCode, err := client.GetEthClient().CodeAt(context.Background(), identityAddress, nil)
	if err != nil {
		return err
	}

	deployedContractByte := common.Bytes2Hex(contractCode)
	identityContractByte := getIdentityByteCode()[2:] // remove 0x prefix
	fmt.Printf("Deployed Hash: %x\n", crypto.Keccak256(common.Hex2Bytes(deployedContractByte)))
	fmt.Printf("IDCFG Hash: %x\n", crypto.Keccak256(common.Hex2Bytes(identityContractByte)))
	if deployedContractByte != identityContractByte {
		return errors.New("deployed identity contract bytecode not correct")
	}
	return nil

}

func (s *factory) CreateIdentity(ctx context.Context) (did *id.DID, err error) {
	tc, err := contextutil.Account(ctx)
	if err != nil {
		return nil, err
	}

	opts, err := s.client.GetTxOpts(tc.GetEthereumDefaultAccountName())
	if err != nil {
		log.Infof("Failed to get txOpts from Ethereum client: %v", err)
		return nil, err
	}

	identityAddress, err := s.CalculateIdentityAddress(ctx)
	if err != nil {
		return nil, err
	}

	createdDID := id.NewDID(*identityAddress)

	txID, done, err := s.txManager.ExecuteWithinTX(context.Background(), createdDID, uuid.Nil, "Check TX for create identity status", s.createIdentityTX(opts))
	if err != nil {
		return nil, err
	}

	isDone := <-done
	// non async task
	if !isDone {
		return nil, errors.New("Create Identity TX failed: txID:%s", txID.String())

	}

	err = isIdentityContract(*identityAddress, s.client)
	if err != nil {
		return nil, err
	}

	return &createdDID, nil
}

// CreateIdentity creates an identity contract
func CreateIdentity(ctx map[string]interface{}, cfg config.Configuration) (*id.DID, error) {
	tc, err := configstore.TempAccount(cfg.GetEthereumDefaultAccountName(), cfg)
	if err != nil {
		return nil, err
	}

	tctx, err := contextutil.New(context.Background(), tc)
	if err != nil {
		return nil, err
	}

	identityFactory := ctx[BootstrappedDIDFactory].(id.Factory)

	did, err := identityFactory.CreateIdentity(tctx)
	if err != nil {
		return nil, err
	}

	return did, nil

}

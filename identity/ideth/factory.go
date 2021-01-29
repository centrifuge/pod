package ideth

import (
	"context"
	"fmt"
	"math/big"

	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("eth-identity")

const identityCreatedEventName = "IdentityCreated(address)"

// CalculateCreatedAddress calculates the Ethereum address based on address and nonce
func CalculateCreatedAddress(address common.Address, nonce uint64) common.Address {
	// How is a Ethereum address calculated:
	// See https://ethereum.stackexchange.com/questions/760/how-is-the-address-of-an-ethereum-contract-computed
	return crypto.CreateAddress(address, nonce)
}

func isIdentityContract(identityAddress common.Address, client ethereum.Client) error {
	contractCode, err := client.GetEthClient().CodeAt(context.Background(), identityAddress, nil)
	if err != nil {
		return err
	}

	if len(contractCode) == 0 {
		return errors.New("bytecode for deployed identity contract %s not correct", identityAddress.String())
	}

	return nil
}

type factoryV2 struct {
	factoryAddress  common.Address
	factoryContract *FactoryContract
	client          ethereum.Client
	config          identity.Config
}

func (f factoryV2) IdentityExists(did identity.DID) (exists bool, err error) {
	opts, cancel := f.client.GetGethCallOpts(false)
	defer cancel()
	valid, err := f.factoryContract.CreatedIdentity(opts, did.ToAddress())
	if err != nil {
		return false, err
	}
	return valid, nil
}

func (f factoryV2) NextIdentityAddress() (did identity.DID, err error) {
	nonce, err := f.client.GetEthClient().PendingNonceAt(context.Background(), f.factoryAddress)
	if err != nil {
		return did, fmt.Errorf("failed to fetch identity factory nonce: %w", err)
	}

	addr := CalculateCreatedAddress(f.factoryAddress, nonce)
	return identity.NewDID(addr), nil
}

func (f factoryV2) CreateIdentity(ethAccount string, keys []identity.Key) (transaction *types.
	Transaction, err error) {
	opts, err := f.client.GetTxOpts(context.Background(), ethAccount)
	if err != nil {
		log.Infof("Failed to get txOpts from Ethereum client: %v", err)
		return nil, err
	}

	ethKeys, purposes := convertKeysToEth(keys)
	ethTX, err := f.client.SubmitTransactionWithRetries(
		f.factoryContract.CreateIdentityFor, opts, opts.From, ethKeys, purposes)
	if err != nil {
		return nil, fmt.Errorf("failed to submit eth transaction: %w", err)
	}

	log.Infof("Sent off identity creation Ethereum transaction hash [%x] and Nonce [%v] and Check [%v]", ethTX.Hash(), ethTX.Nonce(), ethTX.CheckNonce())
	return ethTX, nil
}

func convertKeysToEth(keys []identity.Key) (ethKeys [][32]byte, purposes []*big.Int) {
	for _, k := range keys {
		ethKeys = append(ethKeys, k.GetKey())
		purposes = append(purposes, k.GetPurpose())
	}

	return ethKeys, purposes
}

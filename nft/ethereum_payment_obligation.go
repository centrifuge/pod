package nft

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"math/big"
	"time"

	"github.com/centrifuge/go-centrifuge/coredocument"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/transactions"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/precise-proofs/proofs"
	"github.com/centrifuge/precise-proofs/proofs/proto"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	logging "github.com/ipfs/go-log"
	"github.com/satori/go.uuid"
)

var log = logging.Logger("nft")

const (
	// ErrNFTMinted error for NFT already minted for registry
	ErrNFTMinted = errors.Error("NFT already minted")

	// ErrNFTRoleMissing errors when role to generate proof doesn't exist
	ErrNFTRoleMissing = errors.Error("Role doesn't exist")
)

// ethereumPaymentObligationContract is an abstraction over the contract code to help in mocking it out
type ethereumPaymentObligationContract interface {

	// Mint method abstracts Mint method on the contract
	Mint(opts *bind.TransactOpts, to common.Address, tokenID *big.Int, tokenURI string, anchorID *big.Int, merkleRoot [32]byte, values []string, salts [][32]byte, proofs [][][32]byte) (*types.Transaction, error)

	// OwnerOf to retrieve owner of the tokenID
	OwnerOf(opts *bind.CallOpts, tokenID *big.Int) (common.Address, error)
}

// Config is the config interface for nft package
type Config interface {
	GetEthereumContextWaitTimeout() time.Duration
}

// ethereumPaymentObligation handles all interactions related to minting of NFTs for payment obligations on Ethereum
type ethereumPaymentObligation struct {
	cfg             Config
	identityService identity.Service
	ethClient       ethereum.Client
	queue           queue.TaskQueuer
	docSrv          documents.Service
	bindContract    func(address common.Address, client ethereum.Client) (*EthereumPaymentObligationContract, error)
	txManager       transactions.Manager
	blockHeightFunc func() (height uint64, err error)
}

// newEthereumPaymentObligation creates ethereumPaymentObligation given the parameters
func newEthereumPaymentObligation(
	cfg Config,
	identityService identity.Service,
	ethClient ethereum.Client,
	queue queue.TaskQueuer,
	docSrv documents.Service,
	bindContract func(address common.Address, client ethereum.Client) (*EthereumPaymentObligationContract, error),
	txManager transactions.Manager,
	blockHeightFunc func() (uint64, error)) *ethereumPaymentObligation {
	return &ethereumPaymentObligation{
		cfg:             cfg,
		identityService: identityService,
		ethClient:       ethClient,
		bindContract:    bindContract,
		queue:           queue,
		docSrv:          docSrv,
		txManager:       txManager,
		blockHeightFunc: blockHeightFunc,
	}
}

func (s *ethereumPaymentObligation) prepareMintRequest(ctx context.Context, tokenID TokenID, cid identity.CentID, req MintNFTRequest) (mreq MintRequest, err error) {
	model, err := s.docSrv.GetCurrentVersion(ctx, req.DocumentID)
	if err != nil {
		return mreq, err
	}

	corDocModel, err := model.PackCoreDocument()
	if err != nil {
		return mreq, err
	}

	corDoc := corDocModel.Document
	docProofs, err := s.docSrv.CreateProofs(ctx, req.DocumentID, req.ProofFields)
	if err != nil {
		return mreq, err
	}

	nftProofs, err := generateNFTProofs(corDoc, req, tokenID, cid)
	if err != nil {
		return mreq, err
	}

	var pfs []*proofspb.Proof
	pfs = append(pfs, docProofs.FieldProofs...)
	for _, p := range nftProofs {
		p := p
		pfs = append(pfs, &p)
	}

	toAddress := common.HexToAddress(req.DepositAddress)

	anchorID, err := anchors.ToAnchorID(corDoc.CurrentVersion)
	if err != nil {
		return mreq, err
	}

	rootHash, err := anchors.ToDocumentRoot(corDoc.DocumentRoot)
	if err != nil {
		return mreq, err
	}

	requestData, err := NewMintRequest(tokenID, toAddress, anchorID, pfs, rootHash)
	if err != nil {
		return mreq, err
	}

	return requestData, nil

}

// generateNFTProofs generates required NFT uniqueness proofs
func generateNFTProofs(cd *coredocumentpb.CoreDocument, req MintNFTRequest, tokenID TokenID, cid identity.CentID) ([]proofspb.Proof, error) {
	cdTree, err := coredocument.GetCoreDocTree(cd)
	if err != nil {
		return nil, err
	}

	var pfs []proofspb.Proof
	registry := common.HexToAddress(req.RegistryAddress)
	if req.SubmitTokenProof {
		pf, err := createTokenProof(cd, cdTree, registry)
		if err != nil {
			return nil, err
		}

		pfs = append(pfs, pf)
	}

	if req.GrantNFTReadAccess && req.SubmitNFTReadAccessProof {
		pf, err := createNFTReadAccessProof(cd, cdTree, registry, tokenID[:])
		if err != nil {
			return nil, err
		}

		pfs = append(pfs, pf...)
	}

	if req.SubmitRoleProof != "" {
		pf, err := createRoleProof(cd, cdTree, req.SubmitRoleProof, cid)
		if err != nil {
			return nil, err
		}

		pfs = append(pfs, pf)
	}

	return pfs, nil
}

func createRoleProof(cd *coredocumentpb.CoreDocument, cdTree *proofs.DocumentTree, roleKey string, cid identity.CentID) (proof proofspb.Proof, err error) {
	idx, role := getRoleForAccount(cd, roleKey, cid)
	if role == nil {
		return proof, ErrNFTRoleMissing
	}

	pk := fmt.Sprintf("roles[%s].collaborators[%d]", hexutil.Encode(role.RoleKey), idx)
	return cdTree.CreateProof(pk)
}

// createNFTReadAccessProof creates proof that nft exists in the role
func createNFTReadAccessProof(cd *coredocumentpb.CoreDocument, cdTree *proofs.DocumentTree, registry common.Address, tokenID []byte) (proofs []proofspb.Proof, err error) {

	var rridx, ridx, nftidx int
	var rk []byte
	found := coredocument.FindRole(cd, coredocumentpb.Action_ACTION_READ, func(i, j int, role *coredocumentpb.Role) bool {
		si, found := coredocument.IsNFTInRole(role, registry, tokenID)
		if found {
			rridx = i
			ridx = j
			rk = role.RoleKey
			nftidx = si
		}

		return found
	})

	if !found {
		return nil, ErrNFTRoleMissing
	}

	fields := []string{
		fmt.Sprintf("read_rules[%d].roles[%d]", rridx, ridx),          // proof that a read rule exists with the nft role
		fmt.Sprintf("roles[%s].nfts[%d]", hexutil.Encode(rk), nftidx), // proof that role with nft exists
		fmt.Sprintf("read_rules[%d].action", rridx),                   // proof that this read rule has read access
	}

	for _, f := range fields {
		pf, err := cdTree.CreateProof(f)
		if err != nil {
			return nil, err
		}
		proofs = append(proofs, pf)
	}

	return proofs, nil
}

func createTokenProof(cd *coredocumentpb.CoreDocument, cdTree *proofs.DocumentTree, registry common.Address) (proof proofspb.Proof, err error) {
	nft := getStoredNFT(cd.Nfts, registry.Bytes())
	if nft == nil {
		return proof, errors.New("nft is missing from the p2p document")
	}

	key := hexutil.Encode(nft.RegistryId)
	pk := fmt.Sprintf("nfts[%s]", key)
	return cdTree.CreateProof(pk)
}

// MintNFT mints an NFT
func (s *ethereumPaymentObligation) MintNFT(ctx context.Context, request MintNFTRequest) (*MintNFTResponse, chan bool, error) {
	tc, err := contextutil.Account(ctx)
	if err != nil {
		return nil, nil, err
	}

	cidBytes, err := tc.GetIdentityID()
	if err != nil {
		return nil, nil, err
	}

	cid, err := identity.ToCentID(cidBytes)
	if err != nil {
		return nil, nil, err
	}

	tokenID := NewTokenID()
	model, err := s.docSrv.GetCurrentVersion(ctx, request.DocumentID)
	if err != nil {
		return nil, nil, err
	}

	dm, err := model.PackCoreDocument()
	if err != nil {
		return nil, nil, err
	}

	cd := dm.Document
	// if the role_proof is given check if the role exists with account id as collaborator
	// at the moment, since we generate role key randomly, this will not work
	// but once we figure out acls on the client API, this will continue working with out any chnages
	if request.SubmitRoleProof != "" {
		_, role := getRoleForAccount(cd, request.SubmitRoleProof, cid)
		if role == nil {
			return nil, nil, ErrNFTRoleMissing
		}
	}

	if !request.GrantNFTReadAccess && request.SubmitNFTReadAccessProof {
		return nil, nil, errors.New("enabled grant nft access to generate NFT read access proof")
	}

	registry := common.HexToAddress(request.RegistryAddress)
	mt := getStoredNFT(cd.Nfts, registry.Bytes())
	// check if the nft is successfully minted
	if mt != nil && s.isNFTMinted(registry, mt.TokenId) {
		return nil, nil, errors.NewTypedError(ErrNFTMinted, errors.New("registry %v", registry.String()))
	}

	// Mint NFT within transaction
	// We use context.Background() for now so that the transaction is only limited by ethereum timeouts
	txID, done, err := s.txManager.ExecuteWithinTX(context.Background(), cid, uuid.Nil, "Minting NFT",
		s.minter(ctx, tokenID, model, request))
	if err != nil {
		return nil, nil, err
	}

	return &MintNFTResponse{
		TransactionID: txID.String(),
		TokenID:       tokenID.String(),
	}, done, nil
}

func (s *ethereumPaymentObligation) isNFTMinted(registry common.Address, tokenID []byte) bool {
	// since OwnerOf throws when owner is zero address,
	// if err is not thrown, we can assume that NFT is minted
	_, err := s.OwnerOf(registry, tokenID)
	return err == nil
}

func (s *ethereumPaymentObligation) minter(ctx context.Context, tokenID TokenID, model documents.Model, req MintNFTRequest) func(accountID identity.CentID, txID uuid.UUID, txMan transactions.Manager, errOut chan<- error) {
	return func(accountID identity.CentID, txID uuid.UUID, txMan transactions.Manager, errOut chan<- error) {
		tc, err := contextutil.Account(ctx)
		if err != nil {
			errOut <- err
			return
		}

		dm, err := model.PackCoreDocument()
		if err != nil {
			errOut <- err
			return
		}

		data := dm.Document.EmbeddedData
		newDM, err := dm.PrepareNewVersion(nil)
		if err != nil {
			errOut <- err
			return
		}

		cd := newDM.Document
		cd.EmbeddedData = data
		registry := common.HexToAddress(req.RegistryAddress)
		addNFT(cd, registry.Bytes(), tokenID[:])
		if req.GrantNFTReadAccess {
			err = coredocument.AddNFTToReadRules(cd, registry, tokenID.BigInt().Bytes())
			if err != nil {
				errOut <- err
				return
			}
		}

		err = coredocument.FillSalts(cd)
		if err != nil {
			errOut <- err
			return
		}

		model, err = s.docSrv.DeriveFromCoreDocumentModel(newDM)
		if err != nil {
			errOut <- err
			return
		}

		_, _, done, err := s.docSrv.Update(contextutil.WithTX(ctx, txID), model)
		if err != nil {
			errOut <- err
			return
		}

		isDone := <-done
		if !isDone {
			// some problem occured in a child task
			errOut <- errors.New("update document failed for document %s and transaction %s", hexutil.Encode(newCD.DocumentIdentifier), txID)
			return
		}

		requestData, err := s.prepareMintRequest(ctx, tokenID, accountID, req)
		if err != nil {
			errOut <- errors.New("failed to prepare mint request: %v", err)
			return
		}

		opts, err := s.ethClient.GetTxOpts(tc.GetEthereumDefaultAccountName())
		if err != nil {
			errOut <- err
			return
		}

		contract, err := s.bindContract(registry, s.ethClient)
		if err != nil {
			errOut <- err
			return
		}

		ethTX, err := s.ethClient.SubmitTransactionWithRetries(contract.Mint, opts, requestData.To, requestData.TokenID, requestData.TokenURI, requestData.AnchorID,
			requestData.MerkleRoot, requestData.Values, requestData.Salts, requestData.Proofs)
		if err != nil {
			errOut <- err
			return
		}

		log.Infof("Sent off ethTX to mint [tokenID: %s, anchor: %x, registry: %s] to payment obligation contract. Ethereum transaction hash [%s] and Nonce [%d] and Check [%v]",
			requestData.TokenID, requestData.AnchorID, requestData.To.String(), ethTX.Hash().String(), ethTX.Nonce(), ethTX.CheckNonce())
		log.Infof("Transfer pending: %s\n", ethTX.Hash().String())

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

func getStoredNFT(nfts []*coredocumentpb.NFT, registry []byte) *coredocumentpb.NFT {
	for _, nft := range nfts {
		if bytes.Equal(nft.RegistryId[:20], registry) {
			return nft
		}
	}

	return nil
}

// getRoleForAccount returns the matching role and index of the account
func getRoleForAccount(cd *coredocumentpb.CoreDocument, roleName string, id identity.CentID) (idx int, role *coredocumentpb.Role) {
	sha := sha256.New()
	sha.Write([]byte(roleName))
	role, err := coredocument.GetRole(sha.Sum(nil), cd.Roles)
	if err != nil {
		return idx, nil
	}

	idx, found := coredocument.IsAccountInRole(role, id)
	if !found {
		return 0, nil
	}

	return idx, role
}

// addNFT adds/replaces the NFT
// Note: this is replace operation. Ensure existing token is not minted
func addNFT(dm *documents.CoreDocumentModel, registry, tokenID []byte) {
	cd := dm.Document
	nft := getStoredNFT(cd.Nfts, registry)
	if nft == nil {
		nft = new(coredocumentpb.NFT)
		// add 12 empty bytes
		eb := make([]byte, 12, 12)
		nft.RegistryId = append(registry, eb...)
		cd.Nfts = append(cd.Nfts, nft)
	}

	nft.TokenId = tokenID
}

// OwnerOf returns the owner of the NFT token on ethereum chain
func (s *ethereumPaymentObligation) OwnerOf(registry common.Address, tokenID []byte) (owner common.Address, err error) {
	contract, err := s.bindContract(registry, s.ethClient)
	if err != nil {
		return owner, errors.New("failed to bind the registry contract: %v", err)
	}

	opts, cancF := s.ethClient.GetGethCallOpts(false)
	defer cancF()

	return contract.OwnerOf(opts, utils.ByteSliceToBigInt(tokenID))
}

// MintRequest holds the data needed to mint and NFT from a Centrifuge document
type MintRequest struct {

	// To is the address of the recipient of the minted token
	To common.Address

	// TokenID is the ID for the minted token
	TokenID *big.Int

	// TokenURI is the metadata uri
	TokenURI string

	// AnchorID is the ID of the document as identified by the set up anchorRepository.
	AnchorID *big.Int

	// MerkleRoot is the root hash of the merkle proof/doc
	MerkleRoot [32]byte

	// Values are the values of the leafs that is being proved Will be converted to string and concatenated for proof verification as outlined in precise-proofs library.
	Values [][]byte

	// salts are the salts for the field that is being proved Will be concatenated for proof verification as outlined in precise-proofs library.
	Salts [][32]byte

	// Proofs are the documents proofs that are needed
	Proofs [][][32]byte
}

// NewMintRequest converts the parameters and returns a struct with needed parameter for minting
func NewMintRequest(tokenID TokenID, to common.Address, anchorID anchors.AnchorID, proofs []*proofspb.Proof, rootHash [32]byte) (MintRequest, error) {
	proofData, err := createProofData(proofs)
	if err != nil {
		return MintRequest{}, err
	}

	return MintRequest{
		To:         to,
		TokenID:    tokenID.BigInt(),
		TokenURI:   tokenID.URI(),
		AnchorID:   anchorID.BigInt(),
		MerkleRoot: rootHash,
		Values:     proofData.Values,
		Salts:      proofData.Salts,
		Proofs:     proofData.Proofs}, nil
}

type proofData struct {
	Values [][]byte
	Salts  [][32]byte
	Proofs [][][32]byte
}

func createProofData(proofspb []*proofspb.Proof) (*proofData, error) {
	var values = make([][]byte, len(proofspb))
	var salts = make([][32]byte, len(proofspb))
	var proofs = make([][][32]byte, len(proofspb))

	for i, p := range proofspb {
		values[i] = p.Value
		salt32, err := utils.SliceToByte32(p.Salt)
		if err != nil {
			return nil, err
		}

		salts[i] = salt32
		property, err := convertProofProperty(p.SortedHashes)
		if err != nil {
			return nil, err
		}
		proofs[i] = property
	}

	return &proofData{Values: values, Salts: salts, Proofs: proofs}, nil
}

func convertProofProperty(sortedHashes [][]byte) ([][32]byte, error) {
	var property [][32]byte
	for _, hash := range sortedHashes {
		hash32, err := utils.SliceToByte32(hash)
		if err != nil {
			return nil, err
		}
		property = append(property, hash32)
	}

	return property, nil
}

func bindContract(address common.Address, client ethereum.Client) (*EthereumPaymentObligationContract, error) {
	return NewEthereumPaymentObligationContract(address, client.GetEthClient())
}

package v3

import (
	"context"
	"errors"
	"fmt"

	"github.com/centrifuge/pod/pallets/utility"

	"github.com/centrifuge/pod/centchain"

	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/gocelery/v2"
	"github.com/centrifuge/pod/config"
	"github.com/centrifuge/pod/contextutil"
	"github.com/centrifuge/pod/documents"
	"github.com/centrifuge/pod/ipfs"
	"github.com/centrifuge/pod/jobs"
	"github.com/centrifuge/pod/pallets/uniques"
	"github.com/centrifuge/pod/pending"
	logging "github.com/ipfs/go-log"
	"github.com/ipfs/interface-go-ipfs-core/path"
)

const (
	commitAndMintNFTV3Job    = "Commit and mint NFT V3 Job"
	mintNFTV3Job             = "Mint NFT V3 Job"
	createNFTCollectionV3Job = "Create NFT collection V3 Job"
)

var (
	log = logging.Logger("nft_v3_dispatcher")
)

type CreateCollectionJobRunner struct {
	jobs.Base

	accountsSrv config.Service
	docSrv      documents.Service
	dispatcher  jobs.Dispatcher
	api         uniques.API
}

// New returns a new instance of CreateCollectionJobRunner
func (c *CreateCollectionJobRunner) New() gocelery.Runner {
	cj := &CreateCollectionJobRunner{
		accountsSrv: c.accountsSrv,
		docSrv:      c.docSrv,
		dispatcher:  c.dispatcher,
		api:         c.api,
	}

	cj.Base = jobs.NewBase(cj.loadTasks())
	return cj
}

func (c *CreateCollectionJobRunner) convertArgs(
	args []interface{},
) (
	config.Account,
	types.U64,
	error,
) {
	account, ok := args[0].(config.Account)

	if !ok {
		return nil, 0, errors.New("account not provided in args")
	}

	collectionID, ok := args[1].(types.U64)

	if !ok {
		return nil, 0, errors.New("collection ID not provided in args")
	}

	return account, collectionID, nil
}

func (c *CreateCollectionJobRunner) loadTasks() map[string]jobs.Task {
	return map[string]jobs.Task{
		"create_nft_collection_v3": {
			RunnerFunc: func(args []interface{}, overrides map[string]interface{}) (result interface{}, err error) {
				account, collectionID, err := c.convertArgs(args)

				if err != nil {
					log.Errorf("Couldn't convert args: %s", err)

					return nil, err
				}

				ctx := contextutil.WithAccount(context.Background(), account)

				extInfo, err := c.api.CreateCollection(ctx, collectionID)

				if err != nil {
					log.Errorf("Couldn't create collection: %s", err)

					return nil, err
				}

				log.Infof("Successfully created collection with ID - %d, ext hash - %s", collectionID, extInfo.Hash.Hex())

				overrides["ext_info"] = extInfo

				return nil, nil
			},
		},
	}
}

type CommitAndMintNFTJobRunner struct {
	jobs.Base

	accountsSrv    config.Service
	pendingDocsSrv pending.Service
	pendingRepo    pending.Repository
	docSrv         documents.Service
	dispatcher     jobs.Dispatcher
	utilityAPI     utility.API
	ipfsPinningSrv ipfs.PinningServiceClient
}

// New returns a new instance of MintNFTJobRunner
func (c *CommitAndMintNFTJobRunner) New() gocelery.Runner {
	mj := &CommitAndMintNFTJobRunner{
		accountsSrv:    c.accountsSrv,
		pendingDocsSrv: c.pendingDocsSrv,
		pendingRepo:    c.pendingRepo,
		docSrv:         c.docSrv,
		dispatcher:     c.dispatcher,
		utilityAPI:     c.utilityAPI,
		ipfsPinningSrv: c.ipfsPinningSrv,
	}

	commitAndMintNFTTasks := mergeTaskMaps(
		loadAnchoringTasksForPendingDocument(c.pendingDocsSrv, c.pendingRepo, c.dispatcher),
		loadNFTMintTasks(c.docSrv, c.utilityAPI, c.ipfsPinningSrv),
	)

	mj.Base = jobs.NewBase(commitAndMintNFTTasks)

	return mj
}

func mergeTaskMaps[K comparable, V any](taskMaps ...map[K]V) map[K]V {
	res := make(map[K]V)

	for _, taskMap := range taskMaps {
		for k, v := range taskMap {
			res[k] = v
		}
	}

	return res
}

type MintNFTJobRunner struct {
	jobs.Base

	accountsSrv    config.Service
	docSrv         documents.Service
	dispatcher     jobs.Dispatcher
	utilityAPI     utility.API
	ipfsPinningSrv ipfs.PinningServiceClient
}

// New returns a new instance of MintNFTJobRunner
func (m *MintNFTJobRunner) New() gocelery.Runner {
	mj := &MintNFTJobRunner{
		accountsSrv:    m.accountsSrv,
		docSrv:         m.docSrv,
		dispatcher:     m.dispatcher,
		utilityAPI:     m.utilityAPI,
		ipfsPinningSrv: m.ipfsPinningSrv,
	}

	nftMintTasks := mergeTaskMaps(
		loadAnchoringTasksForCommittedDocument(m.docSrv, m.dispatcher),
		loadNFTMintTasks(m.docSrv, m.utilityAPI, m.ipfsPinningSrv),
	)

	mj.Base = jobs.NewBase(nftMintTasks)
	return mj
}

func convertArgs(
	args []interface{},
) (
	config.Account,
	types.U128,
	*MintNFTRequest,
	error,
) {
	account, ok := args[0].(config.Account)

	if !ok {
		return nil, types.U128{}, nil, errors.New("account not provided in args")
	}

	itemID, ok := args[1].(types.U128)

	if !ok {
		return nil, types.U128{}, nil, errors.New("item ID not provided in args")
	}

	req, ok := args[2].(*MintNFTRequest)

	if !ok {
		return nil, types.U128{}, nil, errors.New("request not provided in args")
	}

	return account, itemID, req, nil
}

const (
	DocumentIDAttributeKey      = "document_id"
	DocumentVersionAttributeKey = "document_version"
)

func loadAnchoringTasksForCommittedDocument(
	docSrv documents.Service,
	dispatcher jobs.Dispatcher,
) map[string]jobs.Task {
	return map[string]jobs.Task{
		"add_nft_v3_to_document": {
			RunnerFunc: func(args []interface{}, overrides map[string]interface{}) (result interface{}, err error) {
				account, itemID, req, err := convertArgs(args)

				if err != nil {
					log.Errorf("Couldn't convert args: %s", err)

					return nil, err
				}

				ctx := contextutil.WithAccount(context.Background(), account)

				doc, err := docSrv.GetCurrentVersion(ctx, req.DocumentID)

				if err != nil {
					log.Errorf("Couldn't get document: %s", err)

					return nil, err
				}

				err = doc.AddNFT(req.GrantReadAccess, req.CollectionID, itemID)

				if err != nil {
					log.Errorf("Couldn't add NFT to document: %s", err)

					return nil, fmt.Errorf("failed to add nft to document: %w", err)
				}

				jobID, err := docSrv.Commit(ctx, doc)

				if err != nil {
					log.Errorf("Couldn't commit document: %s", err)

					return nil, fmt.Errorf("failed to commit document: %w", err)
				}

				overrides["document_commit_job"] = jobID

				return nil, nil
			},
			Next: "wait_for_document_commit",
		},
		"wait_for_document_commit": {
			RunnerFunc: func(args []interface{}, overrides map[string]interface{}) (result interface{}, err error) {
				account, ok := args[0].(config.Account)

				if !ok {
					return nil, errors.New("account not provided in args")
				}

				jobID := overrides["document_commit_job"].(gocelery.JobID)

				log.Info("Waiting for document to be anchored")

				job, err := dispatcher.Job(account.GetIdentity(), jobID)

				if err != nil {
					log.Errorf("Couldn't dispatch job: %s", err)

					return nil, fmt.Errorf("failed to fetch job: %w", err)
				}

				if !job.IsSuccessful() {
					log.Info("Document not committed yet")

					return nil, errors.New("document not committed yet")
				}

				return nil, nil
			},
			Next: "store_nft_on_ipfs",
		},
	}
}

func loadAnchoringTasksForPendingDocument(
	pendingDocsSrv pending.Service,
	pendingRepo pending.Repository,
	dispatcher jobs.Dispatcher,
) map[string]jobs.Task {
	return map[string]jobs.Task{
		"add_nft_v3_to_pending_document": {
			RunnerFunc: func(args []interface{}, overrides map[string]interface{}) (result interface{}, err error) {
				account, itemID, req, err := convertArgs(args)

				if err != nil {
					log.Errorf("Couldn't convert args: %s", err)

					return nil, err
				}

				ctx := contextutil.WithAccount(context.Background(), account)

				doc, err := pendingDocsSrv.Get(ctx, req.DocumentID, documents.Pending)

				if err != nil {
					log.Errorf("Couldn't get pending document: %s", err)

					return nil, err
				}

				err = doc.AddNFT(req.GrantReadAccess, req.CollectionID, itemID)

				if err != nil {
					log.Errorf("Couldn't add NFT to document: %s", err)

					return nil, fmt.Errorf("failed to add nft to document: %w", err)
				}

				if err = pendingRepo.Update(account.GetIdentity().ToBytes(), req.DocumentID, doc); err != nil {
					log.Errorf("Couldn't update document: %s", err)

					return nil, fmt.Errorf("couldn't update document: %w", err)
				}

				_, jobID, err := pendingDocsSrv.Commit(ctx, req.DocumentID)

				if err != nil {
					log.Errorf("Couldn't commit pending document: %s", err)

					return nil, fmt.Errorf("failed to commit document: %w", err)
				}

				overrides["document_commit_job"] = jobID

				return nil, nil
			},
			Next: "wait_for_pending_document_commit",
		},
		"wait_for_pending_document_commit": {
			RunnerFunc: func(args []interface{}, overrides map[string]interface{}) (result interface{}, err error) {
				account, ok := args[0].(config.Account)

				if !ok {
					return nil, errors.New("account not provided in args")
				}

				jobID := overrides["document_commit_job"].(gocelery.JobID)

				log.Info("Waiting for pending document to be anchored")

				job, err := dispatcher.Job(account.GetIdentity(), jobID)

				if err != nil {
					log.Errorf("Couldn't dispatch job: %s", err)

					return nil, fmt.Errorf("failed to fetch job: %w", err)
				}

				if !job.IsSuccessful() {
					log.Info("Document not committed yet")

					return nil, errors.New("document not committed yet")
				}

				return nil, nil
			},
			Next: "store_nft_on_ipfs",
		},
	}
}

func loadNFTMintTasks(
	docSrv documents.Service,
	utilityAPI utility.API,
	ipfsPinningSrv ipfs.PinningServiceClient,
) map[string]jobs.Task {
	return map[string]jobs.Task{
		"store_nft_on_ipfs": {
			RunnerFunc: func(args []interface{}, overrides map[string]interface{}) (result interface{}, err error) {
				account, _, req, err := convertArgs(args)

				if err != nil {
					return nil, err
				}

				ctx := contextutil.WithAccount(context.Background(), account)

				doc, err := docSrv.GetCurrentVersion(ctx, req.DocumentID)

				if err != nil {
					log.Errorf("Couldn't get document: %s", err)

					return nil, fmt.Errorf("failed to get document: %w", err)
				}

				docAttributes, err := GetDocAttributes(doc, req.IPFSMetadata.DocumentAttributeKeys)

				if err != nil {
					log.Errorf("Couldn't get doc attributes: %s", err)

					return nil, fmt.Errorf("couldn't get doc attributes: %w", err)
				}

				nftMetadata := ipfs.NFTMetadata{
					Name:        req.IPFSMetadata.Name,
					Description: req.IPFSMetadata.Description,
					Image:       req.IPFSMetadata.Image,
					Properties:  docAttributes,
				}

				log.Info("Storing NFT metadata in IPFS")

				ipfsPinningRes, err := ipfsPinningSrv.PinData(ctx, &ipfs.PinRequest{
					CIDVersion: 1,
					Data:       nftMetadata,
				})

				if err != nil {
					log.Errorf("Couldn't store NFT metadata in IPFS: %s", err)

					return nil, err
				}

				ipfsPath := path.New(ipfsPinningRes.CID).String()

				overrides["ipfsPath"] = ipfsPath

				return nil, nil
			},
			Next: "execute_nft_batch",
		},
		"execute_nft_batch": {
			RunnerFunc: func(args []interface{}, overrides map[string]interface{}) (result interface{}, err error) {
				account, itemID, req, err := convertArgs(args)

				if err != nil {
					return nil, err
				}

				ipfsPath, ok := overrides["ipfsPath"].(string)

				if !ok {
					return nil, errors.New("invalid IPFS path detected")
				}

				ctx := contextutil.WithAccount(context.Background(), account)

				doc, err := docSrv.GetCurrentVersion(ctx, req.DocumentID)

				if err != nil {
					log.Errorf("Couldn't get current document version: %s", err)

					return nil, err
				}

				_, err = utilityAPI.BatchAll(ctx,
					getMintNFTCallProviderFn(req.CollectionID, itemID, req.Owner),
					getSetMetadataCallProviderFn(req.CollectionID, itemID, []byte(ipfsPath), false),
					getSetAttributeCallProviderFn(req.CollectionID, itemID, []byte(DocumentIDAttributeKey), doc.ID()),
					getSetAttributeCallProviderFn(req.CollectionID, itemID, []byte(DocumentVersionAttributeKey), doc.CurrentVersion()),
				)

				if err != nil {
					log.Errorf("Couldn't execute NFT batch: %s", err)

					return nil, err
				}

				return nil, nil
			},
		},
	}
}

func getMintNFTCallProviderFn(
	collectionID types.U64,
	itemID types.U128,
	owner *types.AccountID,
) centchain.CallProviderFn {
	return func(meta *types.Metadata) (*types.Call, error) {
		ownerMultiAddress, err := types.NewMultiAddressFromAccountID(owner.ToBytes())

		if err != nil {
			return nil, fmt.Errorf("couldn't create owner multi address: %w", err)
		}

		call, err := types.NewCall(
			meta,
			uniques.MintCall,
			collectionID,
			itemID,
			ownerMultiAddress,
		)

		if err != nil {
			return nil, fmt.Errorf("couldn't create MintNFT call: %w", err)
		}

		return &call, nil
	}
}

func getSetMetadataCallProviderFn(
	collectionID types.U64,
	itemID types.U128,
	ipfsPath []byte,
	freezeMetadata bool,
) centchain.CallProviderFn {
	return func(meta *types.Metadata) (*types.Call, error) {
		call, err := types.NewCall(
			meta,
			uniques.SetMetadataCall,
			collectionID,
			itemID,
			ipfsPath,
			freezeMetadata,
		)

		if err != nil {
			return nil, fmt.Errorf("couldn't create SetMetadata call: %w", err)
		}

		return &call, nil
	}
}

func getSetAttributeCallProviderFn(
	collectionID types.U64,
	itemID types.U128,
	attributeKey []byte,
	attributeValue []byte,
) centchain.CallProviderFn {
	return func(meta *types.Metadata) (*types.Call, error) {
		call, err := types.NewCall(
			meta,
			uniques.SetAttributeCall,
			collectionID,
			types.NewOption(itemID),
			attributeKey,
			attributeValue,
		)

		if err != nil {
			return nil, fmt.Errorf("couldn't create SetAttribute call: %w", err)
		}

		return &call, nil
	}
}

func GetDocAttributes(doc documents.Document, attrLabels []string) (map[string]string, error) {
	attrMap := make(map[string]string)

	for _, attrLabel := range attrLabels {
		attrKey, err := documents.AttrKeyFromLabel(attrLabel)

		if err != nil {
			return nil, fmt.Errorf("couldn't create attribute key: %w", err)
		}

		attr, err := doc.GetAttribute(attrKey)

		if err != nil {
			return nil, fmt.Errorf("couldn't get document attribute: %w", err)
		}

		valStr, err := attr.Value.String()

		if err != nil {
			return nil, fmt.Errorf("couldn't get attribute as string: %w", err)
		}

		if attr.Value.Type == documents.AttrMonetary {
			valStr = attr.Value.Monetary.Value.String()
		}

		attrMap[attr.KeyLabel] = valStr
	}

	return attrMap, nil
}

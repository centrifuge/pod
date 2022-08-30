package v3

import (
	"context"
	"errors"
	"fmt"

	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/ipfs_pinning"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/nft/v3/uniques"
	"github.com/centrifuge/go-centrifuge/pending"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/gocelery/v2"
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
	docSrv         documents.Service
	dispatcher     jobs.Dispatcher
	api            uniques.API
	ipfsPinningSrv ipfs_pinning.PinningServiceClient
}

// New returns a new instance of MintNFTJobRunner
func (c *CommitAndMintNFTJobRunner) New() gocelery.Runner {
	mj := &CommitAndMintNFTJobRunner{
		accountsSrv:    c.accountsSrv,
		pendingDocsSrv: c.pendingDocsSrv,
		docSrv:         c.docSrv,
		dispatcher:     c.dispatcher,
		api:            c.api,
		ipfsPinningSrv: c.ipfsPinningSrv,
	}

	documentPendingTasks := loadCommitAndMintTasks(
		c.pendingDocsSrv,
		c.docSrv,
		c.dispatcher,
		c.api,
		c.ipfsPinningSrv,
	)

	mj.Base = jobs.NewBase(documentPendingTasks)
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

func loadCommitAndMintTasks(
	pendingDocsSrv pending.Service,
	docSrv documents.Service,
	dispatcher jobs.Dispatcher,
	api uniques.API,
	ipfsPinningSrv ipfs_pinning.PinningServiceClient,
) map[string]jobs.Task {
	commitTasks := map[string]jobs.Task{
		"commit_pending_document": {
			RunnerFunc: func(args []interface{}, overrides map[string]interface{}) (interface{}, error) {
				account, _, req, err := convertArgs(args)

				if err != nil {
					log.Errorf("Couldn't convert args: %s", err)

					return nil, err
				}

				ctx := contextutil.WithAccount(context.Background(), account)

				_, jobID, err := pendingDocsSrv.Commit(ctx, req.DocumentID)

				if err != nil {
					log.Errorf("Couldn't commit pending document: %s", err)

					return nil, err
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
			Next: "add_nft_v3_to_document",
		},
	}

	return mergeTaskMaps(commitTasks, loadNFTMintTasks(docSrv, dispatcher, api, ipfsPinningSrv))
}

type MintNFTJobRunner struct {
	jobs.Base

	accountsSrv    config.Service
	docSrv         documents.Service
	dispatcher     jobs.Dispatcher
	api            uniques.API
	ipfsPinningSrv ipfs_pinning.PinningServiceClient
}

// New returns a new instance of MintNFTJobRunner
func (m *MintNFTJobRunner) New() gocelery.Runner {
	mj := &MintNFTJobRunner{
		accountsSrv:    m.accountsSrv,
		docSrv:         m.docSrv,
		dispatcher:     m.dispatcher,
		api:            m.api,
		ipfsPinningSrv: m.ipfsPinningSrv,
	}

	nftMintTasks := loadNFTMintTasks(m.docSrv, m.dispatcher, m.api, m.ipfsPinningSrv)

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

func loadNFTMintTasks(
	docSrv documents.Service,
	dispatcher jobs.Dispatcher,
	api uniques.API,
	ipfsPinningSrv ipfs_pinning.PinningServiceClient,
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

				err = doc.AddNFT(req.CollectionID, itemID)

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
			Next: "mint_nft_v3",
		},
		"mint_nft_v3": {
			RunnerFunc: func(args []interface{}, overrides map[string]interface{}) (result interface{}, err error) {
				account, itemID, req, err := convertArgs(args)

				if err != nil {
					return nil, err
				}

				log.Info("Minting NFT on Centrifuge chain...")

				ctx := contextutil.WithAccount(context.Background(), account)

				doc, err := docSrv.GetCurrentVersion(ctx, req.DocumentID)

				if err != nil {
					log.Errorf("Couldn't get current document version: %s", err)

					return nil, err
				}

				anchorID, err := anchors.ToAnchorID(doc.CurrentVersion())

				if err != nil {
					log.Errorf("Couldn't get anchor for document: %s", err)

					return nil, err
				}

				extInfo, err := api.Mint(ctx, req.CollectionID, itemID, req.Owner)

				if err != nil {
					log.Errorf("Couldn't mint item: %s", err)

					return nil, err
				}

				log.Infof(
					"Successfully minted NFT on Centrifuge chain, collection ID - %d, item ID - %d, anchor ID - %s, ext hash - %s",
					req.CollectionID,
					itemID,
					anchorID.String(),
					extInfo.Hash.Hex(),
				)

				return nil, nil
			},
			Next: "store_nft_v3_metadata",
		},
		"store_nft_v3_metadata": {
			RunnerFunc: func(args []interface{}, overrides map[string]interface{}) (result interface{}, err error) {
				account, itemID, req, err := convertArgs(args)

				if err != nil {
					return nil, err
				}

				ctx := contextutil.WithAccount(context.Background(), account)

				doc, err := docSrv.GetCurrentVersion(ctx, req.DocumentID)

				if err != nil {
					log.Errorf("Couldn't get document: %s", err)

					return nil, fmt.Errorf("failed to get document: %w", err)
				}

				docAttributes, err := getDocAttributes(doc, req.IPFSMetadata.DocumentAttributeKeys)

				if err != nil {
					log.Errorf("Couldn't get doc attributes: %s", err)

					return nil, fmt.Errorf("couldn't get doc attributes: %w", err)
				}

				nftMetadata := NFTMetadata{
					Name:        req.IPFSMetadata.Name,
					Description: req.IPFSMetadata.Description,
					Image:       req.IPFSMetadata.Image,
					Properties:  docAttributes,
				}

				log.Info("Storing NFT metadata in IPFS")

				ipfsPinningRes, err := ipfsPinningSrv.PinData(ctx, &ipfs_pinning.PinRequest{
					CIDVersion: 1,
					Data:       nftMetadata,
				})

				if err != nil {
					log.Errorf("Couldn't store NFT metadata in IPFS: %s", err)

					return nil, err
				}

				ipfsPath := path.New(ipfsPinningRes.CID).String()

				log.Infof("Setting the IPFS path as NFT metadata in Centrifuge chain, IPFS path - %s", ipfsPath)

				_, err = api.SetMetadata(ctx, req.CollectionID, itemID, []byte(ipfsPath), req.FreezeMetadata)

				if err != nil {
					log.Errorf("Couldn't set IPFS CID: %s", err)

					return nil, err
				}

				log.Infof(
					"Successfully stored NFT metadata, collection ID - %d, item ID - %d, IPFS path - %s",
					req.CollectionID,
					itemID,
					ipfsPath,
				)

				return nil, nil
			},
			Next: "set_nft_v3_attributes",
		},
		"set_nft_v3_attributes": {
			RunnerFunc: func(args []interface{}, overrides map[string]interface{}) (result interface{}, err error) {
				account, itemID, req, err := convertArgs(args)

				if err != nil {
					return nil, err
				}

				ctx := contextutil.WithAccount(context.Background(), account)

				doc, err := docSrv.GetCurrentVersion(ctx, req.DocumentID)

				if err != nil {
					log.Errorf("Couldn't get document: %s", err)

					return nil, fmt.Errorf("failed to get document: %w", err)
				}

				_, err = api.SetAttribute(ctx, req.CollectionID, itemID, []byte(DocumentIDAttributeKey), doc.ID())

				if err != nil {
					log.Errorf("Couldn't set document ID attribute: %s", err)

					return nil, fmt.Errorf("couldn't set document ID attribute")
				}

				_, err = api.SetAttribute(ctx, req.CollectionID, itemID, []byte(DocumentVersionAttributeKey), doc.CurrentVersion())

				if err != nil {
					log.Errorf("Couldn't set document version attribute: %s", err)

					return nil, fmt.Errorf("couldn't set document version attribute")
				}

				return nil, nil
			},
		},
	}
}

func getDocAttributes(doc documents.Document, attrLabels []string) (map[string]string, error) {
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

		attrMap[attr.KeyLabel] = valStr
	}

	return attrMap, nil
}

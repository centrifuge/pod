package v3

import (
	"context"
	"errors"
	"fmt"

	"github.com/centrifuge/go-centrifuge/nft/v3/uniques"

	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/ipfs_pinning"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/gocelery/v2"
	logging "github.com/ipfs/go-log"
	"github.com/ipfs/interface-go-ipfs-core/path"
)

const (
	mintNFTV3Job        = "Mint NFT V3 Job"
	createNFTClassV3Job = "Create NFT collection V3 Job"
)

type CreateCollectionJob struct {
	jobs.Base

	accountsSrv config.Service
	docSrv      documents.Service
	dispatcher  jobs.Dispatcher
	api         uniques.API
	log         *logging.ZapEventLogger
}

// New returns a new instance of CreateCollectionJob
func (c *CreateCollectionJob) New() gocelery.Runner {
	log := logging.Logger("create_nft_collection_v3_dispatcher")

	cj := &CreateCollectionJob{
		accountsSrv: c.accountsSrv,
		docSrv:      c.docSrv,
		dispatcher:  c.dispatcher,
		api:         c.api,
		log:         log,
	}

	cj.Base = jobs.NewBase(cj.loadTasks())
	return cj
}

func (c *CreateCollectionJob) convertArgs(
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

func (c *CreateCollectionJob) loadTasks() map[string]jobs.Task {
	return map[string]jobs.Task{
		"create_nft_collection_v3": {
			RunnerFunc: func(args []interface{}, overrides map[string]interface{}) (result interface{}, err error) {
				account, collectionID, err := c.convertArgs(args)

				if err != nil {
					c.log.Errorf("Couldn't convert args: %s", err)

					return nil, err
				}

				ctx := contextutil.WithAccount(context.Background(), account)

				extInfo, err := c.api.CreateCollection(ctx, collectionID)

				if err != nil {
					c.log.Errorf("Couldn't create collection: %s", err)

					return nil, err
				}

				c.log.Infof("Successfully created collection with ID - %d, ext hash - %s", collectionID, extInfo.Hash.Hex())

				overrides["ext_info"] = extInfo

				return nil, nil
			},
		},
	}
}

type MintNFTJob struct {
	jobs.Base

	accountsSrv    config.Service
	docSrv         documents.Service
	dispatcher     jobs.Dispatcher
	api            uniques.API
	ipfsPinningSrv ipfs_pinning.PinningServiceClient
	log            *logging.ZapEventLogger
}

// New returns a new instance of MintNFTJob
func (m *MintNFTJob) New() gocelery.Runner {
	log := logging.Logger("mint_nft_v3_dispatcher")

	mj := &MintNFTJob{
		accountsSrv:    m.accountsSrv,
		docSrv:         m.docSrv,
		dispatcher:     m.dispatcher,
		api:            m.api,
		ipfsPinningSrv: m.ipfsPinningSrv,
		log:            log,
	}

	mj.Base = jobs.NewBase(mj.loadTasks())
	return mj
}

func (m *MintNFTJob) convertArgs(
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
		return nil, types.U128{}, nil, errors.New("mint NFT request not provided in args")
	}

	return account, itemID, req, nil
}

func (m *MintNFTJob) loadTasks() map[string]jobs.Task {
	return map[string]jobs.Task{
		"add_nft_v3_to_document": {
			RunnerFunc: func(args []interface{}, overrides map[string]interface{}) (result interface{}, err error) {
				account, itemID, req, err := m.convertArgs(args)

				if err != nil {
					m.log.Errorf("Couldn't convert args: %s", err)

					return nil, err
				}

				ctx := contextutil.WithAccount(context.Background(), account)

				doc, err := m.docSrv.GetCurrentVersion(ctx, req.DocumentID)

				if err != nil {
					m.log.Errorf("Couldn't get document: %s", err)

					return nil, err
				}

				err = doc.AddNFT(req.CollectionID, itemID)

				if err != nil {
					m.log.Errorf("Couldn't add NFT to document: %s", err)

					return nil, fmt.Errorf("failed to add nft to document: %w", err)
				}

				jobID, err := m.docSrv.Commit(ctx, doc)

				if err != nil {
					m.log.Errorf("Couldn't commit document: %s", err)

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

				m.log.Info("Waiting for document to be anchored")

				job, err := m.dispatcher.Job(account.GetIdentity(), jobID)

				if err != nil {
					m.log.Errorf("Couldn't dispatch job: %s", err)

					return nil, fmt.Errorf("failed to fetch job: %w", err)
				}

				if !job.IsSuccessful() {
					m.log.Info("Document not committed yet")

					return nil, errors.New("document not committed yet")
				}

				return nil, nil
			},
			Next: "mint_nft_v3",
		},
		"mint_nft_v3": {
			RunnerFunc: func(args []interface{}, overrides map[string]interface{}) (result interface{}, err error) {
				account, itemID, req, err := m.convertArgs(args)

				if err != nil {
					return nil, err
				}

				m.log.Info("Minting NFT on Centrifuge chain...")

				ctx := contextutil.WithAccount(context.Background(), account)

				doc, err := m.docSrv.GetCurrentVersion(ctx, req.DocumentID)

				if err != nil {
					m.log.Errorf("Couldn't get current document version: %s", err)

					return nil, err
				}

				anchorID, err := anchors.ToAnchorID(doc.CurrentVersion())

				if err != nil {
					m.log.Errorf("Couldn't get anchor for document: %s", err)

					return nil, err
				}

				extInfo, err := m.api.Mint(ctx, req.CollectionID, itemID, req.Owner)

				if err != nil {
					m.log.Errorf("Couldn't mint instance: %s", err)

					return nil, err
				}

				m.log.Infof(
					"Successfully minted NFT on Centrifuge chain, collection ID - %d, instance ID - %d, anchor ID - %s, ext hash - %s",
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
				account, itemID, req, err := m.convertArgs(args)

				if err != nil {
					return nil, err
				}

				if len(req.DocAttributes) == 0 {
					m.log.Info("No document attributes provided")

					return nil, nil
				}

				ctx := contextutil.WithAccount(context.Background(), account)

				doc, err := m.docSrv.GetCurrentVersion(ctx, req.DocumentID)

				if err != nil {
					m.log.Errorf("Couldn't get document: %s", err)

					return nil, fmt.Errorf("failed to get document: %w", err)
				}

				docAttributes, err := getDocAttributes(doc, req.DocAttributes)

				if err != nil {
					m.log.Errorf("Couldn't get doc attributes: %s", err)

					return nil, fmt.Errorf("couldn't get doc attributes: %w", err)
				}

				nftMetadata := NFTMetadata{
					DocID:         doc.ID(),
					DocVersion:    doc.CurrentVersion(),
					DocAttributes: docAttributes,
				}

				m.log.Info("Storing NFT metadata in IPFS")

				ipfsPinningRes, err := m.ipfsPinningSrv.PinData(ctx, &ipfs_pinning.PinRequest{
					CIDVersion: 1,
					Data:       nftMetadata,
				})

				if err != nil {
					m.log.Errorf("Couldn't store NFT metadata in IPFS: %s", err)

					return nil, err
				}

				ipfsPath := path.New(ipfsPinningRes.CID).String()

				m.log.Info("Setting the IPFS path as NFT metadata in Centrifuge chain, IPFS path - %s", ipfsPath)

				_, err = m.api.SetMetadata(ctx, req.CollectionID, itemID, []byte(ipfsPath), req.FreezeMetadata)

				if err != nil {
					m.log.Errorf("Couldn't set IPFS CID: %s", err)

					return nil, err
				}

				m.log.Infof(
					"Successfully stored NFT metadata, collection ID - %d, item ID - %d, IPFS path - %s",
					req.CollectionID,
					itemID,
					ipfsPath,
				)

				return nil, nil
			},
		},
	}
}

func getDocAttributes(doc documents.Document, attrKeys []documents.AttrKey) (DocAttributes, error) {
	attrMap := make(DocAttributes)

	for _, attrKey := range attrKeys {
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

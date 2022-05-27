package v3

import (
	"context"
	"errors"
	"fmt"

	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/ipfs_pinning"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/gocelery/v2"
	logging "github.com/ipfs/go-log"
	"github.com/ipfs/interface-go-ipfs-core/path"
)

const (
	mintNFTV3Job        = "Mint NFT V3 Job"
	createNFTClassV3Job = "Create NFT class V3 Job"
)

type CreateClassJob struct {
	jobs.Base

	accountsSrv config.Service
	docSrv      documents.Service
	dispatcher  jobs.Dispatcher
	api         UniquesAPI
	log         *logging.ZapEventLogger
}

// New returns a new instance of CreateClassJob
func (c *CreateClassJob) New() gocelery.Runner {
	log := logging.Logger("create_nft_class_v3_dispatcher")

	cj := &CreateClassJob{
		accountsSrv: c.accountsSrv,
		docSrv:      c.docSrv,
		dispatcher:  c.dispatcher,
		api:         c.api,
		log:         log,
	}

	cj.Base = jobs.NewBase(cj.loadTasks())
	return cj
}

func (c *CreateClassJob) convertArgs(
	args []interface{},
) (
	ctx context.Context,
	classID types.U64,
	err error,
) {
	did := args[0].(identity.DID)
	classID = args[1].(types.U64)

	acc, err := c.accountsSrv.GetAccount(did[:])
	if err != nil {
		err = fmt.Errorf("failed to get account: %w", err)
		return
	}

	ctx = contextutil.WithAccount(context.Background(), acc)

	return ctx, classID, nil
}

func (c *CreateClassJob) loadTasks() map[string]jobs.Task {
	return map[string]jobs.Task{
		"create_nft_class_v3": {
			RunnerFunc: func(args []interface{}, overrides map[string]interface{}) (result interface{}, err error) {
				ctx, classID, err := c.convertArgs(args)

				if err != nil {
					return nil, err
				}

				extInfo, err := c.api.CreateClass(ctx, classID)

				if err != nil {
					return nil, err
				}

				c.log.Infof("Successfully created class with ID - %d, ext hash - %s", classID, extInfo.Hash.Hex())

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
	api            UniquesAPI
	ipfsPinningSrv ipfs_pinning.PinataServiceClient
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
	ctx context.Context,
	did identity.DID,
	instanceID types.U128,
	req MintNFTRequest,
	err error,
) {
	did = args[0].(identity.DID)
	instanceID = args[1].(types.U128)
	req = args[2].(MintNFTRequest)

	acc, err := m.accountsSrv.GetAccount(did[:])
	if err != nil {
		err = fmt.Errorf("failed to get account: %w", err)
		return
	}

	ctx = contextutil.WithAccount(context.Background(), acc)
	return ctx, did, instanceID, req, nil
}

func (m *MintNFTJob) loadTasks() map[string]jobs.Task {
	return map[string]jobs.Task{
		"add_nft_v3_to_document": {
			RunnerFunc: func(args []interface{}, overrides map[string]interface{}) (result interface{}, err error) {
				ctx, _, instanceID, req, err := m.convertArgs(args)

				if err != nil {
					return nil, err
				}

				doc, err := m.docSrv.GetCurrentVersion(ctx, req.DocumentID)

				if err != nil {
					m.log.Errorf("Couldn't get document: %s", err)

					return nil, fmt.Errorf("failed to get document: %w", err)
				}

				err = doc.AddCcNft(req.ClassID, instanceID)

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
				did := args[0].(identity.DID)

				jobID := overrides["document_commit_job"].(gocelery.JobID)

				m.log.Info("Waiting for document to be anchored")

				job, err := m.dispatcher.Job(did, jobID)

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
				ctx, _, instanceID, req, err := m.convertArgs(args)

				if err != nil {
					return nil, err
				}

				m.log.Info("Minting NFT on Centrifuge chain...")

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

				extInfo, err := m.api.MintInstance(ctx, req.ClassID, instanceID, req.Owner)

				if err != nil {
					m.log.Errorf("Couldn't mint instance: %s", err)

					return nil, err
				}

				m.log.Infof(
					"Successfully minted NFT on Centrifuge chain, class ID - %d, instance ID - %d, anchor ID - %s, ext hash - %s",
					req.ClassID,
					instanceID,
					anchorID.String(),
					extInfo.Hash.Hex(),
				)

				return nil, nil
			},
			Next: "store_nft_v3_metadata",
		},
		"store_nft_v3_metadata": {
			RunnerFunc: func(args []interface{}, overrides map[string]interface{}) (result interface{}, err error) {
				ctx, _, instanceID, req, err := m.convertArgs(args)

				if err != nil {
					return nil, err
				}

				if len(req.DocAttributes) == 0 {
					m.log.Info("No document attributes provided")

					return nil, nil
				}

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

				nftMeta := NFTMetadata{
					DocID:         doc.ID(),
					DocVersion:    doc.CurrentVersion(),
					DocAttributes: docAttributes,
				}

				m.log.Info("Storing NFT metadata in IPFS")

				ipfsPinningRes, err := m.ipfsPinningSrv.PinJSONToIPFS(ctx, nftMeta, &ipfs_pinning.PinataOptions{CIDVersion: 1}, nil)

				if err != nil {
					m.log.Errorf("Couldn't store NFT metadata in IPFS: %s", err)

					return nil, err
				}

				ipfsPath := path.New(ipfsPinningRes.IpfsHash).String()

				m.log.Info("Setting the IPFS path as NFT metadata in Centrifuge chain, IPFS path - %s", ipfsPath)

				_, err = m.api.SetMetadata(ctx, req.ClassID, instanceID, []byte(ipfsPath), req.FreezeMetadata)

				if err != nil {
					m.log.Errorf("Couldn't set IPFS CID: %s", err)

					return nil, err
				}

				m.log.Infof(
					"Successfully stored NFT metadata, class ID - %d, instance ID - %d, IPFS path - %s",
					req.ClassID,
					instanceID,
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

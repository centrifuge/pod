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
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/gocelery/v2"
	logging "github.com/ipfs/go-log"
)

const (
	mintNFTV3Job = "Mint NFT V3 Job"
)

type MintNFTJob struct {
	jobs.Base

	accountsSrv config.Service
	docSrv      documents.Service
	dispatcher  jobs.Dispatcher
	api         UniquesAPI
	log         *logging.ZapEventLogger
}

// New returns a new instance of MintNFTOnCCJob
func (m *MintNFTJob) New() gocelery.Runner {
	log := logging.Logger("nft_v3_dispatcher")

	nm := &MintNFTJob{
		accountsSrv: m.accountsSrv,
		docSrv:      m.docSrv,
		dispatcher:  m.dispatcher,
		api:         m.api,
		log:         log,
	}

	nm.Base = jobs.NewBase(nm.loadTasks())
	return nm
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
		// TODO(cdamian): Insert IPFS step
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

				classID := req.ClassID

				classExists, err := m.classExists(ctx, classID)

				if err != nil {
					m.log.Errorf("Couldn't check if class exists: %s", err)

					return nil, err
				}

				if !classExists {
					m.log.Info("Class does not exist, creating it now")

					_, err := m.api.CreateClass(ctx, classID)

					if err != nil {
						m.log.Errorf("Couldn't create class: %s", err)

						return nil, err
					}
				}

				extInfo, err := m.api.MintInstance(ctx, classID, instanceID, req.Owner)

				if err != nil {
					m.log.Errorf("Couldn't mint instance: %s", err)

					return nil, err
				}

				overrides["ext_info"] = extInfo

				m.log.Infof(
					"Successfully minted NFT on Centrifuge chain, class ID - %d, instance ID - %d, anchor ID - %s, ext hash - %s",
					classID,
					instanceID,
					anchorID.String(),
					extInfo.Hash.Hex(),
				)

				return nil, nil
			},
		},
	}
}

func (m *MintNFTJob) classExists(ctx context.Context, classID types.U64) (bool, error) {
	classDetails, err := m.api.GetClassDetails(ctx, classID)

	if err != nil {
		return false, err
	}

	return classDetails != nil, nil
}

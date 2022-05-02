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
	"go.uber.org/zap"
)

const (
	mintNFTJob = "Mint NFT Job"
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
	instanceID = args[2].(types.U128)
	req = args[3].(MintNFTRequest)

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
		"add_nft_to_document": {
			RunnerFunc: func(args []interface{}, overrides map[string]interface{}) (result interface{}, err error) {
				ctx, _, instanceID, req, err := m.convertArgs(args)

				if err != nil {
					return nil, err
				}

				log := m.log.With(
					zap.ByteString("doc_id", req.DocumentID),
					zap.Uint64("class_id", uint64(req.ClassID)),
					zap.Any("owner", req.Owner),
				)

				doc, err := m.docSrv.GetCurrentVersion(ctx, req.DocumentID)

				if err != nil {
					log.Error("Couldn't get document", zap.Error(err))

					return nil, fmt.Errorf("failed to get document: %w", err)
				}

				err = doc.AddNFTV2(req.ClassID, instanceID)

				if err != nil {
					log.Error("Couldn't add NFT to document", zap.Error(err))

					return nil, fmt.Errorf("failed to add nft to document: %w", err)
				}

				jobID, err := m.docSrv.Commit(ctx, doc)

				if err != nil {
					log.Error("Couldn't commit document", zap.Error(err))

					return nil, fmt.Errorf("failed to commit document: %w", err)
				}

				overrides["document_commit_job"] = jobID

				return nil, nil
			},
			Next: "wait_for_document_commit",
		},
		"wait_for_document_commit": {
			RunnerFunc: func(args []interface{}, overrides map[string]interface{}) (result interface{}, err error) {
				did := args[1].(identity.DID)

				jobID := overrides["document_commit_job"].(gocelery.JobID)

				log := m.log.With(zap.String("job_id", jobID.Hex()))

				log.Info("Waiting for document to be anchored")

				job, err := m.dispatcher.Job(did, jobID)

				if err != nil {
					log.Error("Couldn't dispatch job", zap.Error(err))

					return nil, fmt.Errorf("failed to fetch job: %w", err)
				}

				if !job.IsSuccessful() {
					log.Info("Document not committed yet")

					return nil, errors.New("document not committed yet")
				}

				return nil, nil
			},
			Next: "mint_nft",
		},
		// TODO(cdamian): Insert IPFS step
		"mint_nft": {
			RunnerFunc: func(args []interface{}, overrides map[string]interface{}) (result interface{}, err error) {
				ctx, _, instanceID, req, err := m.convertArgs(args)

				if err != nil {
					return nil, err
				}

				log := m.log.With(
					zap.ByteString("doc_id", req.DocumentID),
					zap.Uint64("class_id", uint64(req.ClassID)),
					zap.Int64("instance_id", instanceID.Int64()),
					zap.Any("owner", req.Owner),
				)

				log.Info("Minting NFT on Centrifuge chain...")

				doc, err := m.docSrv.GetCurrentVersion(ctx, req.DocumentID)

				if err != nil {
					log.Error("Couldn't get current document version", zap.Error(err))

					return nil, err
				}

				anchorID, err := anchors.ToAnchorID(doc.CurrentVersion())

				if err != nil {
					log.Error("Couldn't get anchor for document", zap.Error(err))

					return nil, err
				}

				classID := req.ClassID

				classExists, err := m.classExists(ctx, classID)

				if err != nil {
					log.Error("Couldn't check if class exists", zap.Error(err))

					return nil, err
				}

				if !classExists {
					log.Info("Class does not exist, creating it now")

					_, err := m.api.CreateClass(ctx, classID)

					if err != nil {
						log.Error("Couldn't create class", zap.Error(err))

						return nil, err
					}
				}

				extInfo, err := m.api.MintInstance(ctx, classID, instanceID, req.Owner)

				if err != nil {
					log.Error("Couldn't mint instance", zap.Error(err))

					return nil, err
				}

				overrides["ext_info"] = extInfo

				log.Info(
					"Successfully minted NFT on Centrifuge chain",
					zap.String("anchor_id", anchorID.String()),
					zap.String("extrinsic_hash", extInfo.Hash.Hex()),
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

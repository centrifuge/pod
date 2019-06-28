package httpapi

import (
	"context"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/nft"
	"github.com/ethereum/go-ethereum/common"
)

// CoreService exposes the core services in the node to user specific APIs or to the outside world.
type CoreService interface {
	// CreateDocument creates the document from the payload and anchors it.
	CreateDocument(ctx context.Context, payload documents.CreatePayload) (documents.Model, jobs.JobID, error)

	// UpdateDocument updates the document from the payload and anchors the next version.
	UpdateDocument(ctx context.Context, payload documents.UpdatePayload) (documents.Model, jobs.JobID, error)

	// GetJobStatus returns the job status.
	GetJobStatus(account identity.DID, id jobs.JobID) (jobs.StatusResponse, error)

	// GetDocument returns the latest version of the document.
	GetDocument(ctx context.Context, docID []byte) (documents.Model, error)

	// GetDocumentVersion returns the specific version of the document
	GetDocumentVersion(ctx context.Context, docID, versionID []byte) (documents.Model, error)

	// GenerateProofs returns the proofs for the latest version of the document.
	GenerateProofs(ctx context.Context, docID []byte, fields []string) (*documents.DocumentProof, error)

	// GenerateProofsForVersion returns the proofs for the specific version of the document.
	GenerateProofsForVersion(ctx context.Context, docID, versionID []byte, fields []string) (*documents.DocumentProof, error)

	// MintNFT mints an NFT.
	MintNFT(ctx context.Context, request nft.MintNFTRequest) (*nft.TokenResponse, error)

	// TransferNFT transfers NFT with tokenID in a given registry to `to` address.
	TransferNFT(ctx context.Context, to, registry common.Address, tokenID nft.TokenID) (*nft.TokenResponse, error)

	// OwnerOfNFT returns the owner of the NFT.
	OwnerOfNFT(registry common.Address, tokenID nft.TokenID) (common.Address, error)

	// SignPayload uses the accountID's secret key to sign the payload and returns the signature
	SignPayload(accountID, payload []byte) (*coredocumentpb.Signature, error)
}

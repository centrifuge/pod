package ipfs_pinning

import "context"

//go:generate mockery --name PinningServiceClient --structname PinningServiceClientMock --filename ipfs_mock.go --inpackage

type PinningServiceClient interface {
	PinData(ctx context.Context, req *PinRequest) (*PinResponse, error)
	UnpinData(ctx context.Context, CID string) error
}

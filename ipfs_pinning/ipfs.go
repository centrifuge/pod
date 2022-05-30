package ipfs_pinning

import "context"

type PinningServiceClient interface {
	PinData(ctx context.Context, req *PinRequest) (*PinResponse, error)
	UnpinData(ctx context.Context, CID string) error
}

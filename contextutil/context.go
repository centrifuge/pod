package contextutil

import (
	"context"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

type contextKey string

const (
	// ErrSelfNotFound must be used when self value is not found in the context
	ErrSelfNotFound = errors.Error("self value not found in the context")

	// ErrNonceNotFound must be used when nonce value is not found in context
	ErrNonceNotFound = errors.Error("nonce value not found in the context")

	// ErrDIDMissingFromContext sentinel error when did is missing from the context.
	ErrDIDMissingFromContext = errors.Error("failed to extract did from context")

	self = contextKey("self")

	job = contextKey("job")

	nonce = contextKey("nonce")
)

// New creates new instance of the request headers.
// TODO(ved): this doesn't deserve to be here. Causes too many implicit requirements
func New(ctx context.Context, cfg config.Account) (context.Context, error) {
	return context.WithValue(ctx, self, cfg), nil
}

// WithJob returns a context with Job ID
func WithJob(ctx context.Context, jobID jobs.JobID) context.Context {
	return context.WithValue(ctx, job, jobID)
}

// WithNonce returns a context with Nonce value
func WithNonce(ctx context.Context, n uint32) context.Context {
	return context.WithValue(ctx, nonce, n)
}

// Job returns current jobID
func Job(ctx context.Context) jobs.JobID {
	jobID, ok := ctx.Value(job).(jobs.JobID)
	if !ok {
		return jobs.NilJobID()
	}
	return jobID
}

// AccountDID extracts the AccountConfig DID from the given context value
func AccountDID(ctx context.Context) (identity.DID, error) {
	acc, err := Account(ctx)
	if err != nil {
		return identity.DID{}, err
	}
	didBytes := acc.GetIdentityID()
	did, err := identity.NewDIDFromBytes(didBytes)
	if err != nil {
		return identity.DID{}, err
	}
	return did, nil
}

// Account extracts the TenantConfig from the given context value
func Account(ctx context.Context) (config.Account, error) {
	tc, ok := ctx.Value(self).(config.Account)
	if !ok {
		return nil, ErrSelfNotFound
	}
	return tc, nil
}

// Nonce extracts the nonce value from the given context
func Nonce(ctx context.Context) (uint32, error) {
	n, ok := ctx.Value(nonce).(uint32)
	if !ok {
		return 0, ErrNonceNotFound
	}
	return n, nil
}

// Context updates a context with account info using the configstore, must only be used for api handlers
func Context(ctx context.Context, cs config.Service) (context.Context, error) {
	tcIDHex, ok := ctx.Value(config.AccountHeaderKey).(string)
	if !ok {
		return nil, errors.New("failed to get header %v", config.AccountHeaderKey)
	}

	tcID, err := hexutil.Decode(tcIDHex)
	if err != nil {
		return nil, errors.New("failed to get header: %v", err)
	}

	tc, err := cs.GetAccount(tcID)
	if err != nil {
		return nil, errors.New("failed to get header: %v", err)
	}

	ctxHeader, err := New(ctx, tc)
	if err != nil {
		return nil, errors.New("failed to get header: %v", err)
	}
	return ctxHeader, nil
}

// Copy creates a copy of the given context with relevant values
func Copy(ctx context.Context) context.Context {
	nctx := context.WithValue(context.Background(), self, ctx.Value(self))
	nctx = context.WithValue(nctx, job, ctx.Value(job))
	nctx = context.WithValue(nctx, nonce, ctx.Value(nonce))
	return nctx
}

// DIDFromContext returns did from the context.
func DIDFromContext(ctx context.Context) (did identity.DID, err error) {
	didStr, ok := ctx.Value(config.AccountHeaderKey).(string)
	if !ok {
		return did, ErrDIDMissingFromContext
	}

	didBytes, err := hexutil.Decode(didStr)
	if err != nil {
		return did, err
	}

	return identity.NewDIDFromBytes(didBytes)
}

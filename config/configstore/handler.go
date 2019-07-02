package configstore

import (
	"context"

	"github.com/centrifuge/go-centrifuge/errors"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/account"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/golang/protobuf/ptypes/empty"
	logging "github.com/ipfs/go-log"
)

// ErrDerivingAccount used as generic account deriver type
const ErrDerivingAccount = errors.Error("error deriving account")

var apiLog = logging.Logger("account-api")

type grpcHandler struct {
	service config.Service
}

// GRPCAccountHandler returns an implementation of accountpb.AccountServiceServer
func GRPCAccountHandler(svc config.Service) accountpb.AccountServiceServer {
	return &grpcHandler{service: svc}
}

// deriveAllAccountResponse derives all valid accounts, will not return accounts that fail at load time
func (h grpcHandler) deriveAllAccountResponse(cfgs []config.Account) (*accountpb.GetAllAccountResponse, error) {
	response := new(accountpb.GetAllAccountResponse)
	for _, t := range cfgs {
		tpb, err := t.CreateProtobuf()
		if err != nil {
			bID := t.GetIdentityID()
			apiLog.Errorf("%v", errors.NewTypedError(ErrDerivingAccount, errors.New("account [%s]: %v", hexutil.Encode(bID), err)))
			continue
		}
		response.Data = append(response.Data, tpb)
	}
	return response, nil
}

func (h grpcHandler) GetAllAccounts(ctx context.Context, req *empty.Empty) (*accountpb.GetAllAccountResponse, error) {
	cfgs, err := h.service.GetAllAccounts()
	if err != nil {
		return nil, err
	}
	return h.deriveAllAccountResponse(cfgs)
}

func (h grpcHandler) CreateAccount(ctx context.Context, data *accountpb.AccountData) (*accountpb.AccountData, error) {
	apiLog.Infof("Creating account: %v", data)
	accountConfig := new(Account)
	err := accountConfig.loadFromProtobuf(data)
	if err != nil {
		return nil, err
	}
	tc, err := h.service.CreateAccount(accountConfig)
	if err != nil {
		return nil, err
	}
	return tc.CreateProtobuf()
}

func (h grpcHandler) UpdateAccount(ctx context.Context, req *accountpb.UpdateAccountRequest) (*accountpb.AccountData, error) {
	apiLog.Infof("Updating account: %v", req)
	accountConfig := new(Account)
	err := accountConfig.loadFromProtobuf(req.Data)
	if err != nil {
		return nil, err
	}
	tc, err := h.service.UpdateAccount(accountConfig)
	if err != nil {
		return nil, err
	}
	return tc.CreateProtobuf()
}

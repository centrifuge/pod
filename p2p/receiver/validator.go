package receiver

import (
	"fmt"

	"github.com/centrifuge/go-centrifuge/identity"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	"github.com/centrifuge/go-centrifuge/centerrors"
	"github.com/centrifuge/go-centrifuge/code"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/version"
	libp2pPeer "github.com/libp2p/go-libp2p-peer"
)

// Validator defines method that must be implemented by any validator type.
type Validator interface {
	// Validate validates p2p requests
	Validate(header *p2ppb.Header, centID *identity.CentID, peerID *libp2pPeer.ID) error
}

// ValidatorGroup implements Validator for validating a set of validators.
type ValidatorGroup []Validator

// Validate will execute all group specific atomic validations
func (group ValidatorGroup) Validate(header *p2ppb.Header, centID *identity.CentID, peerID *libp2pPeer.ID) (errs error) {
	for _, v := range group {
		if err := v.Validate(header, centID, peerID); err != nil {
			errs = errors.AppendError(errs, err)
		}
	}
	return errs
}

// ValidatorFunc implements Validator and can be used as a adaptor for functions
// with specific function signature
type ValidatorFunc func(header *p2ppb.Header, centID *identity.CentID, peerID *libp2pPeer.ID) error

// Validate passes the arguments to the underlying validator
// function and returns the results
func (vf ValidatorFunc) Validate(header *p2ppb.Header, centID *identity.CentID, peerID *libp2pPeer.ID) error {
	return vf(header, centID, peerID)
}

func versionValidator() Validator {
	return ValidatorFunc(func(header *p2ppb.Header, centID *identity.CentID, peerID *libp2pPeer.ID) error {
		if header == nil {
			return errors.New("nil header")
		}
		if !version.CheckVersion(header.NodeVersion) {
			return version.IncompatibleVersionError(header.NodeVersion)
		}
		return nil
	})
}

func networkValidator(networkID uint32) Validator {
	return ValidatorFunc(func(header *p2ppb.Header, centID *identity.CentID, peerID *libp2pPeer.ID) error {
		if header == nil {
			return errors.New("nil header")
		}
		if networkID != header.NetworkIdentifier {
			return incompatibleNetworkError(networkID, header.NetworkIdentifier)
		}
		return nil
	})
}

func peerValidator(idService identity.Service) Validator {
	return ValidatorFunc(func(header *p2ppb.Header, centID *identity.CentID, peerID *libp2pPeer.ID) error {
		if header == nil {
			return errors.New("nil header")
		}
		if centID == nil {
			return errors.New("nil centID")
		}
		if peerID == nil {
			return errors.New("nil peerID")
		}
		pk, err := peerID.ExtractPublicKey()
		if err != nil {
			return err
		}
		if pk == nil {
			return errors.New("cannot extract public key out of peer ID")
		}
		idKey, err := pk.Raw()
		if err != nil {
			return err
		}
		return idService.ValidateKey(*centID, idKey, identity.KeyPurposeP2P)
	})
}

// HandshakeValidator validates the p2p handshake details
func HandshakeValidator(networkID uint32, idService identity.Service) ValidatorGroup {
	return ValidatorGroup{
		versionValidator(),
		networkValidator(networkID),
		peerValidator(idService),
	}
}

// DocumentAccessValidator validates the GetDocument request against the AccessType indicated in the request
//func DocumentAccessValidator(dm *documents.CoreDocumentModel, docReq *p2ppb.GetDocumentRequest, requesterCentID identity.CentID) error {
//av := coredocument.AccountValidator()
//// checks which access type is relevant for the request
//switch docReq.GetAccessType() {
//case p2ppb.AccessType_ACCESS_TYPE_REQUESTER_VERIFICATION:
//	if !av.AccountCanRead(doc, requesterCentID) {
//		return errors.New("requester does not have access")
//	}
//case p2ppb.AccessType_ACCESS_TYPE_NFT_OWNER_VERIFICATION:
//	registry := common.BytesToAddress(docReq.NftRegistryAddress)
//	if av.NFTOwnerCanRead(doc, registry, docReq.NftTokenId, requesterCentID) != nil {
//		return errors.New("requester does not have access")
//	}
////// case AccessTokenValidation
//// case p2ppb.AccessType_ACCESS_TYPE_ACCESS_TOKEN_VERIFICATION:
////
//// case p2ppb.AccessType_ACCESS_TYPE_INVALID:
//default:
//	return errors.New("invalid access type ")
//}
//return nil
//}

func incompatibleNetworkError(configNetwork uint32, nodeNetwork uint32) error {
	return centerrors.New(code.NetworkMismatch, fmt.Sprintf("Incompatible network id: node network: %d, client network: %d", configNetwork, nodeNetwork))
}

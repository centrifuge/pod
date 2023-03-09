package p2p

import "github.com/centrifuge/pod/errors"

const (
	ErrInvalidReceiverAccount       = errors.Error("invalid receiver account")
	ErrInvalidRequesterAccount      = errors.Error("invalid requester account")
	ErrInvalidCollaboratorAccount   = errors.Error("invalid collaborator account")
	ErrPeerIDRetrieval              = errors.Error("couldn't retrieve peer ID")
	ErrP2PEnvelopePreparation       = errors.Error("couldn't prepare P2P envelope")
	ErrP2PMessageSending            = errors.Error("couldn't send P2P message")
	ErrP2PDataEnvelopeResolving     = errors.Error("couldn't resolve P2P data envelope")
	ErrP2PClient                    = errors.Error("P2P client error")
	ErrIncorrectResponseMessageType = errors.Error("incorrect response message type")
	ErrResponseDecodeError          = errors.Error("couldn't decode response message")
	ErrSignerCollaboratorsRetrieval = errors.Error("couldn't get signer collaborators")
	ErrP2PKeyRetrievalError         = errors.Error("couldn't retrieve P2P key")
	ErrInvalidP2PKey                = errors.Error("invalid P2P key")
	ErrPeerIDParsing                = errors.Error("couldn't parse peer ID")
	ErrPeerNotFound                 = errors.Error("peer not found")
	ErrCoreDocumentPacking          = errors.Error("couldn't pack core document")
	ErrDocumentSignatureRequest     = errors.Error("couldn't request document signature")
	ErrInvalidSignatureResponse     = errors.Error("invalid signature response")
)

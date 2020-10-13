package centchain

import "github.com/centrifuge/go-substrate-rpc-client/types"

// EventNFTDeposited is emitted when NFT is ready to be deposited to other chain.
type EventNFTDeposited struct {
	Phase  types.Phase
	Asset  types.Hash
	Topics []types.Hash
}

// EventFeeChanged is emitted when a fee for a given key is changed.
type EventFeeChanged struct {
	Phase    types.Phase
	Key      types.Hash
	NewPrice types.U128
	Topics   []types.Hash
}

// EventNewMultiAccount is emitted when a multi account has been created.
// First param is the account that created it, second is the multisig account.
type EventNewMultiAccount struct {
	Phase   types.Phase
	Who, ID types.AccountID
	Topics  []types.Hash
}

// EventMultiAccountUpdated is emitted when a multi account has been updated. First param is the multisig account.
type EventMultiAccountUpdated struct {
	Phase  types.Phase
	Who    types.AccountID
	Topics []types.Hash
}

// EventMultiAccountRemoved is emitted when a multi account has been removed. First param is the multisig account.
type EventMultiAccountRemoved struct {
	Phase  types.Phase
	Who    types.AccountID
	Topics []types.Hash
}

// EventNewMultisig is emitted when a new multisig operation has begun.
// First param is the account that is approving, second is the multisig account.
type EventNewMultisig struct {
	Phase   types.Phase
	Who, ID types.AccountID
	Topics  []types.Hash
}

// TimePoint contains height and index
type TimePoint struct {
	Height types.BlockNumber
	Index  types.U32
}

// EventMultisigApproval is emitted when a multisig operation has been approved by someone.
// First param is the account that is approving, third is the multisig account.
type EventMultisigApproval struct {
	Phase     types.Phase
	Who       types.AccountID
	TimePoint TimePoint
	ID        types.AccountID
	Topics    []types.Hash
}

// EventMultisigExecuted is emitted when a multisig operation has been executed by someone.
// First param is the account that is approving, third is the multisig account.
type EventMultisigExecuted struct {
	Phase     types.Phase
	Who       types.AccountID
	TimePoint TimePoint
	ID        types.AccountID
	Result    types.DispatchResult
	Topics    []types.Hash
}

// EventMultisigCancelled is emitted when a multisig operation has been cancelled by someone.
// First param is the account that is approving, third is the multisig account.
type EventMultisigCancelled struct {
	Phase     types.Phase
	Who       types.AccountID
	TimePoint TimePoint
	ID        types.AccountID
	Topics    []types.Hash
}

// EventFungibleTransfer is emitted when a bridge fungible token transfer is executed
type EventFungibleTransfer struct {
	Phase        types.Phase
	Destination  types.U8
	DepositNonce types.U64
	ResourceID   types.Bytes32
	Amount       types.U32
	Recipient    types.Bytes
	Topics       []types.Hash
}

// EventNonFungibleTransfer is emitted when a bridge non fungible token transfer is executed
type EventNonFungibleTransfer struct {
	Phase        types.Phase
	Destination  types.U8
	DepositNonce types.U64
	ResourceID   types.Bytes32
	TokenID      types.Bytes
	Recipient    types.Bytes
	Metadata     types.Bytes
	Topics       []types.Hash
}

// EventGenericTransfer is emitted when a bridge generic transfer is executed
type EventGenericTransfer struct {
	Phase        types.Phase
	Destination  types.U8
	DepositNonce types.U64
	ResourceID   types.Bytes32
	Metadata     types.Bytes
	Topics       []types.Hash
}

// EventChainWhitelisted is emitted when a new chain has been whitelisted to interact with the bridge
type EventChainWhitelisted struct {
	Phase       types.Phase
	Destination types.U8
	Topics      []types.Hash
}

// EventRelayerAdded is emitted when a new bridge relayer has been whitelisted
type EventRelayerAdded struct {
	Phase   types.Phase
	Relayer types.AccountID
	Topics  []types.Hash
}

// EventRelayerThresholdChanged is emitted when the relayer threshold is changed
type EventRelayerThresholdChanged struct {
	Phase     types.Phase
	Threshold types.U32
	Topics    []types.Hash
}

// Events holds the default events and custom events for centrifuge chain
type Events struct {
	types.EventRecords
	ChainBridge_FungibleTransfer        []EventFungibleTransfer               //nolint:stylecheck,golint
	ChainBridge_NonFungibleTransfer     []EventNonFungibleTransfer            //nolint:stylecheck,golint
	ChainBridge_GenericTransfer         []EventGenericTransfer                //nolint:stylecheck,golint
	ChainBridge_ChainWhitelisted        []EventChainWhitelisted               //nolint:stylecheck,golint
	ChainBridge_RelayerAdded            []EventRelayerAdded                   //nolint:stylecheck,golint
	ChainBridge_RelayerThresholdChanged []EventRelayerThresholdChanged        //nolint:stylecheck,golint
	Nfts_DepositAsset                   []EventNFTDeposited                   //nolint:stylecheck,golint
	Council_Proposed                    []types.EventCollectiveProposed       //nolint:stylecheck,golint
	Council_Voted                       []types.EventCollectiveProposed       //nolint:stylecheck,golint
	Council_Approved                    []types.EventCollectiveApproved       //nolint:stylecheck,golint
	Council_Disapproved                 []types.EventCollectiveDisapproved    //nolint:stylecheck,golint
	Council_Executed                    []types.EventCollectiveExecuted       //nolint:stylecheck,golint
	Council_MemberExecuted              []types.EventCollectiveMemberExecuted //nolint:stylecheck,golint
	Council_Closed                      []types.EventCollectiveClosed         //nolint:stylecheck,golint
	Fees_FeeChanged                     []EventFeeChanged                     //nolint:stylecheck,golint
	MultiAccount_NewMultiAccount        []EventNewMultiAccount                //nolint:stylecheck,golint
	MultiAccount_MultiAccountUpdated    []EventMultiAccountUpdated            //nolint:stylecheck,golint
	MultiAccount_MultiAccountRemoved    []EventMultiAccountRemoved            //nolint:stylecheck,golint
	MultiAccount_NewMultisig            []EventNewMultisig                    //nolint:stylecheck,golint
	MultiAccount_MultisigApproval       []EventMultisigApproval               //nolint:stylecheck,golint
	MultiAccount_MultisigExecuted       []EventMultisigExecuted               //nolint:stylecheck,golint
	MultiAccount_MultisigCancelled      []EventMultisigCancelled              //nolint:stylecheck,golint
}

package centchain

import (
	centEvents "github.com/centrifuge/chain-custom-types"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

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

// EventNFTMint is emitted when an NFT with tokenID is minted in a given registryID
type EventNFTMint struct {
	Phase      types.Phase
	RegistryID types.H160
	TokenID    types.U256
	Topics     []types.Hash
}

// EventRegistryCreated is emitted when a new registry is created
type EventRegistryCreated struct {
	Phase      types.Phase
	RegistryID types.H160
	Topics     []types.Hash
}

// AssetID is a combination of RegistryID and TokenID
type AssetID struct {
	RegistryID types.H160
	TokenID    types.U256
}

// EventNFTTransferred is emitted when an NFT is transferred to a new owner
type EventNFTTransferred struct {
	Phase      types.Phase
	RegistryID types.H160
	AssetID    AssetID
	AccountID  types.AccountID
	Topics     []types.Hash
}

type EventTransactionPaymentTransactionFeePaid struct {
	Phase     types.Phase
	Who       types.AccountID
	ActualFee types.U128
	Tip       types.U128
	Topics    []types.Hash
}

// EventFeesFeeChanged is emitted when a new fee has been set for a key
type EventFeesFeeChanged struct {
	Phase  types.Phase
	Key    types.U8
	Fee    types.U128
	Topics []types.Hash
}

// EventFeesFeeToAuthor is emitted when a fee has been sent to the block author
type EventFeesFeeToAuthor struct {
	Phase   types.Phase
	From    types.AccountID
	Balance types.U128
	Topics  []types.Hash
}

// EventFeesFeeToBurn is emitted when a fee has been burnt
type EventFeesFeeToBurn struct {
	Phase   types.Phase
	From    types.AccountID
	Balance types.U128
	Topics  []types.Hash
}

// EventFeesFeeToTreasury is emitted when a fee has been sent to the treasury
type EventFeesFeeToTreasury struct {
	Phase   types.Phase
	From    types.AccountID
	Balance types.U128
	Topics  []types.Hash
}

type cEvents = centEvents.Events

// Events holds the default events and custom events for centrifuge chain
type Events struct {
	types.EventRecords
	cEvents

	// Ensure that the centrifuge Claims_Claimed event is used.
	Claims_Claimed []centEvents.EventClaimsClaimed //nolint:stylecheck,revive

	ChainBridge_FungibleTransfer          []EventFungibleTransfer                     //nolint:stylecheck,revive
	ChainBridge_NonFungibleTransfer       []EventNonFungibleTransfer                  //nolint:stylecheck,revive
	ChainBridge_GenericTransfer           []EventGenericTransfer                      //nolint:stylecheck,revive
	ChainBridge_ChainWhitelisted          []EventChainWhitelisted                     //nolint:stylecheck,revive
	ChainBridge_RelayerAdded              []EventRelayerAdded                         //nolint:stylecheck,revive
	ChainBridge_RelayerThresholdChanged   []EventRelayerThresholdChanged              //nolint:stylecheck,revive
	Fees_FeeChanged                       []EventFeesFeeChanged                       //nolint:stylecheck,revive
	Fees_FeeToAuthor                      []EventFeesFeeToAuthor                      //nolint:stylecheck,revive
	Fees_FeeToBurn                        []EventFeesFeeToBurn                        //nolint:stylecheck,revive
	Fees_FeeToTreasury                    []EventFeesFeeToTreasury                    //nolint:stylecheck,revive
	Registry_RegistryCreated              []EventRegistryCreated                      //nolint:stylecheck,revive
	Registry_Mint                         []EventNFTMint                              //nolint:stylecheck,revive
	Nft_Transferred                       []EventNFTTransferred                       //nolint:stylecheck,revive
	TransactionPayment_TransactionFeePaid []EventTransactionPaymentTransactionFeePaid //nolint:stylecheck,revive
}

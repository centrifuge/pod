package anchor

import "github.com/spf13/viper"

type Anchor struct {
	AnchorID      string
	RootHash      string
	SchemaVersion uint
}

/*
This function will be more complex in the future, to check if the document should be anchored or not.
Arguably this should be part of each document struct, as it is an inherent property of each document.
TODO Move this to each document type.
 */
func IsAnchoringRequired() bool {
	return viper.GetBool("anchor.ethereum.enabled")
}
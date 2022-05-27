package ipfs_pinning

import "time"

type PinJSONToIPFSResponse struct {
	IpfsHash  string    `json:"IpfsHash"`
	PinSize   int       `json:"PinSize"`
	Timestamp time.Time `json:"Timestamp"`
}

type Region struct {
	ID                      string `json:"id"`
	DesiredReplicationCount int    `json:"desiredReplicationCount"`
}

type CustomPinPolicy struct {
	Regions []Region `json:"regions"`
}

type PinataOptions struct {
	CIDVersion      int              `json:"cidVersion"`
	CustomPinPolicy *CustomPinPolicy `json:"customPinPolicy,omitempty"`
}

type PinataMetadata struct {
	Name      *string           `json:"name,omitempty"`
	KeyValues map[string]string `json:"keyvalues,omitempty"`
}

type PinJSONToIPFSRequest struct {
	PinataOptions  *PinataOptions  `json:"pinataOptions,omitempty"`
	PinataMetadata *PinataMetadata `json:"pinataMetadata,omitempty"`
	PinataContent  any             `json:"pinataContent"`
}

package ipfs

import "time"

var (
	pinReqValidationFn = func(req *PinRequest) error {
		if req == nil {
			return ErrMissingRequest
		}

		if req.Data == nil {
			return ErrMissingPinningData
		}

		if req.CIDVersion < 0 || req.CIDVersion > 1 {
			return ErrInvalidCIDVersion
		}

		return nil
	}
)

type PinRequest struct {
	Data       any
	CIDVersion int
	Metadata   map[string]string
}

type PinResponse struct {
	CID string
}

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

type NFTMetadata struct {
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Image       string            `json:"image,omitempty"`
	Properties  map[string]string `json:"properties,omitempty"`
}

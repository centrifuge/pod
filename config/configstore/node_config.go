package configstore

import (
	"encoding/json"
	"reflect"
	"time"

	"github.com/centrifuge/go-centrifuge/config"
)

// NodeConfig exposes configs specific to the node
type NodeConfig struct {
	StoragePath             string
	ConfigStoragePath       string
	P2PPort                 int
	P2PExternalIP           string
	P2PConnectionTimeout    time.Duration
	P2PResponseDelay        time.Duration
	P2PPublicKey            string
	P2PPrivateKey           string
	ServerPort              int
	ServerAddress           string
	NumWorkers              int
	WorkerWaitTimeMS        int
	TaskValidDuration       time.Duration
	NetworkString           string
	BootstrapPeers          []string
	NetworkID               uint32
	PprofEnabled            bool
	DebugLogEnabled         bool
	AuthenticationEnabled   bool
	CentChainNodeURL        string
	CentChainIntervalRetry  time.Duration
	CentChainMaxRetries     int
	CentChainAnchorLifespan time.Duration
	IPFSPinningServiceName  string
	IPFSPinningServiceURL   string
	IPFSPinningServiceAuth  string
	PodOperatorSecretSeed   string
	PodAdminSecretSeed      string
}

// GetStoragePath refer the interface
func (nc *NodeConfig) GetStoragePath() string {
	return nc.StoragePath
}

// GetConfigStoragePath refer the interface
func (nc *NodeConfig) GetConfigStoragePath() string {
	return nc.ConfigStoragePath
}

// GetP2PPort refer the interface
func (nc *NodeConfig) GetP2PPort() int {
	return nc.P2PPort
}

// GetP2PExternalIP refer the interface
func (nc *NodeConfig) GetP2PExternalIP() string {
	return nc.P2PExternalIP
}

// GetP2PConnectionTimeout refer the interface
func (nc *NodeConfig) GetP2PConnectionTimeout() time.Duration {
	return nc.P2PConnectionTimeout
}

// GetP2PResponseDelay refer the interface
func (nc *NodeConfig) GetP2PResponseDelay() time.Duration {
	return nc.P2PResponseDelay
}

// GetServerPort refer the interface
func (nc *NodeConfig) GetServerPort() int {
	return nc.ServerPort
}

// GetServerAddress refer the interface
func (nc *NodeConfig) GetServerAddress() string {
	return nc.ServerAddress
}

// GetNumWorkers refer the interface
func (nc *NodeConfig) GetNumWorkers() int {
	return nc.NumWorkers
}

// GetWorkerWaitTimeMS refer the interface
func (nc *NodeConfig) GetWorkerWaitTimeMS() int {
	return nc.WorkerWaitTimeMS
}

// GetTaskValidDuration returns the time duration until which task is valid
func (nc *NodeConfig) GetTaskValidDuration() time.Duration {
	return nc.TaskValidDuration
}

// GetNetworkString refer the interface
func (nc *NodeConfig) GetNetworkString() string {
	return nc.NetworkString
}

// GetBootstrapPeers refer the interface
func (nc *NodeConfig) GetBootstrapPeers() []string {
	return nc.BootstrapPeers
}

// GetNetworkID refer the interface
func (nc *NodeConfig) GetNetworkID() uint32 {
	return nc.NetworkID
}

// GetCentChainNodeURL returns the URL of the CentChain Node.
func (nc *NodeConfig) GetCentChainNodeURL() string {
	return nc.CentChainNodeURL
}

// GetCentChainIntervalRetry returns duration to wait between retries.
func (nc *NodeConfig) GetCentChainIntervalRetry() time.Duration {
	return nc.CentChainIntervalRetry
}

// GetCentChainMaxRetries returns the max acceptable retries.
func (nc *NodeConfig) GetCentChainMaxRetries() int {
	return nc.CentChainMaxRetries
}

// GetCentChainAnchorLifespan returns the default lifespan of an anchor.
func (nc *NodeConfig) GetCentChainAnchorLifespan() time.Duration {
	return nc.CentChainAnchorLifespan
}

// GetP2PKeyPair refer the interface
func (nc *NodeConfig) GetP2PKeyPair() (pub, priv string) {
	return nc.P2PPublicKey, nc.P2PPrivateKey
}

// IsPProfEnabled refer the interface
func (nc *NodeConfig) IsPProfEnabled() bool {
	return nc.PprofEnabled
}

// IsDebugLogEnabled refer the interface
func (nc *NodeConfig) IsDebugLogEnabled() bool {
	return nc.DebugLogEnabled
}

// IsAuthenticationEnabled refer the interface
func (nc *NodeConfig) IsAuthenticationEnabled() bool {
	return nc.AuthenticationEnabled
}

func (nc *NodeConfig) GetIPFSPinningServiceName() string {
	return nc.IPFSPinningServiceName
}

func (nc *NodeConfig) GetIPFSPinningServiceURL() string {
	return nc.IPFSPinningServiceURL
}

func (nc *NodeConfig) GetIPFSPinningServiceAuth() string {
	return nc.IPFSPinningServiceAuth
}

func (nc *NodeConfig) GetPodOperatorSecretSeed() string {
	return nc.PodOperatorSecretSeed
}

func (nc *NodeConfig) GetPodAdminSecretSeed() string {
	return nc.PodAdminSecretSeed
}

// Type Returns the underlying type of the NodeConfig
func (nc *NodeConfig) Type() reflect.Type {
	return reflect.TypeOf(nc)
}

// JSON return the json representation of the model
func (nc *NodeConfig) JSON() ([]byte, error) {
	return json.Marshal(nc)
}

// FromJSON initialize the model with a json
func (nc *NodeConfig) FromJSON(data []byte) error {
	return json.Unmarshal(data, nc)
}

// NewNodeConfig creates a new NodeConfig instance with configs
func NewNodeConfig(c config.Configuration) config.Configuration {
	p2pPub, p2pPriv := c.GetP2PKeyPair()

	return &NodeConfig{
		AuthenticationEnabled:   c.IsAuthenticationEnabled(),
		StoragePath:             c.GetStoragePath(),
		ConfigStoragePath:       c.GetConfigStoragePath(),
		P2PPort:                 c.GetP2PPort(),
		P2PExternalIP:           c.GetP2PExternalIP(),
		P2PConnectionTimeout:    c.GetP2PConnectionTimeout(),
		P2PResponseDelay:        c.GetP2PResponseDelay(),
		P2PPublicKey:            p2pPub,
		P2PPrivateKey:           p2pPriv,
		ServerPort:              c.GetServerPort(),
		ServerAddress:           c.GetServerAddress(),
		NumWorkers:              c.GetNumWorkers(),
		WorkerWaitTimeMS:        c.GetWorkerWaitTimeMS(),
		TaskValidDuration:       c.GetTaskValidDuration(),
		NetworkString:           c.GetNetworkString(),
		BootstrapPeers:          c.GetBootstrapPeers(),
		NetworkID:               c.GetNetworkID(),
		PprofEnabled:            c.IsPProfEnabled(),
		DebugLogEnabled:         c.IsDebugLogEnabled(),
		CentChainMaxRetries:     c.GetCentChainMaxRetries(),
		CentChainIntervalRetry:  c.GetCentChainIntervalRetry(),
		CentChainAnchorLifespan: c.GetCentChainAnchorLifespan(),
		CentChainNodeURL:        c.GetCentChainNodeURL(),
		IPFSPinningServiceName:  c.GetIPFSPinningServiceName(),
		IPFSPinningServiceURL:   c.GetIPFSPinningServiceURL(),
		IPFSPinningServiceAuth:  c.GetIPFSPinningServiceAuth(),
		PodOperatorSecretSeed:   c.GetPodOperatorSecretSeed(),
		PodAdminSecretSeed:      c.GetPodAdminSecretSeed(),
	}
}

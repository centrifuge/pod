package configstore

//// NodeConfig exposes configs specific to the node
//type NodeConfig struct {
//	MainIdentity            Account
//	StoragePath             string
//	AccountsKeystore        string
//	P2PPort                 int
//	P2PExternalIP           string
//	P2PConnectionTimeout    time.Duration
//	P2PResponseDelay        time.Duration
//	P2PPublicKeyPath        string
//	P2PPrivateKeyPath       string
//	ServerPort              int
//	ServerAddress           string
//	NumWorkers              int
//	WorkerWaitTimeMS        int
//	TaskValidDuration       time.Duration
//	NetworkString           string
//	BootstrapPeers          []string
//	NetworkID               uint32
//	PprofEnabled            bool
//	DebugLogEnabled         bool
//	CentChainNodeURL        string
//	CentChainIntervalRetry  time.Duration
//	CentChainMaxRetries     int
//	CentChainAnchorLifespan time.Duration
//}
//
//// IsSet refer the interface
//func (nc *NodeConfig) IsSet(key string) bool {
//	panic("irrelevant, NodeConfig#IsSet must not be used")
//}
//
//// Set refer the interface
//func (nc *NodeConfig) Set(key string, value interface{}) {
//	panic("irrelevant, NodeConfig#Set must not be used")
//}
//
//// SetDefault refer the interface
//func (nc *NodeConfig) SetDefault(key string, value interface{}) {
//	panic("irrelevant, NodeConfig#SetDefault must not be used")
//}
//
//// Get refer the interface
//func (nc *NodeConfig) Get(key string) interface{} {
//	panic("irrelevant, NodeConfig#Get must not be used")
//}
//
//// GetString refer the interface
//func (nc *NodeConfig) GetString(key string) string {
//	panic("irrelevant, NodeConfig#GetString must not be used")
//}
//
//// GetBool refer the interface
//func (nc *NodeConfig) GetBool(key string) bool {
//	panic("irrelevant, NodeConfig#GetBool must not be used")
//}
//
//// GetInt refer the interface
//func (nc *NodeConfig) GetInt(key string) int {
//	panic("irrelevant, NodeConfig#GetInt must not be used")
//}
//
//// GetFloat refer the interface
//func (nc *NodeConfig) GetFloat(key string) float64 {
//	panic("irrelevant, NodeConfig#GetFloat32 must not be used")
//}
//
//// GetDuration refer the interface
//func (nc *NodeConfig) GetDuration(key string) time.Duration {
//	panic("irrelevant, NodeConfig#GetDuration must not be used")
//}
//
//// GetStoragePath refer the interface
//func (nc *NodeConfig) GetStoragePath() string {
//	return nc.StoragePath
//}
//
//// GetConfigStoragePath refer the interface
//func (nc *NodeConfig) GetConfigStoragePath() string {
//	panic("irrelevant, NodeConfig#GetConfigStoragePath must not be used")
//}
//
//// GetAccountsKeystore returns the accounts keystore path.
//func (nc *NodeConfig) GetAccountsKeystore() string {
//	return nc.AccountsKeystore
//}
//
//// GetP2PPort refer the interface
//func (nc *NodeConfig) GetP2PPort() int {
//	return nc.P2PPort
//}
//
//// GetP2PExternalIP refer the interface
//func (nc *NodeConfig) GetP2PExternalIP() string {
//	return nc.P2PExternalIP
//}
//
//// GetP2PConnectionTimeout refer the interface
//func (nc *NodeConfig) GetP2PConnectionTimeout() time.Duration {
//	return nc.P2PConnectionTimeout
//}
//
//// GetP2PResponseDelay refer the interface
//func (nc *NodeConfig) GetP2PResponseDelay() time.Duration {
//	return nc.P2PResponseDelay
//}
//
//func (nc *NodeConfig) GetP2PPublicKeyPath() string {
//	return nc.P2PPublicKeyPath
//}
//
//func (nc *NodeConfig) GetP2PPrivateKeyPath() string {
//	return nc.P2PPrivateKeyPath
//}
//
//// GetServerPort refer the interface
//func (nc *NodeConfig) GetServerPort() int {
//	return nc.ServerPort
//}
//
//// GetServerAddress refer the interface
//func (nc *NodeConfig) GetServerAddress() string {
//	return nc.ServerAddress
//}
//
//// GetNumWorkers refer the interface
//func (nc *NodeConfig) GetNumWorkers() int {
//	return nc.NumWorkers
//}
//
//// GetWorkerWaitTimeMS refer the interface
//func (nc *NodeConfig) GetWorkerWaitTimeMS() int {
//	return nc.WorkerWaitTimeMS
//}
//
//// GetTaskValidDuration returns the time duration until which task is valid
//func (nc *NodeConfig) GetTaskValidDuration() time.Duration {
//	return nc.TaskValidDuration
//}
//
//// GetNetworkString refer the interface
//func (nc *NodeConfig) GetNetworkString() string {
//	return nc.NetworkString
//}
//
//// GetNetworkKey refer the interface
//func (nc *NodeConfig) GetNetworkKey(k string) string {
//	panic("irrelevant, NodeConfig#GetNetworkKey must not be used")
//}
//
//// GetContractAddressString refer the interface
//func (nc *NodeConfig) GetContractAddressString(address string) string {
//	panic("irrelevant, NodeConfig#GetContractAddressString must not be used")
//}
//
//// GetBootstrapPeers refer the interface
//func (nc *NodeConfig) GetBootstrapPeers() []string {
//	return nc.BootstrapPeers
//}
//
//// GetNetworkID refer the interface
//func (nc *NodeConfig) GetNetworkID() uint32 {
//	return nc.NetworkID
//}
//
//// GetCentChainNodeURL returns the URL of the CentChain Node.
//func (nc *NodeConfig) GetCentChainNodeURL() string {
//	return nc.CentChainNodeURL
//}
//
//// GetCentChainIntervalRetry returns duration to wait between retries.
//func (nc *NodeConfig) GetCentChainIntervalRetry() time.Duration {
//	return nc.CentChainIntervalRetry
//}
//
//// GetCentChainMaxRetries returns the max acceptable retries.
//func (nc *NodeConfig) GetCentChainMaxRetries() int {
//	return nc.CentChainMaxRetries
//}
//
//// GetCentChainAnchorLifespan returns the default lifespan of an anchor.
//func (nc *NodeConfig) GetCentChainAnchorLifespan() time.Duration {
//	return nc.CentChainAnchorLifespan
//}
//
//// GetReceiveEventNotificationEndpoint refer the interface
//func (nc *NodeConfig) GetReceiveEventNotificationEndpoint() string {
//	return nc.MainIdentity.ReceiveEventNotificationEndpoint
//}
//
//// GetIdentityID refer the interface
//func (nc *NodeConfig) GetIdentityID() ([]byte, error) {
//	return nc.MainIdentity.IdentityID, nil
//}
//
//// GetP2PKeyPair refer the interface
//func (nc *NodeConfig) GetP2PKeyPair() (pub, priv string) {
//	return nc.MainIdentity.P2PKeyPair.Pub, nc.MainIdentity.P2PKeyPair.Pvt
//}
//
//// GetSigningKeyPair refer the interface
//func (nc *NodeConfig) GetSigningKeyPair() (pub, priv string) {
//	return nc.MainIdentity.SigningKeyPair.Pub, nc.MainIdentity.SigningKeyPair.Pvt
//}
//
//// GetPrecommitEnabled refer the interface
//func (nc *NodeConfig) GetPrecommitEnabled() bool {
//	return nc.MainIdentity.PrecommitEnabled
//}
//
//// IsPProfEnabled refer the interface
//func (nc *NodeConfig) IsPProfEnabled() bool {
//	return nc.PprofEnabled
//}
//
//// IsDebugLogEnabled refer the interface
//func (nc *NodeConfig) IsDebugLogEnabled() bool {
//	return nc.DebugLogEnabled
//}
//
//// ID Gets the ID of the document represented by this model
//func (nc *NodeConfig) ID() ([]byte, error) {
//	return []byte{}, nil
//}
//
//// Type Returns the underlying type of the NodeConfig
//func (nc *NodeConfig) Type() reflect.Type {
//	return reflect.TypeOf(nc)
//}
//
//// JSON return the json representation of the model
//func (nc *NodeConfig) JSON() ([]byte, error) {
//	return json.Marshal(nc)
//}
//
//// FromJSON initialize the model with a json
//func (nc *NodeConfig) FromJSON(data []byte) error {
//	return json.Unmarshal(data, nc)
//}

// NewNodeConfig creates a new NodeConfig instance with configs
//func NewNodeConfig(c config.Configuration) config.Configuration {
//	return &NodeConfig{
//		MainIdentity: Account{
//			//IdentityID:                       mainIdentity,
//			ReceiveEventNotificationEndpoint: c.GetReceiveEventNotificationEndpoint(),
//			//P2PKeyPair: KeyPair{
//			//	Pub: p2pPub,
//			//	Pvt: p2pPriv,
//			//},
//			//SigningKeyPair: KeyPair{
//			//	Pub: signPub,
//			//	Pvt: signPriv,
//			//},
//			//CentChainAccount: centChainAccount,
//		},
//		StoragePath:             c.GetStoragePath(),
//		AccountsKeystore:        c.GetAccountsKeystore(),
//		P2PPort:                 c.GetP2PPort(),
//		P2PExternalIP:           c.GetP2PExternalIP(),
//		P2PConnectionTimeout:    c.GetP2PConnectionTimeout(),
//		P2PResponseDelay:        c.GetP2PResponseDelay(),
//		ServerPort:              c.GetServerPort(),
//		ServerAddress:           c.GetServerAddress(),
//		NumWorkers:              c.GetNumWorkers(),
//		WorkerWaitTimeMS:        c.GetWorkerWaitTimeMS(),
//		TaskValidDuration:       c.GetTaskValidDuration(),
//		NetworkString:           c.GetNetworkString(),
//		BootstrapPeers:          c.GetBootstrapPeers(),
//		NetworkID:               c.GetNetworkID(),
//		PprofEnabled:            c.IsPProfEnabled(),
//		DebugLogEnabled:         c.IsDebugLogEnabled(),
//		CentChainMaxRetries:     c.GetCentChainMaxRetries(),
//		CentChainIntervalRetry:  c.GetCentChainIntervalRetry(),
//		CentChainAnchorLifespan: c.GetCentChainAnchorLifespan(),
//		CentChainNodeURL:        c.GetCentChainNodeURL(),
//	}
//}

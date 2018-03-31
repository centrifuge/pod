package identity

const (
	ACTION_CREATE = "create"
	ACTION_ADDKEY = "addkey"
)

type Identity interface {
	String() string
	GetCentrifugeId() string
	GetLastB58KeyForType(keyType int) (ret string, err error)
	AddKeyToIdentity(keyType int, confirmations chan<- *Identity) (err error)
	CheckIdentityExists() (exists bool, err error)
}
package identity

const (
	ACTION_CREATE = "create"
	ACTION_ADDKEY = "addkey"
	KEY_TYPE_PEERID = 1
	KEY_TYPE_SIGNATURE = 2
	KEY_TYPE_ENCRYPTION = 3
)

type Identity interface {
	String() string
	GetCentrifugeId() string
	GetLastB58KeyForType(keyType int) (ret string, err error)
	AddKeyToIdentity(keyType int, confirmations chan<- *Identity) (err error)
	CheckIdentityExists() (exists bool, err error)
}
package ethereum

type Bootstrapper struct {
}

func (*Bootstrapper) Bootstrap(context map[string]interface{}) error {
	client := NewClientConnection()
	SetConnection(client)
	return nil
}

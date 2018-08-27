package ethereum

type Bootstrapper struct {
}

func (*Bootstrapper) Bootstrap(context map[string]interface{}) error {
	client, err := NewClientConnection()
	if err != nil {
		return err
	}
	SetConnection(client)
	return nil
}

package signatures

type Bootstrapper struct {
}

func (*Bootstrapper) Bootstrap(context map[string]interface{}) error {
	NewSigningService(SigningService{})
	return nil
}

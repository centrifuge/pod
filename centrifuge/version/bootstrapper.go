package version

type Bootstrapper struct {
}

func (*Bootstrapper) Bootstrap(context map[string]interface{}) error {
	log.Infof("Running cent node on version: %s", GetVersion())
	return nil
}

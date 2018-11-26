package version

// Bootstrapper implements bootstrapper.Bootstrapper
type Bootstrapper struct{}

// Bootstrap logs the cent node version.
func (*Bootstrapper) Bootstrap(context map[string]interface{}) error {
	log.Infof("Running cent node on version: %s", GetVersion())
	return nil
}

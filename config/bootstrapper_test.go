//go:build unit

package config

import "github.com/centrifuge/go-centrifuge/bootstrap"

func (*Bootstrapper) TestBootstrap(context map[string]interface{}) error {
	if _, ok := context[bootstrap.BootstrappedConfig]; ok {
		return nil
	}
	// To get the config location, we need to traverse the path to find the `go-centrifuge` folder
	//gp := os.Getenv("BASE_PATH")
	//projDir := path.Join(gp, "centrifuge", "go-centrifuge")
	//context[bootstrap.BootstrappedConfig] = LoadConfiguration(fmt.Sprintf("%s/build/configs/testing_config.yaml", projDir))
	context[bootstrap.BootstrappedConfig] = LoadConfiguration("build/configs/testing_config.yaml")
	return nil
}

func (b *Bootstrapper) TestTearDown() error {
	return nil
}

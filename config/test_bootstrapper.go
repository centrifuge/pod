// +build integration unit

package config

import (
	"fmt"
	"path/filepath"
	"strings"
)

func (*Bootstrapper) TestBootstrap(context map[string]interface{}) error {
	// To get the config location, we need to traverse the path to find the `go-centrifuge` folder
	path, _ := filepath.Abs("./")
	match := ""
	for match == "" {
		path = filepath.Join(path, "../")
		if strings.HasSuffix(path, "go-centrifuge") {
			match = path
		}
		if filepath.Dir(path) == "/" {
			log.Fatal("Current working dir is not in `go-centrifuge`")
		}
	}
	c := LoadConfiguration(fmt.Sprintf("%s/build/configs/testing_config.yaml", match))
	context[BootstrappedConfig] = c
	return nil
}

func (b *Bootstrapper) TestTearDown() error {
	return nil
}

package config

import "github.com/centrifuge/go-centrifuge/errors"

const (
	ErrConfigBootstrap = errors.Error("error when bootstrapping config")

	// ErrConfigFileBootstrap used when config file is not found
	ErrConfigFileBootstrapNotFound = errors.Error("config file hasn't been provided")

	ErrConfigDBBootstrap = errors.Error("error when bootstrapping config DB")
)

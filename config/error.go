package config

import "github.com/centrifuge/pod/errors"

const (
	// ErrConfigBootstrap used as default error type
	ErrConfigBootstrap = errors.Error("error when bootstrapping config")

	// ErrConfigFileBootstrapNotFound used when config file is not found
	ErrConfigFileBootstrapNotFound = errors.Error("config file hasn't been provided")
)

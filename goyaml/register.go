// Copyright 2026 The yamlstar-plugin-yamlfmt Project Contributors
// SPDX-License-Identifier: Apache-2.0

// Package goyaml registers yamlfmt with the go-yaml plugin registry.
package goyaml

import (
	plugin "github.com/yamlstar/yamlstar-plugin-yamlfmt"
	"go.yaml.in/yaml/v4"
)

// Register makes yamlfmt available as the default dumper-format
// implementation.
func Register() error {
	return yaml.RegisterPlugin(yaml.PluginRegistration{
		API:     "dumper-format",
		Name:    "yamlfmt",
		Version: plugin.Version,
		Default: true,
		Factory: func(config map[string]any) (any, error) {
			return plugin.New(config)
		},
	})
}

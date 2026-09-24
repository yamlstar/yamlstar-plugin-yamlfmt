// Copyright 2026 The yamlstar-plugin-yamlfmt Project Contributors
// SPDX-License-Identifier: Apache-2.0

// Package yamlfmtplugin provides a dumper-format plugin backed by yamlfmt.
package yamlfmtplugin

import (
	_ "embed"
	"fmt"
	"strings"
	"sync"

	"github.com/google/yamlfmt"
	"github.com/google/yamlfmt/formatters/basic"
	"github.com/google/yamlfmt/formatters/kyaml"
)

// Version is the plugin release version.
const Version = "0.1.0"

// Manifest is the YAMLStar shared-plugin manifest.
//
//go:embed plugin.edn
var Manifest string

// Plugin formats serialized YAML with a configured yamlfmt formatter.
type Plugin struct {
	mu        sync.Mutex
	formatter yamlfmt.Formatter
}

// New constructs a yamlfmt plugin from plugin-specific configuration.
//
// The formatter defaults to the basic formatter. Configuration has this form:
//
//	formatter:
//	  type: basic
//	  indent: 4
func New(config map[string]any) (*Plugin, error) {
	formatterConfig, err := formatterConfig(config)
	if err != nil {
		return nil, err
	}
	formatterType, _ := formatterConfig["type"].(string)
	delete(formatterConfig, "type")

	var formatter yamlfmt.Formatter
	switch formatterType {
	case "basic":
		factory := basic.BasicFormatterFactory{}
		formatter, err = factory.NewFormatter(formatterConfig)
	case "kyaml":
		if len(formatterConfig) != 0 {
			return nil, fmt.Errorf(
				"yamlfmt: kyaml formatter does not accept configuration")
		}
		factory := kyaml.KYAMLFormatterFactory{}
		formatter, err = factory.NewFormatter(nil)
	default:
		return nil, fmt.Errorf("yamlfmt: unknown formatter type %q",
			formatterType)
	}
	if err != nil {
		return nil, fmt.Errorf("yamlfmt: configure %s formatter: %w",
			formatterType, err)
	}
	return &Plugin{formatter: formatter}, nil
}

// Format formats a complete YAML stream.
func (p *Plugin) Format(input []byte) ([]byte, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	output, err := p.formatter.Format(input)
	if err != nil {
		return nil, fmt.Errorf("yamlfmt: %w", err)
	}
	return output, nil
}

var basicKeys = map[string]bool{
	"array_indent":                 true,
	"disable_alias_key_correction": true,
	"disallow_anchors":             true,
	"drop_merge_tag":               true,
	"eof_newline":                  true,
	"force_array_style":            true,
	"force_quote_style":            true,
	"include_document_start":       true,
	"indent":                       true,
	"indent_root_array":            true,
	"indentless_arrays":            true,
	"line_ending":                  true,
	"max_line_length":              true,
	"pad_line_comments":            true,
	"retain_line_breaks":           true,
	"retain_line_breaks_single":    true,
	"scan_folded_as_literal":       true,
	"strip_directives":             true,
	"trim_trailing_whitespace":     true,
}

func formatterConfig(config map[string]any) (map[string]any, error) {
	for key := range config {
		if key != "formatter" {
			return nil, fmt.Errorf("yamlfmt: unknown configuration key %q", key)
		}
	}
	raw, found := config["formatter"]
	if !found || raw == nil {
		return map[string]any{"type": "basic"}, nil
	}
	mapping, ok := raw.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("yamlfmt: formatter must be a mapping")
	}
	normalized := make(map[string]any, len(mapping)+1)
	normalized["type"] = "basic"
	for key, value := range mapping {
		canonical := strings.ReplaceAll(key, "-", "_")
		if _, exists := normalized[canonical]; exists && canonical != "type" {
			return nil, fmt.Errorf(
				"yamlfmt: duplicate formatter key %q", canonical)
		}
		if canonical != "type" && !basicKeys[canonical] {
			return nil, fmt.Errorf("yamlfmt: unknown formatter key %q", key)
		}
		normalized[canonical] = value
	}
	formatterType, ok := normalized["type"].(string)
	if !ok || formatterType == "" {
		return nil, fmt.Errorf(
			"yamlfmt: formatter type must be a non-empty string")
	}
	return normalized, nil
}

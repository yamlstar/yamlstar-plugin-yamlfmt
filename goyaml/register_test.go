// Copyright 2026 The yamlstar-plugin-yamlfmt Project Contributors
// SPDX-License-Identifier: Apache-2.0

package goyaml_test

import (
	"testing"

	"github.com/yamlstar/yamlstar-plugin-yamlfmt/goyaml"
	"go.yaml.in/yaml/v4"
)

func TestRegisteredPlugin(t *testing.T) {
	if err := goyaml.Register(); err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	option, err := yaml.OptsYAML(
		"\nplugin:\n" +
			"  dumper-format:\n" +
			"    name: yamlfmt\n" +
			"    formatter:\n" +
			"      type: basic\n" +
			"      indent: 4\n")
	if err != nil {
		t.Fatalf("OptsYAML failed: %v", err)
	}
	got, err := yaml.Dump(map[string]any{
		"root": map[string]any{"child": []string{"one"}},
	}, option)
	if err != nil {
		t.Fatalf("Dump failed: %v", err)
	}
	want := "root:\n    child:\n        - one\n"
	if string(got) != want {
		t.Fatalf("Dump output = %q, want %q", got, want)
	}
}

// Copyright 2026 The yamlstar-plugin-yamlfmt Project Contributors
// SPDX-License-Identifier: Apache-2.0

package yamlfmtplugin_test

import (
	"strings"
	"sync"
	"testing"

	plugin "github.com/yamlstar/yamlstar-plugin-yamlfmt"
)

func TestBasicFormatter(t *testing.T) {
	formatter, err := plugin.New(map[string]any{
		"formatter": map[string]any{
			"type":   "basic",
			"indent": 4,
		},
	})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	got, err := formatter.Format([]byte("root:\n  child:\n  - one\n"))
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}
	want := "root:\n    child:\n        - one\n"
	if string(got) != want {
		t.Fatalf("Format output = %q, want %q", got, want)
	}
}

func TestHyphenatedKeys(t *testing.T) {
	formatter, err := plugin.New(map[string]any{
		"formatter": map[string]any{
			"include-document-start": true,
		},
	})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	got, err := formatter.Format([]byte("a: b\n"))
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}
	if !strings.HasPrefix(string(got), "---\n") {
		t.Fatalf("Format output = %q, want document start", got)
	}
}

func TestConfigurationErrors(t *testing.T) {
	tests := []struct {
		name   string
		config map[string]any
		match  string
	}{
		{"top-level", map[string]any{"indent": 4}, "unknown configuration"},
		{"mapping", map[string]any{"formatter": "basic"}, "must be a mapping"},
		{"type", map[string]any{"formatter": map[string]any{
			"type": "other"}}, "unknown formatter type"},
		{"unknown", map[string]any{"formatter": map[string]any{
			"unknown": true}}, "unknown formatter key"},
		{"kyaml", map[string]any{"formatter": map[string]any{
			"type": "kyaml", "indent": 4}}, "does not accept"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := plugin.New(test.config)
			if err == nil || !strings.Contains(err.Error(), test.match) {
				t.Fatalf("New error = %v, want match %q", err, test.match)
			}
		})
	}
}

func TestKYAMLFormatter(t *testing.T) {
	formatter, err := plugin.New(map[string]any{
		"formatter": map[string]any{"type": "kyaml"},
	})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	if _, err := formatter.Format([]byte("a: b\n")); err != nil {
		t.Fatalf("Format failed: %v", err)
	}
}

func TestConcurrentFormat(t *testing.T) {
	formatter, err := plugin.New(nil)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	const count = 8
	var wait sync.WaitGroup
	errors := make(chan error, count)
	for range count {
		wait.Add(1)
		go func() {
			defer wait.Done()
			_, err := formatter.Format([]byte("a: {b: 1}\n"))
			errors <- err
		}()
	}
	wait.Wait()
	close(errors)
	for err := range errors {
		if err != nil {
			t.Errorf("Format failed: %v", err)
		}
	}
}

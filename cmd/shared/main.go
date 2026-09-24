// Copyright 2026 The yamlstar-plugin-yamlfmt Project Contributors
// SPDX-License-Identifier: Apache-2.0

package main

/*
#include <stdint.h>
#include <stdlib.h>
*/
import "C"

import (
	"encoding/json"
	"fmt"
	"unsafe"

	plugin "github.com/yamlstar/yamlstar-plugin-yamlfmt"
)

func writeBytes(data []byte, output **C.uint8_t, length *C.size_t) C.int32_t {
	if output == nil || length == nil {
		return 2
	}
	var pointer unsafe.Pointer
	if len(data) != 0 {
		pointer = C.CBytes(data)
	}
	*output = (*C.uint8_t)(pointer)
	*length = C.size_t(len(data))
	return 0
}

func writeError(kind string, err error, output **C.uint8_t,
	length *C.size_t,
) {
	data, marshalErr := json.Marshal(map[string]any{
		"error": map[string]any{
			"type": kind, "message": err.Error(), "data": map[string]any{},
		},
	})
	if marshalErr != nil {
		data = []byte("{\"error\":{\"type\":\"abi\",\"message\":" +
			"\"could not encode error\",\"data\":{}}}")
	}
	_ = writeBytes(data, output, length)
}

func readBytes(pointer *C.uint8_t, length C.size_t) ([]byte, error) {
	if pointer == nil && length != 0 {
		return nil, fmt.Errorf("invalid input buffer")
	}
	maxInt := uint64(^uint(0) >> 1)
	if uint64(length) > maxInt {
		return nil, fmt.Errorf("input buffer is too large")
	}
	return unsafe.Slice((*byte)(unsafe.Pointer(pointer)), int(length)), nil
}

func format(input, options []byte) ([]byte, error) {
	config := map[string]any{}
	if len(options) != 0 {
		if err := json.Unmarshal(options, &config); err != nil {
			return nil, fmt.Errorf("decode options JSON: %w", err)
		}
	}
	formatter, err := plugin.New(config)
	if err != nil {
		return nil, err
	}
	return formatter.Format(input)
}

//export yamlstar_plugin_v1_abi
func yamlstar_plugin_v1_abi() C.uint64_t {
	return 1
}

//export yamlstar_plugin_v1_manifest
func yamlstar_plugin_v1_manifest(output **C.uint8_t,
	length *C.size_t,
) C.int32_t {
	return writeBytes([]byte(plugin.Manifest), output, length)
}

//export yamlstar_plugin_v1_format
func yamlstar_plugin_v1_format(
	input *C.uint8_t, inputLength C.size_t,
	options *C.uint8_t, optionsLength C.size_t,
	output **C.uint8_t, outputLength *C.size_t,
) (status C.int32_t) {
	if output == nil || outputLength == nil {
		return 2
	}
	*output = nil
	*outputLength = 0
	defer func() {
		if value := recover(); value != nil {
			writeError("abi", fmt.Errorf("%v", value), output, outputLength)
			status = 2
		}
	}()
	inputBytes, err := readBytes(input, inputLength)
	if err != nil {
		writeError("abi", err, output, outputLength)
		return 2
	}
	optionBytes, err := readBytes(options, optionsLength)
	if err != nil {
		writeError("abi", err, output, outputLength)
		return 2
	}
	formatted, err := format(inputBytes, optionBytes)
	if err != nil {
		writeError("format", err, output, outputLength)
		return 1
	}
	return writeBytes(formatted, output, outputLength)
}

//export yamlstar_plugin_v1_free
func yamlstar_plugin_v1_free(output *C.uint8_t) {
	C.free(unsafe.Pointer(output))
}

func main() {}

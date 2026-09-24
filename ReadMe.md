# yamlstar-plugin-yamlfmt

This repository provides the yamlfmt implementation of the generic
dumper-format plugin API for go-yaml and YAMLStar.
It uses [google/yamlfmt](https://github.com/google/yamlfmt) v0.21.0.

The go-yaml integration is registered at application startup:

    import "github.com/yamlstar/yamlstar-plugin-yamlfmt/goyaml"

    func init() {
        if err := goyaml.Register(); err != nil {
            panic(err)
        }
    }

It can then be selected in go-yaml options:

    plugin:
      dumper-format:
        name: yamlfmt
        formatter:
          type: basic
          indent: 4

The formatter type can be basic or kyaml.
Basic formatter option names may use underscores or hyphens.
The shared library uses YAMLStar plugin ABI v1 and accepts formatter options
as JSON.

Run `make test` to test the Go adapter and shared ABI.

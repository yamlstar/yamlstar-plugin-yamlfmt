# yamlstar-plugin-yamlfmt

This repository provides the yamlfmt implementation of the generic
dumper-format plugin API for go-yaml and YAMLStar.
It uses [google/yamlfmt](https://github.com/google/yamlfmt) v0.21.0.
The go-yaml adapter currently uses the Go pseudo-version for
[go-yaml pull request 430](https://github.com/yaml/go-yaml/pull/430).
It will move to the next go-yaml release candidate after that pull request
merges.

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

Run `make test` to test the Go adapter and shared ABI without a local Go
workspace.
Run `make test-race` for the Go race tests.

## Release archives

Run `make release-archive` to build and test the archive for the current
platform.
Release `v0.1.0` contains these assets:

- `yamlstar-plugin-yamlfmt-v0.1.0-linux-x64.tar.xz`
- `yamlstar-plugin-yamlfmt-v0.1.0-linux-aarch64.tar.xz`
- `yamlstar-plugin-yamlfmt-v0.1.0-macos-x64.tar.xz`
- `yamlstar-plugin-yamlfmt-v0.1.0-macos-arm64.tar.xz`
- `SHA256SUMS`

The release workflow requires an existing `v`-prefixed tag matching the
version in `Makefile` and `plugin.edn`.
It builds and tests all four archives before publishing them.

Publish an already-versioned and tagged release with:

```sh
make release v=0.1.0 a=1
```

The command checks the source versions, working tree, local and remote tags,
and release state before asking to dispatch and watch the GitHub workflow.
The `a=1` setting allows release from a branch other than `main`.
Use `d=1` to preview the command without starting the workflow.

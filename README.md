# go-build

`go-build` is a library containing common build commands for Go libraries,
based on [goyek](https://github.com/goyek/goyek). While the tasks are
primarily to satisfy build requirements for [CurioSwitch](https://github.com/curioswitch),
they are intended to be generally useful so if you'd like, give them a try.

## Features

The defined tasks attempt to be a baseline that is useful for any Go repository.
One key point is the word "repository", meaning the focus is not only Go code
but any file we would typically see with it in a project. This means that
format and lint tasks target the following languages:

- Go
- Markdown
- YAML

All supporting tasks are executed with `go run` - this means that all languages
can be processed with only a single tool dependency, Go itself.

## Usage

The simplest way to use this library is to copy the contents of [build](./build)
from this repository, which itself is using the defined tasks. You can add it to
a Go workspace to keep build-specific libraries like goyek out of your standard
modules file, or remove the go.mod / go.sum files to include it as a normal
package.

Using the folder `build` is a goyek convention, but any folder name will work,
i.e. if you already use `build` for transient artifacts. Note that these tasks
use `out` for transient artifacts by default but can be configured for different
paths.

A list of all tasks can be seen with `go run ./build -h`. The commonly used tasks
will likely be:

- `go run ./build check` - executes all code checks, including lint and unit tests.
  This should be the command run from a CI script.

- `go run ./build format` - executes all auto-formatting.

Note that for formatting Go code, currently the only tool that is run is
[golangci-lint](https://golangci-lint.run/usage/linters/) with autofixes enabled.
It is recommended to configure your `.golangci.yml` file with the `gofumpt` and
`gci` linters - this way, both will be applied when running `format` and checked
when running `lint`.

VSCode users may want to create a workspace configuration similar to [ours](./go-build.code-workspace),
which is set to allow IDE auto-save to match the result of the tasks in this project
as much as possible.

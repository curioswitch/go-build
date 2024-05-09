# go-build

`go-build` is a library containing common build commands for Go libraries,
based on [goyek](https://github.com/goyek/goyek). While the tasks are
primarily to satisfy build requirements for https://github.com/curioswitch,
they are intended to be generally useful so if you'd like, give them a try.

## Usage

The simplest way to use this library is to copy the contents of [build](./build)
from this repository, which itself is using the defined tasks. You can add it to
a Go workspace to keep build-specific libraries like goyek out of your standard
modules file, or remove the go.mod / go.sum files to include it as a normal
package.

Using the folder `build` is a goyek convention, but any folder name will work,
i.e. if you already use `build` for transient artifacts. Note that these tasks
use `out` for transient artifacts.

A list of all tasks can be seen with `go run ./build -h`. The commonly used tasks
will likely be:

- `go run ./build check` - executes all code checks, including lint and unit tests.
  This should be the command run from a CI script.

- `go run ./build format` - executes all auto-formatting.

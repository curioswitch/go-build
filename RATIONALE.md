# Notable rationale of go-build

## Use goyek

We do not use Makefile, probably the most common tool for builds in Go, because
it is tedious to do complex tasks and hard to make cross-platform.

We have used [Mage](https://magefile.org/) in other projects and it works well -
for projects that currently use Makefile, it can be easier to convince members
to migrate to it vs goyek. It has some quirks though.

For further details, see goyek's [explanation](https://github.com/goyek/goyek?tab=readme-ov-file#alternatives)
which is fair and follows our thoughts.

## Use gofumpt

We prefer to have less bikeshedding in code reviews, and this includes formatting.
While this isn't as prevalent in the Go ecosystem yet, it is common practice in others
such as NodeJS. Where possible, we will prefer auto-formatting that enforces as much
structure as possible. It's automatic, so why not?

## Use gci

We believe there is significant stylistic benefit in having consistent import
ordering and grouping, something the Go standard goimports [cannot do](https://github.com/golang/go/issues/20818).
Both gci and gosimports are great tools for this, and we choose gci because it
is also integrated with golangci-lint, making it simpler to verify in CI.

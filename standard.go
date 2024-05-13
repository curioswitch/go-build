package build

import (
	"fmt"
	"os"

	"github.com/goyek/goyek/v2"
	"github.com/goyek/x/cmd"
)

// DefineTasks defines common tasks for Go projects.
func DefineTasks(opts ...Option) {
	var conf config
	for _, o := range opts {
		o.apply(&conf)
	}

	formatGo := goyek.Define(goyek.Task{
		Name:  "format-go",
		Usage: "Formats Go code.",
		Action: func(a *goyek.A) {
			cmd.Exec(a, fmt.Sprintf("go run mvdan.cc/gofumpt@%s -l -w .", verGoFumpt))

			importSecs := "-s standard -s default"
			for _, prefix := range conf.localPackagePrefixes {
				importSecs += fmt.Sprintf(` -s "prefix(%s)"`, prefix)
			}

			cmd.Exec(a, fmt.Sprintf("go run github.com/daixiang0/gci@%s write %s .", verGci, importSecs))
		},
	})

	goyek.Define(goyek.Task{
		Name:  "format",
		Usage: "Format code in various languages.",
		Deps:  append(goyek.Deps{formatGo}, formatTasks...),
	})

	goyek.Define(goyek.Task{
		Name:  "generate",
		Usage: "Generates code.",
		Deps:  generateTasks,
	})

	lintGo := goyek.Define(goyek.Task{
		Name:  "lint-go",
		Usage: "Lints Go code.",
		Action: func(a *goyek.A) {
			cmd.Exec(a, fmt.Sprintf("go run github.com/golangci/golangci-lint/cmd/golangci-lint@%s run --timeout=20m", verGolangCILint))
		},
	})

	lintYaml := goyek.Define(goyek.Task{
		Name:  "lint-yaml",
		Usage: "Lints Yaml code.",
		Action: func(a *goyek.A) {
			cmd.Exec(a, fmt.Sprintf("go run github.com/wasilibs/go-yamllint/cmd/yamllint@%s .", verGoYamllint))
		},
	})

	lint := goyek.Define(goyek.Task{
		Name:  "lint",
		Usage: "Lints code in various languages.",
		Deps:  append(goyek.Deps{lintGo, lintYaml}, lintTasks...),
	})

	test := goyek.Define(goyek.Task{
		Name:  "test",
		Usage: "Runs unit tests.",
		Action: func(a *goyek.A) {
			if err := os.MkdirAll("out", 0o755); err != nil {
				a.Errorf("failed to create out directory: %v", err)
				return
			}
			cmd.Exec(a, "go test -coverprofile=out/coverage.txt -covermode=atomic -v -timeout=20m ./...")
		},
	})

	goyek.Define(goyek.Task{
		Name:  "check",
		Usage: "Runs all checks.",
		Deps:  goyek.Deps{lint, test},
	})
}

type config struct {
	localPackagePrefixes []string
}

// Option is a configuration option for DefineTasks.
type Option interface {
	apply(conf *config)
}

// LocalPackagePrefix returns an Option to indicate the local package prefix for the project.
// Imports from this prefix will be ordered at the end of other import groups when formatting.
// This option can be provided multiple times to separate multiple sections, in the order
// provided.
func LocalPackagePrefix(prefix string) Option {
	return &localPackagePrefixOption{
		localPackagePrefix: prefix,
	}
}

type localPackagePrefixOption struct {
	localPackagePrefix string
}

func (o *localPackagePrefixOption) apply(c *config) {
	c.localPackagePrefixes = append(c.localPackagePrefixes, o.localPackagePrefix)
}
